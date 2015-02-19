package ortc

type onMessageChannel struct {
	Sender  *OrtcClient
	Channel string
	Message string
}

type channelSubscription struct {
	isSubscribing        bool
	isSubscribed         bool
	subscribeOnReconnect bool
	onMessage            chan onMessageChannel
}

func newChannelSubscription(subscribeOnReconnected bool) *channelSubscription {
	chanSub := new(channelSubscription)
	chanSub.subscribeOnReconnect = subscribeOnReconnected
	chanSub.onMessage = make(chan onMessageChannel)
	chanSub.isSubscribed = false
	chanSub.isSubscribing = false
	return chanSub
}

func (c *channelSubscription) subscribing() bool {
	return c.isSubscribing
}

func (c *channelSubscription) subscribed() bool {
	return c.isSubscribed
}

func (c *channelSubscription) subscribeOnReconnected() bool {
	return c.subscribeOnReconnect
}
