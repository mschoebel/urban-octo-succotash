package uos

import (
	"fmt"
	"net/http"
)

type sitemapInfo struct {
	allowed    []string
	disallowed []string
}

func (si *sitemapInfo) allow(route string) {
	si.allowed = append(si.allowed, route)
}

func (si *sitemapInfo) disallow(route string) {
	si.disallowed = append(si.disallowed, route)
}

func (si *sitemapInfo) addToSitemap(route string, allow bool) {
	if allow {
		si.allow(route)
	} else {
		si.disallow(route)
	}
}

var sitemap = sitemapInfo{
	allowed:    []string{},
	disallowed: []string{},
}

func setupSitemapHandler() {
	Log.DebugContext(
		"register robots/sitemap handler",
		LogContext{"allowed": len(sitemap.allowed), "disallowed": len(sitemap.disallowed)},
	)

	appMux.HandleFunc("/robots.txt", robotsHandler)
	appMux.HandleFunc("/sitemap.txt", sitemapHandler)
}

func robotsHandler(w http.ResponseWriter, r *http.Request) {
	Log.Debug("robots.txt requested")
	if r.Method != http.MethodGet {
		RespondNotFound(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	fmt.Fprintf(w, "User-agent: *\n")
	fmt.Fprintf(w, "Sitemap: %s/sitemap.txt\n\n", Config.Pages["_default"].URL)

	for _, url := range sitemap.disallowed {
		fmt.Fprintf(w, "Disallow: %s%s\n", Config.Pages["_default"].URL, url)
	}
}

func sitemapHandler(w http.ResponseWriter, r *http.Request) {
	Log.Debug("sitemap.txt requested")
	if r.Method != http.MethodGet {
		RespondNotFound(w)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	for _, url := range sitemap.allowed {
		fmt.Fprintf(w, "%s%s\n", Config.Pages["_default"].URL, url)
	}
}
