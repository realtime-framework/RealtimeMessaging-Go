// channelpermissions
package authentication

var permissions = [...]string{"r", "w", "p"}

type ChannelPermissions int

const (
	Read ChannelPermissions = iota
	Write
	Presence
)

func (p ChannelPermissions) String() string {
	return permissions[p]
}
