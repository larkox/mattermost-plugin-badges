package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/larkox/mattermost-plugin-badges/badgesmodel"
	"github.com/mattermost/mattermost-server/v5/model"
)

// HTTPHandlerFuncWithUser is http.HandleFunc but userID is already exported
type HTTPHandlerFuncWithUser func(w http.ResponseWriter, r *http.Request, userID string)

// ResponseType indicates type of response returned by api
type ResponseType string

const (
	// ResponseTypeJSON indicates that response type is json
	ResponseTypeJSON ResponseType = "JSON_RESPONSE"
	// ResponseTypePlain indicates that response type is text plain
	ResponseTypePlain ResponseType = "TEXT_RESPONSE"
	// ResponseTypeDialog indicates that response type is a dialog response
	ResponseTypeDialog ResponseType = "DIALOG"
)

type APIErrorResponse struct {
	ID         string `json:"id"`
	Message    string `json:"message"`
	StatusCode int    `json:"status_code"`
}

func (p *Plugin) initializeAPI() {
	p.router = mux.NewRouter()
	p.router.Use(p.withRecovery)

	apiRouter := p.router.PathPrefix("/api/v1").Subrouter()
	pluginAPIRouter := p.router.PathPrefix(badgesmodel.PluginAPIPath).Subrouter()
	autocompleteRouter := p.router.PathPrefix(AutocompletePath).Subrouter()
	dialogRouter := p.router.PathPrefix(DialogPath).Subrouter()

	apiRouter.HandleFunc("/getUserBadges/{userID}", p.extractUserMiddleWare(p.getUserBadges, ResponseTypeJSON)).Methods(http.MethodGet)
	apiRouter.HandleFunc("/getBadgeDetails/{badgeID}", p.extractUserMiddleWare(p.getBadgeDetails, ResponseTypeJSON)).Methods(http.MethodGet)
	apiRouter.HandleFunc("/getAllBadges", p.extractUserMiddleWare(p.getAllBadges, ResponseTypeJSON)).Methods(http.MethodGet)

	pluginAPIRouter.HandleFunc(badgesmodel.PluginAPIPathEnsure, checkPluginRequest(p.ensureBadges)).Methods(http.MethodPost)
	pluginAPIRouter.HandleFunc(badgesmodel.PluginAPIPathGrant, checkPluginRequest(p.grantBadge)).Methods(http.MethodPost)

	autocompleteRouter.HandleFunc(AutocompletePathBadgeSuggestions, p.extractUserMiddleWare(p.getBadgeSuggestions, ResponseTypeJSON)).Methods(http.MethodGet)
	autocompleteRouter.HandleFunc(AutocompletePathEditBadgeSuggestions, p.extractUserMiddleWare(p.getEditBadgeSuggestions, ResponseTypeJSON)).Methods(http.MethodGet)
	autocompleteRouter.HandleFunc(AutocompletePathTypeSuggestions, p.extractUserMiddleWare(p.getBadgeTypeSuggestions, ResponseTypeJSON)).Methods(http.MethodGet)
	autocompleteRouter.HandleFunc(AutocompletePathEditTypeSuggestions, p.extractUserMiddleWare(p.getEditBadgeTypeSuggestions, ResponseTypeJSON)).Methods(http.MethodGet)

	dialogRouter.HandleFunc(DialogPathCreateBadge, p.extractUserMiddleWare(p.dialogCreateBadge, ResponseTypeDialog)).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathCreateType, p.extractUserMiddleWare(p.dialogCreateType, ResponseTypeDialog)).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathGrant, p.extractUserMiddleWare(p.dialogGrant, ResponseTypeDialog)).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathSelectBadge, p.extractUserMiddleWare(p.dialogSelectBadge, ResponseTypeDialog)).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathSelectType, p.extractUserMiddleWare(p.dialogSelectType, ResponseTypeDialog)).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathEditBadge, p.extractUserMiddleWare(p.dialogEditBadge, ResponseTypeDialog)).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathEditType, p.extractUserMiddleWare(p.dialogEditType, ResponseTypeDialog)).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathCreateSubscription, p.extractUserMiddleWare(p.dialogCreateSubscription, ResponseTypeDialog)).Methods(http.MethodPost)
	dialogRouter.HandleFunc(DialogPathDeleteSubscription, p.extractUserMiddleWare(p.dialogDeleteSubscription, ResponseTypeDialog)).Methods(http.MethodPost)

	p.router.PathPrefix("/").HandlerFunc(p.defaultHandler)
}

