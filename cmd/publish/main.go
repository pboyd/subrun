package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

func main() {
	var message string
	flag.StringVar(&message, "message", "", "The body of the message to publish")
	flag.Parse()

	topicID := flag.Arg(0)
	if topicID == "" {
		fmt.Fprintf(os.Stderr, "usage: %s <topic> -message <message>\n", os.Args[0])
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := connect(ctx)
	if err != nil {
		log.Fatalf("error connecting to pubsub emulator: %v", err)
	}
	defer client.Close()

	topic, err := getOrCreateTopic(ctx, client, topicID)
	if err != nil {
		log.Fatalf("error getting topic: %v", err)
	}
	defer topic.Stop()

	psm := &pubsub.Message{}
	if message != "" {
		psm.Data = []byte(message)
	}

	result := topic.Publish(ctx, psm)

	id, err := result.Get(ctx)
	if err != nil {
		log.Fatalf("error getting publish status: %v", err)
	}

	fmt.Println(id)
}

func connect(ctx context.Context) (*pubsub.Client, error) {
	host := os.Getenv("PUBSUB_EMULATOR_HOST")
	if host == "" {
		return nil, errors.New("PUBSUB_EMULATOR_HOST not set")
	}

	project := os.Getenv("PUBSUB_PROJECT_ID")
	if project == "" {
		project = "fake-project"

		log.Printf("PUBSUB_PROJECT_ID not set, defaulting to %q", project)
	}

	grpcConn, err := grpc.Dial(host, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return pubsub.NewClient(ctx, project, option.WithGRPCConn(grpcConn))
}

func getOrCreateTopic(ctx context.Context, client *pubsub.Client, id string) (*pubsub.Topic, error) {
	topic := client.Topic(id)
	exists, err := topic.Exists(ctx)
	if err != nil {
		return nil, err
	}

	if exists {
		return topic, nil
	}

	topic, err = client.CreateTopic(ctx, id)
	if err != nil {
		return nil, err
	}

	return topic, nil
}
