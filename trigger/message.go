package trigger

type Message struct {
	SubscriptionID string
	Payload        []byte
	Callback       func(success bool)
}
