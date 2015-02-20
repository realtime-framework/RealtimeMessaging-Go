// Copyright 2015 Realtime Framework. 
//
// The ortc package implements the Go lang version of the Realtime Messaging protocol,
//
// If your application has data that needs to be updated in the user’s interface as it changes (e.g. real-time stock quotes or ever changing social news feed) 
// Realtime Messaging is the reliable, easy, unbelievably fast, “works everywhere” solution.
//
// Installation:
//
//	go get github.com/realtime-framework/RealtimeMessaging-Go
//
// Below are examples of use of the ortc package:
//
// - Create a new instance of ortc client:
//
// client, onConnected, onDisconnected, onException, onMessage, onReconnected, onReconnecting, onSubscribed, onUnsubscribed := ortc.NewOrtcClient()
//
// - Using the channels received on the ortc client:
//
// 		//Create a go routine that start listening from the Ortc events channels.
// 		go func() {
//			for {
//			select {
//				case sender := <-onConnected:
//					fmt.Println("CLIENT CONNECTED TO: " + sender.GetUrl())
//				case sender := <-onDisconnected:
//					fmt.Println("CLIENT DISCONNECTED FROM: " + sender.GetUrl())
//				case sender := <-onException:
//					fmt.Println("CLIENT EXCEPTION: " + sender.Err)
//				case msgObj := <-onMessage:
//					fmt.Println("RECEIVED MESSAGE: " + msgObj.Message + " ON CHANNEL: " + msgObj.Channel)
//				case sender := <-onReconnected:
//					fmt.Println("CLIENT RECONNECTED TO: " + sender.GetUrl())
//				case sender := <-onReconnecting:
//					fmt.Println("CLIENT RECONNECTING TO: " + sender.GetUrl())
//				case sender := <-onSubscribed:
//					fmt.Println("CLIENT SUBSCRIBED TO: " + sender.Channel)
//				case sender := <-onUnsubscribed:
//					fmt.Println("CLIENT UNSUBSCRIBED FROM: " + sender.Channel)
//				}
//	 		}
// 		}()
//
// - Connect to a ortc server:
//
// client.Connect("YOUR_APPLICATION_KEY", "myToken", "GoApp", "http://ortc-developers.realtime.co/server/2.1", true, false)
//
// - Disconnect from ortc server:
//
// client.Disconnect()
//
// - Disable presence on a channel:
//
// ch := make(chan ortc.PresenceType)
// go func() {
//		msg := <-ch
//		if msg.Err != nil {
//			fmt.Println(msg.Err.Error())
//		} else {
//			fmt.Println(msg.Response)
//		}
//		channelConsole <- true
//	}()
//
//	ortc.DisablePresence("http://ortc-developers.realtime.co/server/2.1", true, "YOUR_APPLICATION_KEY", "YOUR_PRIVATE_KEY", "my_channel", ch)
//
// - Enable presence on a channel:
//
//  ch := make(chan ortc.PresenceType)
//	go func() {
//		msg := <-ch
//		if msg.Err != nil {
//			fmt.Println(msg.Err.Error())
//		} else {
//			fmt.Println(msg.Response)
//		}
//	}()
//
//	ortc.EnablePresence("http://ortc-developers.realtime.co/server/2.1", true, "YOUR_APPLICATION_KEY", "YOUR_PRIVATE_KEY", "my_channel", ch)
//
// - Get presence on channel:
//
//  ch := make(chan ortc.PresenceStruct)
//  go func() {
//	 msg := <-ch
//	 if msg.Err != nil {
//		fmt.Println(msg.Err.Error())
//	 } else {
//		fmt.Println("Subscriptions - ", msg.Result.Subscriptions)
//
//		for key, value := range msg.Result.Metadata {
//			fmt.Println(key + " - " + strconv.Itoa(value))
//		}
//	  }
//   }()
//
//  ortc.GetPresence("http://ortc-developers.realtime.co/server/2.1", true, "YOUR_APPLICATION_KEY", "myToken", "my_channel", ch)
//
// - Save Authentication:
//
// permissions := make(map[string][]authentication.ChannelPermissions)
//
// yellowPermissions := []authentication.ChannelPermissions{}
// yellowPermissions = append(yellowPermissions, authentication.Write)
// yellowPermissions = append(yellowPermissions, authentication.Presence)
//
// testPermissions := []authentication.ChannelPermissions{}
// testPermissions = append(testPermissions, authentication.Read)
// testPermissions = append(testPermissions, authentication.Presence)
//
// permissions["yellow:*"] = yellowPermissions
// permissions["test:*"] = testPermissions
//
// if ortc.SaveAuthentication("http://ortc-developers.realtime.co/server/2.1", true, "myToken", false, "YOUR_APPLICATION_KEY", 14000, "YOUR_PRIVATE_KEY", permissions) {
//	fmt.Println("Authentication successful")
// } else {
//	fmt.Println("Unable to authenticate")
// }
//
// - Send message to a channel:
//
// client.Send("my_channel", "Hello World!")
//
// - Subscribe to a channel:
//
// client.Subscribe("my_channel", true)
//
// - Unsubscribe from a channel:
//
// client.Unsubscribe("my_channel")
//
// More documentation about the Realtime Messaging service (ORTC) can be found at: 
// http://messaging-public.realtime.co/documentation/starting-guide/overview.html
package ortc
