package uos

import (
	"net/http"
)

type logoutAction struct{}

func (a logoutAction) Name() string {
	return "logout"
}

func (a logoutAction) Do(w http.ResponseWriter, r *http.Request) *ResponseAction {
	return ResponseClearSessionCookie()
}

type languageAction struct{}

func (a languageAction) Name() string {
	return "setLanguage"
}

func (a languageAction) Do(w http.ResponseWriter, r *http.Request) *ResponseAction {
	language := r.Form.Get("lang")

	setLanguage(language, w)

	user, ok := r.Context().Value(ctxAppUser).(AppUser)
	if ok && user.Language != language {
		// update language selection for current user
		err := DB.Model(&user).Update("language", language).Error
		if err != nil {
			Log.ErrorObjR(r, "could not update user language selection", err)
		}
	}

	return ResponseRefresh()
}
