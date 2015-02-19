package ortc

import (
	"fmt"
	"net/url"
	"ortc/authentication"
)

//SaveAuthentication saves the authentication token channels permissions in the ORTC server.
//It returns true if authentication succeeds, otherwise false.
func SaveAuthentication(url1 string, isCluster bool, authenticationToken string, authenticationTokenIsPrivate bool, applicationKey string,
	timeToLive int, privateKey string, permissions map[string][]authentication.ChannelPermissions) bool {

	connectionUrl := url1

	if isCluster {
		connectionUrl = getServerFromBalancer(url1, applicationKey)
	}

	isAuthenticated := false

	u, err := url.Parse(fmt.Sprintf("%s/authenticate", connectionUrl))
	if err != nil {
		fmt.Println("Error /authenticate")
		ortcAuthenticationNotAuthorizedException(err.Error())
	}

	authenticationUrl := u
	//fmt.Println("AuthenticationUrl : " + authenticationUrl.String())

	isAuthenticated, err = authentication.SaveAuthentication(authenticationUrl, authenticationToken,
		authenticationTokenIsPrivate, applicationKey, timeToLive, privateKey, permissions)

	if err != nil {
		ortcAuthenticationNotAuthorizedException(err.Error())
	}

	return isAuthenticated
}
