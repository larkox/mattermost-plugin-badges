package main

import (
	"time"
)

type BadgeType int
type BadgeID int

type Ownership struct {
	User      string    `json:"user"`
	GrantedBy string    `json:"granted_by"`
	Badge     BadgeID   `json:"badge"`
	Time      time.Time `json:"time"`
}

type OwnershipList []Ownership

type Badge struct {
	ID          BadgeID   `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Image       string    `json:"image"`
	ImageType   ImageType `json:"image_type"`
	Multiple    bool      `json:"multiple"`
	Type        BadgeType `json:"type"`
	CreatedBy   string    `json:"created_by"`
}

type UserBadge struct {
	Badge
	Ownership
	GrantedByUsername string `json:"granted_by_name"`
}

type BadgeDetails struct {
	Badge
	Owners            []Ownership `json:"owners"`
	CreatedByUsername string      `json:"created_by_username"`
}

type AllBadgesBadge struct {
	Badge
	Granted      int `json:"granted"`
	GrantedTimes int `json:"granted_times"`
}

type BadgeTypeDefinition struct {
	ID        BadgeType        `json:"id"`
	Name      string           `json:"name"`
	Frame     string           `json:"frame"`
	CreatedBy string           `json:"created_by"`
	CanGrant  PermissionScheme `json:"can_grant"`
	CanCreate PermissionScheme `json:"can_create"`
}

type PermissionScheme struct {
	Everyone  bool            `json:"everyone"`
	Roles     map[string]bool `json:"roles"`
	AllowList map[string]bool `json:"allow_list"`
	BlockList map[string]bool `json:"block_list"`
}

type BadgeTypeList []BadgeTypeDefinition

type ImageType string

func (b Badge) IsValid() bool {
	return len(b.Name) <= NameMaxLength &&
		len(b.Description) <= DescriptionMaxLength &&
		b.Image != ""
}

func (l OwnershipList) IsOwned(user string, badge BadgeID) bool {
	for _, ownership := range l {
		if user == ownership.User && badge == ownership.Badge {
			return true
		}
	}
	return false
}

func (l BadgeTypeList) GetType(id BadgeType) *BadgeTypeDefinition {
	for _, t := range l {
		if t.ID == id {
			return &t
		}
	}

	return nil
}
