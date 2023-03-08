package user

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/user"
	"fleetmanager/api/response"
	"fleetmanager/api/validator"
	"fleetmanager/db/dao"
	"fleetmanager/logger"
	"fmt"
	"net/http"

	"github.com/beego/beego/v2/client/orm"
	"github.com/google/uuid"
)

type ReturnResConf struct {
	Total       int                       `json:"Total"`
	ResConfList []user.UserResourceConfig `json:"ResConfList"`
}

func (c *UserController) InsertResourceConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "user resource config")
	to, err := orm.NewOrm().Begin()
	if err != nil {
		tLogger.Error("Orm Init err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.ServerInternalError))
		return
	}
	resConf := user.UserResourceConfig{}
	body := c.Ctx.Input.RequestBody
	if err := json.Unmarshal(body, &resConf); err != nil {
		tLogger.Error("Json unmarshal err: %+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	// 校验租户信息
	if err := validator.Validate(&resConf); err != nil {
		tLogger.Error("validate resource conf err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorF(errors.InvalidUserinfo, err.Error()))
		return
	}
	// 所有表用同一个id
	u, _ := uuid.NewUUID()
	ResId := u.String()
	resConf.Id = ResId

	// 插入5个表
	if err := InsertRes(to, ResId, &resConf); err != nil {
		to.Rollback()
		tLogger.Error("Insert resource %+v, request: %s", err.Error(), body)
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.OperateResConfigFailed))
		return
	}
	if err := to.Commit(); err != nil {
		tLogger.Error("Commit err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.ServerInternalError))
		return
	}
	// 更新租户数量
	err = updateTotaResNum(tLogger, resConf.UserId)
	if err != nil {
		tLogger.Error("update total num info err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	response.Success(c.Ctx, http.StatusCreated, resConf)
}

func (c *UserController) UpdateResourceConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "user resource config")
	to, err := orm.NewOrm().Begin()
	if err != nil {
		tLogger.Error("Orm Init err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.ServerInternalError)
		return
	}
	resConf := user.UpdateUserResConfig{}
	body := c.Ctx.Input.RequestBody
	if err := json.Unmarshal(body, &resConf); err != nil {
		tLogger.Error("Json unmarshal err: %+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	// 校验租户信息
	if err := validator.Validate(&resConf); err != nil {
		tLogger.Error("validate resource conf err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorF(errors.InvalidUserinfo, err.Error()))
		return
	}
	// 更新5个表
	if err := UpdateRes(to, body, resConf.Id); err != nil {
		tLogger.Error("Update resource %+v, request: %s", err.Error(), body)
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.OperateResConfigFailed))
		return
	}
	if err := to.Commit(); err != nil {
		tLogger.Error("Commit err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.OperateResConfigFailed))
		return
	}
	response.Success(c.Ctx, http.StatusOK, resConf)
}

func (c *UserController) DeleteResConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "user resource config")
	to, err := orm.NewOrm().Begin()
	if err != nil {
		tLogger.Error("Orm Init err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.ServerInternalError)
		return
	}
	id := c.GetString("id")
	resConf, err := dao.GetUserResConfById(id)
	if err != nil {
		tLogger.Error("get user resource config err: %+v, resource id: %s", err, id)
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.OperateResConfigFailed))
		return
	}
	userid := resConf.Userid
	err = DeleteRes(to, id)
	if err != nil {
		tLogger.Error("Delete resource %+v, request: %s", err.Error(), id)
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.OperateResConfigFailed))
		return
	}
	if err := to.Commit(); err != nil {
		tLogger.Error("Commit err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.OperateResConfigFailed))
		return
	}
	// 租户数量-1
	err = updateTotaResNum(tLogger, userid)
	if err != nil {
		tLogger.Error("update total num info err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
	}
	response.Success(c.Ctx, http.StatusNoContent, errors.NewErrorF(errors.NoError, "Delete success"))
}

// 获取用户的所有租户信息
func (c *UserController) GetResourceConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "user get resource config")
	userid := parseTokenToUserId(c.Ctx)
	userinfo, err := dao.GetUser().Get(dao.Filters{"id": userid})
	if err != nil {
		tLogger.Error("Get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	id := c.GetString("id")
	qureyUser, err := dao.GetUser().Get(dao.Filters{"id": id})
	if err != nil {
		tLogger.Error("Get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	if userinfo.UserType != user.Administrator {
		if id != userid {
			tLogger.Error("userid not match the id in params")
			response.Error(c.Ctx, http.StatusBadRequest, errors.NewError(errors.InvalidUserinfo))
			return
		}
	}
	total, resConfList, err := getResConfig(qureyUser.Id, tLogger)
	if err != nil {
		tLogger.Error("get resconf err:%s", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.ServerInternalError))
		return
	}

	tLogger.Info("get user resource config success")
	returninfo := ReturnResConf{Total: total, ResConfList: resConfList}
	response.Success(c.Ctx, http.StatusOK, returninfo)
}

// 管理员获取所有租户信息
func (c *UserController) GetAllResourceConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "admin get resource config")
	users, err := dao.GetUser().GetAll()
	if err != nil {
		tLogger.Error("Get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.ServerInternalError))
		return
	}
	var resConfList []user.UserResourceConfig
	var total int
	for _, u := range users {
		user_id := u.Id
		count, curUserRes, err := getResConfig(user_id, tLogger)
		total += count
		if err != nil {
			tLogger.Error("Get user err:%+v", err.Error())
			continue
		}
		resConfList = append(resConfList, curUserRes...)
	}

	tLogger.Info("get user resource config success")
	returninfo := ReturnResConf{Total: total, ResConfList: resConfList}
	response.Success(c.Ctx, http.StatusOK, returninfo)
}

// 获取单条resource
func (c *UserController) GetOneResourceConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "get one resource config")
	id := c.GetString("id")
	if id == "" {
		tLogger.Error("GetUserResConfById err:%+v", "resource id is empty")
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorF(errors.InvalidParameterValue, "resource id is empty"))
		return
	}
	resConf := user.UserResourceConfig{}
	userconf, err := dao.GetUserResConfById(id)
	if err != nil {
		tLogger.Error("GetUserResConfById err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.ServerInternalError))
		return
	}
	err = GetResConfigHandler(&resConf, userconf)
	if err != nil {
		tLogger.Error("GetResConfigHandler err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.ServerInternalError))
		return
	}
	tLogger.Info("get user resource config %s success", id)
	response.Success(c.Ctx, http.StatusOK, resConf)
}

func getResConfig(user_id string, tLogger *logger.FMLogger) (int, []user.UserResourceConfig, error) {
	userConfList, err := dao.GetAllResConfig(user_id)
	if err != nil {
		return 0, nil, err
	}
	var resConfList []user.UserResourceConfig
	// 循环访问user信息
	for _, userConf := range userConfList {
		resConf := user.UserResourceConfig{}
		err = GetResConfigHandler(&resConf, &userConf)
		if err != nil {
			tLogger.Error("get resource config err:%s", err.Error())
			continue
		}
		resConf.Id = userConf.Id
		resConf.Username = userConf.Username
		resConf.UserId = user_id
		// 去除敏感信息
		resConf.ResUserPassword = ""
		resConf.KeypairData = ""
		resConfList = append(resConfList, resConf)
	}
	total := len(resConfList)
	return total, resConfList, nil
}
func GetResConfigHandler(resConf *user.UserResourceConfig, userConf *dao.UserResConf) error {
	resConf.Id = userConf.Id
	resConf.Username = userConf.Username
	resConf.OriginProjectId = userConf.OriginProjectId
	resConf.OriginDomainId = userConf.OriginDomainId
	resConf.OriginDomainName = userConf.OriginDomainName
	resConf.Region = userConf.Region
	resAgency, err := dao.GetResAgencyById(resConf.Id)
	if err != nil {
		return err
	}
	resProject, err := dao.GetResProjectById(resConf.Id)
	if err != nil {
		return err
	}
	resDomain, err := dao.GetResDomainIdById(resConf.Id)
	if err != nil {
		return err
	}
	resUser, err := dao.GetResUserById(resConf.Id)
	if err != nil {
		return err
	}
	resKeypair, err := dao.GetResKeypairById(resConf.Id)
	if err != nil {
		return err
	}
	resConf.AgencyName = resAgency.AgencyName
	resConf.IamAgencyName = resAgency.IamAgencyName
	resConf.ResProjectId = resProject.ResProjectId
	resConf.ResDomainId = resDomain.ResDomainId
	resConf.ResDomainName = resDomain.ResDomainName
	resConf.ResUserId = resUser.ResUserId
	resConf.ResUserName = resUser.ResUserName
	resConf.KeypairName = resKeypair.KeypairName
	return nil
}

func DeleteRes(to orm.TxOrmer, indexId string) error {
	// 先删user，因为要根据project查user
	if err := dao.DeleteResUserById(to, indexId); err != nil {
		return fmt.Errorf("resuser err:%+v", err.Error())
	}
	if err := dao.DeleteResProjectById(to, indexId); err != nil {
		return fmt.Errorf("raesproject err:%+v", err.Error())
	}
	if err := dao.DeleteResKeypairById(to, indexId); err != nil {
		return fmt.Errorf("reskeypair err:%+v", err.Error())
	}
	if err := dao.DeleteResAgencyById(to, indexId); err != nil {
		return fmt.Errorf("resagency err:%+v", err.Error())
	}
	if err := dao.DeleteResDomainById(to, indexId); err != nil {
		return fmt.Errorf("resdomain err:%+v", err.Error())
	}
	if err := dao.DeleteUserResConfById(to, indexId); err != nil {
		return fmt.Errorf("resuserconf err:%+v", err.Error())
	}
	return nil
}

func InsertRes(to orm.TxOrmer, id string, resConf *user.UserResourceConfig) error {
	if err := dao.InsertUserResConfByJson(to, id, resConf); err != nil {
		return fmt.Errorf("resuserconf err :%+v", err.Error())
	}
	if err := dao.InsertResUserByJson(to, id, resConf); err != nil {
		return fmt.Errorf("resuser err :%+v", err.Error())
	}
	if err := dao.InsertResProjectByJson(to, id, resConf); err != nil {
		return fmt.Errorf("resproject err :%+v", err.Error())
	}
	if err := dao.InsertResDomainByJson(to, id, resConf); err != nil {
		return fmt.Errorf("resdomain err :%+v", err.Error())
	}
	if err := dao.InsertResKeypairByJson(to, id, resConf); err != nil {
		return fmt.Errorf("reskeypair err :%+v", err.Error())
	}
	if err := dao.InsertResAgencyByJson(to, id, resConf); err != nil {
		return fmt.Errorf("resagency err :%+v", err.Error())
	}
	return nil
}

func UpdateRes(to orm.TxOrmer, body []byte, indexId string) error {
	// 在update user_res_conf中做去重验证
	if err := dao.UpdateUserResConfByJson(to, body, indexId); err != nil {
		return fmt.Errorf("res user res conf err :%+v", err.Error())
	}
	if err := dao.UpdateResUserByJson(to, body, indexId); err != nil {
		return fmt.Errorf("res user err :%+v", err.Error())
	}
	if err := dao.UpdateResProjectByJson(to, body, indexId); err != nil {
		return fmt.Errorf("res project err :%+v", err.Error())
	}
	if err := dao.UpdateResDomainByJson(to, body, indexId); err != nil {
		return fmt.Errorf("res domain err :%+v", err.Error())
	}
	if err := dao.UpdateResKeypairByJson(to, body, indexId); err != nil {
		return fmt.Errorf("res keypair err :%+v", err.Error())
	}
	if err := dao.UpdateResAgencyByJson(to, body, indexId); err != nil {
		return fmt.Errorf("res agency err :%+v", err.Error())
	}
	return nil
}

// 统计并更新用户的租户信息数量
func updateTotaResNum(tLogger *logger.FMLogger, userid string) error {
	count, err := dao.GetUserResConf().Count(dao.Filters{"UserId": userid})
	if err != nil {
		tLogger.Error("get res info err: %+v", err.Error())
		return err
	}
	userinfo, err := dao.GetUser().Get(dao.Filters{"Id": userid})
	if err != nil {
		tLogger.Error("get userinfo info err: %+v", err.Error())
		return err
	}
	userinfo.TotalResNumber = int(count)
	err = dao.GetUser().Update(userinfo, "total_res_number")
	if err != nil {
		tLogger.Error("update userinfo info err: %+v", err.Error())
		return err
	}
	return nil
}
