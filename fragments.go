package uos

import (
	"io"
	"net/http"
)

// FragmentSpec describes an interface a web application fragment must provide.
type FragmentSpec interface {
	// Name returns the short name of a fragment. The fragment is available at '/fragments/<name>'
	// and uses the template 'fragment_<name>'.
	Name() string
}

// FragmentSpecRead describes an interface a fragment that supports GET requests must provide.
type FragmentSpecRead interface {
	// GetContextObject returns the fragment template context object for the given URL parameters.
	GetContextObject(params Getter) (interface{}, error)
}

// FragmentHandler returns a handler for the "/fragments/" route providing the specified fragments.
// Creates default handler for fragments (as defined in templates directory).
// The handler can be activated using RegisterAppRequestHandlers.
func FragmentHandler(fragments ...FragmentSpec) AppRequestHandlerMapping {
	return AppRequestHandlerMapping{
		Route:   "/fragments/",
		Handler: getFragmentWebHandlerFunc(fragments),
	}
}

var fragmentRegistry map[string]FragmentSpec

func getFragmentWebHandlerFunc(fragments []FragmentSpec) AppRequestHandler {
	if fragmentRegistry != nil {
		Log.Panic("multiple fragment handler registration")
	}

	fragmentRegistry = map[string]FragmentSpec{}
	for _, f := range fragments {
		fragmentRegistry[f.Name()] = f
		Log.DebugContext("register fragment spec", LogContext{"name": f.Name()})
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// determine fragment
		name := getElementName("fragments", r.URL.Path)
		Log.DebugContextR(
			r, "handle fragment",
			LogContext{
				"name":   name,
				"method": r.Method,
			},
		)

		// only GET requests are supported
		if r.Method != http.MethodGet {
			RespondNotImplemented(w)
			return
		}

		// process request
		status, err := handleFragment(w, r, name, r.Form)
		if err != nil {
			handleFragmentError(w, r, "could not handle fragment", err)
			return
		}

		if status != http.StatusOK {
			respondWithStatusText(w, status)
		}
	}
}

func handleFragment(w io.Writer, r *http.Request, name string, form Getter) (int, error) {
	// get fragment specification
	fragmentSpec, ok := fragmentRegistry[name]
	if !ok {
		// might be fragment without specification -> directly forward to rendering
		Log.DebugContextR(r, "handle fragment without spec", LogContext{"name": name})
		return renderObjectFragment(w, r, name, form.Get("p"))
	}

	// does the fragment support GET method?
	fragmentRead, ok := fragmentSpec.(FragmentSpecRead)
	if !ok {
		return http.StatusNotImplemented, nil
	}

	// process fragment
	obj, err := fragmentRead.GetContextObject(form)
	if err != nil {
		return -1, err
	}

	return renderObjectFragment(w, r, name, obj)
}

func handleFragmentError(w http.ResponseWriter, r *http.Request, message string, err error) {
	switch err {
	case ErrorFragmentNotFound:
		RespondNotFound(w)
		return
	case ErrorFragmentInvalidRequest:
		RespondBadRequest(w)
		return
	}

	// all other cases: log as internal error
	Log.ErrorObjR(r, message, err)
	RespondInternalServerError(w)
}

func renderObjectFragment(w io.Writer, r *http.Request, name string, obj interface{}) (int, error) {
	data := map[string]interface{}{}

	if obj != nil {
		data["Object"] = obj
	}

	return renderFragment(w, r, name, data)
}

func renderFragment(w io.Writer, r *http.Request, name string, data map[string]interface{}) (int, error) {
	fragmentTemplateName := "fragment_" + name

	// initialize template
	tmpl, status := loadTemplate(w, r, name, fragmentTemplateName)
	if tmpl == nil {
		return status, nil
	}

	// render fragment
	err := tmpl.ExecuteTemplate(w, name, data)
	if err != nil {
		Log.ErrorContextR(
			r, "could not execute fragment template",
			LogContext{
				"fragment": name,
				"error":    err,
			},
		)
		return http.StatusInternalServerError, nil
	}

	return http.StatusOK, nil
}
