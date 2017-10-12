package html

import (
	"github.com/Code-Hex/vegeta/model"
)

type (
	Args interface {
		IsAuthed() bool
		IsAdmin() bool
		Year() int
	}

	AdminArgs interface {
		Args
		Token() string
		Users() model.Users
		IsCreated() bool
		Reason() string
	}

	MyPageArgs interface {
		Args
		User() *model.User
		Token() string
	}

	SettingsArgs interface {
		Args
		User() *model.User
		Token() string
	}
)
