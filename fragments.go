package uos

import (
	"net/http"
	"net/url"
)

// FragmentSpec describes an interface a web application fragment must provide.
type FragmentSpec interface {
	// Name returns the short name of a fragment. The fragment is available at '/fragments/<name>'
	// and uses the template 'fragment_<name>'.
	Name() string
}

// FragmentSpecRead describes an interface an fragment that supports GET requests must provide.
type FragmentSpecRead interface {
	// GetContextObject returns the fragment template context object for the given URL parameters.
	GetContextObject(params url.Values) (interface{}, error)
}

// FragmentHandler returns a handler for the "/fragments/" route providing the specified fragments.
// Creates default handler for fragments (as defined in templates directory).
// The handler can be activated using RegisterAppRequestHandlers.
func FragmentHandler(fragments ...FragmentSpec) AppRequestHandlerMapping {
	return AppRequestHandlerMapping{
		Route:   "/fragments/",
		Handler: getFragmentHandlerFunc(fragments),
	}
}

func getFragmentHandlerFunc(fragments []FragmentSpec) AppRequestHandler {
	nameToSpec := map[string]FragmentSpec{}
	for _, f := range fragments {
		nameToSpec[f.Name()] = f
		Log.DebugContext("register fragment spec", LogContext{"name": f.Name()})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// determine fragment
		fragmentName := getElementName("fragments", r.URL.Path)
		Log.DebugContext(
			"handle fragment",
			LogContext{
				"name":   fragmentName,
				"method": r.Method,
			},
		)

		// prepare request processing
		err := r.ParseForm()
		if err != nil {
			Log.WarnError("could not parse form", err)
			RespondBadRequest(w)
			return
		}

		fragmentSpec, ok := nameToSpec[fragmentName]
		if !ok {
			// might be fragment without specification -> directly forward to rendering
			Log.DebugContext("handle fragment without spec", LogContext{"name": fragmentName})
			renderObjectFragment(w, r, fragmentName, r.Form.Get("p"))
			return
		}

		// process request
		switch r.Method {
		case http.MethodGet:
			// does the fragment support GET method?
			fragmentRead, ok := fragmentSpec.(FragmentSpecRead)
			if !ok {
				RespondNotImplemented(w)
				return
			}

			// process fragment
			obj, err := fragmentRead.GetContextObject(r.Form)
			if err != nil {
				handleFragmentError(w, "could not get fragment context", err)
				return
			}

			renderObjectFragment(w, r, fragmentName, obj)
		default:
			RespondNotImplemented(w)
		}
	}
}

func handleFragmentError(w http.ResponseWriter, message string, err error) {
	switch err {
	case ErrorFragmentNotFound:
		RespondNotFound(w)
		return
	case ErrorFragmentInvalidRequest:
		RespondBadRequest(w)
		return
	}

	// all other cases: log as internal error
	Log.ErrorObj(message, err)
	RespondInternalServerError(w)
}

func renderObjectFragment(w http.ResponseWriter, r *http.Request, name string, obj interface{}) {
	data := map[string]interface{}{}

	if obj != nil {
		data["Object"] = obj
	}

	renderFragment(w, r, name, data)
}

func renderFragment(w http.ResponseWriter, r *http.Request, name string, data map[string]interface{}) {
	fragmentTemplateName := "fragment_" + name

	// initialize template
	tmpl := loadTemplate(w, r, name, fragmentTemplateName)
	if tmpl == nil {
		return
	}

	// render fragment
	err := tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		Log.ErrorContext(
			"could not execute fragment template",
			LogContext{
				"fragment": name,
				"error":    err,
			},
		)
		RespondInternalServerError(w)
		return
	}
}
