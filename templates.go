package uos

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func hxResolve(r *http.Request, element, method, url, trigger, swap, attributes string) template.HTML {
	if config.Tuning.ActivateHTMXPreloading && method == "get" && trigger == "load" {
		var (
			itemType = getElementBase(url)
			itemName = getElementName(itemType, url)

			urlData = newFormData(url)

			content bytes.Buffer
		)

		switch itemType {
		case "fragments":
			status, err := handleFragment(&content, r, itemName, urlData)
			if status == http.StatusOK && err == nil {
				Log.TraceContext(
					"HTMX preloading successful",
					LogContext{"type": itemType, "name": itemName},
				)
				return template.HTML(content.String())
			}
		case "markdown":
			status := markdownHandler(&content, r, itemName)
			if status == http.StatusOK {
				Log.TraceContext(
					"HTMX preloading successful",
					LogContext{"type": itemType, "name": itemName},
				)
				return template.HTML(content.String())
			}
		default:
			Log.TraceContext("HTMX preloading not supported", LogContext{"type": itemType})
		}
	}

	// return directly as HTMX element -> no pre-loading
	return template.HTML(
		fmt.Sprintf(
			`<%[1]s hx-%[2]s="%[3]s" hx-trigger="%[4]s" hx-swap="%[5]s" %[6]s></%[1]s>`,
			element, method, url, trigger, swap, attributes,
		),
	)
}

func getTemplateFuncMap(r *http.Request) template.FuncMap {
	return template.FuncMap{
		"loop": func(from, to int) <-chan int {
			ch := make(chan int)
			go func() {
				for i := from; i <= to; i++ {
					ch <- i
				}
				close(ch)
			}()
			return ch
		},

		"contains":  strings.Contains,
		"hasPrefix": strings.HasPrefix,
		"hasSuffix": strings.HasSuffix,

		"inc": func(x int) int { return x + 1 },
		"dec": func(x int) int { return x - 1 },

		"hx": func(element, method, url, trigger, swap, attributes string) template.HTML {
			return hxResolve(r, element, method, url, trigger, swap, attributes)
		},
		"hxLoad": func(url string) template.HTML {
			return hxResolve(r, "div", "get", url, "load", "outerHTML", "")
		},

		"csrf": func() string {
			if user, ok := r.Context().Value(ctxAppUser).(AppUser); ok {
				return user.csrfToken
			}
			return ""
		},

		"app": func(key string) interface{} {
			if info, ok := config.AppInfo[key]; ok {
				return info
			}
			return ""
		},
		"user": func(key string) interface{} {
			user, ok := r.Context().Value(ctxAppUser).(AppUser)
			if !ok {
				return ""
			}

			switch key {
			case "isAuthenticated":
				return true
			case "isAdmin":
				return user.IsAdmin
			case "name":
				return user.Name
			}

			return ""
		},
	}
}

func preprocessTemplate(name string, template []byte) string {
	return fmt.Sprintf(`{{define "%s"}}%s{{end}}`, name, string(template))
}

func renderTemplate(
	w io.Writer,
	r *http.Request,
	name string,
	data interface{},
	templateName string,
) error {
	templateFile, err := ReadFile(filepath.Join(config.Assets.Templates, templateName))
	if err != nil {
		return err
	}
	templateString := preprocessTemplate(name, templateFile)

	tmpl, err := template.New("").Funcs(getTemplateFuncMap(r)).Parse(templateString)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

func renderInternalTemplate(
	w io.Writer,
	r *http.Request,
	name string,
	data interface{},
) error {
	templateFile, err := templateFS.ReadFile("templates/" + name)
	if err != nil {
		return err
	}
	templateString := preprocessTemplate(name, templateFile)

	tmpl, err := template.New("").Funcs(getTemplateFuncMap(r)).Parse(templateString)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, name, data)
}

func loadTemplate(
	w io.Writer,
	r *http.Request,
	name,
	templateName string,
) (*template.Template, int) {
	templatePath := filepath.Join(config.Assets.Templates, templateName)

	// template not found?
	info, err := os.Stat(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			Log.WarnContextR(
				r, "file not found",
				LogContext{"file": templatePath},
			)
			return nil, http.StatusNotFound
		}

		Log.ErrorContextR(
			r, "could not os.Stat template file",
			LogContext{
				"file":  templatePath,
				"error": err,
			},
		)
		return nil, http.StatusInternalServerError
	}

	// requestes a directory? (trailing slash)
	if info.IsDir() {
		return nil, http.StatusNotFound
	}

	// load template
	templateFile, err := ReadFile(templatePath)
	if err != nil {
		Log.ErrorContextR(
			r, "could not read template file",
			LogContext{
				"file":  templatePath,
				"error": err,
			},
		)
		return nil, http.StatusInternalServerError
	}
	templateString := preprocessTemplate(name, templateFile)

	// parse template
	tmpl, err := template.New("").Funcs(getTemplateFuncMap(r)).Parse(templateString)
	if err != nil {
		Log.ErrorContextR(
			r, "could not parse template file",
			LogContext{
				"file":  templatePath,
				"error": err,
			},
		)
		return nil, http.StatusInternalServerError
	}

	return tmpl, http.StatusOK
}
