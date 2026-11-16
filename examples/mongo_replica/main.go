// A MongoDB replica set with eight members and a Mongo Express console.
// Members past the seventh have votes and priority set to zero.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

const (
	numReplicas = 4

	mongoExpressUser = "admin"
	mongoExpressPass = "password"
)

type replicaSet struct {
	ID      string   `json:"_id"`
	Members []member `json:"members"`
}

type member struct {
	ID       int      `json:"_id"`
	Host     string   `json:"host"`
	Votes    *int     `json:"votes,omitempty"`
	Priority *float64 `json:"priority,omitempty"`
}

func main() {
	members := make([]member, 0, numReplicas)
	urlParts := make([]string, 0, numReplicas)
	app := NewCompose(
		"mongo-db-cluster",
		AddNetwork(
			Name("mongo-cluster"),
			Driver("bridge"),
		),
		AddVolume(
			Name("mongo-data"),
		),
	)
	for i := range numReplicas {
		serviceName := fmt.Sprintf("db-%d", i)
		opts := append(
			mongoReplica(i),
			Name(serviceName),
			ContainerName(fmt.Sprintf("mongodb-%d", i)),
		)
		if i > 0 {
			opts = append(opts, DependsOn(Name(fmt.Sprintf("db-%d", i-1))))
		}
		app.WithService(NewService(opts...))

		next := member{Host: serviceName + ":27017", ID: i}
		if i >= numReplicas-1 {
			votes, priority := 0, 0.0
			next.Votes = &votes
			next.Priority = &priority
		}
		members = append(members, next)
		urlParts = append(urlParts, serviceName+":27017")
	}
	url := fmt.Sprintf("mongodb://%s/?replicaSet=rs0", strings.Join(urlParts, ","))

	express := append(mongoExpress(url), Name("mongo-express"))
	for i := range numReplicas {
		express = append(express, DependsOn(Name(fmt.Sprintf("db-%d", i)), Condition(Healthy)))
	}
	app.WithService(NewService(express...))

	cli, err := Connect(FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Close()
	ctx := context.Background()
	api := cli.Composer()

	replicas := make([]string, 0, numReplicas)
	for i := range numReplicas {
		replicas = append(replicas, fmt.Sprintf("db-%d", i))
	}
	if err := app.Up(ctx, api, Recreate("force"), RemoveOrphans(true), Wait(true), Services(replicas...)); err != nil {
		log.Fatal(err)
	}
	if err := initialize(ctx, cli, "mongodb-0", members); err != nil {
		log.Fatal(err)
	}
	if err := app.Up(ctx, api, Wait(true), Services("mongo-express")); err != nil {
		log.Fatal(err)
	}

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt)
	runCtx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signals
		defer cancel()
		fmt.Println("Received interrupt signal, shutting down...")
		if err := app.Down(context.Background(), api); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	yml, err := app.ExportYAML()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.WriteFile("./examples/mongo_replica/docker-compose.yaml", yml, 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("MongoDB replica set", "rs0", "url", url)
	fmt.Println("MongoDB replica set initialized and running.")
	fmt.Println("You can access Mongo Express at http://localhost:8081")
	fmt.Printf("Mongo Express login: %s / %s\n", mongoExpressUser, mongoExpressPass)
	fmt.Println("Press Ctrl+C to stop the containers.")
	<-runCtx.Done()
}

func mongoReplica(index int) []ServiceOption {
	return []ServiceOption{
		Image("mongo:latest"),
		Hostname(fmt.Sprintf("db-%d", index)),
		Command("mongod", "--replSet", "rs0", "--bind_ip_all"),
		Health(
			Test("CMD", "mongosh", "--eval", `db.adminCommand("ping")`),
			Interval(time.Second),
			Timeout(10*time.Second),
			Retries(5),
		),
		Expose(27017, TCP),
		Restart(Policy(UnlessStopped)),
		Network(Name("mongo-cluster")),
	}
}

func mongoExpress(url string) []ServiceOption {
	return []ServiceOption{
		Image("mongo-express:latest"),
		Env("VCAP_APP_HOST", "0.0.0.0"),
		Env("ME_CONFIG_MONGODB_URL", url),
		Env("ME_CONFIG_MONGODB_ENABLE_ADMIN", "true"),
		Env("ME_CONFIG_BASICAUTH_ENABLED", "true"),
		Env("ME_CONFIG_BASICAUTH_USERNAME", mongoExpressUser),
		Env("ME_CONFIG_BASICAUTH_PASSWORD", mongoExpressPass),
		Env("ME_CONFIG_SITE_COOKIESECRET", "containerkit-mongo-express"),
		Env("ME_CONFIG_SITE_SESSIONSECRET", "containerkit-mongo-express"),
		Publish(Protocol(TCP), HostIP("0.0.0.0"), HostPort(8081), ContainerPort(8081)),
		Restart(Policy(Always)),
		Network(Name("mongo-cluster")),
	}
}

func initialize(ctx context.Context, cli *Client, containerID string, members []member) error {
	if len(members) == 0 {
		return fmt.Errorf("no members provided for replica set initialization")
	}
	payload, err := json.Marshal(replicaSet{ID: "rs0", Members: members})
	if err != nil {
		return fmt.Errorf("marshal members: %w", err)
	}
	script := fmt.Sprintf(
		`rs.initiate(%s); const deadline = Date.now() + 30000; while (Date.now() < deadline) { try { if (rs.hello().isWritablePrimary) { quit(0); } } catch (e) {} sleep(500); } throw new Error("replica set did not elect a primary");`,
		string(payload),
	)
	fmt.Println("mongosh --eval rs.initiate(...)")
	created, err := CreateExec(ctx, cli.Engine(), containerID,
		Command("mongosh", "--eval", script),
		AttachStdout(true),
		AttachStderr(true),
	)
	if err != nil {
		return fmt.Errorf("create exec: %w", err)
	}
	attached, err := AttachExec(ctx, cli.Engine(), created.ID)
	if err != nil {
		return fmt.Errorf("attach exec: %w", err)
	}
	defer attached.Close()
	_, err = CopyMultiplexed(os.Stdout, os.Stderr, attached.Reader)
	if err != nil {
		return fmt.Errorf("copy multiplexed: %w", err)
	}
	inspected, err := InspectExec(ctx, cli.Engine(), created.ID)
	if err != nil {
		return fmt.Errorf("inspect exec: %w", err)
	}
	if inspected.ExitCode != 0 {
		return fmt.Errorf("rs.initiate failed: exit %d", inspected.ExitCode)
	}
	return nil
}
