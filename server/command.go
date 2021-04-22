package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

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
	_ = p.API.SendEphemeralPost(args.UserId, post)
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
	case "test-kv":
		handler = p.runTestKV
	case "test-kv-one-more":
		handler = p.runTestKVOneMore
	case "test-clean":
		handler = p.runClean
	case "test-add-badge":
		handler = p.runTestAddBadge
	case "test-initial-badges":
		handler = p.runTestInitialBadges
	case "test-list-types":
		handler = p.runTestListTypes
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
			p.postCommandResponse(args, fmt.Sprintf("__Error: %s.__\n\nRun `/todo help` for usage instructions.", err.Error()))
		} else {
			p.API.LogError(err.Error())
			p.postCommandResponse(args, "An unknown error occurred. Please talk to your system administrator for help.")
		}
	}

	if resp != nil {
		return resp, nil
	}

	return &model.CommandResponse{}, nil
}

func (p *Plugin) runTestListTypes(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	t := p.store.DebugGetTypes()
	b, _ := json.MarshalIndent(t, "", "    ")
	return false, &model.CommandResponse{Text: "```\n" + string(b) + "\n```"}, nil
}

func (p *Plugin) runClean(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	_ = p.API.KVDeleteAll()
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
	b := Badge{
		ImageType: ImageTypeEmoji,
		CreatedBy: extra.UserId,
	}
	var badgeType int
	var multiple string
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&b.Name, "name", "", "Name of the badge")
	fs.StringVar(&b.Description, "description", "", "Description of the badge")
	fs.StringVar(&b.Image, "image", "", "Image of the badge")
	fs.IntVar(&badgeType, "type", 0, "Type of the badge")
	fs.StringVar(&multiple, "multiple", "", "Whether the badge can be granted multiple times")
	if err := fs.Parse(args); err != nil {
		return false, &model.CommandResponse{Text: err.Error()}, nil
	}

	b.Type = BadgeType(badgeType)
	b.Multiple = multiple == TrueString
	if b.Image[0] == ':' {
		b.Image = b.Image[1 : len(b.Image)-1]
	}

	_, err := p.store.AddBadge(b)
	if err != nil {
		return false, &model.CommandResponse{Text: err.Error()}, nil
	}

	return false, &model.CommandResponse{Text: "Created"}, nil
}

func (p *Plugin) runCreateType(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	t := BadgeTypeDefinition{
		CreatedBy: extra.UserId,
	}

	var everyoneCanGrant string
	var everyoneCanCreate string
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&t.Name, "name", "", "Name of the type")
	fs.StringVar(&everyoneCanCreate, "everyoneCanCreate", "", "Whether everyone can create badges of this type")
	fs.StringVar(&everyoneCanGrant, "everyoneCanGrant", "", "Whether everyone can grant badge")
	if err := fs.Parse(args); err != nil {
		return false, &model.CommandResponse{Text: err.Error()}, nil
	}

	t.CanGrant.Everyone = everyoneCanGrant == TrueString
	t.CanCreate.Everyone = everyoneCanCreate == TrueString

	_, err := p.store.AddType(t)
	if err != nil {
		return false, &model.CommandResponse{Text: err.Error()}, nil
	}

	return false, &model.CommandResponse{Text: "Created"}, nil
}

func (p *Plugin) runGrant(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	badgeStr := ""
	username := ""
	fs := pflag.NewFlagSet("", pflag.ContinueOnError)
	fs.StringVar(&badgeStr, "badge", "", "ID of the badge")
	fs.StringVar(&username, "user", "", "Username to grant to")
	if err := fs.Parse(args); err != nil {
		return false, &model.CommandResponse{Text: err.Error()}, nil
	}

	if username[0] == '@' {
		username = username[1:]
	}

	user, appErr := p.API.GetUserByUsername(username)
	if appErr != nil {
		return false, &model.CommandResponse{Text: appErr.Error()}, nil
	}

	badgeID, err := strconv.Atoi(badgeStr)
	if err != nil {
		return false, &model.CommandResponse{Text: err.Error()}, nil
	}

	err = p.store.GrantBadge(BadgeID(badgeID), user.Id, extra.UserId)
	if err != nil {
		return false, &model.CommandResponse{Text: err.Error()}, nil
	}

	return false, &model.CommandResponse{Text: "Granted"}, nil
}

