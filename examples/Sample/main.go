// Sample project main.go
package main

import (
	"bufio"
	"fmt"
	"github.com/realtime-framework/RealtimeMessaging-Go"
	"os"
)

const defaultServerUrl = "http://ortc-developers.realtime.co/server/2.1"
const defaultIsCluster = true
const defaultApplicationKey = "YOUR_APPLICATION_KEY"
const defaultPrivateKey = "YOUR_PRIVATE_KEY"
const defaultAuthenticationToken = "myToken"
const defaultNeedsAuthentication = false
const defaultMetadata = "GoApp"

func main() {

	//Creates a new Ortc client instance with the respective channels events.
	client, onConnected, onDisconnected, onException, onMessage, onReconnected, onReconnecting, onSubscribed, onUnsubscribed := ortc.NewOrtcClient()

	//Read from the Ortc events channels.
	go func() {
		for {
			select {
			case sender := <-onConnected:
				fmt.Println("CLIENT CONNECTED TO: " + sender.GetUrl())
				client.Subscribe("my_channel", true)
			case sender := <-onDisconnected:
				fmt.Println("CLIENT DISCONNECTED FROM: " + sender.GetUrl())
			case sender := <-onException:
				fmt.Println("CLIENT EXCEPTION: " + sender.Err)
			case msgObj := <-onMessage:
				fmt.Println("RECEIVED MESSAGE: " + msgObj.Message + " ON CHANNEL: " + msgObj.Channel)
			case sender := <-onReconnected:
				fmt.Println("CLIENT RECONNECTED TO: " + sender.GetUrl())
			case sender := <-onReconnecting:
				fmt.Println("CLIENT RECONNECTING TO: " + sender.GetUrl())
			case sender := <-onSubscribed:
				fmt.Println("CLIENT SUBSCRIBED TO: " + sender.Channel)
				client.Send("my_channel", "Hello World!")
			case sender := <-onUnsubscribed:
				fmt.Println("CLIENT UNSUBSCRIBED FROM: " + sender.Channel)
			}
		}
	}()

	//Connects the client to the Ortc
	client.Connect(defaultApplicationKey, defaultAuthenticationToken, defaultMetadata, defaultServerUrl, defaultIsCluster, defaultNeedsAuthentication)

	fmt.Println("Press Enter to Exit")

	consolereader := bufio.NewReader(os.Stdin)

	_, err := consolereader.ReadString('\n')

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
