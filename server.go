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
	LogInfoContext("start listening", LogContext{"port": config.Port})

	err := http.ListenAndServe(fmt.Sprintf(":%d", config.Port), appMux)
	if err != nil {
		LogPanicError("application error", err)
	}
}