func (p *Plugin) runTestAddBadge(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	b := Badge{
		Name:        "test",
		Description: "test",
		Image:       "sweat_smile",
		ImageType:   ImageTypeEmoji,
		Type:        0,
		Multiple:    true,
		CreatedBy:   extra.UserId,
	}
	_, _ = p.store.AddBadge(b)
	return false, &model.CommandResponse{Text: "Added"}, nil
}

func (p *Plugin) runTestInitialBadges(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	_ = p.API.KVDeleteAll()

	t := &BadgeTypeDefinition{
		Name:      "Demo",
		Frame:     "",
		CreatedBy: extra.UserId,
		CanGrant: PermissionScheme{
			Everyone: true,
		},
		CanCreate: PermissionScheme{
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
		b := Badge{
			Name:        i.name,
			Description: i.description,
			Image:       i.image,
			ImageType:   ImageTypeEmoji,
			Type:        t.ID,
			Multiple:    true,
			CreatedBy:   extra.UserId,
		}
		_, _ = p.store.AddBadge(b)
	}

	_ = p.store.GrantBadge(0, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(1, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(2, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(1, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(3, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(4, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(0, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(5, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(6, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(7, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(4, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(3, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(6, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(1, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(3, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(4, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(6, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(5, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(7, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(7, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(6, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(7, extra.UserId, extra.UserId)
	_ = p.store.GrantBadge(4, extra.UserId, extra.UserId)
	return false, &model.CommandResponse{Text: "Added"}, nil
}

func (p *Plugin) runTestKV(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	_ = p.API.KVDeleteAll()
	for i := 0; i < 1000; i++ {
		b := Badge{}
		b.Image = ":sweat_smile:"
		b.ImageType = ImageTypeEmoji
		for j := 0; j < NameMaxLength; j++ {
			b.Name += "a"
		}
		for j := 0; j < DescriptionMaxLength; j++ {
			b.Description += "a"
		}
		b.Type = 99
		b.Multiple = true

		_, err := p.store.AddBadge(b)
		if err != nil {
			p.API.LogError("WE REACHED THE LIMIT!", "limit", i, "error", err)
			p.postCommandResponse(extra, "Error")
			return false, &model.CommandResponse{}, nil
		}
	}

	for i := 0; i < 1000*1000; i++ {
		err := p.store.GrantBadge(0, extra.UserId, extra.UserId)
		if err != nil {
			p.API.LogError("WE REACHED THE LIMIT! (ownership)", "limit", i, "error", err)
			p.postCommandResponse(extra, "Error")
			return false, &model.CommandResponse{}, nil
		}
	}
	p.postCommandResponse(extra, "allfine")
	return false, &model.CommandResponse{}, nil
}

func (p *Plugin) runTestKVOneMore(args []string, extra *model.CommandArgs) (bool, *model.CommandResponse, error) {
	before := time.Now()
	_ = p.store.GrantBadge(0, extra.UserId, extra.UserId)
	after := time.Now()
	p.postCommandResponse(extra, fmt.Sprintf("Lasted: %s", after.Sub(before).String()))
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
		"--name badgeName --description badgeDescription --image :image: --type typeID --multiple true|false",
		"Create a badge",
	)
	badge.AddNamedTextArgument("name", "Name of the badge", "--name badgeName", "", true)
	badge.AddNamedTextArgument("description", "Description of the badge", "--description description", "", true)
	badge.AddNamedTextArgument("image", "Image of the badge", "--image :image:", "", true)
	badge.AddNamedDynamicListArgument("type", "Type of the badge", getAutocompletePath(AutocompletePathTypeSuggestions), true)
	badge.AddNamedStaticListArgument("multiple", "Whether the badge can be granted multiple times", true, []model.AutocompleteListItem{
		{Item: TrueString},
		{Item: FalseString},
	})
	create.AddCommand(badge)

	createType := model.NewAutocompleteData(
		"type",
		"--name typeName --everyoneCanCreate true|false --everyoneCanGrant true|false",
		"Create a badge type",
	)
	createType.AddNamedTextArgument("name", "Name of the type", "--name typeName", "", true)
	createType.AddNamedStaticListArgument("everyoneCanCreate", "Whether the badge can be granted by everyone", true, []model.AutocompleteListItem{
		{Item: TrueString},
		{Item: FalseString},
	})
	createType.AddNamedStaticListArgument("everyoneCanGrant", "Whether the badge can be created by everyone", true, []model.AutocompleteListItem{
		{Item: TrueString},
		{Item: FalseString},
	})
	create.AddCommand(createType)

	badges.AddCommand(create)
	return badges
}

func getAutocompletePath(path string) string {
	return "plugins/" + manifest.Id + AutocompletePath + path
}
