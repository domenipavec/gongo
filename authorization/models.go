package authorization

import (
	"time"

	"github.com/jinzhu/gorm"
)

type UserID struct {
	ID     string
	User   User
	UserID uint
}

type User struct {
	gorm.Model
	Name        string `valid:"required"`
	Email       string
	AvatarURL   string
	LastLogin   time.Time
	Active      bool
	Permissions []Permission `gorm:"many2many:user_permission"`
	Groups      []Group      `gorm:"many2many:user_group"`
}

type Group struct {
	gorm.Model
	Name        string       `valid:"required"`
	Permissions []Permission `gorm:"many2many:group_permission"`
}

type Permission struct {
	gorm.Model
	Name string
	Code string
}

func (u User) DisplayName() string {
	return u.Name
}

func (u User) HasPermissions(permissions ...string) bool {
	// TODO: optimize
	for _, requiredPerm := range permissions {
		hasThis := false

		for _, directPerm := range u.Permissions {
			if directPerm.Code == requiredPerm {
				hasThis = true
				break
			}
		}
		if hasThis {
			continue
		}

		for _, group := range u.Groups {
			for _, groupPerm := range group.Permissions {
				if groupPerm.Code == requiredPerm {
					hasThis = true
					break
				}
			}
			if hasThis {
				break
			}
		}
		if hasThis {
			continue
		}

		return false
	}

	return true
}
