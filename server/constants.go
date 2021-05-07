package main

const (
	KVKeyBadges    = "badges"
	KVKeyOwnership = "ownership"
	KVKeyTypes     = "types"

	AutocompletePath                 = "/autocomplete"
	AutocompletePathBadgeSuggestions = "/getBadgeSuggestions"
	AutocompletePathTypeSuggestions  = "/getBadgeTypeSuggestions"

	DialogPath            = "/dialog"
	DialogPathCreateBadge = "/createBadge"
	DialogPathCreateType  = "/createType"
	DialogPathGrant       = "/grant"

	DialogFieldBadgeName              = "name"
	DialogFieldBadgeMultiple          = "multiple"
	DialogFieldBadgeDescription       = "description"
	DialogFieldBadgeType              = "type"
	DialogFieldBadgeImage             = "image"
	DialogFieldTypeName               = "name"
	DialogFieldTypeEveryoneCanGrant   = "everyoneCanGrant"
	DialogFieldTypeAllowlistCanGrant  = "whitelistCanGrant"
	DialogFieldTypeEveryoneCanCreate  = "everyoneCanCreate"
	DialogFieldTypeAllowlistCanCreate = "whitelistCanCreate"
	DialogFieldUser                   = "user"
	DialogFieldBadge                  = "badge"
	DialogFieldNotifyHere             = "notify_here"

	TrueString  = "true"
	FalseString = "false"
)
