package uos

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// FormSpec describes the interface every web application form must implement.
type FormSpec interface {
	// Name returns the short name of the form. The form is available at '/forms/<name>'.
	Name() string
}

// FormSpecRead must be implemented by a web application form to support GET requests.
type FormSpecRead interface {
	// Read returns the list of form items. If an id is specified, the form items for an existing
	// entity is returned.
	Read(id string) (FormItems, error)
}

// FormSpecSave must be implemented by a web application form to support POST requests.
type FormSpecSave interface {
	// Read returns the list of form items. If an id is specified, the form items for an existing
	// entity is returned.
	Read(id string) (FormItems, error)
	// Save writes the specified items to the database.
	Save(id string, items FormItems) (*ResponseAction, error)
}

// FormSpecDelete must be implemented by a web application form to support DELETE requests.
type FormSpecDelete interface {
	// Delete removes the specified item from the database.
	Delete(id string) (*ResponseAction, error)
}

// FormItem describes a single form entry, e.g. an input box.
type FormItem struct {
	ID string

	InputType     string
	InputTypeHTML string

	Name         string
	Value        string
	DefaultValue string

	Constraints *FormItemConstraints

	Class       string
	Placeholder string
	Min         string
	Max         string

	Label string

	Help      string
	HelpClass string

	Message      string
	MessageClass string

	IsHorizontal bool
	IsHidden     bool
	HasFocus     bool
}

type FormItemConstraints struct {
	IsMandatory bool
	IsNumber    bool

	MinValue float64
	MaxValue float64

	MinLength int
	MaxLength int

	Regexp string
}

func (fi *FormItem) validate(value string) bool {
	if fi.Constraints == nil {
		return true
	}

	// check constraints ..

	// .. field required?
	if len(value) == 0 && fi.Constraints.IsMandatory {
		fi.Help = "required"
		fi.HelpClass = "is-danger"
		return false
	}

	// .. value length
	if len(value) < fi.Constraints.MinLength {
		fi.Help = fmt.Sprintf("too short - at least %d characters", fi.Constraints.MinLength)
		fi.HelpClass = "is-danger"
		return false
	}
	if fi.Constraints.MaxLength > 0 && len(value) > fi.Constraints.MaxLength {
		fi.Help = fmt.Sprintf("too long - at most %d characters", fi.Constraints.MaxLength)
		fi.HelpClass = "is-danger"
		return false
	}

	// .. number? -> check min/max
	if fi.Constraints.IsNumber {
		f, err := strconv.ParseFloat(value, 64)
		if err != nil {
			fi.Help = fmt.Sprint("not a number")
			fi.HelpClass = "is-danger"
			return false
		}

		if f < fi.Constraints.MinValue {
			fi.Help = fmt.Sprintf("value too small, must be >= %f", fi.Constraints.MinValue)
			fi.HelpClass = "is-danger"
			return false
		}
		if f > fi.Constraints.MaxValue {
			fi.Help = fmt.Sprintf("value too big, must be <= %f", fi.Constraints.MaxValue)
			fi.HelpClass = "is-danger"
			return false
		}
	}

	// TODO: check regular expression

	return true
}

// FormItems is a list of form items.
type FormItems []FormItem

func (fi *FormItems) setValues(v url.Values) bool {
	var isValid = true

	for i, item := range *fi {
		// get provided (URL) value
		value := strings.TrimSpace(v.Get(item.Name))
		if value == "" {
			value = item.DefaultValue
		}

		// validate item
		itemIsValid := item.validate(value)

		// update item ..
		// .. set value (independent of validity)
		item.Value = value
		// .. set focus on first invalid form item
		item.HasFocus = isValid && !itemIsValid

		(*fi)[i] = item

		// update overall validation result
		isValid = isValid && itemIsValid
	}

	return isValid
}

func (fi *FormItems) Get(name string) *FormItem {
	for _, item := range *fi {
		if item.Name == name {
			return &item
		}
	}

	return nil
}

