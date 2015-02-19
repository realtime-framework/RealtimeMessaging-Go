package ortc

import "reflect"

type bufferedMessage struct {
	messagePart int
	content     string
}

func (this *bufferedMessage) compareTo(b *bufferedMessage) int {
	if b.messagePart == this.messagePart {
		return 0
	} else if b.messagePart > this.messagePart {
		return -1
	} else {
		return 1
	}
}

func (this *bufferedMessage) equals(b *bufferedMessage) bool {
	return reflect.DeepEqual(this, b)
}
