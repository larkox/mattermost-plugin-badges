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

func canEditType(user model.User, badgeType badgesmodel.BadgeTypeDefinition) bool {
	return user.IsSystemAdmin()
}

func canEditBadge(user model.User, badge badgesmodel.Badge) bool {
	return user.IsSystemAdmin() || user.Id == badge.CreatedBy
}

func canCreateType(user model.User, isPlugin bool) bool {
	if isPlugin {
		return true
	}

	return user.IsSystemAdmin()
}

func canCreateSubscription(user model.User, channelID string) bool {
	return user.IsSystemAdmin()
}

func dumpObject(o interface{}) {
	b, err := json.MarshalIndent(o, "", "    ")
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(b))
}

func (p *Plugin) notifyGrant(badgeID badgesmodel.BadgeID, granter string, granted *model.User, inChannel bool, channelID string) {
	b, errBadge := p.store.GetBadgeDetails(badgeID)
	granterUser, errUser := p.mm.User.Get(granter)
	if errBadge != nil {
		p.mm.Log.Debug("badge error", "err", errBadge)
	}
	if errUser != nil {
		p.mm.Log.Debug("user error", "err", errUser)
	}

	subs, _ := p.store.GetTypeSubscriptions(b.Type)

	if errBadge == nil && errUser == nil {
		err := p.mm.Post.DM(p.BotUserID, granted.Id, &model.Post{
			Message: fmt.Sprintf("@%s granted you the `%s` badge.", granterUser.Username, b.Name),
		})
		if err != nil {
			p.mm.Log.Debug("dm error", "err", err)
		}
		basePost := model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
			Message:   fmt.Sprintf("@%s granted @%s the `%s` badge.", granterUser.Username, granted.Username, b.Name),
		}
		for _, sub := range subs {
			post := basePost.Clone()
			post.ChannelId = sub
			err := p.mm.Post.CreatePost(post)
			if err != nil {
				p.mm.Log.Debug("notify subscription error", "err", err)
			}
		}
		if inChannel {
			post := basePost.Clone()
			post.ChannelId = channelID
			err := p.mm.Post.CreatePost(post)
			if err != nil {
				p.mm.Log.Debug("notify here error", "err", err)
			}
		}
	}
}

func getBooleanString(in bool) string {
	if in {
		return TrueString
	}
	return FalseString
}
