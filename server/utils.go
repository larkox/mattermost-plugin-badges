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

func canGrantBadge(user *model.User, badgeAdminID string, badge *badgesmodel.Badge, badgeType *badgesmodel.BadgeTypeDefinition) bool {
	if badgeAdminID != "" && user.Id == badgeAdminID {
		return true
	}

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

func canCreateBadge(user *model.User, badgeAdminID string, badgeType *badgesmodel.BadgeTypeDefinition) bool {
	if badgeAdminID != "" && user.Id == badgeAdminID {
		return true
	}

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

func canEditType(user *model.User, badgeAdminID string, badgeType *badgesmodel.BadgeTypeDefinition) bool {
	if badgeAdminID != "" && user.Id == badgeAdminID {
		return true
	}

	return user.IsSystemAdmin()
}

func canEditBadge(user *model.User, badgeAdminID string, badge *badgesmodel.Badge) bool {
	if badgeAdminID != "" && user.Id == badgeAdminID {
		return true
	}

	return user.IsSystemAdmin() || user.Id == badge.CreatedBy
}

func canCreateType(user *model.User, badgeAdminID string, isPlugin bool) bool {
	if isPlugin {
		return true
	}

	if badgeAdminID != "" && user.Id == badgeAdminID {
		return true
	}

	return user.IsSystemAdmin()
}

func canCreateSubscription(user *model.User, badgeAdminID string, channelID string) bool {
	if badgeAdminID != "" && user.Id == badgeAdminID {
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

func (p *Plugin) notifyGrant(badgeID badgesmodel.BadgeID, granter string, granted *model.User, inChannel bool, channelID string, reason string) {
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
		image := ""
		switch b.ImageType {
		case badgesmodel.ImageTypeEmoji:
			image = fmt.Sprintf(":%s: ", b.Image)
		case badgesmodel.ImageTypeAbsoluteURL:
			image = fmt.Sprintf("![icon](%s) ", b.Image)
		}

		dmPost := &model.Post{}
		dmText := fmt.Sprintf("@%s granted you the %s`%s` badge.", granterUser.Username, image, b.Name)
		if reason != "" {
			dmText += "\nWhy? " + reason
		}
		dmAttachment := model.SlackAttachment{
			Title: fmt.Sprintf("%sbadge granted!", image),
			Text:  dmText,
		}
		model.ParseSlackAttachment(dmPost, []*model.SlackAttachment{&dmAttachment})
		err := p.mm.Post.DM(p.BotUserID, granted.Id, dmPost)
		if err != nil {
			p.mm.Log.Debug("dm error", "err", err)
		}

		basePost := model.Post{
			UserId:    p.BotUserID,
			ChannelId: channelID,
		}
		text := fmt.Sprintf("@%s granted @%s the %s`%s` badge.", granterUser.Username, granted.Username, image, b.Name)
		if reason != "" {
			text += "\nWhy? " + reason
		}
		attachment := model.SlackAttachment{
			Title: fmt.Sprintf("%sbadge granted!", image),
			Text:  text,
		}
		model.ParseSlackAttachment(&basePost, []*model.SlackAttachment{&attachment})
		for _, sub := range subs {
			post := basePost.Clone()
			post.ChannelId = sub
			err := p.mm.Post.CreatePost(post)
			if err != nil {
				p.mm.Log.Debug("notify subscription error", "err", err)
			}
		}
		if inChannel {
			if !p.API.HasPermissionToChannel(granter, channelID, model.PERMISSION_CREATE_POST) {
				p.mm.Post.SendEphemeralPost(granter, &model.Post{Message: "You don't have permissions to notify the grant on this channel.", ChannelId: channelID})
			} else {
				post := basePost.Clone()
				post.ChannelId = channelID
				err := p.mm.Post.CreatePost(post)
				if err != nil {
					p.mm.Log.Debug("notify here error", "err", err)
				}
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
