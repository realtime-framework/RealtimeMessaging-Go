package ortc

import (
	"encoding/json"
	"errors"
	//"fmt"
	"regexp"
	"strconv"
	"strings"
)

const operation_pattern = `^a\["{\\"op\\":\\"([^\"]+)\\",(.*)\}"\]$`
const channel_pattern = `^\\"ch\\":\\"(.*)\\"$`
const received_pattern = `^a\["{\\"ch\\":\\"([^\"]+)\\",\\"m\\":\\"([\s\S]*?)\\"}"]$`
const multi_part_message_pattern = `^(.[^_]*)_(.[^-]*)-(.[^_]*)_([\s\S]*?)$`
const exception_pattern = `^\\"ex\\":(\{.*\})$`
const permissions_pattern = `^\\"up\\":{1}(.*),\\"set\\":(.*)$`

type ortcOperation int

const (
	validated ortcOperation = iota
	subscribed
	unsubscribed
	errorOp
	received
)

var operationIndex = make(map[string]ortcOperation)
var errorOperationIndex = make(map[string]ortcServerErrorOperation)

type ortcMessage struct {
	operation         ortcOperation
	message           string
	messageChannel    string
	messageId         string
	messagePart       int
	messageTotalParts int
}

func newOrtcMessage(operation ortcOperation, message, messageChannel, messageId string, messagePart, messageTotalParts int) *ortcMessage {

	ortcMsg := new(ortcMessage)
	ortcMsg.operation = operation
	ortcMsg.message = message
	ortcMsg.messageChannel = messageChannel
	ortcMsg.messageId = messageId
	ortcMsg.messagePart = messagePart
	ortcMsg.messageTotalParts = messageTotalParts

	return ortcMsg
}

func (o *ortcMessage) channelSubscribed() string {

	matcher := regexp.MustCompile(channel_pattern)

	if !matcher.MatchString(o.message) {
		return "Subscribe channel match to channel not found"
	}

	result := matcher.FindStringSubmatch(o.message)
	return result[1]
}

func (o *ortcMessage) channelUnsubscribed() string {

	matcher := regexp.MustCompile(channel_pattern)

	if !matcher.MatchString(o.message) {
		return "Unsubscribe channel match to channel not found"
	}

	result := matcher.FindStringSubmatch(o.message)
	return result[1]
}

func (o *ortcMessage) getPermissions() map[string]string {
	//fmt.Println("o.message :" + o.message)
	permissionsMap := make(map[string]string)
	matcher := regexp.MustCompile(permissions_pattern)
	if matcher.MatchString(o.message) {
		subMatches := matcher.FindStringSubmatch(o.message)
		content := subMatches[1]
		content = strings.Replace(content, "\\", "", -1)
		//fmt.Println("content: " + content)
		var jsonBlob = []byte(content)
		err := json.Unmarshal(jsonBlob, &permissionsMap)
		if err != nil {
			return permissionsMap
		}
	}

	//fmt.Printf("%+v", permissionsMap)
	return permissionsMap
}

func parseMessage(message string) (*ortcMessage, error) {

	//fmt.Println("parseMessage: " + message)

	operationIndex["ortc-validated"] = validated
	operationIndex["ortc-subscribed"] = subscribed
	operationIndex["ortc-unsubscribed"] = unsubscribed
	operationIndex["ortc-error"] = errorOp

	errorOperationIndex["ex"] = unexpected
	errorOperationIndex["validate"] = validate
	errorOperationIndex["subscribe"] = subscribe
	errorOperationIndex["subscribe_maxsize"] = subscribe_maxSize
	errorOperationIndex["unsubscribe_maxsize"] = unsubscribe_maxSize
	errorOperationIndex["send_maxsize"] = send_maxSize

	var operation ortcOperation
	var parsedMessage string
	var messageChannel string
	var messageId string
	messagePart := -1
	messageTotalParts := -1

	matcher := regexp.MustCompile(operation_pattern)

	operationPatternMatch := matcher.MatchString(message)

	if operationPatternMatch {
		stringSubMatches := matcher.FindStringSubmatch(message)
		o := stringSubMatches[1]
		//fmt.Println("Operation: " + o)
		operation = operationIndex[o]
		parsedMessage = stringSubMatches[2]
		//fmt.Println("Parsed Message: " + parsedMessage)
	} else {
		matcher = regexp.MustCompile(received_pattern)
		if matcher.MatchString(message) {

			operation = received

			/*fmt.Println("Operation Received")*/

			stringSubMatches := matcher.FindStringSubmatch(message)

			/*fmt.Printf("Submatches %v", stringSubMatches)*/

			parsedMessage = stringSubMatches[2]

			messageChannel = stringSubMatches[1]

			multPartMatcher := regexp.MustCompile(multi_part_message_pattern)

			if multPartMatcher.MatchString(parsedMessage) {
				subMatches := multPartMatcher.FindStringSubmatch(parsedMessage)
				parsedMessage = subMatches[4]
				messageId = subMatches[1]

				if len(subMatches[2]) > 0 {
					var err error
					messagePart, err = strconv.Atoi(subMatches[1])
					if err != nil {
						messagePart = -1
					}
				} else {
					messagePart = -1
				}
				if len(subMatches[3]) > 0 {
					var err error
					messageTotalParts, err = strconv.Atoi(subMatches[3])
					if err != nil {
						messageTotalParts = -1
					}
				} else {
					messageTotalParts = -1
				}
			}
		} else {
			return nil, errors.New(ortcInvalidMessageException("Invalid message format: " + message))
		}
	}

	//fmt.Println("New Ortc Msg: parsedMessage:" + parsedMessage + " messageChannel: " + messageChannel + "messageId:" + messageId)
	newMsg := newOrtcMessage(operation, parsedMessage, messageChannel, messageId, messagePart, messageTotalParts)
	return newMsg, nil
}

func (o *ortcMessage) serverError() (*ortcServerErrorException, string) {
	matcher := regexp.MustCompile(exception_pattern)

	if !matcher.MatchString(o.message) {
		return nil, "Exception match not found"
	}

	stringSubMatches := matcher.FindStringSubmatch(o.message)
	content := strings.Replace(stringSubMatches[1], "\\\"", "\"", -1)

	jsonObj, errorJson := json.Marshal(content)

	if errorJson != nil {
		return nil, "Error marshall server error"
	}

	var m *ortcServerErrorException

	err := json.Unmarshal(jsonObj, &m)

	if err != nil {
		return nil, "Error unmarshall server error"
	}

	return m, ""

}
