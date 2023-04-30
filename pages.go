package uos

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
)

// PageHandler returns a standard GET handler for a template based page.
//
// Called with a single parameter: specifiy the page template name, will be provided at "/<name>".
//
// Calles with two parameters: specify route and page name.
//
// Panics if no or more than two parameters are provided. The returned handler can be activated
// using RegisterAppRequestHandlers.
func PageHandler(page ...string) AppRequestHandlerMapping {
	if len(page) == 0 || len(page) > 2 {
		panic("PageHandler must be called with one or two parameters.")
	}

	var route, name string

	if len(page) == 1 {
		route = fmt.Sprintf("/%s", page[0])
		name = page[0]
	} else {
		route = page[0]
		name = page[1]
	}

	return AppRequestHandlerMapping{
		Route: route,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet || r.URL.Path != route {
				RespondNotFound(w)
				return
			}

			renderPage(
				w, r, name,
				map[string]interface{}{
					"Object": r.Form.Get("p"),
				},
			)
		},
	}
}

// renderPage loads the base page template and the content template with the specified name
// and writes the combined result to the specified writer using the given data context.
func renderPage(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	pageTemplateName := "page_" + name

	// render page content
	var content bytes.Buffer
	err := renderTemplate(&content, r, name, data, pageTemplateName)
	if err != nil {
		Log.ErrorContextR(
			r, "could not render page content template",
			LogContext{"name": name, "error": err},
		)
		RespondInternalServerError(w)
		return
	}

	// integrate content in base page
	if data == nil {
		data = map[string]interface{}{}
	}
	data["Content"] = template.HTML(content.String())
	data["Page"] = config.getPageConfig(name)

	err = renderInternalTemplate(w, r, "page", data)
	if err != nil {
		Log.ErrorContextR(
			r, "could not render page",
			LogContext{"name": name, "error": err},
		)
		RespondInternalServerError(w)
		return
	}
}
