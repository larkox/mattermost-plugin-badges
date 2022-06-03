package main

import (
	"errors"
	"fmt"

	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	commandparser "github.com/larkox/mattermost-plugin-badges/server/command_parser"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/plugin"
	"github.com/spf13/pflag"
)

func getHelp() string {
	return `Available Commands:
`
}

func (p *Plugin) getCommand() *model.Command {
	return &model.Command{
		Trigger:          "badges",
		DisplayName:      "Badges Bot",
		Description:      "Badges",
		AutoComplete:     true,
		AutoCompleteDesc: "Available commands:",
		AutoCompleteHint: "[command]",
		AutocompleteData: p.getAutocompleteData(),
	}
}

func (p *Plugin) postCommandResponse(args *model.CommandArgs, text string) {
	post := &model.Post{
		UserId:    p.BotUserID,
		ChannelId: args.ChannelId,
		Message:   text,
	}
	p.mm.Post.SendEphemeralPost(args.UserId, post)
}

func commandError(text string) (bool, *model.CommandResponse, error) {
	return true, &model.CommandResponse{}, errors.New(text)
}

// ExecuteCommand executes a given command and returns a command response.
func (p *Plugin) ExecuteCommand(c *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	stringArgs := commandparser.Parse(args.Command)
	lengthOfArgs := len(stringArgs)
	restOfArgs := []string{}

	var handler func([]string, *model.CommandArgs) (bool, *model.CommandResponse, error)
	if lengthOfArgs == 1 {
		p.postCommandResponse(args, getHelp())
		return &model.CommandResponse{}, nil
	}
	command := stringArgs[1]
	if lengthOfArgs > 2 {
		restOfArgs = stringArgs[2:]
	}
	switch command {
	case "test-clean":
		handler = p.runClean
	case "grant":
		handler = p.runGrant
	case "edit":
		handler = p.runEdit
	case "create":
		handler = p.runCreate
	case "subscription":
		handler = p.runSubscription
	default:
		p.postCommandResponse(args, getHelp())
		return &model.CommandResponse{}, nil
	}
	isUserError, resp, err := handler(restOfArgs, args)
	if err != nil {
		if isUserError {
			p.postCommandResponse(args, fmt.Sprintf("__Error: %s__", err.Error()))
		} else {
			p.mm.Log.Error(err.Error())
			p.postCommandResponse(args, "An unknown error occurred. Please talk to your system administrator for help.")
		}
	}

	if resp != nil {
		return resp, nil
	}

	return &model.CommandResponse{}, nil
}

func (p *Plugin) runClean(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	user, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return false, &model.CommandResponse{Text: "Cannot get user."}, nil
	}
	if !user.IsSystemAdmin() {
		return false, &model.CommandResponse{Text: "Only a system admin can clean the badges database."}, nil
	}
	_ = p.mm.KV.DeleteAll()
	return false, &model.CommandResponse{Text: "Clean"}, nil
}

func (p *Plugin) runCreate(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	lengthOfArgs := len(args)
	restOfArgs := []string{}
	var handler func([]string, *model.CommandArgs) (bool, *model.CommandResponse, error)
	if lengthOfArgs == 0 {
		return false, &model.CommandResponse{Text: "Specify what you want to create."}, nil
	}
	command := args[0]
	if lengthOfArgs > 1 {
		restOfArgs = args[1:]
	}
	switch command {
	case "badge":
		handler = p.runCreateBadge
	case "type":
		handler = p.runCreateType
	default:
		return false, &model.CommandResponse{Text: "You can create either badge or type"}, nil
	}

	return handler(restOfArgs, extra)
}

