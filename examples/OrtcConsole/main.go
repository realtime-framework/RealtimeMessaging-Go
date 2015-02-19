// OrtcConsole project main.go
package main

import (
	"bufio"
	"fmt"
	"github.com/realtime-framework/RealtimeMessaging-Go"
	"github.com/realtime-framework/RealtimeMessaging-Go/authentication"
	"os"
	"strconv"
	"strings"
)

const defaultServerUrl = "http://ortc-developers.realtime.co/server/2.1"
const defaultIsCluster = true
const defaultApplicationKey = "YOUR_APPLICATION_KEY"
const defaultPrivateKey = "YOUR_PRIVATE_KEY"
const defaultAuthenticationToken = "myToken"
const defaultNeedsAuthentication = false
const defaultMetadata = "GoApp"

var serverUrl string
var isCluster bool
var applicationKey string
var authenticationToken string
var metadata string

var channelConsole chan bool

func main() {

	fmt.Println("Welcome to Realtime Demo")

	readConnectionInfo()

	client := initClient()

	interfaceMenu()

	channelConsole = make(chan bool)

	//for {
	readMenuCommand(client)
	//<-channelConsole
	//}

}

func readInput() string {
	consolereader := bufio.NewReader(os.Stdin)

	input, err := consolereader.ReadString('\n')

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	return input
}

func readCommandIndex() int {
	var i int
	_, err := fmt.Scanf("%d\n", &i)
	if err != nil {
		fmt.Println("Error reading command: " + err.Error())
	}
	return i
}

func readConnectionInfo() {
	fmt.Println("Insert the server URL or press ENTER to use default (" + defaultServerUrl + "):")

	serverUrl = readInput()

	if serverUrl == "\n" {
		serverUrl = defaultServerUrl
		fmt.Println("Using default server URL: " + serverUrl)
	}

	strDefaultisCluster := ""

	if defaultIsCluster {
		strDefaultisCluster = "yes"
	} else {
		strDefaultisCluster = "no"
	}

	fmt.Println("Is it a cluster? Press y/n or Enter to use default (" + strDefaultisCluster + "):")

	readIsCluster := readInput()

	if readIsCluster == "\n" {
		isCluster = defaultIsCluster
		if isCluster {
			strDefaultisCluster = "yes"
		} else {
			strDefaultisCluster = "no"
		}
		fmt.Println("Using default is cluster: " + strDefaultisCluster)
	} else {
		isCluster = strings.EqualFold(readIsCluster, "y")
		if isCluster {
			fmt.Println("Yes")
		} else {
			fmt.Println("No")
		}
	}

	fmt.Println("Insert the application key or press ENTER to use default (" + defaultApplicationKey + "):")

	applicationKey = readInput()

	if applicationKey == "\n" {
		applicationKey = defaultApplicationKey
		fmt.Println("Using default application key: " + applicationKey)
	}

	fmt.Println("Insert the authentication token or press ENTER to use default (" + defaultAuthenticationToken + "):")

	authenticationToken = readInput()

	if authenticationToken == "\n" {
		authenticationToken = defaultAuthenticationToken
		fmt.Println("Using default authentication token: " + authenticationToken)
	}

	fmt.Println("Insert metadata or press ENTER to use default (" + defaultMetadata + "):")
	metadata = readInput()

	if metadata == "\n" {
		metadata = defaultMetadata
		fmt.Println("Using default metadata: " + metadata)
	}
}

func interfaceMenu() {
	fmt.Println("========== Ortc Integration Test Menu ==========")
	fmt.Println("")
	fmt.Println(" Commands List:")
	fmt.Println("")
	fmt.Println(" 0 - Connect ")
	fmt.Println(" 1 - Subscribe to channel ")
	fmt.Println(" 2 - Unsubscribe from channel ")
	fmt.Println(" 3 - Send message to channel ")
	fmt.Println(" 4 - Disconnect ")
	fmt.Println(" 5 - Presence ")
	fmt.Println(" 6 - Enable Presence ")
	fmt.Println(" 7 - Disable Presence ")
	fmt.Println("")
	fmt.Println("========== Ortc Integration Test Menu ==========")
}

