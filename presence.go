package ortc

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type presence struct {
	Subscriptions int
	Metadata      map[string]int
}

//PresenceStruct is the callback channel type of GetPresence() method.
//It contains the field Err of type error and Result of with the number of subscriptions and the map of metadata.
//If there was an error Err is not nil and Result empty, otherwise Err is nil and Result not empty.
type PresenceStruct struct {
	Err    error
	Result presence
}

//PresenceType is the callback channel type of Disable/EnablePresence() methods.
//It contains the field  Err of type error and a Response string.
//If there was an error Err is not nil and Response empty, otherwise Err is nil and Response is not empty.
type PresenceType struct {
	Err      error
	Response string
}

type httpResponse struct {
	url      string
	response *http.Response
	err      error
}

//Gets the subscriptions in the specified channel and if active the first 100 unique metadata.
//If success writes the result to the callback channel.
//An error is returned if there was an error on the request.
func GetPresence(url string, isCluster bool, applicationKey string, authenticationToken string, channel string, callback chan<- PresenceStruct) {
	presenceUrl := getServerFromBalancer(url, applicationKey)
	result := getPresence(presenceUrl, isCluster, applicationKey, authenticationToken, channel, callback)
	if result.err != nil {
		callback <- PresenceStruct{result.err, presence{}}
	} else {
		defer result.response.Body.Close()
		contents, err := ioutil.ReadAll(result.response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			return
		}
		callback <- PresenceStruct{nil, deserialize(string(contents[:]))}
	}
}

func getPresence(url string, isCluster bool, applicationKey string, authenticationToken string, channel string, callback chan<- PresenceStruct) *httpResponse {
	presenceUrl := ""
	if len(url) > 0 {
		if url[len(url)-1] == '/' {
			presenceUrl = url
		} else {
			presenceUrl = fmt.Sprintf("%s/", url)
		}
	}

	presenceUrl = presenceUrl + fmt.Sprintf("presence/%s/%s/%s", applicationKey, authenticationToken, channel)
	//fmt.Println("presenceUrl: " + presenceUrl)
	result := asyncHttpGet(presenceUrl)
	return result
}

// EnablePresence enables presence for the specified channel.
// If success writes the result to the callback channel.
// An error is returned if there was an error on the request.
func EnablePresence(url string, isCluster bool, applicationKey string, privateKey string, channel string, metadata bool, callback chan<- PresenceType) {
	presenceUrl := getServerFromBalancer(url, applicationKey)
	result := enablePresence(presenceUrl, isCluster, applicationKey, privateKey, channel, metadata, callback)
	if result.err != nil {
		callback <- PresenceType{result.err, ""}
	} else {
		defer result.response.Body.Close()
		contents, err := ioutil.ReadAll(result.response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			return
		}
		callback <- PresenceType{nil, string(contents[:])}
	}
}

func enablePresence(url1 string, isCluster bool, applicationKey string, privateKey string, channel string, metadata bool, callback chan<- PresenceType) *httpResponse {
	presenceUrl := ""
	if len(url1) > 0 {
		if url1[len(url1)-1] == '/' {
			presenceUrl = url1
		} else {
			presenceUrl = fmt.Sprintf("%s/", url1)
		}

	}

	presenceUrl = presenceUrl + fmt.Sprintf("presence/enable/%s/%s", applicationKey, channel)
	postBody := url.Values{"privatekey": {privateKey}}

	if metadata {
		postBody.Add("metadata", "1")
	}

	result := asyncHttpPost(presenceUrl, postBody)
	return result

}

// DisablePresence disables presence for the specified channel.
// If success writes the result to the callback channel.
// An error is returned if there was an error on the request.
func DisablePresence(url string, isCluster bool, applicationKey string, privateKey string, channel string, callback chan<- PresenceType) {
	presenceUrl := getServerFromBalancer(url, applicationKey)
	result := disablePresence(presenceUrl, isCluster, applicationKey, privateKey, channel, callback)
	if result.err != nil {
		callback <- PresenceType{result.err, ""}
	} else {
		defer result.response.Body.Close()
		contents, err := ioutil.ReadAll(result.response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			return
		}
		callback <- PresenceType{nil, string(contents[:])}
	}
}

func disablePresence(url1 string, isCluster bool, applicationKey string, privateKey string, channel string, onDisablePresence chan<- PresenceType) *httpResponse {
	presenceUrl := ""

	if len(url1) > 0 {
		if url1[len(url1)-1] == '/' {
			presenceUrl = url1
		} else {
			presenceUrl = fmt.Sprintf("%s/", url1)
		}

	}

	presenceUrl = presenceUrl + fmt.Sprintf("presence/disable/%s/%s", applicationKey, channel)
	postBody := url.Values{"privatekey": {privateKey}}

	result := asyncHttpPost(presenceUrl, postBody)
	return result
}

func asyncHttpGet(url string) *httpResponse {
	ch := make(chan *httpResponse)
	var response *httpResponse
	go func(url string) {
		//fmt.Printf("Fetching %s \n", url)
		resp, err := http.Get(url)
		ch <- &httpResponse{url, resp, err}
	}(url)

	for {
		select {
		case r := <-ch:
			//fmt.Printf("%s was fetched\n", r.url)
			response = r
			return response
		}
	}
	return response
}

func asyncHttpPost(url string, postBody url.Values) *httpResponse {
	ch := make(chan *httpResponse)
	var response *httpResponse
	go func(url string) {
		//fmt.Printf("Fetching %s \n", url)
		resp, err := http.PostForm(url, postBody)
		ch <- &httpResponse{url, resp, err}
	}(url)

	for {
		select {
		case r := <-ch:
			//fmt.Printf("%s was fetched\n", r.url)
			response = r
			return response
		}
	}
	return response
}

func deserialize(message string) presence {
	//fmt.Println("Deserialize message " + message)
	var result presence
	if len(message) > 0 {
		json.Unmarshal([]byte(message), &result)
	}
	return result
}