func (p *Plugin) runCreateBadge(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	u, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return commandError(err.Error())
	}

	typeSuggestions, err := p.filterCreateBadgeTypes(u)
	if err != nil {
		return commandError(err.Error())
	}

	typeOptions := []*model.PostActionOptions{}
	for _, typeSuggestion := range typeSuggestions {
		id := string(typeSuggestion.ID)
		typeOptions = append(typeOptions, &model.PostActionOptions{Text: typeSuggestion.Name, Value: id})
	}

	if len(typeOptions) == 0 {
		return commandError("You cannot create badges from any type.")
	}

	err = p.mm.Frontend.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: extra.TriggerId,
		URL:       p.getDialogURL() + DialogPathCreateBadge,
		Dialog: model.Dialog{
			Title:       "Create badge",
			SubmitLabel: "Create",
			Elements: []model.DialogElement{
				{
					DisplayName: "Name",
					Type:        "text",
					Name:        DialogFieldBadgeName,
					MaxLength:   badgesmodel.NameMaxLength,
				},
				{
					DisplayName: "Description",
					Type:        "text",
					Name:        DialogFieldBadgeDescription,
					MaxLength:   badgesmodel.DescriptionMaxLength,
				},
				{
					DisplayName: "Image",
					Type:        "text",
					Name:        DialogFieldBadgeImage,
					HelpText:    "Insert a emoticon name",
				},
				{
					DisplayName: "Type",
					Type:        "select",
					Name:        DialogFieldBadgeType,
					Options:     typeOptions,
				},
				{
					DisplayName: "Multiple",
					Type:        "bool",
					Name:        DialogFieldBadgeMultiple,
					HelpText:    "Whether the badge can be granted multiple times",
					Optional:    true,
				},
			},
		},
	})

	if err != nil {
		return commandError(err.Error())
	}

	return false, &model.CommandResponse{}, nil
}

func (p *Plugin) runEdit(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	lengthOfArgs := len(args)
	restOfArgs := []string{}
	var handler func([]string, *model.CommandArgs) (bool, *model.CommandResponse, error)
	if lengthOfArgs == 0 {
		return false, &model.CommandResponse{Text: "Specify what you want to create."}, nil
	}
	command := args[0]
	if lengthOfArgs > 1 {
		restOfArgs = args[1:]
	}
	switch command {
	case "badge":
		handler = p.runEditBadge
	case "type":
		handler = p.runEditType
	default:
		return false, &model.CommandResponse{Text: "You can create either badge or type"}, nil
	}

	return handler(restOfArgs, extra)
}

func (p *Plugin) runEditBadge(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	u, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return commandError(err.Error())
	}

	var badgeIDStr string
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&badgeIDStr, "id", "", "ID of the badge")
	if err = fs.Parse(args); err != nil {
		return commandError(err.Error())
	}

	if badgeIDStr == "" {
		return commandError("You must set the badge ID")
	}

	badge, err := p.store.GetBadge(badgesmodel.BadgeID(badgeIDStr))
	if err != nil {
		return commandError(err.Error())
	}

	if !canEditBadge(u, p.badgeAdminUserID, badge) {
		return commandError("you cannot edit this badge")
	}

	typeSuggestions, err := p.filterCreateBadgeTypes(u)
	if err != nil {
		return commandError(err.Error())
	}

	typeOptions := []*model.PostActionOptions{}
	for _, typeSuggestion := range typeSuggestions {
		id := string(typeSuggestion.ID)
		typeOptions = append(typeOptions, &model.PostActionOptions{Text: typeSuggestion.Name, Value: id})
	}

	if len(typeOptions) == 0 {
		return commandError("You cannot create badges from any type.")
	}

	err = p.mm.Frontend.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: extra.TriggerId,
		URL:       p.getDialogURL() + DialogPathEditBadge,
		Dialog: model.Dialog{
			Title:       "Create badge",
			SubmitLabel: "Edit",
			State:       string(badge.ID),
			Elements: []model.DialogElement{
				{
					DisplayName: "Name",
					Type:        "text",
					Name:        DialogFieldBadgeName,
					MaxLength:   badgesmodel.NameMaxLength,
					Default:     badge.Name,
				},
				{
					DisplayName: "Description",
					Type:        "text",
					Name:        DialogFieldBadgeDescription,
					MaxLength:   badgesmodel.DescriptionMaxLength,
					Default:     badge.Description,
				},
				{
					DisplayName: "Image",
					Type:        "text",
					Name:        DialogFieldBadgeImage,
					HelpText:    "Insert a emoticon name",
					Default:     badge.Image,
				},
				{
					DisplayName: "Type",
					Type:        "select",
					Name:        DialogFieldBadgeType,
					Options:     typeOptions,
					Default:     string(badge.Type),
				},
				{
					DisplayName: "Multiple",
					Type:        "bool",
					Name:        DialogFieldBadgeMultiple,
					HelpText:    "Whether the badge can be granted multiple times",
					Optional:    true,
					Default:     getBooleanString(badge.Multiple),
				},
				{
					DisplayName: "Delete badge",
					Type:        "bool",
					Name:        DialogFieldBadgeDelete,
					HelpText:    "WARNING: Checking this will remove this badge permanently.",
					Optional:    true,
				},
			},
		},
	})

	if err != nil {
		return commandError(err.Error())
	}

	return false, &model.CommandResponse{}, nil
}

