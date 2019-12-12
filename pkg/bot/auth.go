package bot

// Role is a role user in system
// user can be Gueest (without permissions) and User with all permissions
type Role int

const (
	// Guest User without any permissions
	Guest Role = iota
	// User user allowed to have any access to bot functions
	User
)

// Authorisation interface allow's to get users role in system
type Authorisation interface {
	// GetRoleByLogin get user's role by telegramm login
	GetRoleByLogin(telegramLogin string) (Role, error)
}

// DummyAuthorisation allows to all users to get access to bot
type DummyAuthorisation struct{}

// GetRoleByLogin for Dummy Authorisation return user role for all logins
func (DummyAuthorisation) GetRoleByLogin(_ string) (Role, error) {
	return User, nil
}