func (p *Plugin) defaultHandler(w http.ResponseWriter, r *http.Request) {
	p.mm.Log.Debug("Unexpected call", "url", r.URL)
	w.WriteHeader(http.StatusNotFound)
}

func dialogError(w http.ResponseWriter, text string, errors map[string]string) {
	resp := &model.SubmitDialogResponse{
		Error:  "Error: " + text,
		Errors: errors,
	}
	_, _ = w.Write(resp.ToJson())
}

func dialogOK(w http.ResponseWriter) {
	resp := &model.SubmitDialogResponse{}
	_, _ = w.Write(resp.ToJson())
}

func dialogKeepOpen(w http.ResponseWriter) {
	resp := &model.SubmitDialogResponse{
		Error: "_",
	}
	_, _ = w.Write(resp.ToJson())
}

func (p *Plugin) dialogCreateBadge(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}

	toCreate := badgesmodel.Badge{}
	toCreate.CreatedBy = userID
	toCreate.ImageType = badgesmodel.ImageTypeEmoji
	name, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeName)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}
	toCreate.Name = name

	description, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeDescription)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}
	toCreate.Description = description

	image, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeImage)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	if image[0] == ':' {
		image = image[1 : len(image)-1]
	}
	if image == "" {
		dialogError(w, "Invalid field", map[string]string{"image": "Empty emoji"})
		return
	}
	toCreate.Image = image

	badgeTypeStr, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeType)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	badgeType, err := strconv.Atoi(badgeTypeStr)
	if err != nil {
		dialogError(w, "Invalid field", map[string]string{"type": "Invalid type"})
		return
	}
	toCreate.Type = badgesmodel.BadgeType(badgeType)

	toCreate.Multiple = getDialogSubmissionBoolField(req, DialogFieldBadgeMultiple)

	_, err = p.store.AddBadge(toCreate)
	if err != nil {
		dialogError(w, err.Error(), nil)
		return
	}

	p.mm.Post.SendEphemeralPost(userID, &model.Post{
		UserId:    p.BotUserID,
		ChannelId: req.ChannelId,
		Message:   fmt.Sprintf("Badge `%s` created.", toCreate.Name),
	})

	dialogOK(w)
}

func (p *Plugin) dialogCreateType(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}
	toCreate := badgesmodel.BadgeTypeDefinition{}
	toCreate.CreatedBy = userID
	toCreate.CanCreate.Everyone = getDialogSubmissionBoolField(req, DialogFieldTypeEveryoneCanCreate)
	toCreate.CanGrant.Everyone = getDialogSubmissionBoolField(req, DialogFieldTypeEveryoneCanGrant)
	name, errText, errors := getDialogSubmissionTextField(req, DialogFieldTypeName)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}
	toCreate.Name = name

	createAllowList, _ := req.Submission[DialogFieldTypeAllowlistCanCreate].(string)
	grantAllowList, _ := req.Submission[DialogFieldTypeAllowlistCanGrant].(string)

	if createAllowList != "" {
		toCreate.CanCreate.AllowList = map[string]bool{}
		usernames := strings.Split(createAllowList, ",")
		for _, username := range usernames {
			username = strings.TrimSpace(username)
			if username == "" {
				continue
			}
			u, err := p.mm.User.GetByUsername(username)
			if err != nil {
				dialogError(w, "Cannot find user", map[string]string{DialogFieldTypeAllowlistCanCreate: fmt.Sprintf("Error getting user %s. Error: %v", username, err)})
				return
			}
			toCreate.CanCreate.AllowList[u.Id] = true
		}
	}

	if grantAllowList != "" {
		toCreate.CanCreate.AllowList = map[string]bool{}
		usernames := strings.Split(createAllowList, ",")
		for _, username := range usernames {
			username = strings.TrimSpace(username)
			if username == "" {
				continue
			}
			u, err := p.mm.User.GetByUsername(username)
			if err != nil {
				dialogError(w, "Cannot find user", map[string]string{DialogFieldTypeAllowlistCanGrant: fmt.Sprintf("Error getting user %s. Error: %v", username, err)})
				return
			}
			toCreate.CanGrant.AllowList[u.Id] = true
		}
	}

	_, err := p.store.AddType(toCreate)
	if err != nil {
		dialogError(w, err.Error(), nil)
		return
	}

	p.mm.Post.SendEphemeralPost(userID, &model.Post{
		UserId:    p.BotUserID,
		ChannelId: req.ChannelId,
		Message:   fmt.Sprintf("Type `%s` created.", toCreate.Name),
	})

	dialogOK(w)
}