func (p *Plugin) runEditType(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	u, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return commandError(err.Error())
	}

	if !canCreateType(u, p.badgeAdminUserID, false) {
		return commandError("You have no permissions to edit a badge type.")
	}

	var badgeTypeStr string
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&badgeTypeStr, "type", "", "ID of the type")
	if err = fs.Parse(args); err != nil {
		return commandError(err.Error())
	}

	if badgeTypeStr == "" {
		return commandError("You must provide a type id")
	}

	typeDefinition, err := p.store.GetType(badgesmodel.BadgeType(badgeTypeStr))
	if err != nil {
		return commandError(err.Error())
	}

	if !canEditType(u, p.badgeAdminUserID, typeDefinition) {
		return commandError("you cannot edit this type")
	}

	canGrantAllowList := ""
	for uID, allowed := range typeDefinition.CanGrant.AllowList {
		if !allowed {
			continue
		}
		var allowedUser *model.User
		allowedUser, err = p.mm.User.Get(uID)
		if err != nil {
			continue
		}
		if canGrantAllowList == "" {
			canGrantAllowList += allowedUser.Username
			continue
		}

		canGrantAllowList += ", " + allowedUser.Username
	}

	canCreateAllowList := ""
	for uID, allowed := range typeDefinition.CanCreate.AllowList {
		if !allowed {
			continue
		}
		var allowedUser *model.User
		allowedUser, err = p.mm.User.Get(uID)
		if err != nil {
			continue
		}
		if canCreateAllowList == "" {
			canCreateAllowList += allowedUser.Username
			continue
		}

		canCreateAllowList += ", " + allowedUser.Username
	}

	err = p.mm.Frontend.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: extra.TriggerId,
		URL:       p.getDialogURL() + DialogPathEditType,
		Dialog: model.Dialog{
			Title:       "Edit type",
			SubmitLabel: "Edit",
			State:       badgeTypeStr,
			Elements: []model.DialogElement{
				{
					DisplayName: "Name",
					Type:        "text",
					Name:        DialogFieldTypeName,
					MaxLength:   badgesmodel.NameMaxLength,
					Default:     typeDefinition.Name,
				},
				{
					DisplayName: "Everyone can create badge",
					Type:        "bool",
					Name:        DialogFieldTypeEveryoneCanCreate,
					HelpText:    "Whether any user can create a badge of this type",
					Optional:    true,
					Default:     getBooleanString(typeDefinition.CanCreate.Everyone),
				},
				{
					DisplayName: "Can create allowlist",
					Type:        "text",
					Name:        DialogFieldTypeAllowlistCanCreate,
					HelpText:    "Fill the usernames separated by comma (,) of the people that can create badges of this type.",
					Placeholder: "user-1, user-2, user-3",
					Optional:    true,
					Default:     canCreateAllowList,
				},
				{
					DisplayName: "Everyone can grant badge",
					Type:        "bool",
					Name:        DialogFieldTypeEveryoneCanGrant,
					HelpText:    "Whether any user can grant a badge of this type",
					Optional:    true,
					Default:     getBooleanString(typeDefinition.CanGrant.Everyone),
				},
				{
					DisplayName: "Can grant allowlist",
					Type:        "text",
					Name:        DialogFieldTypeAllowlistCanGrant,
					HelpText:    "Fill the usernames separated by comma (,) of the people that can grant badges of this type.",
					Placeholder: "user-1, user-2, user-3",
					Optional:    true,
					Default:     canGrantAllowList,
				},
				{
					DisplayName: "Remove type",
					Type:        "bool",
					Name:        DialogFieldTypeDelete,
					HelpText:    "WARNING: checking this will remove this type and all associated badges permanently.",
					Optional:    true,
				},
			},
		},
	})

	if err != nil {
		return commandError(err.Error())
	}

	return false, &model.CommandResponse{}, nil
}

