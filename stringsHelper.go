package ortc

import (
	"math/rand"
	"net/url"
	"regexp"
	"time"
)

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func ortcIsValidUrl(uri string) bool {

	if len(uri) > 0 {
		_, err := url.Parse(uri)
		if err != nil {
			return false
		}
	} else {
		return false
	}
	return true
}

func ortcIsValidInput(input string) bool {
	match, _ := regexp.MatchString(`[\w-:\/.]*$`, input)
	return match
}

func randString(n int) string {
	rand.Seed(time.Now().UTC().UnixNano())
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
