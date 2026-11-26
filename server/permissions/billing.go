package permissions

import "github.com/monetr/monetr/server/models"

func CanManageBilling(me *models.User) bool {
	if me == nil {
		return false
	}
	return me.Role == models.UserRoleOwner
}
