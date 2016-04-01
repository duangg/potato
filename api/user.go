package api

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"time"
	"bytes"
	"net/url"
	"io/ioutil"
	"encoding/json"
	"log"

	"github.com/gorilla/mux"

	"github.com/SpruceX/potato/service"
	"github.com/SpruceX/potato/utils"
)

type JsonResult struct {
	Ret     int `json:"ret"`
	Data    interface{} `json:"data"`
	ErrMsg  string `json:"errmsg"`
	ErrCode int `json:"errcode"`
}

const (
	TTL = 600
	TimeOut = time.Duration(3 * time.Second)
	SsoRetCode = "ret"
	SsoRetErrMsg = "errmsg"
	SsoCookieName = "user_token"
	ContentTypeJson = "application/json;charset=utf-8"
	ContentTypeUrl  = "application/x-www-form-urlencoded"
)

func initUser(r *mux.Router) {
	log.Printf("Initializing user api routes")
	user := r.PathPrefix("/user").Subrouter()
//	user.Handle("/{name}/change", ApiAppHandler(changePasswd)).Methods("POST")
//	user.Handle("/{name}/login", AnonymousApiHandler(login)).Methods("POST")
	user.Handle("/{name}/logout", ApiAppHandler(logout)).Methods("POST")
}

func initAdmin(r *mux.Router) {
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Handle("/", ApiAppHandler(searchAll)).Methods("GET")
	admin.Handle("/", ApiAppHandler(add)).Methods("POST")
	admin.Handle("/{name}", ApiAppHandler(del)).Methods("DELETE")
}

func add(c *Context, w http.ResponseWriter, r *http.Request) {
	uname := r.PostFormValue("username")
	pwd := r.PostFormValue("password")
	if uname == "" || pwd == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	Md5 := md5.New()
	Md5.Write([]byte(pwd))
	pwd = hex.EncodeToString(Md5.Sum(nil))

	if ok := service.CheckUser(uname); !ok {
		c.Err = NewAppError("add error", "This user has already been occupied", 500)
		return
	}

	user := service.NewUser(uname, pwd)
	err := service.AddUser(user)
	if err != nil {
		c.Err = NewAppError("add user failure", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "add user success")
}

func del(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uname := vars["name"]
	if uname == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	err := service.DelUser(uname)
	if err != nil {
		c.Err = NewAppError("del user failure", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "del user success")
}

func changePasswd(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uname := vars["name"]
	newPwd := r.PostFormValue("new_password")
	oldPwd := r.PostFormValue("old_password")
	if uname == "" || newPwd == "" || newPwd == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}
	Md5 := md5.New()
	Md5.Write([]byte(newPwd))
	newPwd = hex.EncodeToString(Md5.Sum(nil))

	Md5 = md5.New()
	Md5.Write([]byte(oldPwd))
	oldPwd = hex.EncodeToString(Md5.Sum(nil))

	err := service.UpdateUserPwd(uname, newPwd, oldPwd)
	if err != nil {
		c.Err = NewAppError("Param error", err.Error(), 500)
		return
	}
	utils.WriteJson(w, "update user success")
}

func searchAll(c *Context, w http.ResponseWriter, r *http.Request) {
	userSet, err := service.SearchAllUser()
	if err != nil {
		c.Err = NewAppError("user search all err:", err.Error(), 500)
		return
	} else {
		utils.WriteJson(w, userSet)
	}
}

func login(c *Context, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	uname := vars["name"]
	pwd := r.PostFormValue("password")
	if uname == "" || pwd == "" {
		c.Err = NewAppError("Param error", "Parameters cannot be empty", 500)
		return
	}

	Md5 := md5.New()
	Md5.Write([]byte(pwd))
	pwd = hex.EncodeToString(Md5.Sum(nil))

	oldSid, newSid, err := Srv.Service.User.Login(uname, pwd)
	if err != nil {
		c.Err = NewAppError("login", err.Error(), http.StatusForbidden)
		return
	}

	cookie := &http.Cookie{
		Name:  UserTokenName,
		Value: newSid,
		Path:  "/",
	}
	http.SetCookie(w, cookie)

	if _, ok := Srv.Service.User.Session[oldSid]; ok {
		delete(Srv.Service.User.Session, oldSid)
		Srv.Service.User.Session[newSid] = time.Now().Unix()
	} else {
		Srv.Service.User.Session[newSid] = time.Now().Unix()
	}

	utils.WriteJson(w, "login ok")
}

func SessionCheck(c *Context, w http.ResponseWriter, r *http.Request) (bool, error) {
	key, err := r.Cookie(UserTokenName)
	if err != nil {
		return false, errors.New("token is not found, no action is performed")
	}
	client := http.Client{
		Timeout: TimeOut,
	}
	v := url.Values{}
	v.Set("sid", utils.Cfg.SystemID)
	v.Set(SsoCookieName, key.Value)
	body := bytes.NewBuffer([]byte(v.Encode()))
	res, err := client.Post(utils.Cfg.SsoVerifyUrl, ContentTypeUrl, body)
	if err != nil {
		log.Printf("failed to post to sso verify server,err:%s", err.Error())
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("failed to verify Session,res.StatusCode-%d", res.StatusCode)
		return false, err
	}
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("failed to read body from response http message,err:%s", err.Error())
		return false, err
	}
	var result JsonResult
	err = json.Unmarshal(b, &result)
	if err != nil {
		log.Printf("failed to unmarshall body json,err:%s", err.Error())
		return false, err
	}

	if result.Ret == 0 {
		log.Printf("faided to verify SessionId, err:%s", result.ErrMsg)
		return false, errors.New(result.ErrMsg)
	}
	return true, nil
}

func logout(c *Context, w http.ResponseWriter, r *http.Request) {
	var infoArray []string = make([]string, 1)
	infoArray[0] = utils.Cfg.RedirectUrl
	utils.WriteJson(w, infoArray)

//	token, err := r.Cookie(UserTokenName)
//	if err != nil {
//		c.Err = NewAppError("logout", "token is not found, no action is performed", http.StatusBadRequest)
//		return
//	}
//	err = Srv.Service.User.Logout(token.Value)
//	if err != nil {
//		utils.WriteJson(w, "logout failure")
//		return
//	}
//	utils.WriteJson(w, "logout ok")
}
