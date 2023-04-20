package uos

import (
	"net/http"
)

// ActionSpec describes the interface every web application action must implement.
type ActionSpec interface {
	// Name returns the short name of the action. The action can be triggered at "/actions/<name>".
	Name() string
	// Do executes the action. The implementation must ensure logging and an appropriate response.
	// The returned action (optional) is executed afterwards.
	Do(http.ResponseWriter, *http.Request) *ResponseAction
}

// ActionHandler returns a handler for the "/actions/" route providing the specified actions.
// The handler can be activated using RegisterAppRequestHandlers.
func ActionHandler(actions ...ActionSpec) AppRequestHandlerMapping {
	return AppRequestHandlerMapping{
		Route:   "/actions/",
		Handler: getActionHandlerFunc(actions),
	}
}

func getActionHandlerFunc(actions []ActionSpec) AppRequestHandler {
	nameToSpec := map[string]ActionSpec{
		// pre-defined actions
		"logout": logoutAction{},
	}
	for _, a := range actions {
		nameToSpec[a.Name()] = a
		LogDebugContext("register action spec", LogContext{"name": a.Name()})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// determine action
		actionName := getElementName("actions", r.URL.Path)
		LogDebugContext(
			"handle action",
			LogContext{
				"name":   actionName,
				"method": r.Method,
			},
		)

		actionSpec, ok := nameToSpec[actionName]
		if !ok {
			RespondNotFound(w)
			return
		}

		// prepare request processing (URL form data might be empty)
		err := r.ParseForm()
		if err != nil {
			LogWarnError("could not parse form", err)
			RespondBadRequest(w)
			return
		}

		LogInfoContext("execute action", LogContext{"name": actionName, "method": r.Method})
		handleResponseAction(w, r, actionSpec.Do(w, r))
	}
}
