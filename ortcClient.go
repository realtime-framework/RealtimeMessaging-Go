package ortc

import (
	"fmt"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

const max_message_size = 800
const max_channel_size = 100
const max_connection_metadata_size = 256
const connection_timeout_default_value = 5000
const secure = "wss"
const unsecure = "ws"
const heartBeatTimeout = 30

var lastHeartBeat time.Time
var socket *websocket.Conn

var isSocketClosed bool

type channelPermission int

type pair struct {
	first  bool
	second string
}

type pairString struct {
	firtsStr  string
	secondStr string
}

const (
	read  channelPermission = iota // read == 0
	write                          // write == 1
)

type eventEnum int

const (
	onConnected eventEnum = iota
	onDisconnected
	onException
	onReconnected
	onReconnecting
	onSubscribed
	onUnsubscribed
	onReceived
)

type exceptionOrtc struct {
	Sender *OrtcClient
	Err    string
}

type subsOrtc struct {
	Sender  *OrtcClient
	Channel string
}

type OrtcClient struct {
	onConnectedChannel    chan *OrtcClient
	onDisconnectedChannel chan *OrtcClient
	onExceptionChannel    chan exceptionOrtc
	onMessageChannel      chan onMessageChannel
	onReconnectedChannel  chan *OrtcClient
	onReconnectingChannel chan *OrtcClient
	onSubscribedChannel   chan subsOrtc
	onUnsubscribedChannel chan subsOrtc

	clusterUrl             string
	serverUrl              string
	connectionMetadata     string
	announcementSubChannel string
	applicationKey         string
	authenticationToken    string
	needsAuthentication    bool

	uri               *url.URL
	connectionTimeout int
	id                int
	protocol          string

	subscribedChannels      map[string]channelSubscription
	channelsPermissions     map[string]string
	multiPartMessagesBuffer map[string][]bufferedMessage

	isCluster       bool
	isConnected     bool
	isDisconnecting bool
	isReconnecting  bool
	isConnecting    bool
}

//NewOrtcClient creates a new instance of OrtcClient.
//It returns the client instance and the correspondant ortc callback event channels for the client.
func NewOrtcClient() (c *OrtcClient, onConnected <-chan *OrtcClient, onDisconnected <-chan *OrtcClient, onException <-chan exceptionOrtc,
	onMessage <-chan onMessageChannel, onReconnected <-chan *OrtcClient, onReconnecting <-chan *OrtcClient, onSubscribed <-chan subsOrtc,
	onUnsubscribed <-chan subsOrtc) {

	c = new(OrtcClient)
	c.connectionTimeout = connection_timeout_default_value
	c.onConnectedChannel = make(chan *OrtcClient)
	c.onDisconnectedChannel = make(chan *OrtcClient)
	c.onExceptionChannel = make(chan exceptionOrtc)
	c.onMessageChannel = make(chan onMessageChannel)
	c.onReconnectedChannel = make(chan *OrtcClient)
	c.onReconnectingChannel = make(chan *OrtcClient)
	c.onSubscribedChannel = make(chan subsOrtc)
	c.onUnsubscribedChannel = make(chan subsOrtc)
	c.subscribedChannels = make(map[string]channelSubscription)
	c.channelsPermissions = make(map[string]string)
	c.multiPartMessagesBuffer = make(map[string][]bufferedMessage)
	return c, c.onConnectedChannel, c.onDisconnectedChannel, c.onExceptionChannel, c.onMessageChannel, c.onReconnectedChannel,
		c.onReconnectingChannel, c.onSubscribedChannel, c.onUnsubscribedChannel
}

//Connect connects the ortc client to the url previously specified.
func (client *OrtcClient) Connect(applicationKey, authenticationToken, metadata, serverUrl string, isCluster, needsAuthentication bool) {

	client.applicationKey = applicationKey
	client.authenticationToken = authenticationToken
	client.needsAuthentication = needsAuthentication
	client.connectionMetadata = metadata

	if isCluster {
		client.clusterUrl = serverUrl
		client.isCluster = true
	} else {
		client.serverUrl = serverUrl
		client.isCluster = false
	}

	if client.isConnected {
		raiseOrtcExceptionEvent(onException, client, ortcAlreadyConnectedException())
		//fmt.Printf("%s already connected", "Ortc")
	} else if len(client.clusterUrl) == 0 && len(client.serverUrl) == 0 {
		raiseOrtcExceptionEvent(onException, client, ortcEmptyFieldException("URL"))
		raiseOrtcExceptionEvent(onException, client, ortcEmptyFieldException("Cluster URL"))
		/*fmt.Printf("%s is null or empty", "URL")
		fmt.Printf("%s is null or empty", "Cluster url")*/
	} else if len(client.applicationKey) == 0 {
		raiseOrtcExceptionEvent(onException, client, ortcEmptyFieldException("Application key"))
		//fmt.Printf("%s is null or empty", "Application key")
	} else if len(client.authenticationToken) == 0 {
		raiseOrtcExceptionEvent(onException, client, ortcEmptyFieldException("Authentication key"))
		//fmt.Printf("%s is null or empty", "Authentication token")
	} else if !client.isCluster && !ortcIsValidUrl(client.serverUrl) {
		raiseOrtcExceptionEvent(onException, client, ortcInvalidCharactersException("URL"))
		//fmt.Printf("%s has invalid characters", "Url")
	} else if client.isCluster && !ortcIsValidUrl(client.clusterUrl) {
		raiseOrtcExceptionEvent(onException, client, ortcInvalidCharactersException("Cluster URL"))
		//fmt.Printf("%s has invalid characters", "Cluster url")
	} else if !ortcIsValidInput(client.applicationKey) {
		raiseOrtcExceptionEvent(onException, client, ortcInvalidCharactersException("Application key"))
		//fmt.Printf("%s has invalid characters", "Application key")
	} else if !ortcIsValidInput(client.authenticationToken) {
		raiseOrtcExceptionEvent(onException, client, ortcInvalidCharactersException("Authentication token"))
		//fmt.Printf("%s has invalid characters", "Authentication token")
	} else if len(client.announcementSubChannel) > 0 && !ortcIsValidInput(client.announcementSubChannel) {
		raiseOrtcExceptionEvent(onException, client, ortcInvalidCharactersException("Announcement Subchannel"))
		//fmt.Printf("%s has invalid characters", "Announcement Subchannel")
	} else if len(client.connectionMetadata) > 0 && len(client.connectionMetadata) > max_connection_metadata_size {
		raiseOrtcExceptionEvent(onException, client, ortcMaxLengthException("Connection metadata", max_connection_metadata_size))
		//fmt.Printf("%s size exceeds the limit of %d characters", "Connection metadata", max_connection_metadata_size)
	} else if client.isConnecting && !client.isReconnecting {
		raiseOrtcExceptionEvent(onException, client, ortcNotConnectedException("Already trying to connect"))
		//fmt.Printf("%s already trying to connect", "Ortc")
	} else {

		go func() {
			if !client.isReconnecting {
				client.isConnecting = true
			}
			client.applicationKey = applicationKey
			client.authenticationToken = authenticationToken

			if client.isCluster {
				clusterServer := getServerFromBalancer(client.clusterUrl, client.applicationKey)
				client.serverUrl = clusterServer
				client.isCluster = true
			}

			u, err := url.Parse(client.serverUrl)
			if err != nil {
				raiseOrtcExceptionEvent(onException, client, ortcInvalidCharactersException("URL"))
			}

			client.uri = u

			if "http" == client.uri.Scheme {
				client.protocol = unsecure
			} else {
				client.protocol = secure
			}

			host := strings.Split(u.Host, ":")[0]

			/*fmt.Println("u : " + u.String())

			fmt.Println("Scheme : " + scheme)
			fmt.Println("Host : " + host)*/

			rand.Seed(time.Now().UTC().UnixNano())

			randomNumber := rand.Intn(1000)
			randomNumberStr := strconv.Itoa(randomNumber)
			randomString := randString(8)
			connectionUrl := fmt.Sprintf("%s://%s/broadcast/%s/%s/websocket", client.protocol, host, randomNumberStr, randomString)

			//fmt.Println("connectionUrl : " + connectionUrl)

			c, _, err := websocket.DefaultDialer.Dial(connectionUrl, nil)

			if err != nil {
				raiseOrtcExceptionEvent(onException, client, ortcNotConnectedException("Could not connect. Check if the server is running correctly"))
				//fmt.Println("websocket.NewClient Error: %s\nResp:%+v", err, resp)
				if client.isReconnecting {
					raiseOrtcEvent(onReconnecting, client)
				}
				return
			}

			socket = c

			isSocketClosed = false

			c.SetReadLimit(max_message_size)

			for {
				_, message, err := c.ReadMessage()
				if err != nil {
					//fmt.Println(err.Error())
					if isSocketClosed {
						return
					}
					//raiseOrtcExceptionEvent(onException, client, err.Error())
					client.isReconnecting = true
					//closeHeartBeatRoutine()
					err = c.Close()

					if err != nil {
						//raiseOrtcExceptionEvent(onException, client, err.Error())
					}

					//closeHeartBeatRoutine()
					//fmt.Println("asddas")
					raiseOnDisconnected(client)
					return
				}

				lastHeartBeat = time.Now()
				wsMessage := string(message[:])
				//fmt.Println("Message from WS :" + wsMessage)
				if strings.EqualFold(wsMessage, "h") {
					//fmt.Println("h")
				} else {
					if strings.EqualFold(wsMessage, "o") {
						validateMessage := fmt.Sprintf("\"validate;%s;%s;%s;%s;%s\"", client.applicationKey, client.authenticationToken,
							client.announcementSubChannel, "", client.connectionMetadata)

						errWritesocket := c.WriteMessage(websocket.TextMessage, []byte(validateMessage))
						if errWritesocket != nil {
							raiseOrtcExceptionEvent(onException, client, errWritesocket.Error())
							//closeHeartBeatRoutine()
							err = c.Close()
							if err != nil {
								//raiseOrtcExceptionEvent(onException, client, err.Error())
							}
							//closeHeartBeatRoutine()
							raiseOnDisconnected(client)
							return
						}
					} else {
						ortcMsg, err := parseMessage(wsMessage)
						if err != nil {
							client.isReconnecting = true
							if c != nil {
								//closeHeartBeatRoutine()
								err = c.Close()
								if err != nil {
									//raiseOrtcExceptionEvent(onException, client, err.Error())
								}
								raiseOnDisconnected(client)
							} else {
								raiseOnDisconnected(client)
							}
							return
						}

						operation := ortcMsg.operation

						switch operation {
						case validated:
							client.channelsPermissions = ortcMsg.getPermissions()
							raiseOrtcEvent(onConnected, client)
							go func() {
								for {
									//fmt.Println("hearbeatRoutine")
									if client.isConnected {
										time.Sleep(heartBeatTimeout * time.Second)
										currentDate := time.Now()
										diff := time.Since(lastHeartBeat).Seconds() - time.Since(currentDate).Seconds()
										if diff > heartBeatTimeout {
											err := c.Close()
											if err != nil {
												//raiseOrtcExceptionEvent(onException, client, err.Error())
											}

											raiseOnDisconnected(client)
											return
										}
									} else {
										return
									}
								}
							}()
						case subscribed:
							raiseOrtcSubsEvent(onSubscribed, client, ortcMsg.channelSubscribed())
						case unsubscribed:
							raiseOrtcSubsEvent(onUnsubscribed, client, ortcMsg.channelUnsubscribed())
						case received:
							raiseOrtcReceivedEvent(onReceived, client, ortcMsg.messageChannel, ortcMsg.message, ortcMsg.messageId,
								ortcMsg.messagePart, ortcMsg.messageTotalParts)
						case errorOp:
							onError(client, ortcMsg)
						}
					}

				}
			}
		}()
	}
}

//GetUrl returns the url of the ortc client connection.
func (client *OrtcClient) GetUrl() string {
	if client.isCluster {
		return client.clusterUrl
	} else {
		return client.serverUrl
	}
}

func (client *OrtcClient) channelHasPermissions(channelName string, permission channelPermission) pair {
	result := pair{first: true, second: ""}
	//fmt.Println("channelHasPermissions: ")
	//fmt.Printf("%v", client.channelsPermissions)
	if len(client.channelsPermissions) > 0 {
		domainChannelCharIndex := strings.Index(channelName, ":")
		channelToValidate := channelName

		if domainChannelCharIndex > 0 {
			channelToValidate = fmt.Sprintf("%s%s", channelName[0:domainChannelCharIndex+1], "*")
		}

		hash := client.channelsPermissions[channelName]
		if len(hash) == 0 {
			hash = client.channelsPermissions[channelToValidate]
		}

		result = pair{first: !(len(hash) == 0), second: hash}
	}

	if !result.first {
		if permission == read {
			noPermission := "subscribe"
			errorMsg := fmt.Sprintf("No permission found to %s to the channel %s", noPermission, channelName)
			raiseOrtcExceptionEvent(onException, client, ortcDoesNotHavePermissionException(errorMsg))
			//fmt.Println(errorMsg)
		} else {
			noPermission := "send"
			errorMsg := fmt.Sprintf("No permission found to %s to the channel %s", noPermission, channelName)
			raiseOrtcExceptionEvent(onException, client, ortcDoesNotHavePermissionException(errorMsg))
			//fmt.Println(errorMsg)
		}
	}

	return result
}

func (client *OrtcClient) isSendValid(channelName, message string) pair {
	result := pair{first: true, second: ""}

	if !client.isConnected {
		//fmt.Println("Raise not connected exception")
		raiseOrtcExceptionEvent(onException, client, ortcNotConnectedException("Not connected"))
		result.first = false
	} else if len(channelName) == 0 {
		//fmt.Println("Channel name is empty")
		raiseOrtcExceptionEvent(onException, client, ortcEmptyFieldException("Channel"))
		result.first = false
	} else if !ortcIsValidInput(channelName) {
		//fmt.Println("Channel name is invalid")
		raiseOrtcExceptionEvent(onException, client, ortcInvalidCharactersException("Channel"))
		result.first = false
	} else if len(message) == 0 {
		//fmt.Println("Message empty")
		raiseOrtcExceptionEvent(onException, client, ortcEmptyFieldException("Message"))
		result.first = false
	} else if len(channelName) > max_channel_size {
		//fmt.Println("Channel name too big")
		raiseOrtcExceptionEvent(onException, client, ortcMaxLengthException("Channel", max_channel_size))
		result.first = false
	}

	if result.first {
		chPermissions := client.channelHasPermissions(channelName, write)
		result.first = chPermissions.first
		result.second = chPermissions.second
	}

	return result
}

func multiPartMessage(message, messageId string) []pairString {

	messageParts := []pairString{}

	messageBytes := []byte(message)

	totalParts := 0

	if len(messageBytes)%max_message_size == 0 {
		totalParts = len(messageBytes) / max_message_size
	} else {
		totalParts = len(messageBytes)/max_message_size + 1
	}

	messagePartIndex := 1
	currentPosition := 0
	messagePartSize := 0

	for currentPosition < len(messageBytes) {
		if len(messageBytes)-currentPosition > max_message_size {
			messagePartSize = max_message_size
		} else {
			messagePartSize = len(messageBytes) - currentPosition
		}

		if messagePartSize > 0 {
			messagePartBytes := make([]byte, messagePartSize)
			slice := messageBytes[currentPosition:]
			copy(messagePartBytes, slice)
			messagePartIdentifier := fmt.Sprintf("%s_%s-%s", messageId, strconv.Itoa(messagePartIndex), strconv.Itoa(totalParts))
			messageParts = append(messageParts, pairString{firtsStr: messagePartIdentifier, secondStr: string(messagePartBytes)})
			//fmt.Println("MultiPartMessage:")
			//fmt.Printf("%v\n", messageParts)
		}

		currentPosition = currentPosition + messagePartSize
		messagePartIndex = messagePartIndex + 1

	}

	return messageParts
}

//Send sends a message to the specified channel.
func (client *OrtcClient) Send(channel, message string) {
	sendValidation := client.isSendValid(channel, message)
	//fmt.Println("Send Validation First: " + strconv.FormatBool(sendValidation.first))
	if sendValidation.first {
		messageId := randString(8)
		messagesToSend := multiPartMessage(message, messageId)

		for _, messageToSend := range messagesToSend {
			client.send(channel, messageToSend.secondStr, messageToSend.firtsStr, sendValidation.second)
		}
	}
	return
}

func (client *OrtcClient) send(channel, message, messagePartIdentifier, permission string) {
	//fmt.Println("Sending message: " + message + " to channel:" + channel + " with partIdentifier: " + messagePartIdentifier + " and permission:" + permission)
	escapedMessage := strconv.Quote(message)
	if strings.EqualFold(escapedMessage[0:1], "\"") {
		escapedMessage = escapedMessage[1:]
	}
	if strings.EqualFold(escapedMessage[len(escapedMessage)-1:], "\"") {
		escapedMessage = escapedMessage[0 : len(escapedMessage)-1]
	}
	messageParsed := fmt.Sprintf("send;%s;%s;%s;%s;%s", client.applicationKey, client.authenticationToken, channel, permission, fmt.Sprintf("%s_%s", messagePartIdentifier, escapedMessage))
	//fmt.Println("Message Parsed: " + messageParsed)
	sendMessage(messageParsed, client)
}

func sendMessage(message string, c *OrtcClient) {
	msgFinal := fmt.Sprintf("\"%s\"", message)
	//fmt.Println(msgFinal)
	err := socket.WriteMessage(websocket.TextMessage, []byte(msgFinal))
	if err != nil {
		raiseOrtcExceptionEvent(onException, c, err.Error())
		//fmt.Println("Error sending message to socket: " + err.Error())
	}

}

//Subscribe subscribes the specified channel in order to receive messages in that channel.
func (c *OrtcClient) Subscribe(channel string, subscribeOnReconnect bool) <-chan onMessageChannel {
	subscribedChannel := c.subscribedChannels[channel]
	subscribeValidation := isSubscribeValid(c, channel, subscribedChannel)

	if subscribeValidation.first {
		subscribedChannel := channelSubscription{}
		subscribedChannel.subscribeOnReconnect = subscribeOnReconnect
		subscribedChannel.isSubscribing = true
		subscribedChannel.isSubscribed = false
		subscribedChannel.onMessage = make(chan onMessageChannel)
		c.subscribedChannels[channel] = subscribedChannel
		c.subscribe(channel, subscribeValidation.second)
		return subscribedChannel.onMessage
	}
	return nil
}

func (c *OrtcClient) subscribe(channel, permission string) {
	subscribeMsg := fmt.Sprintf("subscribe;%s;%s;%s;%s", c.applicationKey, c.authenticationToken, channel, permission)
	sendMessage(subscribeMsg, c)
}

func isSubscribeValid(c *OrtcClient, channelName string, channel channelSubscription) pair {
	result := pair{true, ""}

	if !c.isConnected {
		raiseOrtcExceptionEvent(onException, c, ortcNotConnectedException("Not Connected"))
		result.first = false
	} else if len(channelName) == 0 {
		raiseOrtcExceptionEvent(onException, c, ortcEmptyFieldException("Channel"))
		result.first = false
	} else if !ortcIsValidInput(channelName) {
		raiseOrtcExceptionEvent(onException, c, ortcEmptyFieldException("Channel"))
		result.first = false
	} else if channel.isSubscribing {
		raiseOrtcExceptionEvent(onException, c, ortcSubscribedException(fmt.Sprintf("Already subscribing to the channel %s", channelName)))
		result.first = false
	} else if channel.isSubscribed {
		raiseOrtcExceptionEvent(onException, c, ortcSubscribedException(fmt.Sprintf("Already subscribed to the channel %s", channelName)))
		result.first = false
	} else if len(channelName) > max_channel_size {
		raiseOrtcExceptionEvent(onException, c, ortcMaxLengthException("Channel", max_channel_size))
		result.first = false
	}

	if result.first {
		channelPermission := c.channelHasPermissions(channelName, read)
		result.first = channelPermission.first
		result.second = channelPermission.second
	}

	return result
}

//Stop receiving messages in the specified channel.
func (c *OrtcClient) Unsubscribe(channel string) {
	subscribedChannel := c.subscribedChannels[channel]

	if isUnsubscribeValid(c, channel, subscribedChannel) {
		subscribedChannel.subscribeOnReconnect = false
		unsubscribe(c, channel, true)
	}
}

func unsubscribe(c *OrtcClient, channel string, isValid bool) {
	if isValid {
		unsubscribeMessage := fmt.Sprintf("unsubscribe;%s;%s", c.applicationKey, channel)
		sendMessage(unsubscribeMessage, c)
	}
}

func isUnsubscribeValid(c *OrtcClient, channelName string, channel channelSubscription) bool {
	result := true

	//fmt.Printf("%v", channel)
	if !c.isConnected {
		raiseOrtcExceptionEvent(onException, c, ortcNotConnectedException("Not connected"))
		result = false
	} else if &channel == nil {
		raiseOrtcExceptionEvent(onException, c, ortcEmptyFieldException("Channel"))
		result = false
	} else if !ortcIsValidInput(channelName) {
		raiseOrtcExceptionEvent(onException, c, ortcInvalidCharactersException("Channel"))
		result = false
	} else if !channel.isSubscribed {
		raiseOrtcExceptionEvent(onException, c, ortcNotSubscribedException(channelName))
		result = false
	} else if len(channelName) > max_channel_size {
		raiseOrtcExceptionEvent(onException, c, ortcMaxLengthException("Channel", max_channel_size))
		result = false
	}

	return result
}

//Disconnect closes the current connection of the ortc client.
func (c *OrtcClient) Disconnect() {
	c.disconnect()
}

func (c *OrtcClient) disconnect() {
	c.isConnecting = false
	c.isDisconnecting = true

	if !c.isConnected && !c.isReconnecting {
		raiseOrtcExceptionEvent(onException, c, ortcNotConnectedException("Not connected"))
	} else if socket != nil {
		c.isReconnecting = false
		err := socket.Close()
		if err != nil {
			raiseOrtcExceptionEvent(onException, c, err.Error())
		}
		raiseOnDisconnected(c)
	}
}

func raiseOrtcEvent(ev eventEnum, c *OrtcClient) {
	switch ev {
	case onConnected:
		raiseOnConnected(c)
	case onDisconnected:
		raiseOnDisconnected(c)
	case onReconnected:
		raiseOnReconnected(c)
	case onReconnecting:
		raiseOnReconnecting(c)
	}
}

func raiseOrtcSubsEvent(ev eventEnum, c *OrtcClient, channel string) {
	switch ev {
	case onSubscribed:
		raiseOnSubscribed(c, channel)
	case onUnsubscribed:
		raiseOnUnsubscribed(c, channel)
	}
}

func raiseOrtcExceptionEvent(ev eventEnum, c *OrtcClient, errStr string) {
	switch ev {
	case onException:
		raiseOnException(c, errStr)
	}
}

func raiseOrtcReceivedEvent(ev eventEnum, c *OrtcClient, msgCh, msg, msgId string, msgPart, msgTotalParts int) {
	switch ev {
	case onReceived:
		raiseOnReceived(c, msgCh, msg, msgId, msgPart, msgTotalParts)
	}
}

func raiseOnConnected(c *OrtcClient) {
	c.isConnected = true
	c.isDisconnecting = false
	if c.isReconnecting && !c.isConnecting {
		raiseOrtcEvent(onReconnected, c)
	} else {
		c.isConnecting = false
		if c.onConnectedChannel != nil {
			c.onConnectedChannel <- c
		}
	}
}

func raiseOnDisconnected(c *OrtcClient) {
	c.channelsPermissions = make(map[string]string)
	if c.isDisconnecting || c.isConnecting {
		c.isConnected = false
		c.isDisconnecting = false
		c.isConnecting = false
		c.subscribedChannels = make(map[string]channelSubscription)
		if c.onDisconnectedChannel != nil {
			isSocketClosed = true
			//fmt.Println("Send disconnected")
			c.onDisconnectedChannel <- c
		}
	} else {
		if !c.isReconnecting {
			if c.onDisconnectedChannel != nil {
				isSocketClosed = true
				c.onDisconnectedChannel <- c
			}
		}
		raiseOrtcEvent(onReconnecting, c)
	}
}

func raiseOnException(c *OrtcClient, errStr string) {
	newException := exceptionOrtc{c, errStr}
	c.onExceptionChannel <- newException
}

func raiseOnReconnected(c *OrtcClient) {

	c.isReconnecting = false
	channelsToRemove := []string{}
	subscribedChannelsSet := []string{}
	for k := range c.subscribedChannels {
		subscribedChannelsSet = append(subscribedChannelsSet, k)
	}

	for _, channelName := range subscribedChannelsSet {
		subscribedChannel := c.subscribedChannels[channelName]
		if subscribedChannel.subscribeOnReconnected() {
			subscribedChannel.isSubscribing = true
			subscribedChannel.isSubscribed = false
			chPermission := c.channelHasPermissions(channelName, read)

			if chPermission.first {
				//fmt.Println("Subscribe " + channelName + "with permission " + chPermission.second)
				c.subscribe(channelName, chPermission.second)
			}
		} else {
			channelsToRemove = append(channelsToRemove, channelName)
		}
	}

	for _, channelToRemove := range channelsToRemove {
		delete(c.subscribedChannels, channelToRemove)
	}

	if c.onReconnectedChannel != nil {
		c.onReconnectedChannel <- c
	}

	//fmt.Println("RaiseOnReconnected")
}

func raiseOnReconnecting(c *OrtcClient) {
	if c.isReconnecting {
		time.Sleep(connection_timeout_default_value * time.Millisecond)
	}

	c.isConnected = false
	c.isDisconnecting = false
	c.isReconnecting = true

	if c.onReconnectingChannel != nil {
		c.onReconnectingChannel <- c
	}

	if c.isCluster {
		c.Connect(c.applicationKey, c.authenticationToken, c.connectionMetadata, c.clusterUrl, c.isCluster, c.needsAuthentication)
	} else {
		c.Connect(c.applicationKey, c.authenticationToken, c.connectionMetadata, c.serverUrl, c.isCluster, c.needsAuthentication)
	}

	//fmt.Println("RaiseOnReconnecting")
}

func raiseOnSubscribed(c *OrtcClient, channel string) {
	subscribedChannel := c.subscribedChannels[channel]
	subscribedChannel.isSubscribed = true
	subscribedChannel.isSubscribing = false
	c.subscribedChannels[channel] = subscribedChannel

	if c.onSubscribedChannel != nil {
		c.onSubscribedChannel <- subsOrtc{c, channel}
	}

	//fmt.Println("RaiseOnSubscribed")
}

func raiseOnUnsubscribed(c *OrtcClient, channel string) {
	subscribedChannel := c.subscribedChannels[channel]
	subscribedChannel.isSubscribed = false
	subscribedChannel.isSubscribing = false

	if c.onUnsubscribedChannel != nil {
		c.onUnsubscribedChannel <- subsOrtc{c, channel}
	}

	//fmt.Println("RaiseOnUnsubscribed")
}

// byMessagePart implements sort.Interface for []bufferedMessage based on
// the messagePart field.
type byMessagePart []bufferedMessage

func (a byMessagePart) Len() int           { return len(a) }
func (a byMessagePart) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byMessagePart) Less(i, j int) bool { return a[i].messagePart < a[j].messagePart }