func (p *Plugin) runCreateType(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	u, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return commandError(err.Error())
	}

	if !canCreateType(u, p.badgeAdminUserID, false) {
		return commandError("You have no permissions to create a badge type.")
	}

	err = p.mm.Frontend.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: extra.TriggerId,
		URL:       p.getDialogURL() + DialogPathCreateType,
		Dialog: model.Dialog{
			Title:       "Create type",
			SubmitLabel: "Create",
			Elements: []model.DialogElement{
				{
					DisplayName: "Name",
					Type:        "text",
					Name:        DialogFieldTypeName,
					MaxLength:   badgesmodel.NameMaxLength,
				},
				{
					DisplayName: "Everyone can create badge",
					Type:        "bool",
					Name:        DialogFieldTypeEveryoneCanCreate,
					HelpText:    "Whether any user can create a badge of this type",
					Optional:    true,
				},
				{
					DisplayName: "Can create allowlist",
					Type:        "text",
					Name:        DialogFieldTypeAllowlistCanCreate,
					HelpText:    "Fill the usernames separated by comma (,) of the people that can create badges of this type.",
					Placeholder: "user-1, user-2, user-3",
					Optional:    true,
				},
				{
					DisplayName: "Everyone can grant badge",
					Type:        "bool",
					Name:        DialogFieldTypeEveryoneCanGrant,
					HelpText:    "Whether any user can grant a badge of this type",
					Optional:    true,
				},
				{
					DisplayName: "Can grant allowlist",
					Type:        "text",
					Name:        DialogFieldTypeAllowlistCanGrant,
					HelpText:    "Fill the usernames separated by comma (,) of the people that can grant badges of this type.",
					Placeholder: "user-1, user-2, user-3",
					Optional:    true,
				},
			},
		},
	})

	if err != nil {
		return commandError(err.Error())
	}

	return false, &model.CommandResponse{}, nil
}

func (p *Plugin) runGrant(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	badgeStr := ""
	username := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&badgeStr, "badge", "", "ID of the badge")
	fs.StringVar(&username, "user", "", "Username to grant to")
	if err := fs.Parse(args); err != nil {
		return commandError(err.Error())
	}

	if username != "" && badgeStr != "" {
		if username[0] == '@' {
			username = username[1:]
		}

		granter, err := p.mm.User.Get(extra.UserId)
		if err != nil {
			return commandError(err.Error())
		}

		badge, err := p.store.GetBadge(badgesmodel.BadgeID(badgeStr))
		if err != nil {
			return commandError(err.Error())
		}

		badgeType, err := p.store.GetType(badge.Type)
		if err != nil {
			return commandError(err.Error())
		}

		if !canGrantBadge(granter, p.badgeAdminUserID, badge, badgeType) {
			return commandError("you have no permissions to grant this badge")
		}

		user, err := p.mm.User.GetByUsername(username)
		if err != nil {
			return commandError(err.Error())
		}

		shouldNotify, err := p.store.GrantBadge(badgesmodel.BadgeID(badgeStr), user.Id, extra.UserId, "")
		if err != nil {
			return commandError(err.Error())
		}

		if shouldNotify {
			p.notifyGrant(badgesmodel.BadgeID(badgeStr), extra.UserId, user, false, "", "")
		}

		p.postCommandResponse(extra, "Granted")
		return false, &model.CommandResponse{}, nil
	}

	elements := []model.DialogElement{}

	stateText := ""
	introductionText := ""
	if username != "" {
		if username[0] == '@' {
			username = username[1:]
		}

		user, err := p.mm.User.GetByUsername(username)
		if err != nil {
			return commandError(err.Error())
		}

		introductionText = "Grant badge to @" + username
		stateText = user.Id
	}

	if stateText == "" {
		elements = append(elements, model.DialogElement{
			DisplayName: "User",
			Type:        "select",
			Name:        DialogFieldUser,
			DataSource:  "users",
		})
	}

	actingUser, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return commandError(err.Error())
	}

	options := []*model.PostActionOptions{}
	grantableBadges, err := p.filterGrantBadges(actingUser)
	if err != nil {
		return commandError(err.Error())
	}
	for _, badge := range grantableBadges {
		options = append(options, &model.PostActionOptions{Text: badge.Name, Value: string(badge.ID)})
	}

	badgeElement := model.DialogElement{
		DisplayName: "Badge",
		Type:        "select",
		Name:        DialogFieldBadge,
		Options:     options,
	}

	if badgeStr != "" {
		found := false
		for _, badge := range grantableBadges {
			if badgeStr == string(badge.ID) {
				found = true
				break
			}
		}

		if !found {
			return commandError("You cannot grant that badge")
		}

		badgeElement.Default = badgeStr
	}

	elements = append(elements, badgeElement)

	elements = append(elements, model.DialogElement{
		DisplayName: "Reason",
		Name:        DialogFieldGrantReason,
		Optional:    true,
		HelpText:    "Reason why you are granting this badge. This will be seen by the user, and wherever this grant notification is shown (e.g. subscriptions).",
		Type:        "text",
	})

	elements = append(elements, model.DialogElement{
		DisplayName: "Notify on this channel",
		Name:        DialogFieldNotifyHere,
		Type:        "bool",
		HelpText:    "If you mark this, the bot will send a message to this channel notifying that you granted this badge to this person.",
		Optional:    true,
	})

	err = p.mm.Frontend.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: extra.TriggerId,
		URL:       p.getDialogURL() + DialogPathGrant,
		Dialog: model.Dialog{
			Title:            "Grant badge",
			IntroductionText: introductionText,
			SubmitLabel:      "Grant",
			Elements:         elements,
			State:            stateText,
		},
	})

	if err != nil {
		return commandError(err.Error())
	}

	return false, &model.CommandResponse{}, nil
}

