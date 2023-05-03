package uos

import (
	"net/http"
)

// ResponseAction describes what happens on a successful save or delete form action.
// Use the FormResponse* functions to create a response action.
type ResponseAction struct {
	doPageRefresh bool
	doCloseDialog bool

	isFormError bool

	message      string
	messageClass string

	callback func(http.ResponseWriter)

	redirect string
}

// ResponseRefresh triggers a full frontend page refresh.
func ResponseRefresh() *ResponseAction {
	return &ResponseAction{doPageRefresh: true}
}

// ResponseCloseDialog closes an open modal dialog.
func ResponseCloseDialog() *ResponseAction {
	return &ResponseAction{doCloseDialog: true}
}

// ResponseMessage returns a message element.
func ResponseMessage(message, class string) *ResponseAction {
	return &ResponseAction{
		message:      message,
		messageClass: class,
	}
}

// ResponseFormError renders a form as response - including the specified error message.
func ResponseFormError(message string) *ResponseAction {
	return &ResponseAction{
		isFormError: true,
		message:     message,
	}
}

// ResponseSetSessionCookie sets a session cookie for the specified user and
// triggers a full frontend page refresh or a redirect to the given URL.
func ResponseSetSessionCookie(userID uint, language string) *ResponseAction {
	return &ResponseAction{
		doPageRefresh: true,
		callback: func(w http.ResponseWriter) {
			setSession(userID, w)
			setLanguage(language, w)
		},
	}
}

// ResponseClearSessionCookie clears the session cookie and triggers a full frontend page refresh.
func ResponseClearSessionCookie() *ResponseAction {
	return &ResponseAction{
		doPageRefresh: true,
		callback: func(w http.ResponseWriter) {
			clearSession(w)
		},
	}
}

func handleResponseAction(w http.ResponseWriter, r *http.Request, action *ResponseAction) {
	if action == nil {
		// return - do nothing
		return
	}

	if action.callback != nil {
		action.callback(w)
	}

	if action.doPageRefresh {
		if action.redirect != "" {
			w.Header().Add("HX-Redirect", action.redirect)
		} else {
			w.Header().Add("HX-Refresh", "true")
		}
		return
	}

	var (
		template string
		context  interface{}
	)

	switch {
	case action.doCloseDialog:
		template = "dialog_close"

	case action.message != "":
		template = "message"
		context = map[string]string{
			"Class":   action.messageClass,
			"Message": action.message,
		}
	}

	err := renderInternalTemplate(w, r, template, context)
	if err != nil {
		Log.ErrorContextR(
			r, "could not render form response action",
			LogContext{"name": template, "error": err},
		)
		RespondInternalServerError(w)
	}
}