func readMenuCommand(client *ortc.OrtcClient) {

	var channel string
	var message string

	for {
		fmt.Println("Enter command 0/1/2/3/4/5/6/7:")
		command := readCommandIndex()

		switch command {
		case 0:
			fmt.Println("Connecting...")
			client.Connect(applicationKey, authenticationToken, defaultMetadata, serverUrl, isCluster, defaultNeedsAuthentication)
			<-channelConsole
		case 1:
			fmt.Println("channel:")
			channel = readInput()
			channel = strings.Trim(channel, "\n")
			fmt.Println("Subscribing to " + channel + "...")
			client.Subscribe(channel, true)
			<-channelConsole
		case 2:
			fmt.Println("channel:")
			channel = readInput()
			channel = strings.Trim(channel, "\n")
			fmt.Println("Unsubscribing from " + channel + "...")
			client.Unsubscribe(channel)
			<-channelConsole
		case 3:
			fmt.Println("channel:")
			channel = readInput()
			channel = strings.Trim(channel, "\n")
			fmt.Println("message:")
			message = readInput()
			message = strings.Trim(message, "\n")
			fmt.Println("Sending " + message + " to " + channel + "...")
			client.Send(channel, message)
		case 4:
			fmt.Println("Disconnecting...")
			client.Disconnect()
			<-channelConsole
		case 5:
			fmt.Println("channel:")
			channel = readInput()
			channel = strings.Trim(channel, "\n")
			getPresence(channel)
			<-channelConsole
		case 6:
			fmt.Println("channel:")
			channel = readInput()
			fmt.Println("metadata(1,t, T,true,TRUE,True / 0,f,F,false,FALSE,False):")
			var metadataStr string
			metadataStr = readInput()
			metadataStr = strings.Trim(metadataStr, "\n")
			metadata, _ := strconv.ParseBool(metadataStr)
			enablePresence(channel, metadata)
			<-channelConsole
		case 7:
			fmt.Println("channel:")
			channel = readInput()
			disablePresence(channel)
			<-channelConsole
		default:
			fmt.Println("Invalid command")
		}
	}
}

func getPresence(channel string) {
	ch := make(chan ortc.PresenceStruct)
	go func() {
		msg := <-ch
		if msg.Err != nil {
			fmt.Println(msg.Err.Error())
		} else {
			fmt.Println("Subscriptions - ", msg.Result.Subscriptions)

			for key, value := range msg.Result.Metadata {
				fmt.Println(key + " - " + strconv.Itoa(value))
			}
		}
		channelConsole <- true
	}()
	ortc.GetPresence(serverUrl, isCluster, applicationKey, authenticationToken, channel, ch)
}

func enablePresence(channel string, metadata bool) {
	//fmt.Printf("Metadata value: %v\n", metadata)
	ch := make(chan ortc.PresenceType)
	go func() {
		msg := <-ch
		if msg.Err != nil {
			fmt.Println(msg.Err.Error())
		} else {
			fmt.Println(msg.Response)
		}
		channelConsole <- true
	}()
	fmt.Println("Enabling Presence...")
	ortc.EnablePresence(serverUrl, isCluster, applicationKey, defaultPrivateKey, channel, metadata, ch)
}

func disablePresence(channel string) {
	ch := make(chan ortc.PresenceType)
	go func() {
		msg := <-ch
		if msg.Err != nil {
			fmt.Println(msg.Err.Error())
		} else {
			fmt.Println(msg.Response)
		}
		channelConsole <- true
	}()
	fmt.Println("Disabling Presence...")
	ortc.DisablePresence(serverUrl, isCluster, applicationKey, defaultPrivateKey, channel, ch)
}

