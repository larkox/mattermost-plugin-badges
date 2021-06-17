package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

var errInvalidBadge = errors.New("invalid badge")
var errBadgeNotFound = errors.New("badge not found")

type Store interface {
	// Interface
	GetUserBadges(userID string) ([]*badgesmodel.UserBadge, error)
	GetAllBadges() ([]*badgesmodel.AllBadgesBadge, error)
	GetBadgeDetails(badgeID badgesmodel.BadgeID) (*badgesmodel.BadgeDetails, error)

	// Autocomplete
	GetRawBadges() ([]*badgesmodel.Badge, error)
	GetRawTypes() (badgesmodel.BadgeTypeList, error)

	// API
	AddBadge(badge *badgesmodel.Badge) (*badgesmodel.Badge, error)
	GrantBadge(badgeID badgesmodel.BadgeID, userID string, grantedBy string, reason string) (bool, error)
	AddType(t *badgesmodel.BadgeTypeDefinition) (*badgesmodel.BadgeTypeDefinition, error)
	GetType(tID badgesmodel.BadgeType) (*badgesmodel.BadgeTypeDefinition, error)
	GetBadge(badgeID badgesmodel.BadgeID) (*badgesmodel.Badge, error)
	UpdateType(t *badgesmodel.BadgeTypeDefinition) error
	UpdateBadge(b *badgesmodel.Badge) error
	DeleteType(tID badgesmodel.BadgeType) error
	DeleteBadge(bID badgesmodel.BadgeID) error

	AddSubscription(tID badgesmodel.BadgeType, cID string) error
	RemoveSubscriptions(tID badgesmodel.BadgeType, cID string) error
	GetTypeSubscriptions(tID badgesmodel.BadgeType) ([]string, error)
	GetChannelSubscriptions(cID string) ([]*badgesmodel.BadgeTypeDefinition, error)

	// PAPI
	EnsureBadges(badges []*badgesmodel.Badge, pluginID, botID string) ([]*badgesmodel.Badge, error)
}

type store struct {
	api plugin.API
}

func NewStore(api plugin.API) Store {
	return &store{
		api: api,
	}
}

func (s *store) EnsureBadges(badges []*badgesmodel.Badge, pluginID, botID string) ([]*badgesmodel.Badge, error) {
	l, _, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	var tDef *badgesmodel.BadgeTypeDefinition
	for i, t := range l {
		if t.CreatedBy == botID {
			tDef = l[i]
			break
		}
	}

	if tDef == nil {
		tDef, err = s.addType(&badgesmodel.BadgeTypeDefinition{
			Name:      "Plugin badges: " + pluginID,
			CreatedBy: botID,
		}, true)
		if err != nil {
			return nil, err
		}
	}

	bb, _, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.Badge{}
	for _, pb := range badges {
		found := false
		for _, b := range bb {
			if b.CreatedBy == botID && b.Name == pb.Name {
				found = true
				out = append(out, b)
				break
			}
		}
		if !found {
			pb.Type = tDef.ID
			pb.CreatedBy = botID
			newBadge, err := s.AddBadge(pb)
			if err != nil {
				return nil, err
			}
			out = append(out, newBadge)
		}
	}

	return out, nil
}

