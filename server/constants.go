package main

const (
	KVKeyBadges    = "badges"
	KVKeyOwnership = "ownership"
	KVKeyTypes     = "types"

	AutocompletePath                     = "/autocomplete"
	AutocompletePathBadgeSuggestions     = "/getBadgeSuggestions"
	AutocompletePathTypeSuggestions      = "/getBadgeTypeSuggestions"
	AutocompletePathEditBadgeSuggestions = "/getEditBadgeSuggestions"
	AutocompletePathEditTypeSuggestions  = "/getEditTypeSuggestions"

	DialogPath            = "/dialog"
	DialogPathCreateBadge = "/createBadge"
	DialogPathSelectType  = "/selectType"
	DialogPathCreateType  = "/createType"
	DialogPathEditType    = "/editType"
	DialogPathGrant       = "/grant"
	DialogPathSelectBadge = "/selectBadge"
	DialogPathEditBadge   = "/editBadge"

	DialogFieldBadgeName              = "name"
	DialogFieldBadgeMultiple          = "multiple"
	DialogFieldBadgeDescription       = "description"
	DialogFieldBadgeType              = "type"
	DialogFieldBadgeImage             = "image"
	DialogFieldBadgeDelete            = "delete"
	DialogFieldTypeName               = "name"
	DialogFieldTypeEveryoneCanGrant   = "everyoneCanGrant"
	DialogFieldTypeAllowlistCanGrant  = "whitelistCanGrant"
	DialogFieldTypeEveryoneCanCreate  = "everyoneCanCreate"
	DialogFieldTypeAllowlistCanCreate = "whitelistCanCreate"
	DialogFieldTypeDelete             = "delete"
	DialogFieldUser                   = "user"
	DialogFieldBadge                  = "badge"
	DialogFieldNotifyHere             = "notify_here"

	TrueString  = "true"
	FalseString = "false"
)