func raiseOnReceived(c *OrtcClient, channel, message, messageId string, messagePart, messageTotalParts int) {

	if messagePart == -1 || (messagePart == 1 && messageTotalParts == 1) {

		subscription := c.subscribedChannels[channel]
		//fmt.Printf("Raise on Received Subscription: %v", subscription)
		if subscription.onMessage != nil {
			unescapedStr := strings.Replace(message, "\\\\\\", "", -1)
			//fmt.Println("UnescapedStr : " + unescapedStr)
			c.onMessageChannel <- onMessageChannel{c, channel, unescapedStr}
			if _, ok := c.multiPartMessagesBuffer[messageId]; ok {
				delete(c.multiPartMessagesBuffer, messageId)
			}
		}
	} else {
		if _, ok := c.multiPartMessagesBuffer[messageId]; !ok {
			c.multiPartMessagesBuffer[messageId] = []bufferedMessage{}
		}

		if c.multiPartMessagesBuffer[messageId] != nil {
			c.multiPartMessagesBuffer[messageId] = append(c.multiPartMessagesBuffer[messageId], bufferedMessage{messagePart, message})
		}

		messageParts := c.multiPartMessagesBuffer[messageId]

		if len(messageParts) == messageTotalParts {
			sort.Sort(byMessagePart(messageParts))
			fullMessage := ""
			for _, part := range messageParts {
				fullMessage = fmt.Sprintf("%s%s", fullMessage, part.content)
			}
			raiseOnReceived(c, channel, fullMessage, messageId, -1, -1)
		}
	}

	//fmt.Println("Raise on received")
}

