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
	GetUserBadges(userID string) ([]badgesmodel.UserBadge, error)
	GetAllBadges() ([]badgesmodel.AllBadgesBadge, error)
	GetBadgeDetails(badgeID badgesmodel.BadgeID) (*badgesmodel.BadgeDetails, error)

	// Autocomplete
	GetGrantSuggestions(user model.User) ([]badgesmodel.Badge, error)
	GetEditTypeSuggestions(user model.User) (badgesmodel.BadgeTypeList, error)
	GetTypeSuggestions(user model.User) (badgesmodel.BadgeTypeList, error)
	GetEditBadgeSuggestions(user model.User) ([]badgesmodel.Badge, error)

	// API
	AddBadge(badge badgesmodel.Badge) (*badgesmodel.Badge, error)
	GrantBadge(badgeID badgesmodel.BadgeID, userID string, grantedBy string) (bool, error)
	AddType(t badgesmodel.BadgeTypeDefinition) (*badgesmodel.BadgeTypeDefinition, error)
	GetType(tID badgesmodel.BadgeType) (*badgesmodel.BadgeTypeDefinition, error)
	GetBadge(badgeID badgesmodel.BadgeID) (*badgesmodel.Badge, error)
	UpdateType(t badgesmodel.BadgeTypeDefinition) error
	UpdateBadge(b badgesmodel.Badge) error
	DeleteType(tID badgesmodel.BadgeType) error
	DeleteBadge(bID badgesmodel.BadgeID) error

	AddSubscription(tID badgesmodel.BadgeType, cID string) error
	RemoveSubscriptions(tID badgesmodel.BadgeType, cID string) error
	GetTypeSubscriptions(tID badgesmodel.BadgeType) ([]string, error)
	GetChannelSubscriptions(cID string) ([]badgesmodel.BadgeTypeDefinition, error)

	// PAPI
	EnsureBadges(badges []badgesmodel.Badge, pluginID, botID string) ([]badgesmodel.Badge, error)

	// DEBUG
	DebugGetTypes() badgesmodel.BadgeTypeList
}

type store struct {
	api plugin.API
}

func NewStore(api plugin.API) Store {
	return &store{
		api: api,
	}
}

func (s *store) EnsureBadges(badges []badgesmodel.Badge, pluginID, botID string) ([]badgesmodel.Badge, error) {
	l, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	var tDef *badgesmodel.BadgeTypeDefinition
	for _, t := range l {
		if t.CreatedBy == botID {
			tDef = &t
			break
		}
	}

	if tDef == nil {
		tDef, err = s.addType(badgesmodel.BadgeTypeDefinition{
			Name:      "Plugin badges: " + pluginID,
			CreatedBy: botID,
		}, true)
		if err != nil {
			return nil, err
		}
	}

	bb, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	out := []badgesmodel.Badge{}
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
			out = append(out, *newBadge)
		}
	}

	return out, nil
}
func (s *store) DebugGetTypes() badgesmodel.BadgeTypeList {
	l, _ := s.getAllTypes()
	return l
}

func (s *store) AddBadge(b badgesmodel.Badge) (*badgesmodel.Badge, error) {
	if !b.IsValid() {
		return nil, errInvalidBadge
	}

	u, appErr := s.api.GetUser(b.CreatedBy)
	if appErr != nil {
		return nil, appErr
	}

	badgeTypes, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	t := badgeTypes.GetType(b.Type)
	if t == nil {
		return nil, errors.New("missing badge type")
	}

	if !canCreateBadge(*u, *t) {
		return nil, errors.New("you have no permission to create this type of badge")
	}

	badgeList, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	var lastID badgesmodel.BadgeID = -1
	if len(badgeList) > 0 {
		lastID = badgeList[len(badgeList)-1].ID
	}
	b.ID = lastID + 1
	badgeList = append(badgeList, b)

	data, err := json.Marshal(badgeList)
	if err != nil {
		return nil, err
	}

	appErr = s.api.KVSet(KVKeyBadges, data)
	if appErr != nil {
		return nil, appErr
	}

	return &b, nil
}