// This is not working on the current webapp architecture. A similar approach should be handled using Apps.
func (p *Plugin) dialogSelectType(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}

	badgeTypeStr, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeType)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	badgeType, err := strconv.Atoi(badgeTypeStr)
	if err != nil {
		dialogError(w, "Invalid field", map[string]string{DialogFieldBadgeType: "Invalid type"})
		return
	}

	t, err := p.store.GetType(badgesmodel.BadgeType(badgeType))
	if err != nil {
		dialogError(w, "Cannot get type", map[string]string{DialogFieldBadgeType: "cannot get type"})
		return
	}

	u, err := p.mm.User.Get(userID)
	if err != nil {
		dialogError(w, "Cannot find user", nil)
		return
	}

	if !canEditType(*u, *t) {
		dialogError(w, "You cannot edit this type", nil)
		return
	}

	_, _ = p.mm.SlashCommand.Execute(&model.CommandArgs{
		UserId:    userID,
		ChannelId: req.ChannelId,
		TeamId:    req.TeamId,
		Command:   "/badges edit type --type " + badgeTypeStr,
	})

	dialogKeepOpen(w)
}

func (p *Plugin) dialogEditType(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}

	u, err := p.mm.User.Get(userID)
	if err != nil {
		dialogError(w, "Cannot find user", nil)
		return
	}

	originalTypeID, err := strconv.Atoi(req.State)
	if err != nil {
		dialogError(w, "type id not found", nil)
		return
	}

	originalType, err := p.store.GetType(badgesmodel.BadgeType(originalTypeID))
	if err != nil {
		dialogError(w, "could not get the type", nil)
		return
	}

	if !canEditType(*u, *originalType) {
		dialogError(w, "you have no permissions to edit this type", nil)
		return
	}

	if getDialogSubmissionBoolField(req, DialogFieldTypeDelete) {
		err = p.store.DeleteType(badgesmodel.BadgeType(originalTypeID))
		if err != nil {
			dialogError(w, err.Error(), nil)
		}
		return
	}
	originalType.CanCreate.Everyone = getDialogSubmissionBoolField(req, DialogFieldTypeEveryoneCanCreate)
	originalType.CanGrant.Everyone = getDialogSubmissionBoolField(req, DialogFieldTypeEveryoneCanGrant)
	name, errText, errors := getDialogSubmissionTextField(req, DialogFieldTypeName)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}
	originalType.Name = name

	createAllowList, _ := req.Submission[DialogFieldTypeAllowlistCanCreate].(string)
	grantAllowList, _ := req.Submission[DialogFieldTypeAllowlistCanGrant].(string)

	if createAllowList != "" {
		originalType.CanCreate.AllowList = map[string]bool{}
		usernames := strings.Split(createAllowList, ",")
		for _, username := range usernames {
			username = strings.TrimSpace(username)
			if username == "" {
				continue
			}
			var allowedUser *model.User
			allowedUser, err = p.mm.User.GetByUsername(username)
			if err != nil {
				dialogError(w, "Cannot find user", map[string]string{DialogFieldTypeAllowlistCanCreate: fmt.Sprintf("Error getting user %s. Error: %v", username, err)})
				return
			}
			originalType.CanCreate.AllowList[allowedUser.Id] = true
		}
	}

	if grantAllowList != "" {
		originalType.CanGrant.AllowList = map[string]bool{}
		usernames := strings.Split(createAllowList, ",")
		for _, username := range usernames {
			username = strings.TrimSpace(username)
			if username == "" {
				continue
			}
			var allowedUser *model.User
			allowedUser, err = p.mm.User.GetByUsername(username)
			if err != nil {
				dialogError(w, "Cannot find user", map[string]string{DialogFieldTypeAllowlistCanGrant: fmt.Sprintf("Error getting user %s. Error: %v", username, err)})
				return
			}
			originalType.CanGrant.AllowList[allowedUser.Id] = true
		}
	}

	err = p.store.UpdateType(*originalType)
	if err != nil {
		dialogError(w, err.Error(), nil)
		return
	}

	p.mm.Post.SendEphemeralPost(userID, &model.Post{
		UserId:    p.BotUserID,
		ChannelId: req.ChannelId,
		Message:   fmt.Sprintf("Type `%s` updated.", originalType.Name),
	})

	dialogOK(w)
}

