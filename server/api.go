package main

import (
	"encoding/json"
	"net/http"
	"runtime/debug"
	"strconv"

	"github.com/gorilla/mux"
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
	autocompleteRouter := p.router.PathPrefix(AutocompletePath).Subrouter()

	apiRouter.HandleFunc("/getUserBadges/{userID}", p.extractUserMiddleWare(p.getUserBadges, ResponseTypeJSON)).Methods(http.MethodGet)
	apiRouter.HandleFunc("/getBadgeDetails/{badgeID}", p.extractUserMiddleWare(p.getBadgeDetails, ResponseTypeJSON)).Methods(http.MethodGet)
	apiRouter.HandleFunc("/getAllBadges", p.extractUserMiddleWare(p.getAllBadges, ResponseTypeJSON)).Methods(http.MethodGet)

	// apiRouter.HandleFunc("/config", checkPluginRequest(p.getConfig)).Methods(http.MethodGet)
	// apiRouter.HandleFunc("/token", checkPluginRequest(p.getToken)).Methods(http.MethodGet)

	autocompleteRouter.HandleFunc(AutocompletePathBadgeSuggestions, p.extractUserMiddleWare(p.getBadgeSuggestions, ResponseTypeJSON)).Methods(http.MethodGet)
	autocompleteRouter.HandleFunc(AutocompletePathTypeSuggestions, p.extractUserMiddleWare(p.getBadgeTypeSuggestions, ResponseTypeJSON)).Methods(http.MethodGet)

	p.router.PathPrefix("/").HandlerFunc(p.defaultHandler)
}

func (p *Plugin) defaultHandler(w http.ResponseWriter, r *http.Request) {
	p.API.LogDebug("Unexpected call", "url", r.URL)
	w.WriteHeader(http.StatusNotFound)
}

func (p *Plugin) getBadgeSuggestions(w http.ResponseWriter, r *http.Request, actingUserID string) {
	out := []model.AutocompleteListItem{}
	u, appErr := p.API.GetUser(actingUserID)
	if appErr != nil {
		p.API.LogDebug("Error getting user", "error", appErr)
		w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	bb, err := p.store.GetGrantSuggestions(*u)
	if err != nil {
		p.API.LogDebug("Error getting suggestions", "error", err)
		w.Write(model.AutocompleteStaticListItemsToJSON(out))
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
	w.Write(model.AutocompleteStaticListItemsToJSON(out))
}

func (p *Plugin) getBadgeTypeSuggestions(w http.ResponseWriter, r *http.Request, actingUserID string) {
	out := []model.AutocompleteListItem{}
	u, appErr := p.API.GetUser(actingUserID)
	if appErr != nil {
		p.API.LogDebug("Error getting user", "error", appErr)
		w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	types, err := p.store.GetTypeSuggestions(*u)
	if err != nil {
		p.API.LogDebug("Error getting suggestions", "error", err)
		w.Write(model.AutocompleteStaticListItemsToJSON(out))
		return
	}

	for _, t := range types {
		s := model.AutocompleteListItem{
			Item: strconv.Itoa(int(t.ID)),
			Hint: t.Name,
		}

		out = append(out, s)
	}
	w.Write(model.AutocompleteStaticListItemsToJSON(out))
}

func (p *Plugin) getUserBadges(w http.ResponseWriter, r *http.Request, actingUserID string) {
	userID, ok := mux.Vars(r)["userID"]
	if !ok {
		userID = actingUserID
	}

	badges, err := p.store.GetUserBadges(userID)
	if err != nil {
		p.API.LogDebug("Error getting the badges for user", "error", err, "user", userID)
	}

	b, _ := json.Marshal(badges)
	w.Write(b)
}

func (p *Plugin) getBadgeDetails(w http.ResponseWriter, r *http.Request, actingUserID string) {
	badgeIDString, ok := mux.Vars(r)["badgeID"]
	if !ok {
		errMessage := "Missing badge id"
		p.API.LogDebug(errMessage)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(errMessage))
		return
	}

	badgeIDNumber, err := strconv.Atoi(badgeIDString)
	if err != nil {
		errMessage := "Cannot convert badgeID to number"
		p.API.LogDebug(errMessage, "badgeID", badgeIDString, "err", err)
		w.Write([]byte(errMessage))
		return
	}

	badgeID := BadgeID(badgeIDNumber)

	badge, err := p.store.GetBadgeDetails(badgeID)
	if err != nil {
		p.API.LogDebug("Cannot get badge details", "badgeID", badgeID, "error", err)
	}

	b, _ := json.Marshal(badge)
	w.Write(b)
}

func (p *Plugin) getAllBadges(w http.ResponseWriter, r *http.Request, actingUserID string) {
	badge, err := p.store.GetAllBadges()
	if err != nil {
		p.API.LogDebug("Cannot get all badges", "error", err)
	}

	b, _ := json.Marshal(badge)
	w.Write(b)
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
			default:
				p.API.LogError("Unknown ResponseType detected")
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
				p.API.LogError("Recovered from a panic",
					"url", r.URL.String(),
					"error", x,
					"stack", string(debug.Stack()))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func checkPluginRequest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// All other plugins are allowed
		pluginID := r.Header.Get("Mattermost-Plugin-ID")
		if pluginID == "" {
			http.Error(w, "Not authorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

func (p *Plugin) writeAPIError(w http.ResponseWriter, apiErr *APIErrorResponse) {
	b, err := json.Marshal(apiErr)
	if err != nil {
		p.API.LogWarn("Failed to marshal API error", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(apiErr.StatusCode)

	_, err = w.Write(b)
	if err != nil {
		p.API.LogWarn("Failed to write JSON response", "error", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
