package main

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
)

var errInvalidBadge = errors.New("invalid badge")
var errBadgeNotFound = errors.New("badge not found")

type Store interface {
	// Interface
	GetUserBadges(userID string) ([]UserBadge, error)
	GetAllBadges() ([]AllBadgesBadge, error)
	GetBadgeDetails(badgeID BadgeID) (*BadgeDetails, error)

	// Autocomplete
	GetGrantSuggestions(user model.User) ([]Badge, error)
	GetTypeSuggestions(user model.User) (BadgeTypeList, error)

	// API
	AddBadge(badge Badge) (*Badge, error)
	GrantBadge(badgeID BadgeID, userID string, grantedBy string) error
	AddType(t BadgeTypeDefinition) (*BadgeTypeDefinition, error)

	// DEBUG
	DebugGetTypes() BadgeTypeList
}

type store struct {
	api plugin.API
}

func NewStore(api plugin.API) Store {
	return &store{
		api: api,
	}
}

func (s *store) DebugGetTypes() BadgeTypeList {
	l, _ := s.getAllTypes()
	return l
}

func (s *store) AddBadge(b Badge) (*Badge, error) {
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

	var lastID BadgeID = -1
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

func (s *store) AddType(t BadgeTypeDefinition) (*BadgeTypeDefinition, error) {
	u, appErr := s.api.GetUser(t.CreatedBy)
	if appErr != nil {
		return nil, appErr
	}

	if !canCreateType(*u) {
		return nil, errors.New("you have no permission to create this badge type")
	}

	badgeTypes, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	var lastID BadgeType = -1
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

func (s *store) GetAllBadges() ([]AllBadgesBadge, error) {
	badges, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	ownership, err := s.getOwnershipList()
	if err != nil {
		return nil, err
	}

	out := []AllBadgesBadge{}
	for _, b := range badges {
		badge := AllBadgesBadge{
			Badge: b,
		}
		grantedTo := map[string]bool{}
		for _, o := range ownership {
			if o.Badge == badge.ID {
				badge.GrantedTimes++
			}
			if !grantedTo[o.User] {
				badge.Granted++
				grantedTo[o.User] = true
			}
		}
		out = append(out, badge)
	}

	return out, nil
}

func (s *store) GetGrantSuggestions(user model.User) ([]Badge, error) {
	badges, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	types, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	out := []Badge{}
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

func (s *store) GetTypeSuggestions(user model.User) (BadgeTypeList, error) {
	types, err := s.getAllTypes()
	if err != nil {
		return nil, err
	}

	out := BadgeTypeList{}
	for _, t := range types {
		if canCreateBadge(user, t) {
			out = append(out, t)
		}
	}

	return out, nil
}

func (s *store) getAllTypes() (BadgeTypeList, error) {
	data, appErr := s.api.KVGet(KVKeyTypes)
	if appErr != nil {
		return nil, appErr
	}

	typeList := []BadgeTypeDefinition{}
	if data != nil {
		err := json.Unmarshal(data, &typeList)
		if err != nil {
			return nil, err
		}
	}

	return typeList, nil
}

func (s *store) getAllBadges() ([]Badge, error) {
	data, appErr := s.api.KVGet(KVKeyBadges)
	if appErr != nil {
		return nil, appErr
	}

	badgeList := []Badge{}
	if data != nil {
		err := json.Unmarshal(data, &badgeList)
		if err != nil {
			return nil, err
		}
	}

	return badgeList, nil
}

func (s *store) getBadge(id BadgeID) (*Badge, error) {
	badgeList, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	return s.getBadgeFromList(id, badgeList)
}

func (s *store) GetBadgeDetails(id BadgeID) (*BadgeDetails, error) {
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

	return &BadgeDetails{
		Badge:             *badge,
		Owners:            owners,
		CreatedByUsername: createdByName,
	}, nil
}

func (s *store) getOwnershipList() (OwnershipList, error) {
	data, appErr := s.api.KVGet(KVKeyOwnership)
	if appErr != nil {
		return nil, appErr
	}

	ownership := OwnershipList{}
	if data != nil {
		err := json.Unmarshal(data, &ownership)
		if err != nil {
			return nil, err
		}
	}

	return ownership, nil
}

func (s *store) GrantBadge(id BadgeID, userID string, grantedBy string) error {
	badge, err := s.getBadge(id)
	if err != nil {
		return err
	}

	types, err := s.getAllTypes()
	if err != nil {
		return err
	}

	badgeType := types.GetType(badge.Type)
	if badgeType == nil {
		return errors.New("badge type not found")
	}

	user, appErr := s.api.GetUser(userID)
	if appErr != nil {
		return err
	}

	grantedByUser, appErr := s.api.GetUser(grantedBy)
	if appErr != nil {
		return err
	}

	if !canGrantBadge(*grantedByUser, *badge, *badgeType) {
		return errors.New("you don't have permission to grant this badge")
	}

	ownership, err := s.getOwnershipList()
	if err != nil {
		return err
	}

	if !badge.Multiple && ownership.IsOwned(user.Id, id) {
		return nil
	}

	ownership = append(ownership, Ownership{
		User:      user.Id,
		Badge:     badge.ID,
		Time:      time.Now(),
		GrantedBy: grantedByUser.Id,
	})

	data, err := json.Marshal(ownership)
	if err != nil {
		return err
	}

	appErr = s.api.KVSet(KVKeyOwnership, data)
	if appErr != nil {
		return appErr
	}

	return nil
}

func (s *store) GetUserBadges(userID string) ([]UserBadge, error) {
	ownership, err := s.getOwnershipList()
	if err != nil {
		return nil, err
	}

	badges, err := s.getAllBadges()
	if err != nil {
		return nil, err
	}

	out := []UserBadge{}
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

			out = append([]UserBadge{{Badge: *badge, Ownership: o, GrantedByUsername: grantedByName}}, out...)
		}
	}

	return out, nil
}

func (s *store) getBadgeFromList(badgeID BadgeID, list []Badge) (*Badge, error) {
	for _, badge := range list {
		if badgeID == badge.ID {
			return &badge, nil
		}
	}
	return nil, errBadgeNotFound
}

func (s *store) getBadgeUsers(badgeID BadgeID) (OwnershipList, error) {
	_, err := s.getBadge(badgeID)
	if err != nil {
		return nil, errBadgeNotFound
	}

	ownership, err := s.getOwnershipList()
	if err != nil {
		return nil, err
	}

	out := OwnershipList{}
	for _, o := range ownership {
		if o.Badge == badgeID {
			out = append(out, o)
		}
	}

	return out, nil
}
