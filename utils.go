package uos

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func getElementName(prefix, urlPath string) string {
	urlPath = filepath.Clean(urlPath)
	if !strings.HasPrefix(urlPath, fmt.Sprintf("/%s/", prefix)) {
		return ""
	}
	return urlPath[len(prefix)+2:]
}

func respondWithStatusText(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// RespondNotFound sends "not found" error.
func RespondNotFound(w http.ResponseWriter) {
	respondWithStatusText(w, http.StatusNotFound)
}

// RespondBadRequest sends "bad request" error.
func RespondBadRequest(w http.ResponseWriter) {
	respondWithStatusText(w, http.StatusBadRequest)
}

// RespondNotImplemented sends "not implemented" error.
func RespondNotImplemented(w http.ResponseWriter) {
	respondWithStatusText(w, http.StatusNotImplemented)
}

// RespondInternalServerError sends "internal server error".
func RespondInternalServerError(w http.ResponseWriter) {
	respondWithStatusText(w, http.StatusInternalServerError)
}

func IsCSRFtokenValid(r *http.Request, token string) bool {
	if user, ok := r.Context().Value(ctxAppUser).(AppUser); ok {
		// user is authenticated -> token can be "" (= no token required) or must match
		return token == "" || token == user.csrfToken
	}

	// token must be "" (= no token required)
	return token == ""
}

func stringToInt(s string, defaultValue int) int {
	v, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}

	return v
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func randomString(length int) string {
	var (
		charset    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		charsetLen = len(charset)
	)

	s := make([]byte, length)
	for i := range s {
		s[i] = charset[rand.Intn(charsetLen)]
	}

	return string(s)
}

func base64encode(input []byte) string {
	return base64.URLEncoding.EncodeToString(input)
}

func base64decode(input string) []byte {
	result, err := base64.URLEncoding.DecodeString(input)
	if err != nil {
		return nil
	}
	return result
}

// ReadFile reads the specified file (using 'os.ReadFile'). It checks, whether the given path
// is contained in the configured base directory (e.g. the current working dir) - if not it
// returnes an error. Use this function to guard against malicious paths.
func ReadFile(path string) ([]byte, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}

	Log.TraceContext(
		"base directory check",
		LogContext{"base": config.BaseDir, "file": absPath},
	)
	if !strings.HasPrefix(absPath, config.BaseDir+string(filepath.Separator)) {
		return nil, fmt.Errorf("path outside of base directory")
	}

	return os.ReadFile(path)
}
