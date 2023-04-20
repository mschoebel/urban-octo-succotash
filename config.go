package uos

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// AppConfiguration specifies application/framework configuration.
// Is read from a JSON configuration file in ComponentSetup.
type AppConfiguration struct {
	// web application port
	Port int `json:"port"`

	Logging  LogConfiguration   `json:"logging"`
	Database DBConfiguration    `json:"database"`
	Assets   AssetConfiguration `json:"assets"`

	Auth AuthenticationConfiguration `json:"auth"`

	// page configuration integrated into HTML pages. To define common settings
	// the page "_default" can be specified.
	Pages map[string]PageConfiguration `json:"pages"`
}

// LogConfiguration specifies logging behaviour.
type LogConfiguration struct {
	// log level (panic, fatal, error, warn, info, debug, trace)
	Level string `json:"level"`
	// write logmessages as colored output to stderr - otherwise log as JSON
	UseConsole bool `json:"use_console"`
}

// DBConfiguration specifies the database.
type DBConfiguration struct {
	// SQLite database file
	File string `json:"file"`
}

// AssetConfiguration specifies directories containing different types of static data.
type AssetConfiguration struct {
	// directory containing template files for pages, forms, dialogs, fragements.
	Templates string `json:"templates"`
	// directory containing markdown documents
	Markdown string `json:"markdown"`
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

	StaticBaseURL string `json:"static_base_url"`

	FavIcon string `json:"favicon"`

	Styles []string `json:"styles"`

	ScriptsHead []string `json:"scripts_head"`
	ScriptsBody []string `json:"scripts_body"`
}

var config = AppConfiguration{}

func readConfiguration(configFilePath string) error {
	configFileContent, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(configFileContent, &config)
	if err != nil {
		return err
	}

	// check/create authentification info
	if len(config.Auth.HashKey) == 0 {
		config.Auth.HashKey = randomString(64)
		fmt.Printf("generated  hash key: %s\n", config.Auth.HashKey)
	}
	if len(config.Auth.BlockKey) == 0 {
		config.Auth.BlockKey = randomString(32)
		fmt.Printf("generated block key: %s\n", config.Auth.BlockKey)
	}

	config.Auth.hash = []byte(config.Auth.HashKey)
	config.Auth.block = []byte(config.Auth.BlockKey)

	return nil
}

func (c AppConfiguration) getPageConfig(pageName string) PageConfiguration {
	var (
		result     = c.Pages["_default"]
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