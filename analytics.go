package uos

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/mileusna/useragent"
)

type session struct {
	startTime      time.Time
	expirationTime time.Time
}

type analyticsInfo struct {
	isEnabled bool

	mutex          sync.Mutex
	activeSessions map[string]session

	expirationCheck chan struct{}
}

const (
	sessionExpirationDuration = 5 * time.Minute
	sessionCleanupInterval    = 30 * time.Second
)

var analytics = analyticsInfo{
	activeSessions:  map[string]session{},
	expirationCheck: make(chan struct{}),
}

func setupAnalytics() {
	if Config.Features.Analytics == "" {
		// feature disabled
		return
	}

	// create output directory
	err := os.MkdirAll(Config.Features.Analytics, 0755)
	if err != nil {
		Log.PanicError("could not create analytics output directory", err)
	}
	Log.InfoContext("analytics enabled", LogContext{"dir": Config.Features.Analytics})

	analytics.isEnabled = true
	analytics.log("----- restart -----")

	// start background session expiration check
	ticker := time.NewTicker(sessionCleanupInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				analytics.removeExpiredSessions(false)
			case <-analytics.expirationCheck:
				ticker.Stop()
				return
			}
		}
	}()
}

func cleanupAnalytics() {
	if !analytics.isEnabled {
		return
	}

	// terminate background job
	close(analytics.expirationCheck)
	analytics.removeExpiredSessions(true)
}

func (a *analyticsInfo) add(r *http.Request) string {
	var (
		sid = sessionID(r)
		rid = randomString(8)
	)

	if !a.isEnabled {
		return rid
	}

	var ()

	Log.InfoContext("session request", LogContext{"sid": sid, "rid": rid})

	// register session
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if _, isActive := a.activeSessions[sid]; !isActive {
		// new session
		a.activeSessions[sid] = session{startTime: time.Now()}
		Metrics.GaugeInc(mActiveSessions)

		ua := useragent.Parse(r.UserAgent())
		a.log(
			fmt.Sprintf(
				"N %s '%s %s' '%s %s' - m=%v t=%v d=%v b=%v",
				sid, ua.Name, ua.VersionNoShort(), ua.OS, ua.OSVersionNoShort(),
				ua.Mobile, ua.Tablet, ua.Desktop, ua.Bot,
			),
		)
	}

	s := a.activeSessions[sid]
	s.expirationTime = time.Now().Add(sessionExpirationDuration)
	a.activeSessions[sid] = s

	a.log(fmt.Sprintf("R %s %s - %s %s - ref=%s", sid, rid, r.Method, r.URL.Path, r.Header.Get("Referer")))

	return rid
}

func (a *analyticsInfo) removeExpiredSessions(force bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	Log.DebugContext("start removing expired sessions", LogContext{"active": len(a.activeSessions)})
	now := time.Now()

	for id, s := range a.activeSessions {
		if force || s.expirationTime.Before(now) {
			Metrics.GaugeDec(mActiveSessions)
			delete(a.activeSessions, id)

			duration := s.expirationTime.Sub(s.startTime).Seconds() - float64(sessionExpirationDuration.Seconds())
			a.log(fmt.Sprintf("E %s - %.2f", id, duration))
		}
	}
	Log.DebugContext("done removing expired sessions", LogContext{"active": len(a.activeSessions)})
}

func (a *analyticsInfo) log(entry string) {
	fileName := path.Join(
		Config.Features.Analytics,
		fmt.Sprintf("analytics_%s.log", time.Now().UTC().Format("2006-01-02")),
	)

	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		Log.ErrorObj("could not open analytics log file", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("%v %s\n", time.Now().UTC().Unix(), entry))
	if err != nil {
		Log.ErrorObj("could not write to analytics log file", err)
	}
}

func sessionID(r *http.Request) string {
	// get request parameter
	var (
		ip        = r.RemoteAddr
		userAgent = r.UserAgent()

		xForwardedFor = r.Header.Get("X-Forwarded-For")
		xRealIP       = r.Header.Get("X-Real-IP")
	)

	// combine
	identifier := ip + ":" + xForwardedFor + ":" + xRealIP + ":" + userAgent
	Log.DebugContext("calculate session ID", LogContext{"id": identifier})

	// create hash
	hasher := sha256.New()
	hasher.Write([]byte(identifier))

	return hex.EncodeToString(hasher.Sum(nil))
}
