package uos

import (
	"embed"
	"net/http"
)

// AppRequestHandler represents a application specific HTTP handler function.
type AppRequestHandler func(http.ResponseWriter, *http.Request)

// AppRequestHandlerOptions specify meta information about request handling
type AppRequestHandlerOptions struct {
	// do not expect valid CSRF token on POST/PUT/DELETE
	NoCSRFcheck bool

	// redirect to login page if not authenticated
	IsAuthRequired bool
	// login page (target for "authentication required" pages)
	IsAuthPage bool
}

// AppRequestHandlerMapping represents a path pattern and a corresponding handler.
type AppRequestHandlerMapping struct {
	// URL route
	Route string
	// request handler for specified route
	Handler AppRequestHandler
	// options
	Options AppRequestHandlerOptions
}

// Internal indicates, that the given request handler requires authentication.
func (hm AppRequestHandlerMapping) Internal() AppRequestHandlerMapping {
	hm.Options.IsAuthRequired = true
	return hm
}

// AuthPage indicates, that the given request handler provides the authentication page.
func (hm AppRequestHandlerMapping) AuthPage() AppRequestHandlerMapping {
	hm.Options.IsAuthPage = true
	return hm
}

// RegisterAppRequestHandler registers the given handler for a specific URL pattern/path.
func RegisterAppRequestHandler(pattern string, handler AppRequestHandler, options AppRequestHandlerOptions) {
	Log.DebugContext("register request handler", LogContext{"pattern": pattern})

	if options.IsAuthPage {
		if authenticationPageURL != "" {
			Log.PanicContext(
				"multiple auth pages specified",
				LogContext{"first": authenticationPageURL, "second": pattern},
			)
			panic("multiple auth pages specified")
		}
		authenticationPageURL = pattern
		Log.InfoContext("registered auth page", LogContext{"url": authenticationPageURL})
	}

	appMux.Handle(pattern, mwWrapF(handler, options))
}

// RegisterAppRequesHandlers registers a list of handler.
func RegisterAppRequestHandlers(handlerList ...AppRequestHandlerMapping) {
	for _, mapping := range handlerList {
		RegisterAppRequestHandler(
			mapping.Route,
			mapping.Handler,
			mapping.Options,
		)
	}
}

// RegisterAppRequestHandlersList registers a slice of handler.
func RegisterAppRequestHandlersList(list []AppRequestHandlerMapping) {
	RegisterAppRequestHandlers(list...)
}

// RegisterStaticAssetServer provides the given content file system at the specified location.
func RegisterStaticAssets(route string, content embed.FS) {
	Log.DebugContext(
		"register static assets server",
		LogContext{"route": route},
	)
	appMux.Handle(route, http.FileServer(http.FS(content)))
}