func (p *Plugin) runSubscription(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	lengthOfArgs := len(args)
	restOfArgs := []string{}
	var handler func([]string, *model.CommandArgs) (bool, *model.CommandResponse, error)
	if lengthOfArgs == 0 {
		return false, &model.CommandResponse{Text: "Specify what you want to do."}, nil
	}
	command := args[0]
	if lengthOfArgs > 1 {
		restOfArgs = args[1:]
	}
	switch command {
	case "create":
		handler = p.runCreateSubscription
	case "remove":
		handler = p.runDeleteSubscription
	default:
		return false, &model.CommandResponse{Text: "You can either create or delete subscriptions"}, nil
	}

	return handler(restOfArgs, extra)
}

func (p *Plugin) runCreateSubscription(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	typeStr := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&typeStr, "type", "", "ID of the badge")
	if err := fs.Parse(args); err != nil {
		return commandError(err.Error())
	}

	actingUser, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return commandError(err.Error())
	}

	if !canCreateSubscription(actingUser, p.badgeAdminUserID, extra.ChannelId) {
		return commandError("You cannot create subscriptions")
	}

	if typeStr != "" {

		err = p.store.AddSubscription(badgesmodel.BadgeType(typeStr), extra.ChannelId)
		if err != nil {
			return commandError(err.Error())
		}

		p.postCommandResponse(extra, "Granted")
		return false, &model.CommandResponse{}, nil
	}

	options := []*model.PostActionOptions{}
	typesDefinitions, err := p.filterEditTypes(actingUser)
	if err != nil {
		return commandError(err.Error())
	}
	for _, typeDefinition := range typesDefinitions {
		options = append(options, &model.PostActionOptions{Text: typeDefinition.Name, Value: string(typeDefinition.ID)})
	}

	err = p.mm.Frontend.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: extra.TriggerId,
		URL:       p.getDialogURL() + DialogPathCreateSubscription,
		Dialog: model.Dialog{
			Title:            "Create subscription",
			IntroductionText: "Introduce the badge type you want to subscribe to this channel.",
			SubmitLabel:      "Add",
			Elements: []model.DialogElement{
				{
					DisplayName: "Type",
					Type:        "select",
					Name:        DialogFieldBadgeType,
					Options:     options,
				},
			},
		},
	})

	if err != nil {
		return commandError(err.Error())
	}

	return false, &model.CommandResponse{}, nil
}

