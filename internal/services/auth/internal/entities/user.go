package entities

import (
	"errors"
	"time"
)

type Role string

const (
	RoleUser      Role = "user"
	RoleModerator Role = "moderator"
	RoleAdmin     Role = "admin"
)

type User struct {
	ID          string
	Email       string
	Password    string
	Name        string
	Nickname    string
	PhoneNumber string
	Age         int
	Bio         string
	Roles       []Role
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (u *User) Validate() error {
	if u.Email == "" {
		return errors.New("email is required")
	}
	if u.Name == "" {
		return errors.New("name is required")
	}
	if u.Age < 16 {
		return errors.New("user must be at least 16 years old")
	}

	return nil
}

type AuthenticatedUser struct {
	UserID    string
	Name      string
	Email     string
	Nickname  string
	Age       int
	Bio       string
	Roles     []Role
	ExpiresAt time.Time
	TokenPair TokenPair
}

func (u *User) HasRole(role Role) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func (u *User) AddRole(role Role) {
	if !u.HasRole(role) {
		u.Roles = append(u.Roles, role)
	}
}

func (u *User) RemoveRole(role Role) {
	for i, r := range u.Roles {
		if r == role {
			u.Roles = append(u.Roles[:i], u.Roles[i+1:]...)
			return
		}
	}
}

func (au *AuthenticatedUser) HasRole(role Role) bool {
	for _, r := range au.Roles {
		if r == role {
			return true
		}
	}
	return false
}
