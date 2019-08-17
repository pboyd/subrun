package pubsub

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

// These tests require the PubSub emulator: https://cloud.google.com/pubsub/docs/emulator

var pubSubFakeAddress = os.Getenv("PUBSUB_EMULATOR_HOST")

func testPubSubClient(ctx context.Context, t *testing.T, project string) *pubsub.Client {
	if pubSubFakeAddress == "" {
		t.Skipf("Skipping. Set PUBSUB_EMULATOR_HOST=localhost:8085 to run")
	}

	grpcConn, err := grpc.Dial(pubSubFakeAddress, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("error dialing GRPC: %v", err)
	}

	client, err := pubsub.NewClient(ctx, project, option.WithGRPCConn(grpcConn))
	if err != nil {
		t.Fatalf("error creating client: %v", err)
	}

	return client
}

func testCreateTopic(ctx context.Context, t *testing.T, client *pubsub.Client, id string) *pubsub.Topic {
	topic := client.Topic(id)
	exists, err := topic.Exists(ctx)
	if err != nil {
		t.Fatalf("error checking topic: %v", err)
	}

	if exists {
		return topic
	}

	topic, err = client.CreateTopic(ctx, id)
	if err != nil {
		t.Fatalf("error creating topic: %v", err)
	}

	return topic
}

func TestTrigger(t *testing.T) {
	const (
		projectName = "fake-project"
		topicName   = "foobar"
		subrunID    = "trigger-test"
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := testPubSubClient(ctx, t, projectName)
	topic := testCreateTopic(ctx, t, client, topicName)

	trigger, err := NewTrigger(subrunID, projectName, topicName, TriggerOpts{
		ClientOverride: client,
	})
	if err != nil {
		t.Fatalf("error starting trigger: %v", err)
	}

	data := []byte("test data")
	topic.Publish(ctx, &pubsub.Message{Data: data})
	topic.Stop()

	select {
	case <-ctx.Done():
		t.Fatalf("time out waiting for trigger to fire")
	case msg := <-trigger.C:
		if msg.SubscriptionID != subrunID {
			t.Errorf("got ID %q, want %q", msg.SubscriptionID, subrunID)
		}

		if !bytes.Equal(msg.Payload, data) {
			t.Errorf("got payload %q, want %q", string(msg.Payload), string(data))
		}

		msg.Callback(true)
	}
}