// This is not working on the current webapp architecture. A similar approach should be handled using Apps.
func (p *Plugin) dialogSelectBadge(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}

	badgeIDStr, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadge)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	badgeID, err := strconv.Atoi(badgeIDStr)
	if err != nil {
		dialogError(w, "Invalid field", map[string]string{DialogFieldBadge: "Invalid badge"})
		return
	}

	b, err := p.store.GetBadge(badgesmodel.BadgeID(badgeID))
	if err != nil {
		dialogError(w, "Cannot get type", map[string]string{DialogFieldBadge: "cannot get badge"})
		return
	}

	u, err := p.mm.User.Get(userID)
	if err != nil {
		dialogError(w, "Cannot find user", nil)
		return
	}

	if !canEditBadge(*u, *b) {
		dialogError(w, "You cannot edit this badge", nil)
		return
	}

	_, _ = p.mm.SlashCommand.Execute(&model.CommandArgs{
		UserId:    userID,
		ChannelId: req.ChannelId,
		TeamId:    req.TeamId,
		Command:   "/badges edit badge --id " + badgeIDStr,
	})

	dialogKeepOpen(w)
}

func (p *Plugin) dialogEditBadge(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}

	u, err := p.mm.User.Get(userID)
	if err != nil {
		dialogError(w, "Cannot find user", nil)
		return
	}

	originalBadgeID, err := strconv.Atoi(req.State)
	if err != nil {
		dialogError(w, "badge id not found", nil)
		return
	}

	originalBadge, err := p.store.GetBadge(badgesmodel.BadgeID(originalBadgeID))
	if err != nil {
		dialogError(w, "could not get the badge", nil)
		return
	}

	if !canEditBadge(*u, *originalBadge) {
		dialogError(w, "you have no permissions to edit this type", nil)
		return
	}

	if getDialogSubmissionBoolField(req, DialogFieldBadgeDelete) {
		err = p.store.DeleteBadge(badgesmodel.BadgeID(originalBadgeID))
		if err != nil {
			dialogError(w, err.Error(), nil)
			return
		}
		return
	}
	name, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeName)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}
	originalBadge.Name = name

	description, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeDescription)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}
	originalBadge.Description = description

	image, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeImage)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	if image[0] == ':' {
		image = image[1 : len(image)-1]
	}
	if image == "" {
		dialogError(w, "Invalid field", map[string]string{"image": "Empty emoji"})
		return
	}
	originalBadge.Image = image

	badgeTypeStr, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeType)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	badgeType, err := strconv.Atoi(badgeTypeStr)
	if err != nil {
		dialogError(w, "Invalid field", map[string]string{"type": "Invalid type"})
		return
	}
	originalBadge.Type = badgesmodel.BadgeType(badgeType)

	originalBadge.Multiple = getDialogSubmissionBoolField(req, DialogFieldBadgeMultiple)

	err = p.store.UpdateBadge(*originalBadge)
	if err != nil {
		dialogError(w, err.Error(), nil)
		return
	}

	p.mm.Post.SendEphemeralPost(userID, &model.Post{
		UserId:    p.BotUserID,
		ChannelId: req.ChannelId,
		Message:   fmt.Sprintf("Badge `%s` updated.", originalBadge.Name),
	})

	dialogOK(w)
}

