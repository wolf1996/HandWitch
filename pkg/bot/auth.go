package bot

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

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

// UsersList list of users allowed to talk with bot
type UsersList struct {
	users map[string]struct{}
}

// GetAuthSourceFromJSON parse users names from json
func GetAuthSourceFromJSON(reader io.Reader) (*UsersList, error) {
	fileStructure := struct {
		UsersList []string `json:"users"`
	}{}
	bytes, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bytes, &fileStructure)
	if err != nil {
		return nil, err
	}
	usersSet := map[string]struct{}{}
	for _, userLogin := range fileStructure.UsersList {
		usersSet[userLogin] = struct{}{}
	}
	result := UsersList{
		users: usersSet,
	}
	return &result, nil
}

// GetRoleByLogin returns role User if userLogin contains in lists of logins
func (list *UsersList) GetRoleByLogin(userLogin string) (Role, error) {
	if _, contains := list.users[userLogin]; contains {
		return User, nil
	}
	return Guest, nil
}
