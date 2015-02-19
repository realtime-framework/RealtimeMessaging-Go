package ortc

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

func getServerFromBalancer(balancerUrl, applicationKey string) string {

	match, err := regexp.MatchString(`^(http(s)?).*$`, balancerUrl)

	if err != nil {
		return "Error retrieving server from balancer"
	}

	protocol := ""

	if !match {
		protocol = balancerUrl
	}

	parsedUrl := fmt.Sprintf("%s%s", protocol, balancerUrl)

	//fmt.Println("ParsedUrl " + parsedUrl)

	if len(applicationKey) > 0 {
		parsedUrl = parsedUrl + fmt.Sprintf("?appkey=%s", applicationKey)
	}

	//fmt.Println("ParsedUrl after if: " + parsedUrl)
	clusterServer := unsecureRequest(parsedUrl)

	return clusterServer
}

func unsecureRequest(url string) string {

	resp, err := http.Get(url)

	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return ""
	}

	line := string(body[:])
	//fmt.Println("Line: " + line)

	match, _ := regexp.MatchString(`^var SOCKET_SERVER = "(http.*)";$`, line)

	if !match {
		return "Server returned invalid server"
	}

	balancerUrl := strings.Split(line, "=")[1]
	balancerUrl = strings.Replace(balancerUrl, ";", "", -1)
	balancerUrl = strings.Replace(balancerUrl, "\"", "", -1)
	balancerUrl = strings.TrimSpace(balancerUrl)

	//fmt.Println("Balancer: " + balancerUrl)

	return balancerUrl
}
