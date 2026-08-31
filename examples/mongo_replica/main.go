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
	"github.com/aptd3v/containerkit/pkg/containerkit"
)

const NumReplicas = 3 // Number of MongoDB replicas in the replica set

type RSet struct {
	ID      string     `json:"_id"`
	Members []RSMember `json:"members"`
}

type RSMember struct {
	ID   int    `json:"_id"`
	Host string `json:"host"`
}

func main() {

	project := containerkit.NewProject("mongo-db-cluster")

	members := []RSMember{}
	urlParts := make([]string, 0, NumReplicas)
	for i := range NumReplicas {
		serviceName := fmt.Sprintf("db-%d", i)

		extras := []any{}
		if i > 0 {
			extras = append(extras, containerkit.DependsOn(fmt.Sprintf("db-%d", i-1)))
		}
		project.WithService(serviceName, WithMongoReplica(i), extras...)
		members = append(members, RSMember{
			Host: serviceName,
			ID:   i,
		})
		urlParts = append(urlParts, fmt.Sprintf("%s:27017", serviceName))
	}
	url := fmt.Sprintf("mongodb://%s/?replicaSet=rs0", strings.Join(urlParts, ","))

	project.WithService("mongo-express", WithMongoExpress(url))

	project.WithNetwork("mongo-cluster").WithVolume("mongo-data")

	database := containerkit.NewCompose(project)

	if err := database.Up(
		context.Background(),
		&containerkit.Up{
			ForceRecreate: true,
			RemoveOrphans: true,
			Wait:          true,
		},
	); err != nil {
		log.Fatal(err)
	}

	err := Initialize(context.Background(), "mongodb-0", members)
	if err != nil {
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
	fmt.Println("Press Ctrl+C to stop the containers.")
	<-ctx.Done()
}

func WithMongoReplica(index int) *containerkit.Container {
	containerName := fmt.Sprintf("mongodb-%d", index)
	return containerkit.NewContainer(containerName).
		Image("mongo:latest").
		Command("mongod", "--replSet", "rs0", "--bind_ip_all").
		HealthCheck(containerkit.Health{
			Test:         []string{"CMD", "mongosh", "--eval", `db.adminCommand("ping")`},
			IntervalD:    "1s",
			TimeoutD:     "10s",
			StartPeriodD: "0s",
			Retries:      5,
		}).
		ExposedPort("tcp", "27017").
		RestartUnlessStopped().
		Endpoint("mongo-cluster")
}

func Initialize(ctx context.Context, initContainer string, members []RSMember) error {
	opts := []client.SetClientOption{client.WithAPIVersionNegotiation()}
	if os.Getenv("DOCKER_HOST") != "" {
		opts = append([]client.SetClientOption{client.FromEnv()}, opts...)
	}
	cli, err := client.NewClient(opts...)
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
	command := []string{"mongosh", "--eval", fmt.Sprintf("rs.initiate(%s)", string(init))}

	fmt.Println(strings.Join(command, " "))
	res, err := cli.ContainerExecCreate(ctx, initContainer, &client.Exec{
		Command:      command,
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
	return nil
}

func WithMongoExpress(url string) *containerkit.Container {
	return containerkit.NewContainer("mongo-express").
		Image("mongo-express:latest").
		Env("ME_CONFIG_MONGODB_URL", url).
		Env("ME_CONFIG_MONGODB_AUTH_USERNAME", "admin").
		Env("ME_CONFIG_MONGODB_AUTH_PASSWORD", "password").
		PortBindings("tcp", "0.0.0.0", "8081", "8081").
		RestartAlways().
		Endpoint("mongo-cluster")
}
