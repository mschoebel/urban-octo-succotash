package uos

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AppConfiguration specifies application/framework configuration.
// Is read from a JSON configuration file in ComponentSetup.
type AppConfiguration struct {
	// application context information, available as template context
	AppInfo map[string]interface{} `json:"app"`

	// web application port
	Port int `json:"port"`
	// base deployment directory - only files "below" this directory are used.
	// If empty or not defined, the directory of the executable is set.
	BaseDir string `json:"base_dir"`

	Logging    LogConfiguration        `json:"logging"`
	Monitoring MonitoringConfiguration `json:"monitoring"`
	Database   DBConfiguration         `json:"database"`
	Assets     AssetConfiguration      `json:"assets"`
	I18N       I18NConfiguration       `json:"i18n"`

	Auth AuthenticationConfiguration `json:"auth"`

	// page configuration integrated into HTML pages. To define common settings
	// the page "_default" can be specified.
	Pages map[string]PageConfiguration `json:"pages"`

	Tuning   TuningConfiguration  `json:"tuning"`
	Features FeatureConfiguration `json:"-"`
}

// LogConfiguration specifies logging behaviour.
type LogConfiguration struct {
	// log level (panic, fatal, error, warn, info, debug, trace)
	Level string `json:"level"`
	// write logmessages as colored output to stderr - otherwise log as JSON
	UseConsole bool `json:"use_console"`
}

// MonitoringConfiguration specifies ports for application monitoring.
type MonitoringConfiguration struct {
	// port for pprof web interface
	PortPPROF int `json:"pprof"`
	// port for application metrics (Prometheus)
	PortMetrics int `json:"metrics"`
}

// DBConfiguration specifies the database.
type DBConfiguration struct {
	// SQLite database file
	File string `json:"file"`
}

// AssetConfiguration specifies directories containing different types of static data.
type AssetConfiguration struct {
	// directory containing "dynamic" assets (= assets that are not included in the executable)
	Dynamic string `json:"dynamic"`
	// directory containing template files for pages, forms, dialogs, fragements.
	Templates string `json:"templates"`
	// directory containing markdown documents
	Markdown string `json:"markdown"`
}

type I18NConfiguration struct {
	// directory containing translations (PO files)
	Locale string `json:"locale"`
	// list of supported languages (first entry is primary language)
	Languages []string `json:"languages"`
}

// AuthenticationConfiguration specifies required keys for cookie handling.
// If a propertie is changed, existing cookies are invalidated.
type AuthenticationConfiguration struct {
	HashKey  string `json:"hash"`
	BlockKey string `json:"block"`

	hash  []byte
	block []byte
}

type PageConfiguration struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Author      string `json:"author"`
	URL         string `json:"url"`

	StaticBaseURL string `json:"static_base_url"`

	FavIcon string `json:"favicon"`

	Styles []string `json:"styles"`

	ScriptsHead []string `json:"scripts_head"`
	ScriptsBody []string `json:"scripts_body"`
}

func (pc PageConfiguration) clone() PageConfiguration {
	var (
		buf bytes.Buffer

		enc = gob.NewEncoder(&buf)
		dec = gob.NewDecoder(&buf)
	)

	err := enc.Encode(pc)
	if err != nil {
		Log.ErrorObj("could not clone page configuration (encode)", err)
		return pc
	}

	var cloned PageConfiguration
	err = dec.Decode(&cloned)
	if err != nil {
		Log.ErrorObj("could not clone page configuration (decode)", err)
		return pc
	}

	return cloned
}

type TuningConfiguration struct {
	ActivateHTMXPreloading bool `json:"htmx_preload"`
}

type FeatureConfiguration struct {
	Dialogs bool `json:"-"`
}

var Config = AppConfiguration{}

func readConfiguration(configFilePath string) error {
	configFileContent, err := os.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(configFileContent, &Config)
	if err != nil {
		return err
	}

	// check/create authentification info
	if len(Config.Auth.HashKey) == 0 {
		Config.Auth.HashKey = randomString(64)
		fmt.Printf("generated  hash key: %s\n", Config.Auth.HashKey)
	}
	if len(Config.Auth.BlockKey) == 0 {
		Config.Auth.BlockKey = randomString(32)
		fmt.Printf("generated block key: %s\n", Config.Auth.BlockKey)
	}

	Config.Auth.hash = []byte(Config.Auth.HashKey)
	Config.Auth.block = []byte(Config.Auth.BlockKey)

	// determine base directory
	if Config.BaseDir == "" {
		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		Config.BaseDir, err = filepath.Abs(exePath)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c AppConfiguration) getPageConfig(pageName string) PageConfiguration {
	var (
		result     = c.Pages["_default"].clone()
		pageConfig = c.Pages[pageName]
	)

	// integrate specific page configuration into result
	if pageConfig.Title != "" {
		result.Title = pageConfig.Title
	}
	if pageConfig.Description != "" {
		result.Description = pageConfig.Description
	}
	if pageConfig.Author != "" {
		result.Author = pageConfig.Author
	}
	if pageConfig.URL != "" {
		result.URL = pageConfig.URL
	}
	if pageConfig.StaticBaseURL != "" {
		result.StaticBaseURL = pageConfig.StaticBaseURL
	}
	if pageConfig.FavIcon != "" {
		result.FavIcon = pageConfig.FavIcon
	}
	if len(pageConfig.Styles) != 0 {
		result.Styles = pageConfig.Styles
	}
	if len(pageConfig.ScriptsHead) != 0 {
		result.Styles = pageConfig.ScriptsHead
	}
	if len(pageConfig.ScriptsBody) != 0 {
		result.Styles = pageConfig.ScriptsBody
	}

	return result
}
