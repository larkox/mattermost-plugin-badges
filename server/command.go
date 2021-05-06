package main

import (
	"errors"
	"fmt"
	"strconv"

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
	case "test-initial-badges":
		handler = p.runTestInitialBadges
	case "grant":
		handler = p.runGrant
	case "create":
		handler = p.runCreate
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

	typeSuggestions, err := p.store.GetTypeSuggestions(*u)
	if err != nil {
		return commandError(err.Error())
	}

	typeOptions := []*model.PostActionOptions{}
	for _, typeSuggestion := range typeSuggestions {
		id := strconv.Itoa(int(typeSuggestion.ID))
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

func (p *Plugin) runCreateType(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	u, err := p.mm.User.Get(extra.UserId)
	if err != nil {
		return commandError(err.Error())
	}

	if !canCreateType(*u, false) {
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
					DisplayName: "Everyone can grant badge",
					Type:        "bool",
					Name:        DialogFieldTypeEveryoneCanGrant,
					HelpText:    "Whether any user can grant a badge of this type",
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

		user, err := p.mm.User.GetByUsername(username)
		if err != nil {
			return commandError(err.Error())
		}

		badgeID, err := strconv.Atoi(badgeStr)
		if err != nil {
			return commandError(err.Error())
		}

		shouldNotify, err := p.store.GrantBadge(badgesmodel.BadgeID(badgeID), user.Id, extra.UserId)
		if err != nil {
			return commandError(err.Error())
		}

		if shouldNotify {
			p.notifyGrant(badgesmodel.BadgeID(badgeID), extra.UserId, user.Id)
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
	grantableBadges, err := p.store.GetGrantSuggestions(*actingUser)
	if err != nil {
		return commandError(err.Error())
	}
	for _, badge := range grantableBadges {
		options = append(options, &model.PostActionOptions{Text: badge.Name, Value: strconv.Itoa(int(badge.ID))})
	}

	badgeElement := model.DialogElement{
		DisplayName: "Badge",
		Type:        "select",
		Name:        DialogFieldBadge,
		Options:     options,
	}

	if badgeStr != "" {
		badgeID, err := strconv.Atoi(badgeStr)
		if err != nil {
			return commandError(err.Error())
		}

		found := false
		for _, badge := range grantableBadges {
			if badgeID == int(badge.ID) {
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

func (p *Plugin) runTestInitialBadges(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	_ = p.mm.KV.DeleteAll()

	t := &badgesmodel.BadgeTypeDefinition{
		Name:      "Demo",
		Frame:     "",
		CreatedBy: extra.UserId,
		CanGrant: badgesmodel.PermissionScheme{
			Everyone: true,
		},
		CanCreate: badgesmodel.PermissionScheme{
			Everyone: true,
		},
	}
	t, _ = p.store.AddType(*t)

	info := []struct {
		name        string
		description string
		image       string
	}{
		{
			name:        "Sporty",
			description: "Won a sports event",
			image:       "medal_sports",
		},
		{
			name:        "General",
			description: "Won a manager event",
			image:       "medal_military",
		},
		{
			name:        "1st place",
			description: "Got first place in an event",
			image:       "1st_place_medal",
		},
		{
			name:        "2nd place",
			description: "Got second place in an event",
			image:       "2nd_place_medal",
		},
		{
			name:        "3rd place",
			description: "Got third place in an event",
			image:       "3rd_place_medal",
		},
		{
			name:        "Winner",
			description: "Won a trophy",
			image:       "trophy",
		},
		{
			name:        "Racer",
			description: "Won a car race",
			image:       "racing_car",
		},
		{
			name:        "Poker Pro",
			description: "Won a poker game",
			image:       "spades",
		},
	}
	for _, i := range info {
		b := badgesmodel.Badge{
			Name:        i.name,
			Description: i.description,
			Image:       i.image,
			ImageType:   badgesmodel.ImageTypeEmoji,
			Type:        t.ID,
			Multiple:    true,
			CreatedBy:   extra.UserId,
		}
		_, _ = p.store.AddBadge(b)
	}

	_, _ = p.store.GrantBadge(0, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(1, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(2, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(1, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(3, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(4, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(0, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(5, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(6, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(7, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(4, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(3, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(6, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(1, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(3, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(4, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(6, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(5, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(7, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(7, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(6, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(7, extra.UserId, extra.UserId)
	_, _ = p.store.GrantBadge(4, extra.UserId, extra.UserId)
	return false, &model.CommandResponse{Text: "Added"}, nil
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
	return badges
}

func getAutocompletePath(path string) string {
	return "plugins/" + manifest.Id + AutocompletePath + path
}