func (p *Plugin) dialogGrant(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}

	badgeIDStr, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadge)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	notifyHere := getDialogSubmissionBoolField(req, DialogFieldNotifyHere)

	badgeID, err := strconv.Atoi(badgeIDStr)
	if err != nil {
		dialogError(w, "Invalid field", map[string]string{DialogFieldBadge: "Invalid badge"})
		return
	}

	badge, err := p.store.GetBadgeDetails(badgesmodel.BadgeID(badgeID))
	if err != nil {
		dialogError(w, "badge not found", nil)
		return
	}

	grantToID := req.State
	if grantToID == "" {
		grantToID, errText, errors = getDialogSubmissionTextField(req, DialogFieldUser)
		if errors != nil {
			dialogError(w, errText, errors)
			return
		}
	}

	grantToUser, err := p.mm.User.Get(grantToID)
	if err != nil {
		dialogError(w, "user not found", nil)
		return
	}

	reason, _ := req.Submission[DialogFieldGrantReason].(string)

	shouldNotify, err := p.store.GrantBadge(badgesmodel.BadgeID(badgeID), grantToID, userID, reason)
	if err != nil {
		p.writeAPIError(w, &APIErrorResponse{
			ID:         "cannot grant badge",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	if shouldNotify {
		p.notifyGrant(badgesmodel.BadgeID(badgeID), userID, grantToUser, notifyHere, req.ChannelId, reason)
	}

	p.mm.Post.SendEphemeralPost(userID, &model.Post{
		UserId:    p.BotUserID,
		ChannelId: req.ChannelId,
		Message:   fmt.Sprintf("Badge `%s` granted to @%s.", badge.Name, grantToUser.Username),
	})

	dialogOK(w)
}

func (p *Plugin) dialogCreateSubscription(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}

	u, err := p.mm.User.Get(userID)
	if err != nil {
		dialogError(w, err.Error(), nil)
		return
	}

	if !canCreateSubscription(*u, req.ChannelId) {
		dialogError(w, "You cannot create a subscription", nil)
		return
	}

	typeIDStr, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeType)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		dialogError(w, "Invalid field", map[string]string{DialogFieldBadgeType: "Invalid type"})
		return
	}

	err = p.store.AddSubscription(badgesmodel.BadgeType(typeID), req.ChannelId)
	if err != nil {
		dialogError(w, err.Error(), nil)
	}

	p.mm.Post.SendEphemeralPost(userID, &model.Post{
		UserId:    p.BotUserID,
		ChannelId: req.ChannelId,
		Message:   "Subscription added",
	})

	dialogOK(w)
}

func (p *Plugin) dialogDeleteSubscription(w http.ResponseWriter, r *http.Request, userID string) {
	req := model.SubmitDialogRequestFromJson(r.Body)
	if req == nil {
		dialogError(w, "could not get the dialog request", nil)
		return
	}

	u, err := p.mm.User.Get(userID)
	if err != nil {
		dialogError(w, err.Error(), nil)
		return
	}

	if !canCreateSubscription(*u, req.ChannelId) {
		dialogError(w, "You cannot delete a subscription", nil)
		return
	}

	typeIDStr, errText, errors := getDialogSubmissionTextField(req, DialogFieldBadgeType)
	if errors != nil {
		dialogError(w, errText, errors)
		return
	}

	typeID, err := strconv.Atoi(typeIDStr)
	if err != nil {
		dialogError(w, "Invalid field", map[string]string{DialogFieldBadgeType: "Invalid type"})
		return
	}

	err = p.store.RemoveSubscriptions(badgesmodel.BadgeType(typeID), req.ChannelId)
	if err != nil {
		dialogError(w, err.Error(), nil)
	}

	p.mm.Post.SendEphemeralPost(userID, &model.Post{
		UserId:    p.BotUserID,
		ChannelId: req.ChannelId,
		Message:   "Subscription removed",
	})

	dialogOK(w)
}

func getDialogSubmissionTextField(req *model.SubmitDialogRequest, fieldName string) (value string, errText string, errors map[string]string) {
	value, ok := req.Submission[fieldName].(string)
	value = strings.TrimSpace(value)
	if !ok || value == "" {
		return "", "Invalid argument", map[string]string{fieldName: "Field empty or not recognized."}
	}

	return value, "", nil
}

func getDialogSubmissionBoolField(req *model.SubmitDialogRequest, fieldName string) bool {
	value, _ := req.Submission[fieldName].(bool)
	return value
}

