package main

import (
	"encoding/json"
	"fmt"

	"github.com/mattermost/mattermost-server/v5/model"
)

func areRolesAllowed(userRoles []string, allowedRoles map[string]bool) bool {
	for ar, b := range allowedRoles {
		if !b {
			continue
		}
		for _, ur := range userRoles {
			if ar == ur {
				return true
			}
		}
	}

	return false
}

func canGrantBadge(user model.User, badge Badge, badgeType BadgeTypeDefinition) bool {
	if user.IsSystemAdmin() {
		return true
	}

	if badgeType.CreatedBy == user.Id {
		return true
	}

	if badge.CreatedBy == user.Id {
		return true
	}

	blocked := badgeType.CanGrant.BlockList[user.Id]
	if blocked {
		return false
	}

	if areRolesAllowed(user.GetRoles(), badgeType.CanGrant.Roles) {
		return true
	}

	allowed := badgeType.CanGrant.AllowList[user.Id]
	if allowed {
		return true
	}

	return badgeType.CanGrant.Everyone
}

func canCreateBadge(user model.User, badgeType BadgeTypeDefinition) bool {
	if user.IsSystemAdmin() {
		return true
	}

	if badgeType.CreatedBy == user.Id {
		return true
	}

	blocked := badgeType.CanCreate.BlockList[user.Id]
	if blocked {
		return false
	}

	if areRolesAllowed(user.GetRoles(), badgeType.CanCreate.Roles) {
		return true
	}

	allowed := badgeType.CanCreate.AllowList[user.Id]
	if allowed {
		return true
	}

	return badgeType.CanCreate.Everyone
}

func canCreateType(user model.User) bool {
	return user.IsSystemAdmin()
}

func dumpObject(o interface{}) {
	b, _ := json.MarshalIndent(o, "", "    ")
	fmt.Println(string(b))
}
