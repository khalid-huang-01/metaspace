// 用户-资源账号表
package dao

import (
	"encoding/json"
	"fleetmanager/api/model/user"
	"fleetmanager/db/dbm"
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
)

type UserResConf struct {
	Id               string    `orm:"column(id);size(64);pk" json:"id"`
	Username         string    `orm:"column(username);size(64)" json:"username"`
	Userid           string    `orm:"column(user_id);size(64)" json:"user_id"`
	OriginDomainId   string    `orm:"column(origin_domain_id);size(64)" json:"origin_domain_id"`
	OriginDomainName string    `orm:"column(origin_domain_name);size(64)" json:"origin_domain_name"`
	OriginProjectId  string    `orm:"column(origin_project_id);size(64)" json:"origin_project_id"`
	Region           string    `orm:"column(region);size(64)" json:"region"`
	CreationTime     time.Time `orm:"column(creation_time);type(datetime);auto_now_add" json:"creation_time"`
}

func BuildUserResConf(resConf *user.UserResourceConfig) *UserResConf {
	resUserResConf := UserResConf{
		Id:               resConf.Id,
		Username:         resConf.Username,
		Userid:           resConf.UserId,
		OriginDomainId:   resConf.OriginDomainId,
		OriginDomainName: resConf.OriginDomainName,
		OriginProjectId:  resConf.OriginProjectId,
		Region:           resConf.Region,
	}
	return &resUserResConf
}

type userResConfTable struct{}

var urcTable = userResConfTable{}

func GetUserResConf() *userResConfTable {
	return &urcTable
}
func (u *userResConfTable) Get(f Filters) (*UserResConf, error) {
	var resConf UserResConf
	err := f.Filter(UserResConfTable).One(&resConf)
	if err != nil {
		return nil, err
	}
	return &resConf, nil
}

func (u *userResConfTable) Count(f Filters) (int64, error) {
	count, err := f.Filter(UserResConfTable).Count()
	return count, err
}

func GetUserResConfByProjectId(originProjectId string) (*UserResConf, error) {
	var u UserResConf
	err := Filters{"OriginProjectId": originProjectId}.Filter(UserResConfTable).One(&u)
	return &u, err
}

func GetUserResConfById(id string) (*UserResConf, error) {
	u := UserResConf{}
	err := Filters{"Id": id}.Filter(UserResConfTable).One(&u)
	return &u, err
}

func InsertUserResConfByJson(to orm.TxOrmer, id string, resConf *user.UserResourceConfig) error {
	r := *BuildUserResConf(resConf)
	userinfo := User{Id: r.Userid}
	err := dbm.Ormer.Read(&userinfo)
	if err != nil {
		return err
	}
	r.Username = userinfo.UserName
	resConf.Username = userinfo.UserName
	r.Id = id
	checkDomainExist := Filters{"OriginDomainId": resConf.OriginDomainId}.Filter(UserResConfTable).Exist()
	checkProjectExist := Filters{"OriginProjectId": resConf.OriginProjectId}.Filter(UserResConfTable).Exist()
	if checkDomainExist || checkProjectExist {
		return fmt.Errorf("OriginDomainId or OriginProjectId exist in %s", UserResConfTable)
	}
	_, err = to.Insert(&r)
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

func UpdateUserResConfByJson(to orm.TxOrmer, request []byte, indexId string) error {
	newConfig := UserResConf{}
	if err := json.Unmarshal(request, &newConfig); err != nil {
		return err
	}
	checkDomainExist := Filters{"OriginDomainId": newConfig.OriginDomainId}.Filter(UserResConfTable).Exist()
	checkProjectExist := Filters{"OriginProjectId": newConfig.OriginProjectId}.Filter(UserResConfTable).Exist()
	if checkDomainExist || checkProjectExist {
		return fmt.Errorf("OriginDomainId or OriginProjectId exist in %s", UserResConfTable)
	}
	_, err := to.Update(&newConfig, "OriginDomainId", "OriginDomainName", "OriginProjectId")
	if err != nil {
		_ = to.Rollback()
		return err
	}
	return nil
}

func DeleteUserResConfById(to orm.TxOrmer, indexId string) error {
	userconf := UserResConf{}
	err := Filters{"Id": indexId}.Filter(UserResConfTable).One(&userconf)
	if err != nil {
		return err
	}
	_, err = to.Delete(&userconf)
	if err != nil {
		_ = to.Rollback()
		return fmt.Errorf("delete user resconf err:%s", err.Error())
	}
	return nil
}

func DeleteUserResConf(originProjectId string, region string) error {
	_, err := Filters{"OriginProjectId": originProjectId, "Region": region}.Filter(UserResConfTable).Delete()
	return err
}

func GetAllResConfig(user_id string) ([]UserResConf, error) {
	var userResConfList []UserResConf
	o := orm.NewOrm()
	qs := o.QueryTable(UserResConfTable)
	_, err := qs.Filter("user_id", user_id).All(&userResConfList)
	if err != nil {
		return nil, err
	}
	return userResConfList, err
}

func GetAllResConfigByMap(queryParams map[string]string) ([]UserResConf, error) {
	var userResConfList []UserResConf
	o := orm.NewOrm()
	cond := orm.NewCondition()
	for key, value := range queryParams {
		if value != "" {
			cond = cond.And(key, value)
		}
	}
	_, err := o.QueryTable(UserResConfTable).SetCond(cond).All(&userResConfList)
	if err != nil {
		return nil, err
	}
	return userResConfList, err
}
