package uos

import (
	"embed"
	"net/http"
)

// AppRequestHandler represents a application specific HTTP handler function.
type AppRequestHandler func(http.ResponseWriter, *http.Request)

// AppRequestHandlerMapping represents a path pattern and a corresponding handler.
type AppRequestHandlerMapping struct {
	// URL route
	Route string
	// request handler for specified route
	Handler AppRequestHandler
}

// RegisterAppRequestHandler registers the given handler for a specific URL pattern/path.
func RegisterAppRequestHandler(pattern string, handler AppRequestHandler) {
	Log.DebugContext("register request handler", LogContext{"pattern": pattern})
	appMux.Handle(pattern, mwWrapF(handler))
}

// RegisterAppRequesHandlers registers a list of handler.
func RegisterAppRequestHandlers(handlerList ...AppRequestHandlerMapping) {
	for _, mapping := range handlerList {
		RegisterAppRequestHandler(
			mapping.Route,
			mapping.Handler,
		)
	}
}

// RegisterStaticAssetServer provides the given content file system at the specified location.
func RegisterStaticAssets(route string, content embed.FS) {
	Log.DebugContext(
		"register static assets server",
		LogContext{"route": route},
	)
	appMux.Handle(route, http.FileServer(http.FS(content)))
}