func (p *Plugin) grantBadge(w http.ResponseWriter, r *http.Request, pluginID string) {
	var req *badgesmodel.GrantBadgeRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		p.writeAPIError(w, &APIErrorResponse{
			ID:         "cannot unmarshal request",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
	p.mm.Log.Debug("Granting badge", "req", req)

	if req == nil {
		p.writeAPIError(w, &APIErrorResponse{
			ID:         "missing request",
			Message:    "Missing grant request on request body",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	shouldNotify, err := p.store.GrantBadge(req.BadgeID, req.UserID, req.BotID, req.Reason)
	if err != nil {
		p.writeAPIError(w, &APIErrorResponse{
			ID:         "cannot grant badge",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
	if shouldNotify {
		u, err := p.mm.User.Get(req.UserID)
		if err == nil {
			p.notifyGrant(req.BadgeID, req.BotID, u, false, "", req.Reason)
		}
	}

	_, _ = w.Write([]byte("OK"))
}

func (p *Plugin) ensureBadges(w http.ResponseWriter, r *http.Request, pluginID string) {
	var req *badgesmodel.EnsureBadgesRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		p.writeAPIError(w, &APIErrorResponse{
			ID:         "cannot unmarshal request",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}
	if req == nil {
		p.writeAPIError(w, &APIErrorResponse{
			ID:         "missing request",
			Message:    "Missing ensure request on request body",
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	badges, err := p.store.EnsureBadges(req.Badges, pluginID, req.BotID)
	if err != nil {
		p.writeAPIError(w, &APIErrorResponse{
			ID:         "cannot ensure",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	b, err := json.Marshal(badges)
	if err != nil {
		p.writeAPIError(w, &APIErrorResponse{
			ID:         "cannot marshal",
			Message:    err.Error(),
			StatusCode: http.StatusInternalServerError,
		})
		return
	}

	_, _ = w.Write(b)
}

func (p *Plugin) getBadgeSuggestions(w http.ResponseWriter, r *http.Request, actingUserID string) {
	out := []model.AutocompleteListItem{}
	u, err := p.mm.User.Get(actingUserID)
	if err != nil {
		p.mm.Log.Debug("Error getting user", "error", err)
		_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	bb, err := p.store.GetGrantSuggestions(*u)
	if err != nil {
		p.mm.Log.Debug("Error getting suggestions", "error", err)
		_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	for _, b := range bb {
		s := model.AutocompleteListItem{
			Item:     strconv.Itoa(int(b.ID)),
			Hint:     b.Name,
			HelpText: b.Description,
		}

		out = append(out, s)
	}
	_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
}

func (p *Plugin) getEditBadgeSuggestions(w http.ResponseWriter, r *http.Request, actingUserID string) {
	out := []model.AutocompleteListItem{}
	u, err := p.mm.User.Get(actingUserID)
	if err != nil {
		p.mm.Log.Debug("Error getting user", "error", err)
		_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	bb, err := p.store.GetEditBadgeSuggestions(*u)
	if err != nil {
		p.mm.Log.Debug("Error getting suggestions", "error", err)
		_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	for _, b := range bb {
		s := model.AutocompleteListItem{
			Item:     strconv.Itoa(int(b.ID)),
			Hint:     b.Name,
			HelpText: b.Description,
		}

		out = append(out, s)
	}
	_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
}

func (p *Plugin) getBadgeTypeSuggestions(w http.ResponseWriter, r *http.Request, actingUserID string) {
	out := []model.AutocompleteListItem{}
	u, err := p.mm.User.Get(actingUserID)
	if err != nil {
		p.mm.Log.Debug("Error getting user", "error", err)
		_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	types, err := p.store.GetTypeSuggestions(*u)
	if err != nil {
		p.mm.Log.Debug("Error getting suggestions", "error", err)
		_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	for _, t := range types {
		s := model.AutocompleteListItem{
			Item: strconv.Itoa(int(t.ID)),
			Hint: t.Name,
		}

		out = append(out, s)
	}
	_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
}

func (p *Plugin) getEditBadgeTypeSuggestions(w http.ResponseWriter, r *http.Request, actingUserID string) {
	out := []model.AutocompleteListItem{}
	u, err := p.mm.User.Get(actingUserID)
	if err != nil {
		p.mm.Log.Debug("Error getting user", "error", err)
		_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	types, err := p.store.GetEditTypeSuggestions(*u)
	if err != nil {
		p.mm.Log.Debug("Error getting suggestions", "error", err)
		_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	for _, t := range types {
		s := model.AutocompleteListItem{
			Item: strconv.Itoa(int(t.ID)),
			Hint: t.Name,
		}

		out = append(out, s)
	}
	_, _ = w.Write(model.AutocompleteStaticListItemsToJSON(out))
}

func (p *Plugin) getUserBadges(w http.ResponseWriter, r *http.Request, actingUserID string) {
	userID, ok := mux.Vars(r)["userID"]
	if !ok {
		userID = actingUserID
	}

	badges, err := p.store.GetUserBadges(userID)
	if err != nil {
		p.mm.Log.Debug("Error getting the badges for user", "error", err, "user", userID)
	}

	b, _ := json.Marshal(badges)
	_, _ = w.Write(b)
}

func (p *Plugin) getBadgeDetails(w http.ResponseWriter, r *http.Request, actingUserID string) {
	badgeIDString, ok := mux.Vars(r)["badgeID"]
	if !ok {
		errMessage := "Missing badge id"
		p.mm.Log.Debug(errMessage)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(errMessage))
		return
	}

	badgeIDNumber, err := strconv.Atoi(badgeIDString)
	if err != nil {
		errMessage := "Cannot convert badgeID to number"
		p.mm.Log.Debug(errMessage, "badgeID", badgeIDString, "err", err)
		_, _ = w.Write([]byte(errMessage))
		return
	}

	badgeID := badgesmodel.BadgeID(badgeIDNumber)

	badge, err := p.store.GetBadgeDetails(badgeID)
	if err != nil {
		p.mm.Log.Debug("Cannot get badge details", "badgeID", badgeID, "error", err)
	}

	b, _ := json.Marshal(badge)
	_, _ = w.Write(b)
}

func (p *Plugin) getAllBadges(w http.ResponseWriter, r *http.Request, actingUserID string) {
	badge, err := p.store.GetAllBadges()
	if err != nil {
		p.mm.Log.Debug("Cannot get all badges", "error", err)
	}

	b, _ := json.Marshal(badge)
	_, _ = w.Write(b)
}

func (p *Plugin) extractUserMiddleWare(handler HTTPHandlerFuncWithUser, responseType ResponseType) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.Header.Get("Mattermost-User-ID")
		if userID == "" {
			switch responseType {
			case ResponseTypeJSON:
				p.writeAPIError(w, &APIErrorResponse{ID: "", Message: "Not authorized.", StatusCode: http.StatusUnauthorized})
			case ResponseTypePlain:
				http.Error(w, "Not authorized", http.StatusUnauthorized)
			case ResponseTypeDialog:
				dialogError(w, "Not Authorized", nil)
			default:
				p.mm.Log.Error("Unknown ResponseType detected")
			}
			return
		}

		handler(w, r, userID)
	}
}

func (p *Plugin) withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if x := recover(); x != nil {
				p.mm.Log.Error("Recovered from a panic",
					"url", r.URL.String(),
					"error", x,
					"stack", string(debug.Stack()))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func checkPluginRequest(next HTTPHandlerFuncWithUser) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// All other plugins are allowed
		pluginID := r.Header.Get("Mattermost-Plugin-ID")
		if pluginID == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		next(w, r, pluginID)
	}
}

func (p *Plugin) writeAPIError(w http.ResponseWriter, apiErr *APIErrorResponse) {
	b, err := json.Marshal(apiErr)
	if err != nil {
		p.mm.Log.Warn("Failed to marshal API error", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(apiErr.StatusCode)

	_, err = w.Write(b)
	if err != nil {
		p.mm.Log.Warn("Failed to write JSON response", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (p *Plugin) getPluginURL() string {
	urlP := p.mm.Configuration.GetConfig().ServiceSettings.SiteURL
	url := "/"
	if urlP != nil {
		url = *urlP
	}
	if url[len(url)-1] == '/' {
		url = url[0 : len(url)-1]
	}
	return url + "/plugins/" + manifest.Id
}

func (p *Plugin) getDialogURL() string {
	return p.getPluginURL() + DialogPath
}
