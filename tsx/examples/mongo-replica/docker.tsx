/** @jsxImportSource containerkit */
import { Compose, DependsOn, Env, Expose, Health, Network, Publish, Restart, Service, Volume } from "containerkit";
import { useShutdown } from "containerkit/hooks";
import { root, useEffect, useSession, useState } from "containerkit/live";
import { depend, netDriver, proto, restartPolicy } from "containerkit/vocab";

// Eight MongoDB members. Members past the seventh do not vote. The first render
// is the replica set. After those healthchecks pass, an effect initiates rs0
// and the next render adds Mongo Express.

const { TCP } = proto;
const { UnlessStopped, Always } = restartPolicy;

const numReplicas = 8;
const mongoPort = 27017;
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
const initiate = `
rs.initiate(${JSON.stringify({ _id: "rs0", members }, null, 2)});

const deadline = Date.now() + 30000;
while (Date.now() < deadline) {
	try {
		if (rs.hello().isWritablePrimary) quit(0);
	} catch {
		// hello() throws until the set has a primary.
	}
	sleep(500);
}
throw new Error("replica set did not elect a primary");
`.trim();

function App() {
	const session = useSession();
	useShutdown({ removeOrphans: true, volumes: true });
	const [initiated, setInitiated] = useState(false);

	useEffect(() => {
		void (async () => {
			try {
				const exec = await session.createExec("mongodb-0", {
					command: ["mongosh", "--eval", initiate],
					attachStdout: true,
					attachStderr: true,
				});
				await session.startExec(exec.ID);
				setInitiated(true);
			} catch (err) {
				console.error(err);
			}
		})();
	}, [session]);

	return (
		<Compose name="mongo-db-cluster">
			{members.map((member) => (
				<Volume key={member._id} name={`mongo-${member._id}`} />
			))}
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
					<Health
						test={["CMD", "mongosh", "--eval", 'db.adminCommand("ping")']}
						interval="1s"
						timeout="10s"
						retries={5}
					/>
					<Expose port={mongoPort} proto={TCP} />
					<Restart policy={UnlessStopped} />
					<Network name="mongo-cluster" />
					<Volume source={`mongo-${member._id}`} target="/data/db" />
				</Service>
			))}
			{initiated && (
				<Service name="mongo-express" image="mongo-express:latest">
					<Publish protocol={TCP} hostIP="127.0.0.1" hostPort={8081} containerPort={8081} />
					{members.map((member) => (
						<DependsOn key={member._id} name={`db-${member._id}`} condition={depend.Healthy} />
					))}
					<Env name="ME_CONFIG_MONGODB_URL" value={url} />
					<Env name="ME_CONFIG_MONGODB_ENABLE_ADMIN" value="true" />
					<Env name="VCAP_APP_HOST" value="0.0.0.0" />
					<Env name="ME_CONFIG_BASICAUTH_ENABLED" value="true" />
					<Env name="ME_CONFIG_BASICAUTH_USERNAME" value={mongoExpressUser} />
					<Env name="ME_CONFIG_BASICAUTH_PASSWORD" value={mongoExpressPass} />
					<Env name="ME_CONFIG_SITE_COOKIESECRET" value="containerkit-mongo-express" />
					<Env name="ME_CONFIG_SITE_SESSIONSECRET" value="containerkit-mongo-express" />
					<Restart policy={Always} />
					<Network name="mongo-cluster" />
				</Service>
			)}
		</Compose>
	);
}

export default root({ fromEnv: true, onExit: "down", wait: true }, App);
