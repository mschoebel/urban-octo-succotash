package uos

import (
	"math/rand"
	"time"
)

// ComponentSetup initializes web application framework using the specified configuration file.
// Uses the default configuration file name ("app_config.json") if called without parameters.
// Otherwise, the first parameter is used. Panics if anything fails.
func ComponentSetup(configFilePath ...string) {
	if len(configFilePath) > 1 {
		panic("ComponentSetup called with too many parameters")
	}

	configFile := "app_config.json"
	if len(configFilePath) > 0 {
		configFile = configFilePath[0]
	}

	err := readConfiguration(configFile)
	if err != nil {
		panic(err)
	}

	setupLogging()
	setupMonitoring()
	setupDataAccess()
	setupAuthentication()
	setupInternationalization()

	rand.Seed(time.Now().UnixNano())

	Log.Info("framework initialized")
}

// ComponentCleanup frees all initializes framework resources.
// Should be called on program termination, eg. in a defer call.
func ComponentCleanup() {
	Log.Info("framework cleanup")

	cleanupDataAccess()
}
