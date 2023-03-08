package dao

import (
	"fleetmanager/api/service/constants"
	"fleetmanager/db/dbm"
	"fleetmanager/setting"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/google/uuid"
)

// User 用户表结构体
type User struct {
	Id             string    `orm:"column(id);size(64);pk"`
	UserName       string    `orm:"column(username);unique;size(32)"` // 用户名
	Password       string    `orm:"column(password);size(256)"`       // 密码
	Email          string    `orm:"column(email);size(64)"`           // Email
	Phone          string    `orm:"column(phone);size(32)"`
	LastLogin      time.Time `orm:"column(last_login);auto_now;type(datetime)"` // 上次登录时间
	Activation     int8      `orm:"column(activation)"`                         // 状态 0默认 1激活 -1冻结
	UserType       int8      `orm:"column(usertype)"`                           // 用户类型
	MaxRetry       int8      `orm:"column(max_retry)"`                          // 最大尝试登录次数
	TotalResNumber int       `orm:"column(total_res_number)"`
	FrozenTime     time.Time `orm:"column(frozen_time);auto_now;type(datetime)"`
}

type OutputInfo struct {
	Id             string
	UserName       string
	Email          string
	Phone          string
	LastLogin      string
	UserType       int8
	Activation     int8
	TotalResNumber int
}

type userTable struct{}

var ut = userTable{}

func GetUser() *userTable {
	return &ut
}

func (urt *userTable) Insert(u *User) error {
	_, err := dbm.Ormer.Insert(u)
	return err
}

func (urt *userTable) Update(u *User, cols ...string) error {
	_, err := dbm.Ormer.Update(u, cols...)
	return err
}

func (urt *userTable) Get(f Filters) (*User, error) {
	var user User

	err := f.Filter(UserTable).One(&user)
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (urt *userTable) GetAll() ([]OutputInfo, error) {
	var users []User
	o := orm.NewOrm()
	qs := o.QueryTable(UserTable)
	_, err := qs.All(&users, "id", "username", "email", "phone", "usertype",
		"activation", "last_login", "total_res_number")
	if err != nil {
		return nil, err
	}
	outputs := []OutputInfo{}
	for _, user := range users {
		output := GetUser().ConvertStruct(&user)
		outputs = append(outputs, *output)
	}
	return outputs, nil
}

func (urt *userTable) Delete(username string) error {
	_, err := Filters{"username": username}.Filter(UserTable).Delete()
	return err
}

func (urt *userTable) DeletebyId(id string) error {
	_, err := Filters{"Id": id}.Filter(UserTable).Delete()
	return err
}

func (urt *userTable) ResetPassword(u *User) error {
	u.Password = setting.DefaultGCMPassword
	u.Activation = defaultAvtivation
	u.MaxRetry = 0
	_, err := dbm.Ormer.Update(u, "Password", "Activation", "MaxRetry")
	return err
}

func (urt *userTable) ConvertStruct(user *User) *OutputInfo {
	output := &OutputInfo{
		Id:             user.Id,
		UserName:       user.UserName,
		Email:          user.Email,
		Phone:          user.Phone,
		UserType:       user.UserType,
		Activation:     user.Activation,
		LastLogin:      user.LastLogin.Format(constants.TimeFormatLayout),
		TotalResNumber: user.TotalResNumber,
	}
	return output
}

var defaultUsername string = "admin"
var defaultUserType int8 = 9
var defaultAvtivation int8 = 0

func InitUserTable() error {
	var users []User
	o := orm.NewOrm()
	qs := o.QueryTable(UserTable)
	_, err := qs.All(&users, "id")
	if err != nil {
		return err
	}
	if len(users) != 0 {
		return nil
	}
	u, _ := uuid.NewUUID()
	ResId := u.String()
	DefaultUser := User{
		Id:         ResId,
		UserName:   defaultUsername,
		Password:   setting.DefaultGCMPassword,
		UserType:   defaultUserType,
		Activation: defaultAvtivation,
	}
	if err := GetUser().Insert(&DefaultUser); err != nil {
		return err
	}
	return nil
}
