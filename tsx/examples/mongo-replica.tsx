import { Compose, DependsOn, Env, Expose, Health, Network, Publish, Restart, Service, Volume } from "containerkit";
import { root, runIfMain } from "containerkit/live";
import { depend, netDriver, proto, restartPolicy } from "containerkit/vocab";

// Eight MongoDB replicas and Mongo Express. Members past the seventh do not vote.
// The Go program is examples/mongo_replica. It execs rs.initiate after the
// replicas are healthy. This session does the same with createExec.

const numReplicas = 8;
const mongoExpressUser = "admin";
const mongoExpressPass = "password";

type Member = {
	_id: number;
	host: string;
	votes?: number;
	priority?: number;
};

const members: Member[] = [];
const hosts: string[] = [];
for (let i = 0; i < numReplicas; i++) {
	const host = `db-${i}:27017`;
	hosts.push(host);
	const member: Member = { _id: i, host };
	if (i >= 7) {
		member.votes = 0;
		member.priority = 0;
	}
	members.push(member);
}

const url = `mongodb://${hosts.join(",")}?replicaSet=rs0`;
const replicaServices = members.map((member) => `db-${member._id}`);

const tree = (
	<Compose name="mongo-db-cluster">
		<Volume name="mongo-data" />
		<Network name="mongo-cluster" driver={netDriver.Bridge} />
		{members.map((member) => (
			<Service
				key={member._id}
				name={`db-${member._id}`}
				containerName={`mongodb-${member._id}`}
				image="mongo:latest"
				hostname={`db-${member._id}`}
				command={["mongod", "--replSet", "rs0", "--bind_ip_all"]}
			>
				{member._id > 0 && <DependsOn name={`db-${member._id - 1}`} />}
				<Health test={["CMD", "mongosh", "--eval", 'db.adminCommand("ping")']} interval="1s" timeout="10s" retries={5} />
				<Expose port={27017} proto={proto.TCP} />
				<Restart policy={restartPolicy.UnlessStopped} />
				<Network name="mongo-cluster" />
			</Service>
		))}
		<Service name="mongo-express" image="mongo-express:latest">
			{replicaServices.map((name) => (
				<DependsOn key={name} name={name} condition={depend.Healthy} />
			))}
			<Env name="ME_CONFIG_MONGODB_URL" value={url} />
			<Env name="ME_CONFIG_MONGODB_ENABLE_ADMIN" value="true" />
			<Env name="VCAP_APP_HOST" value="0.0.0.0" />
			<Env name="ME_CONFIG_BASICAUTH_ENABLED" value="true" />
			<Env name="ME_CONFIG_BASICAUTH_USERNAME" value={mongoExpressUser} />
			<Env name="ME_CONFIG_BASICAUTH_PASSWORD" value={mongoExpressPass} />
			<Env name="ME_CONFIG_SITE_COOKIESECRET" value="containerkit-mongo-express" />
			<Env name="ME_CONFIG_SITE_SESSIONSECRET" value="containerkit-mongo-express" />
			<Publish protocol={proto.TCP} hostIP="0.0.0.0" hostPort={8081} containerPort={8081} />
			<Restart policy={restartPolicy.Always} />
			<Network name="mongo-cluster" />
		</Service>
	</Compose>
);

export { tree, url };

const initiate = `rs.initiate(${JSON.stringify({ _id: "rs0", members })}); const deadline = Date.now() + 30000; while (Date.now() < deadline) { try { if (rs.hello().isWritablePrimary) { quit(0); } } catch (e) {} sleep(500); } throw new Error("replica set did not elect a primary");`;

const live = root({ fromEnv: true, onExit: "down" }, async function App(session) {
	await session.composeUp(tree, {
		recreate: "force",
		removeOrphans: true,
		wait: true,
		services: replicaServices,
	});
	const exec = await session.createExec("mongodb-0", {
		command: ["mongosh", "--eval", initiate],
		attachStdout: true,
		attachStderr: true,
	});
	await session.startExec(exec.ID);
	await session.composeUp(tree, { wait: true, services: "mongo-express" });
	console.error(`Mongo Express http://localhost:8081  ${mongoExpressUser} / ${mongoExpressPass}`);
	console.error(url);
	return tree;
});

export default live;

await runIfMain(live, import.meta.url);
