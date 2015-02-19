// authentication
package authentication

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func SaveAuthentication(authenticationUrl *url.URL, authenticationToken string, authenticationTokenIsPrivate bool,
	applicationKey string, timeToLive int, privateKey string, permissions map[string][]ChannelPermissions) (bool, error) {

	strAuthenticationTokenIsPrivate := ""

	if authenticationTokenIsPrivate {
		strAuthenticationTokenIsPrivate = "1"
	} else {
		strAuthenticationTokenIsPrivate = "0"
	}

	postBody := fmt.Sprintf("AT=%s&AK=%s&PK=%s&TTL=%s&TP=%s&PVT=%s", authenticationToken, applicationKey, privateKey,
		strconv.Itoa(timeToLive), strconv.Itoa(len(permissions)), strAuthenticationTokenIsPrivate)

	for channelNamePerms, channelPermissions := range permissions {
		channelPermissionText := ""
		for _, channelPermission := range channelPermissions {
			channelPermissionText = channelPermissionText + channelPermission.String()
		}

		chPermission := fmt.Sprintf("&%s=%s", channelNamePerms, channelPermissionText)
		postBody = fmt.Sprintf("%s%s", postBody, chPermission)
	}

	isAuthenticated, err := postRequest(authenticationUrl, postBody)

	return isAuthenticated, err

}

func postRequest(url *url.URL, postBody string) (bool, error) {
	//fmt.Println("PostBody : " + postBody)
	client := &http.Client{}
	r, err := http.NewRequest("POST", url.String(), bytes.NewBufferString(postBody))
	if err != nil {
		return false, err
	}

	resp, err := client.Do(r)
	if err != nil {
		return false, err
	}
	//fmt.Println(resp.Status)
	return strings.EqualFold(strconv.Itoa(resp.StatusCode), "201"), nil
}

/*apiUrl := "https://api.com"
  resource := "/user/"
  data := url.Values{}
  data.Set("name", "foo")
  data.Add("surname", "bar")

  u, _ := url.ParseRequestURI(apiUrl)
  u.Path = resource
  urlStr := fmt.Sprintf("%v", u) // "https://api.com/user/"

  client := &http.Client{}
  r, _ := http.NewRequest("POST", urlStr, bytes.NewBufferString(data.Encode())) // <-- URL-encoded payload
  r.Header.Add("Authorization", "auth_token=\"XXXXXXX\"")
  r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
  r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

  resp, _ := client.Do(r)
  fmt.Println(resp.Status)*/
