package uos

import (
	"bytes"
	"html/template"
	"net/http"
)

// DialogSpec describes the interface every web application dialog must implement.
type DialogSpec interface {
	// Name returns the short name of the dialog. The form is available at '/dialogs/<name>'.
	Name() string
	// DisplayTitle returns the title string of the dialog.
	DisplayTitle() string
	// Footer returns a dialog footer specification.
	Footer() DialogFooter
}

// DialogFooter speficies the footer of a dialog.
type DialogFooter struct {
	// left-aligned footer elements
	Left []DialogFooterElement
	// right-aligned footer elements
	Right []DialogFooterElement
}

// DialogFooterElement specifies an element shown in a dialog footer.
// Use the DialogFooter* functions to create elements.
type DialogFooterElement struct {
	// shown as button
	IsButton bool
	// close dialog
	IsClosing bool
	// button submits a form
	IsSaving bool

	// text (button or text label)
	Text string
	// class (button or text label)
	TextClass string

	// referenced form (in combination with IsSaving)
	Form string
}

// DialogFooterButton returns a button specification.
func DialogFooterButton(text, class string) DialogFooterElement {
	return DialogFooterElement{
		IsButton:  true,
		Text:      text,
		TextClass: class,
	}
}

// DialogFooterPrimaryButton returns a primary button specification.
func DialogFooterPrimaryButton(text string) DialogFooterElement {
	return DialogFooterButton(text, "primary")
}

// DialogFooterText returns a text label.
func DialogFooterText(text, class string) DialogFooterElement {
	return DialogFooterElement{
		Text:      text,
		TextClass: class,
	}
}

// Closing sets the IsClosing field and returns the element.
func (fe DialogFooterElement) Closing() DialogFooterElement {
	fe.IsClosing = true
	return fe
}

// Saving sets the IsSaving field and returns the element.
func (fe DialogFooterElement) Saving(form string) DialogFooterElement {
	fe.IsSaving = true
	fe.Form = form
	return fe
}

// DialogHandler returns a handler for the "/dialogs/" route providing the specified dialogs.
// The handler can be activated using RegisterAppRequestHandlers.
func DialogHandler(dialogs ...DialogSpec) AppRequestHandlerMapping {
	return AppRequestHandlerMapping{
		Route:   "/dialogs/",
		Handler: getDialogHandlerFunc(dialogs),
	}
}

func getDialogHandlerFunc(dialogs []DialogSpec) AppRequestHandler {
	nameToSpec := map[string]DialogSpec{}
	for _, d := range dialogs {
		nameToSpec[d.Name()] = d
		Log.DebugContext("register dialog spec", LogContext{"name": d.Name()})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// determine dialog
		dialogName := getElementName("dialogs", r.URL.Path)
		Log.DebugContextR(
			r, "handle dialog",
			LogContext{
				"name":   dialogName,
				"method": r.Method,
			},
		)

		dialogSpec, ok := nameToSpec[dialogName]
		if !ok {
			RespondNotFound(w)
			return
		}

		// process request
		switch r.Method {
		case http.MethodGet:
			renderDialog(w, r, dialogName, dialogSpec, r.Form.Get("id"))
		default:
			RespondNotImplemented(w)
		}
	}
}

func renderDialog(w http.ResponseWriter, r *http.Request, name string, spec DialogSpec, id string) {
	dialogTemplateName := "dialog_" + name

	// render dialog content
	var content bytes.Buffer
	err := renderTemplate(&content, r, name, map[string]string{"ID": id}, dialogTemplateName)
	if err != nil {
		Log.ErrorContextR(
			r, "could not render page content template",
			LogContext{"name": name, "error": err},
		)
		RespondInternalServerError(w)
		return
	}

	// integrate content in dialog
	data := map[string]interface{}{
		"Title": spec.DisplayTitle(),

		"Content": template.HTML(content.String()),

		"FooterLeft":  spec.Footer().Left,
		"FooterRight": spec.Footer().Right,
	}

	err = renderInternalTemplate(w, r, "dialog", data)
	if err != nil {
		Log.ErrorContextR(
			r, "could not render dialog",
			LogContext{"name": name, "error": err},
		)
		RespondInternalServerError(w)
		return
	}
}
