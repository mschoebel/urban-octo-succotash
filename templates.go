package uos

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

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

		"user": func(key string) interface{} {
			user, ok := r.Context().Value(ctxAppUser).(AppUser)
			if !ok {
				return false
			}

			switch key {
			case "isAuthenticated":
				return true
			case "isAdmin":
				return user.IsAdmin
			case "name":
				return user.Name
			}

			return false
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
	templateFile, err := os.ReadFile(filepath.Join(config.Assets.Templates, templateName))
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
	w http.ResponseWriter,
	r *http.Request,
	name,
	templateName string,
) *template.Template {
	templatePath := filepath.Join(config.Assets.Templates, templateName)

	// template not found?
	info, err := os.Stat(templatePath)
	if err != nil {
		if os.IsNotExist(err) {
			LogWarnContext(
				"file not found",
				LogContext{"file": templatePath},
			)
			RespondNotFound(w)
			return nil
		}

		LogErrorContext(
			"could not os.Stat template file",
			LogContext{
				"file":  templatePath,
				"error": err,
			},
		)
		RespondInternalServerError(w)
		return nil
	}

	// requestes a directory? (trailing slash)
	if info.IsDir() {
		RespondNotFound(w)
		return nil
	}

	// load template
	templateFile, err := os.ReadFile(templatePath)
	if err != nil {
		LogErrorContext(
			"could not read template file",
			LogContext{
				"file":  templatePath,
				"error": err,
			},
		)
		RespondInternalServerError(w)
		return nil
	}
	templateString := preprocessTemplate(name, templateFile)

	// parse template
	tmpl, err := template.New("").Funcs(getTemplateFuncMap(r)).Parse(templateString)
	if err != nil {
		LogErrorContext(
			"could not parse template file",
			LogContext{
				"file":  templatePath,
				"error": err,
			},
		)
		RespondInternalServerError(w)
		return nil
	}

	return tmpl
}
