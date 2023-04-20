package uos

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/html"
)

// MarkdownHandler returns a handler for the "/markdown/" route providing HTML for the documents
// in the configured markdown directory.
// The handler can be activated using RegisterAppRequestHandlers.
func MarkdownHandler() AppRequestHandlerMapping {
	return AppRequestHandlerMapping{
		Route:   "/markdown/",
		Handler: markdownHandler,
	}
}

func markdownHandler(w http.ResponseWriter, r *http.Request) {
	// determine markdown element name
	mdElementName := getElementName("markdown", r.URL.Path)
	if mdElementName == "" {
		RespondNotFound(w)
		return
	}
	mdFilePath := filepath.Join(config.Assets.Markdown, mdElementName)

	// append ".md" file extension (of not already present)
	if !strings.HasSuffix(mdFilePath, ".md") {
		mdFilePath = mdFilePath + ".md"
	}

	// markdown file not found?
	info, err := os.Stat(mdFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			LogWarnContext(
				"file not found",
				LogContext{"file": mdFilePath},
			)
			RespondNotFound(w)
			return
		}

		LogErrorContext(
			"could not os.Stat markdown file",
			LogContext{
				"file":  mdFilePath,
				"error": err,
			},
		)
		RespondInternalServerError(w)
		return
	}

	// requestes a directory? (trailing slash)
	if info.IsDir() {
		RespondNotFound(w)
		return
	}

	// read markdown file
	md, err := ioutil.ReadFile(mdFilePath)
	if err != nil {
		LogErrorContext(
			"could not read markdown file",
			LogContext{
				"file":  mdFilePath,
				"error": err,
			},
		)
		RespondInternalServerError(w)
		return
	}

	var (
		opts = html.RendererOptions{
			Flags:          html.FlagsNone,
			RenderNodeHook: renderHook,
		}
		renderer = html.NewRenderer(opts)
	)

	// render document and wrap in a content-div (Bulma CSS)
	result := fmt.Sprintf(
		`<div class="content">%s</div>`,
		string(markdown.ToHTML(md, nil, renderer)),
	)
	w.Write([]byte(result))
}

func renderHook(w io.Writer, node ast.Node, entering bool) (ast.WalkStatus, bool) {
	switch astObj := node.(type) {
	case *ast.Heading:
		// customized rendering (Bulma CSS)
		level := astObj.Level

		if entering {
			switch level {
			case 1:
				w.Write([]byte(`<h1 class="title">`))
			default:
				w.Write([]byte(
					fmt.Sprintf(`<h%d class="subtitle is-%d">`, level, level+2),
				))
			}
		} else {
			w.Write([]byte(
				fmt.Sprintf("</h%d>", level),
			))
		}

		return ast.GoToNext, true
	}

	return ast.GoToNext, false
}
