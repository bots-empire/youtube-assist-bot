package regexp

import "regexp"

var (
	// InvitationLink is regular expression for invitation telegram link
	InvitationLink = regexp.MustCompile("^https://t.me/joinchat/[A-z0-9-]{1,30}$")
	// Email is regular expression for paypal email addres
	Email = regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
)
