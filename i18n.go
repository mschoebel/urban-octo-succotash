package uos

import (
	"github.com/vorlif/spreak"
	"golang.org/x/text/language"
)

var i18n *spreak.Bundle

func setupInternationalization() {
	if Config.I18N.Locale == "" {
		Log.Info("no locale specified - skip i18n initialization")
		return
	}
	if len(Config.I18N.Languages) == 0 {
		Log.Info("no languages specified - skip i18n initialization")
		return
	}

	langCodes := make([]interface{}, len(Config.I18N.Languages))
	for i, langConfig := range Config.I18N.Languages {
		code, err := language.Parse(langConfig)
		if err != nil {
			Log.PanicContext(
				"invalid language configuration",
				LogContext{"error": err, "lang": langConfig},
			)
			panic("invalid language configuration")
		}
		langCodes[i] = code
	}

	Log.Info("initialize i18n")
	bundle, err := spreak.NewBundle(
		spreak.WithSourceLanguage(langCodes[0].(language.Tag)),
		spreak.WithDomainPath(spreak.NoDomain, Config.I18N.Locale),
		spreak.WithLanguage(langCodes[1:]...),
	)
	if err != nil {
		Log.PanicError("could not initialize i18n", err)
	}

	i18n = bundle
}

type wrappedLocalizer struct {
	lang string
	loc  *spreak.Localizer
}

func getLocalizer(lang string) *wrappedLocalizer {
	if i18n == nil {
		return nil
	}

	return &wrappedLocalizer{
		lang: lang,
		loc:  spreak.NewLocalizer(i18n, lang),
	}
}

func (l *wrappedLocalizer) Lang() string {
	return l.lang
}

func (l *wrappedLocalizer) Tr(message string, vars ...interface{}) string {
	return l.loc.Getf(message, vars...)
}
