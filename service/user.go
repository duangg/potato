package service

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2/bson"
	"github.com/SpruceX/potato/models"
	"github.com/SpruceX/potato/store"
)

type ApiUserService struct {
	Store   *store.UserStore
	Session map[string]int64
}

func InitUser() *ApiUserService {
	return &ApiUserService{
		Store:   store.Store.User,
		Session: make(map[string]int64),
	}
}

func NewUser(uname, pwd string) *models.UserTable {
	return &models.UserTable{
		Id:       bson.NewObjectId(),
		Status:   0,
		Auth:     0,
		Name:     uname,
		Pwd:      pwd,
		Session:  "",
		DateTime: time.Now(),
	}
}

func CheckUser(name string) bool {
	if _, err := AllService.User.Store.UserFindUserByName(name); err == nil {
		return false
	}

	return true
}

func AddUser(user *models.UserTable) error {
	return AllService.User.Store.UserInsert(user)
}

func DelUser(name string) error {
	return AllService.User.Store.DeleteByName(name)
}

func SearchAllUser() ([]models.UserTable, error) {
	return AllService.User.Store.UserTraversal()
}

func UpdateUserPwd(name, newPwd, oldPwd string) error {
	user, err := AllService.User.Store.UserFindUserByName(name)
	if err != nil {
		return err
	}
	user.Pwd = newPwd
	return AllService.User.Store.UserUpdate(&user)
}

func (u *ApiUserService) Login(username, password string) (string, string, error) {
	user, err := AllService.User.Store.UserFindUserByName(username)
	if err != nil {
		return "", "", errors.New("The user is not exist")
	} else {
		if user.Pwd == password {
			oldSid := user.Session
			user.Session = bson.NewObjectId().Hex()
			err := AllService.User.Store.UserUpdate(&user)
			if err != nil {
				return "", "", err
			}

			return oldSid, user.Session, nil
		} else {
			return "", "", errors.New("User name or password mismatch")
		}
	}
}

func (u *ApiUserService) Logout(sid string) error {
	if _, ok := u.Session[sid]; ok {
		delete(u.Session, sid)
	}

	return nil
}