func (s *store) AddBadge(b *badgesmodel.Badge) (*badgesmodel.Badge, error) {
	if !b.IsValid() {
		return nil, errInvalidBadge
	}

	badgeTypes, _, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	t := badgeTypes.GetType(b.Type)
	if t == nil {
		return nil, errors.New("missing badge type")
	}

	b.ID = badgesmodel.BadgeID(model.NewId())
	err = s.doAtomic(func() (bool, error) { return s.atomicAddBadge(b) })
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (s *store) AddType(t *badgesmodel.BadgeTypeDefinition) (*badgesmodel.BadgeTypeDefinition, error) {
	return s.addType(t, false)
}

func (s *store) addType(t *badgesmodel.BadgeTypeDefinition, isPlugin bool) (*badgesmodel.BadgeTypeDefinition, error) {
	t.ID = badgesmodel.BadgeType(model.NewId())
	err := s.doAtomic(func() (bool, error) { return s.atomicAddType(t) })
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (s *store) GetAllBadges() ([]*badgesmodel.AllBadgesBadge, error) {
	badges, _, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	ownership, _, err := s.getOwnershipList()
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.AllBadgesBadge{}
	for _, b := range badges {
		badge := &badgesmodel.AllBadgesBadge{
			Badge: *b,
		}
		grantedTo := map[string]bool{}
		for _, o := range ownership {
			if o.Badge != badge.ID {
				continue
			}
			badge.GrantedTimes++

			if !grantedTo[o.User] {
				badge.Granted++
				grantedTo[o.User] = true
			}
		}

		badge.TypeName = "unknown"
		t, err := s.GetType(badge.Type)
		if err == nil {
			badge.TypeName = t.Name
		}
		out = append(out, badge)
	}

	return out, nil
}

func (s *store) GetRawBadges() ([]*badgesmodel.Badge, error) {
	bb, _, err := s.getAllBadges()
	return bb, err
}

func (s *store) GetRawTypes() (badgesmodel.BadgeTypeList, error) {
	tt, _, err := s.getAllTypes()
	return tt, err
}

func (s *store) getAllTypes() (badgesmodel.BadgeTypeList, []byte, error) {
	data, appErr := s.api.KVGet(KVKeyTypes)
	if appErr != nil {
		return nil, nil, appErr
	}

	typeList := []*badgesmodel.BadgeTypeDefinition{}
	if data != nil {
		err := json.Unmarshal(data, &typeList)
		if err != nil {
			return nil, nil, err
		}
	}

	return typeList, data, nil
}

func (s *store) getAllBadges() ([]*badgesmodel.Badge, []byte, error) {
	data, appErr := s.api.KVGet(KVKeyBadges)
	if appErr != nil {
		return nil, nil, appErr
	}

	badgeList := []*badgesmodel.Badge{}
	if data != nil {
		err := json.Unmarshal(data, &badgeList)
		if err != nil {
			return nil, nil, err
		}
	}

	return badgeList, data, nil
}

func (s *store) getBadge(id badgesmodel.BadgeID) (*badgesmodel.Badge, error) {
	badgeList, _, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	return s.getBadgeFromList(id, badgeList)
}

func (s *store) GetBadgeDetails(id badgesmodel.BadgeID) (*badgesmodel.BadgeDetails, error) {
	badge, err := s.getBadge(id)
	if err != nil {
		return nil, err
	}

	owners, err := s.getBadgeUsers(id)
	if err != nil {
		return nil, err
	}

	createdByName := "unknown"
	u, appErr := s.api.GetUser(badge.CreatedBy)
	if appErr == nil {
		conf := s.api.GetConfig()
		if conf != nil {
			format := conf.TeamSettings.TeammateNameDisplay
			if format != nil {
				createdByName = u.GetDisplayName(*format)
			}
		}
	}

	typeName := "unknown"
	t, err := s.GetType(badge.Type)
	if err == nil {
		typeName = t.Name
	}
	return &badgesmodel.BadgeDetails{
		Badge:             *badge,
		Owners:            owners,
		CreatedByUsername: createdByName,
		TypeName:          typeName,
	}, nil
}

func (s *store) getOwnershipList() (badgesmodel.OwnershipList, []byte, error) {
	data, appErr := s.api.KVGet(KVKeyOwnership)
	if appErr != nil {
		return nil, nil, appErr
	}

	ownership := badgesmodel.OwnershipList{}
	if data != nil {
		err := json.Unmarshal(data, &ownership)
		if err != nil {
			return nil, nil, err
		}
	}

	return ownership, data, nil
}

func (s *store) GrantBadge(id badgesmodel.BadgeID, userID string, grantedBy string, reason string) (bool, error) {
	badge, err := s.getBadge(id)
	if err != nil {
		return false, err
	}

	types, _, err := s.getAllTypes()
	if err != nil {
		return false, err
	}

	badgeType := types.GetType(badge.Type)
	if badgeType == nil {
		return false, errors.New("badge type not found")
	}

	ownership := badgesmodel.Ownership{
		User:      userID,
		Badge:     badge.ID,
		Time:      time.Now(),
		Reason:    reason,
		GrantedBy: grantedBy,
	}

	shouldNotify := false
	err = s.doAtomic(func() (bool, error) {
		var done bool
		var err error
		shouldNotify, done, err = s.atomicAddBadgeToOwnership(ownership, badge.Multiple)
		return done, err
	})
	if err != nil {
		return false, err
	}

	return shouldNotify, nil
}

func (s *store) GetUserBadges(userID string) ([]*badgesmodel.UserBadge, error) {
	ownership, _, err := s.getOwnershipList()
	if err != nil {
		return nil, err
	}

	badges, _, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.UserBadge{}
	for _, o := range ownership {
		if o.User == userID {
			badge, err := s.getBadgeFromList(o.Badge, badges)
			if err != nil {
				s.api.LogDebug("Badge not found while getting user badges", "badgeID", o.Badge, "userID", userID)
				continue
			}

			grantedByName := "unknown"
			u, appErr := s.api.GetUser(o.GrantedBy)
			if appErr == nil {
				conf := s.api.GetConfig()
				if conf != nil {
					format := conf.TeamSettings.TeammateNameDisplay
					if format != nil {
						grantedByName = u.GetDisplayName(*format)
					}
				}
			}

			typeName := "unknown"
			t, err := s.GetType(badge.Type)
			if err == nil {
				typeName = t.Name
			}

			out = append([]*badgesmodel.UserBadge{{Badge: *badge, Ownership: o, GrantedByUsername: grantedByName, TypeName: typeName}}, out...)
		}
	}

	return out, nil
}

func (s *store) GetType(tID badgesmodel.BadgeType) (*badgesmodel.BadgeTypeDefinition, error) {
	tt, _, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	for _, t := range tt {
		if t.ID == tID {
			return t, nil
		}
	}

	return nil, errors.New("not found")
}

func (s *store) GetBadge(badgeID badgesmodel.BadgeID) (*badgesmodel.Badge, error) {
	return s.getBadge(badgeID)
}

func (s *store) UpdateType(t *badgesmodel.BadgeTypeDefinition) error {
	return s.doAtomic(func() (bool, error) { return s.atomicUpdateType(t) })
}

func (s *store) UpdateBadge(b *badgesmodel.Badge) error {
	return s.doAtomic(func() (bool, error) { return s.atomicUpdateBadge(b) })
}

func (s *store) atomicDeleteType(tID badgesmodel.BadgeType) (bool, error) {
	tt, data, err := s.getAllTypes()
	if err != nil {
		return false, err
	}

	for i, t := range tt {
		if t.ID == tID {
			tt = append(tt[:i], tt[i+1:]...)
			break
		}
	}

	return s.compareAndSet(KVKeyTypes, data, tt)
}

func (s *store) DeleteType(tID badgesmodel.BadgeType) error {
	s.doAtomic(func() (bool, error) { return s.atomicDeleteType(tID) })

	bb, _, err := s.getAllBadges()
	if err != nil {
		return err
	}

	for _, b := range bb {
		if b.Type == tID {
			s.api.LogDebug("Deleting badge", "name", b.Name)
			err := s.DeleteBadge(b.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *store) DeleteBadge(bID badgesmodel.BadgeID) error {
	err := s.doAtomic(func() (bool, error) { return s.atomicRemoveBadge(bID) })
	if err != nil {
		return err
	}

	err = s.doAtomic(func() (bool, error) { return s.atomicRemoveBadgeFromOwnership(bID) })
	if err != nil {
		return err
	}

	return nil
}

func (s *store) getAllSubscriptions() ([]badgesmodel.Subscription, []byte, error) {
	data, appErr := s.api.KVGet(KVKeySubscriptions)
	if appErr != nil {
		return nil, nil, appErr
	}

	subs := []badgesmodel.Subscription{}
	if data != nil {
		err := json.Unmarshal(data, &subs)
		if err != nil {
			return nil, nil, err
		}
	}

	return subs, data, nil
}

func (s *store) AddSubscription(tID badgesmodel.BadgeType, cID string) error {
	toAdd := badgesmodel.Subscription{ChannelID: cID, TypeID: tID}
	return s.doAtomic(func() (bool, error) { return s.atomicAddSubscription(toAdd) })
}

func (s *store) RemoveSubscriptions(tID badgesmodel.BadgeType, cID string) error {
	toRemove := badgesmodel.Subscription{ChannelID: cID, TypeID: tID}
	return s.doAtomic(func() (bool, error) { return s.atomicRemoveSubscription(toRemove) })
}

func (s *store) GetTypeSubscriptions(tID badgesmodel.BadgeType) ([]string, error) {
	subs, _, err := s.getAllSubscriptions()
	if err != nil {
		return nil, err
	}

	out := []string{}
	for _, sub := range subs {
		if sub.TypeID == tID {
			out = append(out, sub.ChannelID)
		}
	}

	return out, nil
}

func (s *store) GetChannelSubscriptions(cID string) ([]*badgesmodel.BadgeTypeDefinition, error) {
	subs, _, err := s.getAllSubscriptions()
	if err != nil {
		return nil, err
	}

	out := []*badgesmodel.BadgeTypeDefinition{}
	for _, sub := range subs {
		if sub.ChannelID == cID {
			t, err := s.GetType(sub.TypeID)
			if err != nil {
				s.api.LogDebug("cannot get type", "err", err)
				continue
			}
			out = append(out, t)
		}
	}

	return out, nil
}

func (s *store) getBadgeFromList(badgeID badgesmodel.BadgeID, list []*badgesmodel.Badge) (*badgesmodel.Badge, error) {
	for _, badge := range list {
		if badgeID == badge.ID {
			return badge, nil
		}
	}
	return nil, errBadgeNotFound
}

func (s *store) getBadgeUsers(badgeID badgesmodel.BadgeID) (badgesmodel.OwnershipList, error) {
	_, err := s.getBadge(badgeID)
	if err != nil {
		return nil, errBadgeNotFound
	}

	ownership, _, err := s.getOwnershipList()
	if err != nil {
		return nil, err
	}

	out := badgesmodel.OwnershipList{}
	for _, o := range ownership {
		if o.Badge == badgeID {
			out = append(out, o)
		}
	}

	return out, nil
}
