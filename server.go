package uos

import (
	"fmt"
	"net/http"
)

var appMux = http.NewServeMux()

// StartApp starts the web application server.
// Starts handling requests at the configured port. Blocks.
// Panics if anything fails.
func StartApp() {
	Log.InfoContext("start listening", LogContext{"port": Config.Port})

	setupSitemapHandler()

	err := http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), appMux)
	if err != nil {
		Log.PanicError("application error", err)
	}
}
