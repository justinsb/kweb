package login

import (
	"github.com/justinsb/kweb/components"
)

type Component struct {
	Provider components.AuthenticationProvider

	UserMapper UserMapper
}
