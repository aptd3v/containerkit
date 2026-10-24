// This example demonstrates how to create a MongoDB replica set with x members
// and a Mongo Express instance to manage the replica set.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/aptd3v/containerkit/pkg/client"
	"github.com/aptd3v/containerkit/pkg/kit"
	"github.com/aptd3v/containerkit/pkg/kit/types"
)

const (
	NumReplicas = 8 // Number of MongoDB replicas in the replica set

	mongoExpressUser = "admin"
	mongoExpressPass = "password"
)

type RSet struct {
	ID      string     `json:"_id"`
	Members []RSMember `json:"members"`
}

type RSMember struct {
	ID       int      `json:"_id"`
	Host     string   `json:"host"`
	Votes    *int     `json:"votes,omitempty"`
	Priority *float64 `json:"priority,omitempty"`
}

func main() {

	project := kit.NewProject("mongo-db-cluster")

	members := []RSMember{}
	urlParts := make([]string, 0, NumReplicas)
	for i := range NumReplicas {
		serviceName := fmt.Sprintf("db-%d", i)

		db := kit.NewService(serviceName, WithMongoReplica(i))
		if i > 0 {
			db.DependsOn(fmt.Sprintf("db-%d", i-1))
		}
		project.WithService(db)
		member := RSMember{
			Host: serviceName + ":27017",
			ID:   i,
		}
		// MongoDB allows at most 7 voting members.
		if i >= 7 {
			votes, priority := 0, 0.0
			member.Votes = &votes
			member.Priority = &priority
		}
		members = append(members, member)
		urlParts = append(urlParts, fmt.Sprintf("%s:27017", serviceName))
	}
	url := fmt.Sprintf("mongodb://%s/?replicaSet=rs0", strings.Join(urlParts, ","))

	express := kit.NewService("mongo-express", WithMongoExpress(url))
	for i := range NumReplicas {
		express.DependsOnHealthy(fmt.Sprintf("db-%d", i))
	}
	project.WithService(express)

	project.WithNetwork("mongo-cluster").WithVolume("mongo-data")

	database := kit.NewCompose(project)
	ctx := context.Background()

	replicaServices := make([]string, 0, NumReplicas)
	for i := range NumReplicas {
		replicaServices = append(replicaServices, fmt.Sprintf("db-%d", i))
	}
	if err := database.Up(ctx, &kit.Up{
		ForceRecreate: true,
		RemoveOrphans: true,
		Wait:          true,
		ServiceNames:  replicaServices,
	}); err != nil {
		log.Fatal(err)
	}

	err := Initialize(ctx, "mongodb-0", members)
	if err != nil {
		log.Fatal(err)
	}
	if err := database.Up(ctx, &kit.Up{
		Wait:         true,
		ServiceNames: []string{"mongo-express"},
	}); err != nil {
		log.Fatal(err)
	}
	signalsChan := make(chan os.Signal, 1)
	signal.Notify(signalsChan, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signalsChan
		defer cancel()
		fmt.Println("Received interrupt signal, shutting down...")
		if err := database.Down(context.Background(), nil); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	err = project.Export("./examples/mongo_replica/docker-compose.yaml", 0644)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("MongoDB replica set", "rs0", "url", url)
	fmt.Println("MongoDB replica set initialized and running.")
	fmt.Println("You can access Mongo Express at http://localhost:8081")
	fmt.Printf("Mongo Express login: %s / %s\n", mongoExpressUser, mongoExpressPass)
	fmt.Println("Press Ctrl+C to stop the containers.")
	<-ctx.Done()
}

func WithMongoReplica(index int) *kit.Container {
	containerName := fmt.Sprintf("mongodb-%d", index)
	return kit.NewContainer(containerName).
		Image("mongo:latest").
		HostName(fmt.Sprintf("db-%d", index)).
		Command("mongod", "--replSet", "rs0", "--bind_ip_all").
		HealthCheck(kit.NewHealth().
			CMD("mongosh", "--eval", `db.adminCommand("ping")`).
			IntervalD("1s").
			TimeoutD("10s").
			StartPeriodD("0s").
			Retries(5)).
		ExposedPort("tcp", "27017").
		RestartUnlessStopped().
		Endpoint("mongo-cluster")
}

func Initialize(ctx context.Context, initContainer string, members []RSMember) error {
	cli, err := client.NewClient(client.FromEnv(), client.WithAPIVersionFromEnv())
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %w", err)
	}
	if len(members) == 0 {
		return fmt.Errorf("no members provided for replica set initialization")
	}
	initiate := RSet{
		ID:      "rs0",
		Members: members,
	}

	init, err := json.Marshal(initiate)
	if err != nil {
		return fmt.Errorf("failed to marshal members: %w", err)
	}
	command := []string{"mongosh", "--eval", fmt.Sprintf(
		`rs.initiate(%s); const deadline = Date.now() + 30000; while (Date.now() < deadline) { try { if (rs.hello().isWritablePrimary) { quit(0); } } catch (e) {} sleep(500); } throw new Error("replica set did not elect a primary");`,
		string(init),
	)}

	fmt.Println("mongosh --eval rs.initiate(...)")
	res, err := cli.ContainerExecCreate(ctx, initContainer, &client.Exec{
		Cmd:          command,
		AttachStdout: true,
		AttachStderr: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec command: %w", err)
	}
	attached, err := cli.ContainerExecAttach(ctx, res.ID, nil)
	if err != nil {
		return fmt.Errorf("failed to start exec command: %w", err)
	}
	defer attached.Close()
	_, _ = io.Copy(os.Stdout, attached.Reader)
	inspect, err := cli.ContainerExecInspect(ctx, res.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect exec command: %w", err)
	}
	if inspect.ExitCode != nil && *inspect.ExitCode != 0 {
		return fmt.Errorf("rs.initiate failed: exit %d", *inspect.ExitCode)
	}
	return nil
}

func WithMongoExpress(url string) *kit.Container {
	return kit.NewContainer("mongo-express").
		Image("mongo-express:latest").
		Env("ME_CONFIG_MONGODB_URL", url).
		Env("ME_CONFIG_MONGODB_ENABLE_ADMIN", "true").
		Env("VCAP_APP_HOST", "0.0.0.0").
		Env("ME_CONFIG_BASICAUTH_ENABLED", "true").
		Env("ME_CONFIG_BASICAUTH_USERNAME", mongoExpressUser).
		Env("ME_CONFIG_BASICAUTH_PASSWORD", mongoExpressPass).
		Env("ME_CONFIG_SITE_COOKIESECRET", "containerkit-mongo-express").
		Env("ME_CONFIG_SITE_SESSIONSECRET", "containerkit-mongo-express").
		PortBinding(types.PortBinding{Protocol: types.TCP, HostIP: "0.0.0.0", HostPort: "8081", ContainerPort: "8081"}).
		RestartAlways().
		Endpoint("mongo-cluster")
}
