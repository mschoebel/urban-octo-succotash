package uos

import (
	"fmt"
	"net/http"
	"strings"
)

// ResourceSpec describes an interface a web application resource must provide.
type ResourceSpec interface {
	// Name returns the short name of a resource. The resource is available at '/<name>'.
	Name() string
}

// ResourceSpecRead describes an interface a resource with list access must provide.
type ResourceSpecList interface {
	// List return a filtered list of resources. Used as context to resource_list_<name> template.
	List(filter string, page, count int) ([]interface{}, error)
}

// ResourceSpecRead describes an interface a resource with ID access must provide.
type ResourceSpecRead interface {
	// Read returns the resource as template context for resource_<name>.
	Read(id string) (interface{}, error)
}

// ResourceHandler returns a request handler for resources.
// - GET /<name>
// - GET /<name>/<id>
//
// PUT/POST is not supported, assuming that resources will be created/modified via a form.
func ResourceHandler(resource ResourceSpec) AppRequestHandlerMapping {
	route := fmt.Sprintf("/%s/", resource.Name())

	return AppRequestHandlerMapping{
		Route: route,
		Handler: func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				RespondNotImplemented(w)
				return
			}

			var (
				templateName string      = resource.Name()
				context      interface{} = nil

				err error
			)

			id := strings.TrimPrefix(r.URL.Path, route)
			if id == "" {
				// list request
				resourceList, ok := resource.(ResourceSpecList)
				if !ok {
					RespondNotImplemented(w)
					return
				}

				context, err = resourceList.List(
					r.Form.Get("q"),
					stringToInt(r.Form.Get("page"), -1),
					stringToInt(r.Form.Get("count"), -1),
				)
				if err != nil {
					Log.ErrorObjR(r, "could not get resource list", err)
					RespondInternalServerError(w)
					return
				}

				templateName = "list_" + templateName
			} else {
				// resource request by ID
				resourceRead, ok := resource.(ResourceSpecRead)
				if !ok {
					RespondNotImplemented(w)
					return
				}

				context, err = resourceRead.Read(id)
				if err != nil {
					Log.ErrorContextR(
						r, "could not get resource",
						LogContext{"id": id, "error": err},
					)
					RespondInternalServerError(w)
					return
				}
			}

			if context == nil {
				RespondNotFound(w)
				return
			}

			if r.Form.Get("format") != "" {
				templateName = fmt.Sprintf("%s_%s", templateName, r.Form.Get("format"))
			}

			Log.TraceContextR(r, "resource context", LogContext{"context": context})
			renderResource(w, r, templateName, context)
		},
	}
}

func renderResource(w http.ResponseWriter, r *http.Request, name string, context interface{}) {
	resourceTemplateName := "resource_" + name

	// initialize template
	tmpl, _ := loadTemplate(w, r, name, resourceTemplateName)
	if tmpl == nil {
		return
	}

	// render resource
	err := tmpl.ExecuteTemplate(w, name, context)
	if err != nil {
		Log.ErrorContextR(
			r, "could not execute resource template",
			LogContext{
				"resource": name,
				"error":    err,
			},
		)
		RespondInternalServerError(w)
		return
	}
}