func initClient() *ortc.OrtcClient {

	if defaultNeedsAuthentication {
		fmt.Println("Authenticating...")

		permissions := make(map[string][]authentication.ChannelPermissions)

		yellowPermissions := []authentication.ChannelPermissions{}
		yellowPermissions = append(yellowPermissions, authentication.Write)
		yellowPermissions = append(yellowPermissions, authentication.Presence)

		testPermissions := []authentication.ChannelPermissions{}
		testPermissions = append(testPermissions, authentication.Read)
		testPermissions = append(testPermissions, authentication.Presence)

		permissions["yellow:*"] = yellowPermissions
		permissions["test:*"] = testPermissions

		if ortc.SaveAuthentication(serverUrl, isCluster, authenticationToken, false, applicationKey, 14000, defaultPrivateKey, permissions) {
			fmt.Println("Authentication successful")
		} else {
			fmt.Println("Unable to authenticate")
		}

	}

	client, onConnected, onDisconnected, onException, onMessage, onReconnected, onReconnecting, onSubscribed, onUnsubscribed := ortc.NewOrtcClient()

	go func() {
		for {
			select {
			case sender := <-onConnected:
				fmt.Println("CLIENT CONNECTED TO: " + sender.GetUrl())
				channelConsole <- true
			case sender := <-onDisconnected:
				fmt.Println("CLIENT DISCONNECTED FROM: " + sender.GetUrl())
				channelConsole <- true
			case sender := <-onException:
				fmt.Println("CLIENT EXCEPTION: " + sender.Err)
				channelConsole <- true
			case msgObj := <-onMessage:
				fmt.Println("RECEIVED MESSAGE: " + msgObj.Message + " ON CHANNEL: " + msgObj.Channel)
				//channelConsole <- true
			case sender := <-onReconnected:
				fmt.Println("CLIENT RECONNECTED TO: " + sender.GetUrl())
				channelConsole <- true
			case sender := <-onReconnecting:
				fmt.Println("CLIENT RECONNECTING TO: " + sender.GetUrl())
				//channelConsole <- true
			case sender := <-onSubscribed:
				fmt.Println("CLIENT SUBSCRIBED TO: " + sender.Channel)
				channelConsole <- true
			case sender := <-onUnsubscribed:
				fmt.Println("CLIENT UNSUBSCRIBED FROM: " + sender.Channel)
				channelConsole <- true
			}
		}
	}()

	return client
}

/*func connectOrtcClient() *ortcLib.OrtcClient {

	if defaultNeedsAuthentication {
		fmt.Println("Authenticating...")

		permissions := make(map[string][]authentication.ChannelPermissions)

		yellowPermissions := []authentication.ChannelPermissions{}
		yellowPermissions = append(yellowPermissions, authentication.Write)
		yellowPermissions = append(yellowPermissions, authentication.Presence)

		testPermissions := []authentication.ChannelPermissions{}
		testPermissions = append(testPermissions, authentication.Read)
		testPermissions = append(testPermissions, authentication.Presence)

		permissions["yellow:*"] = yellowPermissions
		permissions["test:*"] = testPermissions

		if ortcLib.SaveAuthentication(serverUrl, isCluster, authenticationToken, false, applicationKey, 14000, defaultPrivateKey, permissions) {
			fmt.Println("Authentication successful")
		} else {
			fmt.Println("Unable to authenticate")
		}

	}

	client, onConnected, onDisconnected, onException, onMessage, onReconnected, onReconnecting, onSubscribed, onUnsubscribed := ortcLib.NewOrtcClient()

	go func() {
		for {
			select {
			case sender := <-onConnected:
				fmt.Println("CLIENT CONNECTED TO: " + sender.GetUrl())

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

			case sender := <-onUnsubscribed:
				fmt.Println("CLIENT UNSUBSCRIBED FROM: " + sender.Channel)
			}
		}
	}()

	return client

	//fmt.Println("Connecting...")
	//client.Connect(applicationKey, authenticationToken, serverUrl, isCluster, defaultNeedsAuthentication)
}*/
