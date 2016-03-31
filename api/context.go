package api

import (
	"net/http"
	"github.com/SpruceX/utils"
)

const (
	UserTokenName = "user_token"
)

const (
	UsernameOrPasswordUnMatch = 10001
	TokenNotFound             = 10002
)

type Context struct {
	Err interface{}
}

func (c *Context) SetUnknownError(where, message string) {
	c.Err = &AppError{where, message, http.StatusInternalServerError}
}

func NewAppError(where, message string, code int) *AppError {
	return &AppError{where, message, code}
}

type AppError struct {
	Where      string
	Message    string
	StatusCode int
}

type Page struct {
	TemplateName string
	Title        string
	SiteName     string
	SiteURL      string
	Props        map[string]string
}

func ApiAppHandler(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	return &handler{h, true}
}

func AppHandler(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	return &handler{h, true}
}

func AnonymousApiHandler(h func(*Context, http.ResponseWriter, *http.Request)) http.Handler {
	return &handler{h, false}
}

type handler struct {
	handleFunc   func(*Context, http.ResponseWriter, *http.Request)
	requireLogin bool
	//	requireUser bool
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer recoverError(w)
	c := &Context{}
	if h.requireLogin {
		result, _ := SessionCheck(c, w, r)
		if !result {
			http.Redirect(w, r, utils.Cfg.RedirectUrl, http.StatusFound)
			return
		}
	}
	h.handleFunc(c, w, r)
	if c.Err != nil {
		appErr := c.Err.(*AppError)
		utils.WriteError(w, appErr.StatusCode, appErr.Where+": "+appErr.Message)
	}
}

func recoverError(w http.ResponseWriter) {
	if err := recover(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		utils.WriteError(w, http.StatusInternalServerError, err.(error).Error())
	}
}
