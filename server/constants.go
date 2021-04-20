package main

const (
	NameMaxLength        = 20
	DescriptionMaxLength = 120

	KVKeyBadges    = "badges"
	KVKeyOwnership = "ownership"
	KVKeyTypes     = "types"

	ImageTypeEmoji       ImageType = "emoji"
	ImageTypeRelativeURL ImageType = "rel_url"
	ImageTypeAbsoluteURL ImageType = "abs_url"

	BadgeTypeManual BadgeType = 0
	BadgeTypePlugin BadgeType = 1

	AutocompletePath                 = "/autocomplete"
	AutocompletePathBadgeSuggestions = "/getBadgeSuggestions"
	AutocompletePathTypeSuggestions  = "/getBadgeTypeSuggestions"

	TrueString  = "true"
	FalseString = "false"
)