// FormHandler returns a handler for the "/forms/" route providing the specified forms.
// The handler can be activated using RegisterAppRequestHandlers.
func FormHandler(forms ...FormSpec) AppRequestHandlerMapping {
	return AppRequestHandlerMapping{
		Route:   "/forms/",
		Handler: getFormsHandlerFunc(forms),
	}
}

func getFormsHandlerFunc(forms []FormSpec) AppRequestHandler {
	nameToSpec := map[string]FormSpec{}
	for _, f := range forms {
		nameToSpec[f.Name()] = f
		Log.DebugContext("register form spec", LogContext{"name": f.Name()})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// determine form
		formName := getElementName("forms", r.URL.Path)
		Log.DebugContextR(
			r, "handle form",
			LogContext{
				"name":   formName,
				"method": r.Method,
			},
		)

		formSpec, ok := nameToSpec[formName]
		if !ok {
			RespondNotFound(w)
			return
		}

		// prepare request processing (URL form data might be empty)
		var (
			id           = r.Form.Get("id")
			submitButton = r.Form.Get("btn")
			csrf         = r.Form.Get("csrf")
		)

		// process request
		switch r.Method {
		case http.MethodGet:
			// does the form support GET method?
			formRead, ok := formSpec.(FormSpecRead)
			if !ok {
				RespondNotImplemented(w)
				return
			}

			items, err := formRead.Read(id)
			if err != nil {
				handleFormError(w, r, "could not read/initialize form", err)
				return
			}

			renderForm(w, r, formName, items, submitButton, "")
		case http.MethodPost:
			// does the form support POST method?
			formSave, ok := formSpec.(FormSpecSave)
			if !ok {
				RespondNotImplemented(w)
				return
			}

			// CSRF protection
			if !IsCSRFtokenValid(r, csrf) {
				Log.DebugR(r, "CSRF token mismatch")
				RespondBadRequest(w)
				return
			}

			// initialize (empty) form
			items, err := formSave.Read("")
			if err != nil {
				handleFormError(w, r, "could not initialize form", err)
				return
			}

			// integrate values form posted form data and validate
			isValid := items.setValues(r.Form)
			if isValid {
				action, err := formSave.Save(id, items)
				if err != nil {
					Log.ErrorObjR(r, "could not save form item", err)
					RespondInternalServerError(w)
					return
				}

				if action.isFormError {
					renderForm(w, r, formName, items, submitButton, action.message)
					return
				}

				action.doCloseDialog = r.Form.Get("dialog") == "true"
				action.redirect = r.Form.Get("ref")

				handleResponseAction(w, r, action)
				return
			}

			renderForm(w, r, formName, items, submitButton, "")
		case http.MethodDelete:
			// does the form support DELETE method?
			formDelete, ok := formSpec.(FormSpecDelete)
			if !ok {
				RespondNotImplemented(w)
				return
			}

			// CSRF protection
			if !IsCSRFtokenValid(r, csrf) {
				Log.DebugR(r, "CSRF token mismatch")
				RespondBadRequest(w)
				return
			}

			action, err := formDelete.Delete(id)
			if err != nil {
				handleFormError(w, r, "could not delete form item", err)
				return
			}

			action.doCloseDialog = r.Form.Get("dialog") == "true"
			handleResponseAction(w, r, action)
		default:
			RespondNotImplemented(w)
		}
	}
}

func handleFormError(w http.ResponseWriter, r *http.Request, message string, err error) {
	switch err {
	case ErrorFormItemNotFound:
		RespondNotFound(w)
		return
	case ErrorFormInvalidRequest:
		RespondBadRequest(w)
		return
	}

	// all other cases: log error and respond
	Log.ErrorObjR(r, message, err)
	RespondInternalServerError(w)
}

func renderForm(w http.ResponseWriter, r *http.Request, name string, form FormItems, submitButton, errorMessage string) {
	// initialize form context
	context := struct {
		ID     string
		Items  FormItems
		Button string
		Error  string
	}{name, form, submitButton, errorMessage}

	err := renderInternalTemplate(w, r, "form", context)
	if err != nil {
		Log.ErrorContextR(
			r, "could not render form",
			LogContext{"name": name, "error": err},
		)
		RespondInternalServerError(w)
	}
}