func (s *store) AddType(t badgesmodel.BadgeTypeDefinition) (*badgesmodel.BadgeTypeDefinition, error) {
	return s.addType(t, false)
}

func (s *store) addType(t badgesmodel.BadgeTypeDefinition, isPlugin bool) (*badgesmodel.BadgeTypeDefinition, error) {
	u, appErr := s.api.GetUser(t.CreatedBy)
	if appErr != nil {
		return nil, appErr
	}

	if !canCreateType(*u, isPlugin) {
		return nil, errors.New("you have no permission to create this badge type")
	}

	badgeTypes, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	var lastID badgesmodel.BadgeType = -1
	if len(badgeTypes) > 0 {
		lastID = badgeTypes[len(badgeTypes)-1].ID
	}
	t.ID = lastID + 1
	badgeTypes = append(badgeTypes, t)

	data, err := json.Marshal(badgeTypes)
	if err != nil {
		return nil, err
	}

	appErr = s.api.KVSet(KVKeyTypes, data)
	if appErr != nil {
		return nil, appErr
	}

	return &t, nil
}

func (s *store) GetAllBadges() ([]badgesmodel.AllBadgesBadge, error) {
	badges, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	ownership, err := s.getOwnershipList()
	if err != nil {
		return nil, err
	}

	out := []badgesmodel.AllBadgesBadge{}
	for _, b := range badges {
		badge := badgesmodel.AllBadgesBadge{
			Badge: b,
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

func (s *store) GetGrantSuggestions(user model.User) ([]badgesmodel.Badge, error) {
	badges, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	types, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	out := []badgesmodel.Badge{}
	for _, b := range badges {
		badgeType := types.GetType(b.Type)
		if badgeType == nil {
			s.api.LogDebug("Badge with missing type", "badge", b)
			break
		}
		if canGrantBadge(user, b, *badgeType) {
			out = append(out, b)
		}
	}

	return out, nil
}

func (s *store) GetTypeSuggestions(user model.User) (badgesmodel.BadgeTypeList, error) {
	types, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	out := badgesmodel.BadgeTypeList{}
	for _, t := range types {
		if canCreateBadge(user, t) {
			out = append(out, t)
		}
	}

	return out, nil
}

func (s *store) GetEditTypeSuggestions(user model.User) (badgesmodel.BadgeTypeList, error) {
	types, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	out := badgesmodel.BadgeTypeList{}
	for _, t := range types {
		if canEditType(user, t) {
			out = append(out, t)
		}
	}

	return out, nil
}

func (s *store) GetEditBadgeSuggestions(user model.User) ([]badgesmodel.Badge, error) {
	bb, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	out := []badgesmodel.Badge{}
	for _, b := range bb {
		if canEditBadge(user, b) {
			out = append(out, b)
		}
	}

	return out, nil
}

func (s *store) getAllTypes() (badgesmodel.BadgeTypeList, error) {
	data, appErr := s.api.KVGet(KVKeyTypes)
	if appErr != nil {
		return nil, appErr
	}

	typeList := []badgesmodel.BadgeTypeDefinition{}
	if data != nil {
		err := json.Unmarshal(data, &typeList)
		if err != nil {
			return nil, err
		}
	}

	return typeList, nil
}

func (s *store) getAllBadges() ([]badgesmodel.Badge, error) {
	data, appErr := s.api.KVGet(KVKeyBadges)
	if appErr != nil {
		return nil, appErr
	}

	badgeList := []badgesmodel.Badge{}
	if data != nil {
		err := json.Unmarshal(data, &badgeList)
		if err != nil {
			return nil, err
		}
	}

	return badgeList, nil
}

func (s *store) getBadge(id badgesmodel.BadgeID) (*badgesmodel.Badge, error) {
	badgeList, err := s.getAllBadges()
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

func (s *store) getOwnershipList() (badgesmodel.OwnershipList, error) {
	data, appErr := s.api.KVGet(KVKeyOwnership)
	if appErr != nil {
		return nil, appErr
	}

	ownership := badgesmodel.OwnershipList{}
	if data != nil {
		err := json.Unmarshal(data, &ownership)
		if err != nil {
			return nil, err
		}
	}

	return ownership, nil
}

func (s *store) GrantBadge(id badgesmodel.BadgeID, userID string, grantedBy string) (bool, error) {
	badge, err := s.getBadge(id)
	if err != nil {
		return false, err
	}

	types, err := s.getAllTypes()
	if err != nil {
		return false, err
	}

	badgeType := types.GetType(badge.Type)
	if badgeType == nil {
		return false, errors.New("badge type not found")
	}

	user, appErr := s.api.GetUser(userID)
	if appErr != nil {
		return false, err
	}

	grantedByUser, appErr := s.api.GetUser(grantedBy)
	if appErr != nil {
		return false, err
	}

	if !canGrantBadge(*grantedByUser, *badge, *badgeType) {
		return false, errors.New("you don't have permission to grant this badge")
	}

	ownership, err := s.getOwnershipList()
	if err != nil {
		return false, err
	}

	if !badge.Multiple && ownership.IsOwned(user.Id, id) {
		return false, nil
	}

	ownership = append(ownership, badgesmodel.Ownership{
		User:      user.Id,
		Badge:     badge.ID,
		Time:      time.Now(),
		GrantedBy: grantedByUser.Id,
	})

	data, err := json.Marshal(ownership)
	if err != nil {
		return false, err
	}

	appErr = s.api.KVSet(KVKeyOwnership, data)
	if appErr != nil {
		return false, appErr
	}

	return true, nil
}

func (s *store) GetUserBadges(userID string) ([]badgesmodel.UserBadge, error) {
	ownership, err := s.getOwnershipList()
	if err != nil {
		return nil, err
	}

	badges, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	out := []badgesmodel.UserBadge{}
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

			out = append([]badgesmodel.UserBadge{{Badge: *badge, Ownership: o, GrantedByUsername: grantedByName, TypeName: typeName}}, out...)
		}
	}

	return out, nil
}

func (s *store) GetType(tID badgesmodel.BadgeType) (*badgesmodel.BadgeTypeDefinition, error) {
	tt, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	for _, t := range tt {
		if t.ID == tID {
			return &t, nil
		}
	}

	return nil, errors.New("not found")
}

func (s *store) GetBadge(badgeID badgesmodel.BadgeID) (*badgesmodel.Badge, error) {
	return s.getBadge(badgeID)
}

func (s *store) UpdateType(b badgesmodel.BadgeTypeDefinition) error {
	bb, err := s.getAllTypes()
	if err != nil {
		return err
	}

	found := false
	for i, bOld := range bb {
		if bOld.ID == b.ID {
			bb[i] = b
			found = true
			break
		}
	}

	if !found {
		return errors.New("not found")
	}

	data, err := json.Marshal(bb)
	if err != nil {
		return err
	}

	appErr := s.api.KVSet(KVKeyTypes, data)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (s *store) UpdateBadge(t badgesmodel.Badge) error {
	tt, err := s.getAllBadges()
	if err != nil {
		return err
	}

	found := false
	for i, tOld := range tt {
		if tOld.ID == t.ID {
			tt[i] = t
			found = true
			break
		}
	}

	if !found {
		return errors.New("not found")
	}

	data, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	appErr := s.api.KVSet(KVKeyBadges, data)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (s *store) DeleteType(tID badgesmodel.BadgeType) error {
	tt, err := s.getAllTypes()
	if err != nil {
		return err
	}

	for i, t := range tt {
		if t.ID == tID {
			tt = append(tt[:i], tt[i+1:]...)
			break
		}
	}

	data, err := json.Marshal(tt)
	if err != nil {
		return err
	}

	appErr := s.api.KVSet(KVKeyTypes, data)
	if appErr != nil {
		return appErr
	}

	bb, err := s.getAllBadges()
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
	bb, err := s.getAllBadges()
	if err != nil {
		return err
	}

	for i, b := range bb {
		if b.ID == bID {
			bb = append(bb[:i], bb[i+1:]...)
			break
		}
	}

	data, err := json.Marshal(bb)
	if err != nil {
		return err
	}

	appErr := s.api.KVSet(KVKeyBadges, data)
	if appErr != nil {
		return appErr
	}

	ownership, err := s.getOwnershipList()
	if err != nil {
		return err
	}

	toDelete := []int{}
	for i, o := range ownership {
		if o.Badge == bID {
			toDelete = append([]int{i}, toDelete...)
		}
	}

	for _, index := range toDelete {
		ownership = append(ownership[:index], ownership[index+1:]...)
	}

	data, err = json.Marshal(ownership)
	if err != nil {
		return err
	}

	appErr = s.api.KVSet(KVKeyOwnership, data)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (s *store) getAllSubscriptions() ([]badgesmodel.Subscription, error) {
	data, appErr := s.api.KVGet(KVKeySubscriptions)
	if appErr != nil {
		return nil, appErr
	}

	subs := []badgesmodel.Subscription{}
	if data != nil {
		err := json.Unmarshal(data, &subs)
		if err != nil {
			return nil, err
		}
	}

	return subs, nil
}

func (s *store) storeSubscriptionList(subs []badgesmodel.Subscription) error {
	data, err := json.Marshal(subs)
	if err != nil {
		return err
	}

	appErr := s.api.KVSet(KVKeySubscriptions, data)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (s *store) AddSubscription(tID badgesmodel.BadgeType, cID string) error {
	subs, err := s.getAllSubscriptions()
	if err != nil {
		return err
	}

	for _, sub := range subs {
		if sub.ChannelID == cID && sub.TypeID == tID {
			return nil
		}
	}

	subs = append(subs, badgesmodel.Subscription{ChannelID: cID, TypeID: tID})

	err = s.storeSubscriptionList(subs)
	if err != nil {
		return err
	}

	return nil
}

func (s *store) RemoveSubscriptions(tID badgesmodel.BadgeType, cID string) error {
	subs, err := s.getAllSubscriptions()
	if err != nil {
		return err
	}

	found := false
	for i, sub := range subs {
		if sub.ChannelID == cID && sub.TypeID == tID {
			subs = append(subs[:i], subs[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return nil
	}

	err = s.storeSubscriptionList(subs)
	if err != nil {
		return err
	}

	return nil
}

func (s *store) GetTypeSubscriptions(tID badgesmodel.BadgeType) ([]string, error) {
	subs, err := s.getAllSubscriptions()
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
func (s *store) GetChannelSubscriptions(cID string) ([]badgesmodel.BadgeTypeDefinition, error) {
	subs, err := s.getAllSubscriptions()
	if err != nil {
		return nil, err
	}

	out := []badgesmodel.BadgeTypeDefinition{}
	for _, sub := range subs {
		if sub.ChannelID == cID {
			t, err := s.GetType(sub.TypeID)
			if err != nil {
				s.api.LogDebug("cannot get type", "err", err)
				continue
			}
			out = append(out, *t)
		}
	}

	return out, nil
}

func (s *store) getBadgeFromList(badgeID badgesmodel.BadgeID, list []badgesmodel.Badge) (*badgesmodel.Badge, error) {
	for _, badge := range list {
		if badgeID == badge.ID {
			return &badge, nil
		}
	}
	return nil, errBadgeNotFound
}

func (s *store) getBadgeUsers(badgeID badgesmodel.BadgeID) (badgesmodel.OwnershipList, error) {
	_, err := s.getBadge(badgeID)
	if err != nil {
		return nil, errBadgeNotFound
	}

	ownership, err := s.getOwnershipList()
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
