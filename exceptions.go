package ortc

import "fmt"

type ortcServerErrorOperation int

const (
	unexpected ortcServerErrorOperation = iota
	validate
	subscribe
	subscribe_maxSize
	unsubscribe_maxSize
	send_maxSize
)

type ortcServerErrorException struct {
	operation ortcServerErrorOperation
	channel   string
	message   string
}

func ortcAlreadyConnectedException() string {
	return "Already Connected"
}

func ortcAuthenticationNotAuthorizedException(message string) string {
	return message
}

func ortcDoesNotHavePermissionException(message string) string {
	return message
}

func ortcEmptyFieldException(field string) string {
	return fmt.Sprintf("%s is null or empty", field)
}

func ortcInvalidCharactersException(field string) string {
	return fmt.Sprintf("%s has invalid characters", field)
}

func ortcInvalidMessageException(message string) string {
	return message
}

func ortcMaxLengthException(field string, maxValue int) string {
	return fmt.Sprintf("%s size exceed the limit of %d characters", field, maxValue)
}

func ortcNotConnectedException(message string) string {
	return message
}

func ortcNotSubscribedException(channel string) string {
	return fmt.Sprintf("Not subscribed to channel %s", channel)
}

func ortcSubscribedException(message string) string {
	return message
}
