package session

import (
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/knchan0x/umeshu/log"
)

// Session manager manages session and session store.
type SessionManager struct {
	store    Store           // session store
	settings SessionSettings // settings

	// GC
	isGCStarted bool
	gcStop      chan struct{}
}

// Settings of session manager.
type SessionSettings struct {
	Name        string // name of session cookie
	StoreType   string // session store type
	MaxLifeTime int    // max lifetime for session, in second
	LapseTime   int    // lapse time for session id, in second

	// key value pair for session authentication
	// optional, will use default values if nil
	// default value:
	// 		TokenKey = "Token"
	// 		TokenValue = "User-Agent"
	TokenKey   string // token key
	TokenValue string // token value, must be part of HTTP request header
}

const (
	createTime string = "CreateTime" // create time of session cookie
)

// DefaultSettings provides default values for session manager.
var DefaultSettings = SessionSettings{
	Name:        "SID",
	StoreType:   "InMemory",
	MaxLifeTime: 86400, // one day
	LapseTime:   1800,  // half hour
	TokenKey:    "Token",
	TokenValue:  "User-Agent",
}

// Global session manager
var Manager *SessionManager

// NewManager creates and returns new session manager object.
// It will start running GC at separate goroutine.
// SessionManager can only be created once.
func NewManager(settings SessionSettings) *SessionManager {
	if Manager != nil {
		return Manager
	}

	if settings.Name == "" {
		settings.Name = DefaultSettings.Name
	}
	if settings.StoreType == "" {
		settings.StoreType = DefaultSettings.StoreType
	}
	if settings.MaxLifeTime == 0 {
		settings.MaxLifeTime = DefaultSettings.MaxLifeTime
	}
	if settings.LapseTime == 0 {
		settings.LapseTime = DefaultSettings.MaxLifeTime
	}
	if settings.TokenKey == "" {
		settings.Name = DefaultSettings.TokenKey
	}
	if settings.TokenValue == "" {
		settings.Name = DefaultSettings.TokenValue
	}

	Manager = &SessionManager{
		store:    NewStore(settings.StoreType, settings),
		settings: settings,
	}

	Manager.StartGC()

	return Manager
}

// StartSession returns existing session or creates new session if no matched.
func (m *SessionManager) StartSession(rw http.ResponseWriter, r *http.Request) Session {
	if m == nil {
		log.Panic("session manager does not exists.")
	}

	cookie, err := r.Cookie(m.settings.Name)
	value := r.Header.Get(m.settings.TokenValue)
	if err != nil || cookie.Value == "" {
		return m.newSession(rw, value)
	}
	if err != nil {
		log.Error("session manager unable to load session cookie: %s", err.Error())
	}

	sid, _ := url.QueryUnescape(cookie.Value)
	session, _ := m.store.Read(sid)
	if session == nil {
		return m.newSession(rw, value)
	}

	if tokenValue := session.Get(m.settings.TokenKey); tokenValue != value {
		return m.newSession(rw, value)
	}

	if createTime := session.Get(createTime); createTime != nil {
		if createTime.(time.Time).Unix()+int64(m.settings.LapseTime) < time.Now().Unix() {
			// change session id to prevent session hijacking
			newSID := m.newSessionID()
			m.store.UpdateSID(sid, newSID)
			cookie := m.newSessionCookie(newSID)
			http.SetCookie(rw, cookie)
			return session
		}
	}

	return session
}

// EndSession closes session and remove it from session store.
func (m *SessionManager) EndSession(rw http.ResponseWriter, r *http.Request) {
	if m == nil {
		log.Panic("session manager does not exists")
	}

	cookie, err := r.Cookie(m.settings.Name)
	if err != nil || cookie.Value == "" {
		return
	}

	if err := m.store.Delete(cookie.Value); err != nil {
		log.Error("unable to delete session from store, error: %s", err.Error())
	}

	http.SetCookie(rw, m.newEndSessionCookie())
}

// StartGC starts the GC for session manager.
func (m *SessionManager) StartGC() {
	if m.isGCStarted {
		return
	}

	m.isGCStarted = true
	m.gcStop = make(chan struct{}, 1)
	m.gc()
}

// StopGC sends signal to stop the GC of session manager.
func (m *SessionManager) StopGC() {
	if !m.isGCStarted {
		return
	}
	m.isGCStarted = false
	close(m.gcStop)
}

// gc runs GC in a new goroutine.
func (m *SessionManager) gc() {
	ticker := time.NewTicker(time.Duration(m.settings.MaxLifeTime) * time.Second)
	go func() {
		for {
			select {
			case _, ok := <-m.gcStop:
				if !ok {
					return
				}
			case <-ticker.C:
				m.store.GC(m.settings.MaxLifeTime)
			}
		}
	}()
}

// newSessionCookie creates new *http.Cookie object with settings
// defined in session manager.
func (m *SessionManager) newSessionCookie(sid string) *http.Cookie {
	return &http.Cookie{
		Name:     m.settings.Name,
		Value:    url.QueryEscape(sid),
		Path:     "/",
		HttpOnly: true,
		MaxAge:   m.settings.MaxLifeTime,
	}
}

// newEndSessionCookie creates new empty *http.Cookie object.
// with -1 max age.
func (m *SessionManager) newEndSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     m.settings.Name,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
		MaxAge:   -1,
	}
}

// newSession creates and returns new session object.
func (m *SessionManager) newSession(rw http.ResponseWriter, agent string) Session {
	sid := m.newSessionID()
	session, _ := m.store.Insert(sid, agent)
	session.Set(createTime, time.Now())
	cookie := m.newSessionCookie(sid)
	http.SetCookie(rw, cookie)
	return session
}

// newSessionID returns new UUID for session.
func (m *SessionManager) newSessionID() string {
	return uuid.New().String()
}