func onError(c *OrtcClient, message *ortcMessage) {
	serverError, exceptionStr := message.serverError()
	var errorStr = ""
	if exceptionStr != "" {
		errorStr = exceptionStr
	} else {
		errorStr = serverError.message
	}

	if serverError != nil {
		so := serverError.operation
		switch so {
		case validate:
			c.validateServerError()
		case subscribe:
			c.cancelSubscription(serverError.channel)
		case subscribe_maxSize:
			c.channelMaxSizeError(serverError.channel)
		case unsubscribe_maxSize:
			c.channelMaxSizeError(serverError.channel)
		case send_maxSize:
			c.messageMaxSize()
		}
	}

	raiseOrtcExceptionEvent(onException, c, errorStr)
	//fmt.Println("Raise on error")
}

func (c *OrtcClient) stopReconnecting() {
	c.isDisconnecting = false
	c.isReconnecting = false
}

func (c *OrtcClient) validateServerError() {
	c.stopReconnecting()
	//closeHeartBeatRoutine()
	err := socket.Close()
	if err != nil {
		raiseOrtcExceptionEvent(onException, c, err.Error())
	}
	raiseOnDisconnected(c)
}

func (c *OrtcClient) cancelSubscription(channel string) {
	if len(channel) > 0 {
		if val, ok := c.subscribedChannels[channel]; ok {
			val.isSubscribing = false
		}
	}
}

func (c *OrtcClient) channelMaxSizeError(channel string) {
	c.cancelSubscription(channel)
	c.stopReconnecting()
	//closeHeartBeatRoutine()
	err := socket.Close()
	if err != nil {
		raiseOrtcExceptionEvent(onException, c, err.Error())
	}

	raiseOnDisconnected(c)
}

func (c *OrtcClient) messageMaxSize() {
	c.stopReconnecting()
	//closeHeartBeatRoutine()
	err := socket.Close()
	if err != nil {
		raiseOrtcExceptionEvent(onException, c, err.Error())
	}

	raiseOnDisconnected(c)
}
