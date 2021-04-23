package main

import (
	"encoding/json"
	"fmt"

	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
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

func canGrantBadge(user model.User, badge badgesmodel.Badge, badgeType badgesmodel.BadgeTypeDefinition) bool {
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

func canCreateBadge(user model.User, badgeType badgesmodel.BadgeTypeDefinition) bool {
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

func canCreateType(user model.User, isPlugin bool) bool {
	if isPlugin {
		return true
	}

	return user.IsSystemAdmin()
}

func dumpObject(o interface{}) {
	b, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(b))
}

func (p *Plugin) notifyGrant(badgeID badgesmodel.BadgeID, granter, granted string) {
	b, errBadge := p.store.GetBadgeDetails(badgeID)
	u, errUser := p.mm.User.Get(granter)
	if errBadge != nil {
		p.mm.Log.Debug("badge error", "err", errBadge)
	}
	if errUser != nil {
		p.mm.Log.Debug("user error", "err", errUser)
	}
	if errBadge == nil && errUser == nil {
		err := p.mm.Post.DM(p.BotUserID, granted, &model.Post{
			Message: fmt.Sprintf("@%s granted you the `%s` badge.", u.Username, b.Name),
		})
		if err != nil {
			p.mm.Log.Debug("dm error", "err", err)
		}
	}
}
