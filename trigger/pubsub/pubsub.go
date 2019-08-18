package pubsub

import (
	"context"
	"fmt"

	"github.com/pboyd/subrun/trigger"

	"cloud.google.com/go/pubsub"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
)

type Trigger struct {
	ctx    context.Context
	cancel func()

	client *pubsub.Client
	id     string
	out    chan trigger.Message
	C      <-chan trigger.Message
}

func NewTrigger(id, project, topic string, options TriggerOpts) (*Trigger, error) {
	ctx, cancel := context.WithCancel(context.Background())

	client := options.ClientOverride
	if client == nil {
		var err error
		client, err = pubsub.NewClient(ctx, project, options.clientOpts()...)
		if err != nil {
			return nil, err
		}
	}

	trigger := &Trigger{
		ctx:    ctx,
		cancel: cancel,
		id:     id,
		client: client,
	}

	err := trigger.start(id, topic)
	if err != nil {
		trigger.Close()
		return nil, err
	}

	return trigger, nil
}

func (t *Trigger) Close() {
	t.cancel()
	t.client.Close()
	close(t.out)
}

func (t *Trigger) start(id, topic string) error {
	t.out = make(chan trigger.Message)
	t.C = t.out

	sub, err := t.getSubscription(topic)
	if err != nil {
		return err
	}

	go sub.Receive(t.ctx, func(ctx context.Context, msg *pubsub.Message) {
		t.processMessage(id, msg)
	})

	return nil
}

func (t *Trigger) getSubscription(topicName string) (*pubsub.Subscription, error) {
	topic := t.client.Topic(topicName)
	exists, err := topic.Exists(t.ctx)
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("subrun-%s-%s", t.id, topicName)
	sub := t.client.Subscription(id)

	exists, err = sub.Exists(t.ctx)
	if err != nil {
		return nil, err
	}

	if !exists {
		return t.client.CreateSubscription(t.ctx, id, pubsub.SubscriptionConfig{
			Topic: topic,
		})
	}

	return sub, nil
}

func (t *Trigger) processMessage(id string, msg *pubsub.Message) {
	tMsg := trigger.Message{
		SubscriptionID: id,
		Payload:        msg.Data,
		Callback: func(success bool) {
			if success {
				msg.Ack()
			} else {
				msg.Nack()
			}
		},
	}

	select {
	case <-t.ctx.Done():
	case t.out <- tMsg:
	}
}

type TriggerOpts struct {
	CredentialsFile string
	ClientOverride  *pubsub.Client
}

func (o TriggerOpts) clientOpts() []option.ClientOption {
	opts := []option.ClientOption{}

	if o.CredentialsFile != "" {
		opts = append(opts, option.WithCredentialsFile(o.CredentialsFile))
	}

	return opts
}

func PubSubEmulatorOpts(ctx context.Context, addr, project string) (TriggerOpts, error) {
	grpcConn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return TriggerOpts{}, err
	}

	client, err := pubsub.NewClient(ctx, project, option.WithGRPCConn(grpcConn))
	if err != nil {
		return TriggerOpts{}, err
	}

	return TriggerOpts{
		ClientOverride: client,
	}, nil
}