func (p *Plugin) runDeleteSubscription(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	typeStr := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&typeStr, "type", "", "ID of the badge")
	if err := fs.Parse(args); err != nil {
		return commandError(err.Error())
	}

	actingUser, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return commandError(err.Error())
	}

	if !canCreateSubscription(actingUser, p.badgeAdminUserID, extra.ChannelId) {
		return commandError("You cannot create subscriptions")
	}

	if typeStr != "" {
		err = p.store.RemoveSubscriptions(badgesmodel.BadgeType(typeStr), extra.ChannelId)
		if err != nil {
			return commandError(err.Error())
		}

		p.postCommandResponse(extra, "Removed")
		return false, &model.CommandResponse{}, nil
	}

	options := []*model.PostActionOptions{}
	typesDefinitions, err := p.store.GetChannelSubscriptions(extra.ChannelId)
	if err != nil {
		return commandError(err.Error())
	}
	for _, typeDefinition := range typesDefinitions {
		options = append(options, &model.PostActionOptions{Text: typeDefinition.Name, Value: string(typeDefinition.ID)})
	}

	err = p.mm.Frontend.OpenInteractiveDialog(model.OpenDialogRequest{
		TriggerId: extra.TriggerId,
		URL:       p.getDialogURL() + DialogPathDeleteSubscription,
		Dialog: model.Dialog{
			Title:            "Delete subscription",
			IntroductionText: "Introduce the badge type you want to remove from this channel.",
			SubmitLabel:      "Remove",
			Elements: []model.DialogElement{
				{
					DisplayName: "Type",
					Type:        "select",
					Name:        DialogFieldBadgeType,
					Options:     options,
				},
			},
		},
	})

	if err != nil {
		return commandError(err.Error())
	}

	return false, &model.CommandResponse{}, nil
}

func (p *Plugin) getAutocompleteData() *model.AutocompleteData {
	badges := model.NewAutocompleteData("badges", "[command]", "Available commands: grant")

	grant := model.NewAutocompleteData("grant", "--user @username --badge id", "Grant a badge to a user")
	grant.AddNamedDynamicListArgument("badge", "--badge badgeID", getAutocompletePath(AutocompletePathBadgeSuggestions), true)
	grant.AddNamedTextArgument("user", "User to grant the badge to", "--user @username", "", true)
	badges.AddCommand(grant)

	create := model.NewAutocompleteData("create", "badge | type", "Create a badge or a type")

	badge := model.NewAutocompleteData(
		"badge",
		"",
		"Create a badge",
	)
	create.AddCommand(badge)

	createType := model.NewAutocompleteData(
		"type",
		"",
		"Create a badge type",
	)
	create.AddCommand(createType)

	badges.AddCommand(create)

	edit := model.NewAutocompleteData("edit", "badge | type", "Edit a badge or a type")

	editBadge := model.NewAutocompleteData(
		"badge",
		"",
		"Edit a badge",
	)
	editBadge.AddNamedDynamicListArgument("id", "--id badgeID", getAutocompletePath(AutocompletePathEditBadgeSuggestions), true)
	edit.AddCommand(editBadge)

	editType := model.NewAutocompleteData(
		"type",
		"",
		"Edit a badge type",
	)
	editType.AddNamedDynamicListArgument("type", "--type typeID", getAutocompletePath(AutocompletePathEditTypeSuggestions), true)
	edit.AddCommand(editType)

	badges.AddCommand(edit)

	subscription := model.NewAutocompleteData("subscription", "create | remove", "Manage this channel subscriptions")

	createSubscription := model.NewAutocompleteData(
		"create",
		"",
		"Create a subscription",
	)
	subscription.AddCommand(createSubscription)

	deleteSubscription := model.NewAutocompleteData(
		"remove",
		"",
		"Remove a subscription",
	)
	subscription.AddCommand(deleteSubscription)

	badges.AddCommand(subscription)

	return badges
}

func getAutocompletePath(path string) string {
	return "plugins/" + manifest.Id + AutocompletePath + path
}
