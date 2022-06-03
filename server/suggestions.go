package main

import (
	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/mattermost/mattermost-server/v5/model"
)

func (p *Plugin) filterGrantBadges(user *model.User) ([]*badgesmodel.Badge, error) {
	badges, err := p.store.GetRawBadges()
	if err != nil {
		return nil, err
	}

	types, err := p.store.GetRawTypes()
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.Badge{}
	for _, b := range badges {
		badgeType := types.GetType(b.Type)
		if badgeType == nil {
			p.mm.Log.Debug("Badge with missing type", "badge", b)
			continue
		}
		if canGrantBadge(user, p.badgeAdminUserID, b, badgeType) {
			out = append(out, b)
		}
	}

	return out, nil
}

func (p *Plugin) filterCreateBadgeTypes(user *model.User) (badgesmodel.BadgeTypeList, error) {
	types, err := p.store.GetRawTypes()
	if err != nil {
		return nil, err
	}

	out := badgesmodel.BadgeTypeList{}
	for _, t := range types {
		if canCreateBadge(user, p.badgeAdminUserID, t) {
			out = append(out, t)
		}
	}

	return out, nil
}

func (p *Plugin) filterEditTypes(user *model.User) (badgesmodel.BadgeTypeList, error) {
	types, err := p.store.GetRawTypes()
	if err != nil {
		return nil, err
	}

	out := badgesmodel.BadgeTypeList{}
	for _, t := range types {
		if canEditType(user, p.badgeAdminUserID, t) {
			out = append(out, t)
		}
	}

	return out, nil
}

func (p *Plugin) filterEditBadges(user *model.User) ([]*badgesmodel.Badge, error) {
	bb, err := p.store.GetRawBadges()
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.Badge{}
	for _, b := range bb {
		if canEditBadge(user, p.badgeAdminUserID, b) {
			out = append(out, b)
		}
	}

	return out, nil
}
