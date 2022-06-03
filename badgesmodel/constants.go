package badgesmodel

const (
	NameMaxLength        = 20
	DescriptionMaxLength = 120

	ImageTypeEmoji       ImageType = "emoji"
	ImageTypeRelativeURL ImageType = "rel_url"
	ImageTypeAbsoluteURL ImageType = "abs_url"

	PluginPath          = "/com.mattermost.badges"
	PluginAPIPath       = "/papi/v1"
	PluginAPIPathEnsure = "/ensure"
	PluginAPIPathGrant  = "/grant"
)
