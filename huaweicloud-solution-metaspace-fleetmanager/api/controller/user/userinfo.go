package user

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/errors"
	"fleetmanager/api/model/user"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/build"
	"fleetmanager/api/validator"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"fleetmanager/logger"
	"fleetmanager/security"
	"fleetmanager/setting"
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/google/uuid"
)

type UserController struct {
	web.Controller
}

const AuthToken = "Auth-token"

// 用户登录
func (c *UserController) Login() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "user login")
	u := user.UserLogin{}
	body := c.Ctx.Input.RequestBody
	if err := json.Unmarshal(body, &u); err != nil {
		tLogger.Error("Json unmarshal err:%+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	username := u.Username
	passwordGCM, err := transformPass(u.Password, false)
	if err != nil {
		tLogger.Error("Password invalid: %+v", err.Error())
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorF(errors.InvalidUserinfo, err.Error()))
		return
	}
	userinfo, err := dao.GetUser().Get(dao.Filters{"username": username})
	if err != nil {
		tLogger.Error("Get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	if err := checkUser(userinfo, passwordGCM, tLogger); err != nil {
		tLogger.Error("check user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorF(errors.InvalidUserinfo, err.Error()))
		return
	}

	sessionid := c.Ctx.Input.CruSession.SessionID(c.Ctx.Request.Context())
	token, err := GetJWTToken(sessionid, userinfo.Id,
		time.Now().Add(time.Second*time.Duration(setting.JwtTokenLifeTime)))
	if err != nil {
		tLogger.Error("Get JWTToken err:%+v", err.Error())
		response.ParamsError(c.Ctx, err)
		return
	}
	// 登录后设置redis中的值为[token, token]，刷新token时更新value
	if err := dbm.RedisClient.Set(token, token, time.Second*time.Duration(setting.JwtTokenLifeTime)).Err(); err != nil {
		tLogger.Error("Set redis err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.ServerInternalError))
		return
	}
	c.Ctx.SetCookie("sessionid", sessionid)
	c.Ctx.SetCookie("username", username)
	c.Ctx.SetCookie("Auth-token", token)
	c.Ctx.SetCookie("ExpireTime", fmt.Sprint(time.Now().Unix()+int64(setting.JwtTokenLifeTime)))
	tLogger.Info("login success")
	returnInfo := user.LoginResponse{AuthToken: token, Username: username, UserType: userinfo.UserType,
		Activation: userinfo.Activation, UserId: userinfo.Id, TotalResCount: userinfo.TotalResNumber}
	returnInfo.AuthToken = token
	returnInfo.Username = username
	response.Success(c.Ctx, http.StatusCreated, returnInfo)
}

func transformPass(RSAPassword string, checkRegular bool) (string, error) {
	plain_text, err := security.RSA_Decrypt(RSAPassword, setting.RSAPrivateFile)
	if err != nil {
		return "", err
	}
	// 密码正则
	if checkRegular {
		flag := user.CheckPassrordFun(plain_text)
		if !flag {
			return "", fmt.Errorf("password invalid")
		}
	}

	// GCM 加密密文
	passwordGCM, _ := security.GCM_Encrypt(plain_text, setting.GCMKey, setting.GCMNonce)
	return passwordGCM, nil
}

func checkUser(userinfo *dao.User, passwordGCM string, tLogger *logger.FMLogger) error {

	if userinfo.Password == passwordGCM {
		userinfo.LastLogin = time.Now()
		userinfo.MaxRetry = 0
		userinfo.FrozenTime = time.Now()
		dao.GetUser().Update(userinfo, "LastLogin", "MaxRetry", "FrozenTime")
		return nil
	} else {
		// 输入密码错误，尝试次数+1,错误次数大于5次 冻结
		if userinfo.MaxRetry >= user.MaxRetry {
			userinfo.Activation = user.STATUS_INACTIVATED
			userinfo.FrozenTime = time.Now().Add(15 * time.Minute)
			dao.GetUser().Update(userinfo, "Activation", "FrozenTime")
			return fmt.Errorf("user is frozen, please retry after 15 minutes")
		}
		tLogger.Error("user password wrong")
		userinfo.MaxRetry += 1
		dao.GetUser().Update(userinfo, "MaxRetry")
		return fmt.Errorf("user password wrong")
	}
}

// 用户登出
func (c *UserController) Logout() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "user logout")
	c.DestroySession()
	c.Ctx.SetCookie("Auth-token", "")
	c.Ctx.SetCookie("sessionid", "")
	c.Ctx.SetCookie("username", "")
	c.Ctx.SetCookie("ExpireTime", "")
	tLogger.Info("user logout success")
	response.Success(c.Ctx, http.StatusNoContent, errors.NewErrorF(errors.NoError, "Logout success"))
}

// 管理员创建用户
func (c *UserController) AddUser() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "add user")
	u := user.AddUserInfo{}
	body := c.Ctx.Input.RequestBody
	if err := json.Unmarshal(body, &u); err != nil {
		tLogger.Error("Json unmarshal err:%+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	// 校验输入字段
	if err := validator.Validate(&u); err != nil {
		tLogger.Error("validate userinfo err:%+v", err.Error())
		response.ParamsError(c.Ctx, err)
		return
	}

	username := u.Username
	passwordStr, err := transformPass(u.Password, true)
	if err != nil {
		tLogger.Error("Password invalid: %+v", err.Error())
		response.ParamsError(c.Ctx, err)
		return
	}
	// 判断用户是否存在
	_, err = dao.GetUser().Get(dao.Filters{"username": username})
	if err == nil {
		tLogger.Error("username is existed")
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewError(errors.UserExist))
		return
	}
	uuid, _ := uuid.NewUUID()
	uidstring := uuid.String()
	// 插入新用户
	insertFilter := dao.User{
		Id:         uidstring,
		UserName:   username,
		Password:   passwordStr,
		Email:      u.Email,
		Phone:      u.Phone,
		Activation: u.Activation,
		UserType:   user.General_User,
	}
	err = dao.GetUser().Insert(&insertFilter)
	if err != nil {
		tLogger.Error("Insert new user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	insertFilter.Password = ""
	tLogger.Info("Add user success")
	output := dao.GetUser().ConvertStruct(&insertFilter)
	response.Success(c.Ctx, http.StatusCreated, output)
}

// 修改密码
func (c *UserController) ChangePass() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "change passord")
	req := user.ChangePassReq{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &req); err != nil {
		tLogger.Error("Json unmarshal err:%+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	// 获取用户信息，验证旧密码
	userinfo, err := dao.GetUser().Get(dao.Filters{"username": req.Username})
	if err != nil {
		tLogger.Error("Get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.DBError))
		return
	}
	//输入密码与数据库密码不一致
	oldPasswordStrGCM, err := transformPass(req.OldPass, false)
	if err != nil {
		tLogger.Error("parse password wrong")
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewError(errors.PasswordWrong))
		return
	}
	if oldPasswordStrGCM != userinfo.Password {
		tLogger.Error("password wrong")
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewError(errors.PasswordWrong))
		return
	}
	newPassrwordStrGCM, err := transformPass(req.NewPass, true)
	if err != nil {
		tLogger.Error("password parse err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewError(errors.InvalidUserinfo))
		return
	}
	userinfo.Password = newPassrwordStrGCM
	userinfo.Activation = user.STATUS_ACTIVATED
	err = dao.GetUser().Update(userinfo, "password", "activation")
	if err != nil {
		tLogger.Error("Update user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	response.Success(c.Ctx, http.StatusNoContent, errors.NewError(errors.NoError))
	c.Logout()
}

// 用户修改基本信息
func (c *UserController) UserReviseInfo() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "change userinfo")
	changeReq := user.UserChangeInfo{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &changeReq); err != nil {
		tLogger.Error("Json unmarshal err:%+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	// 校验输入用户
	if err := validator.Validate(&changeReq); err != nil {
		tLogger.Error("validate userinfo err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorF(errors.InvalidUserinfo, err.Error()))
		return
	}
	userid := parseTokenToUserId(c.Ctx)
	userinfo, err := dao.GetUser().Get(dao.Filters{"id": userid})
	if err != nil {
		tLogger.Error("get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	// 只更新传送了的字段，changeReq只用于索引与校验，管理员修改同理
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &userinfo); err != nil {
		tLogger.Error("Json unmarshal err:%+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	err = dao.GetUser().Update(userinfo, "email", "phone")
	if err != nil {
		tLogger.Error("Update user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	output := dao.GetUser().ConvertStruct(userinfo)
	tLogger.Info("update userinfo success")
	response.Success(c.Ctx, http.StatusOK, output)
}

// 获取用户信息
func (c *UserController) UserGetInfo() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "get user infomation")
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
	output := dao.GetUser().ConvertStruct(qureyUser)
	tLogger.Info("get userinfo success")
	response.Success(c.Ctx, http.StatusOK, output)
}

// 管理员修改用户信息
func (c *UserController) AdminReviseUser() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "change userinfo")
	changeReq := user.AdminChangeUserInfo{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &changeReq); err != nil {
		tLogger.Error("Json unmarshal err:%+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	// 校验输入用户
	if err := validator.Validate(&changeReq); err != nil {
		tLogger.Error("validate userinfo err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusBadRequest,
			errors.NewErrorF(errors.InvalidUserinfo, err.Error()))
		return
	}
	userinfo, err := dao.GetUser().Get(dao.Filters{"id": changeReq.Id})
	if err != nil {
		tLogger.Error("Get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	// 实际更新的操作
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &userinfo); err != nil {
		tLogger.Error("Json unmarshal err:%+v", err.Error())
		response.InputError(c.Ctx)
		return
	}
	err = dao.GetUser().Update(userinfo, "email", "phone", "activation")
	if err != nil {
		tLogger.Error("Update user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	userinfo, _ = dao.GetUser().Get(dao.Filters{"id": changeReq.Id})
	tLogger.Info("update userinfo success")
	output := dao.GetUser().ConvertStruct(userinfo)
	response.Success(c.Ctx, http.StatusOK, output)
}

// 管理员获取用户列表
func (c *UserController) AdminGetAllUser() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "get all userinfo")
	users, err := dao.GetUser().GetAll()
	if err != nil {
		tLogger.Error("Get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	tLogger.Info("get all userinfo success")
	response.Success(c.Ctx, http.StatusOK, users)
}

// 管理员删除用户
func (c *UserController) AdminDeleteUser() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "delete user")
	id := c.GetString("id")
	err := dao.GetUser().DeletebyId(id)
	if err != nil {
		tLogger.Error("get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	// 删除相关的租户信息
	userConfList, err := dao.GetAllResConfig(id)
	if err != nil {
		tLogger.Error("get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.NewError(errors.DBError))
		return
	}
	to, err := orm.NewOrm().Begin()
	if err != nil {
		tLogger.Error("Orm Init err: %+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errors.ServerInternalError)
		return
	}
	for _, userConf := range userConfList {
		err = DeleteRes(to, userConf.Id)
		if err != nil {
			tLogger.Error("Delete resource %+v, request: %s", err.Error(), id)
			response.Error(c.Ctx, http.StatusInternalServerError,
				errors.NewError(errors.OperateResConfigFailed))
			continue
		}
	}
	if err = to.Commit(); err != nil {
		tLogger.Error("Delete resource %+v, request: %s", err.Error(), id)
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.OperateResConfigFailed))
		return
	}
	tLogger.Info(fmt.Sprintf("delete user %s success", id))
	response.Success(c.Ctx, http.StatusNoContent, errors.NewErrorF(errors.NoError, "Delete user success"))
}

// 管理员重置用户密码
func (c *UserController) AdminResetPassword() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "reset password")
	userid := c.GetString("id")
	userinfo, err := dao.GetUser().Get(dao.Filters{"Id": userid})
	if err != nil {
		tLogger.Error("Get user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusBadRequest, errors.NewErrorF(errors.InvalidUserinfo, err.Error()))
		return
	}
	if err := dao.GetUser().ResetPassword(userinfo); err != nil {
		tLogger.Error("Update user err:%+v", err.Error())
		response.Error(c.Ctx, http.StatusInternalServerError,
			errors.NewError(errors.DBError))
		return
	}
	tLogger.Info(fmt.Sprintf("reset user %s password success", userid))
	response.Success(c.Ctx, http.StatusOK, errors.NewErrorF(errors.NoError, "Reset password success"))

}

func (c *UserController) Bucket() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_build")
	s := service.NewBuildService(c.Ctx, tLogger)
	bucket, err := s.Bucket(c.Ctx)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, err)
		tLogger.WithField(logger.Error, err.Error()).Error("query bucket name from obs error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, bucket)
}

func (c *UserController) Image() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_ims")
	s := service.NewBuildService(c.Ctx, tLogger)
	image, err := s.Image(c.Ctx)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, err)
		tLogger.WithField(logger.Error, err.Error()).Error("query image from IMS error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, image)
}

func (c *UserController) Vpc() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_vpc")
	s := service.NewBuildService(c.Ctx, tLogger)
	vpc, err := s.Vpc(c.Ctx)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, err)
		tLogger.WithField(logger.Error, err.Error()).Error("query vpc error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, vpc)
}

func (c *UserController) Subnet() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "show_subnet")
	s := service.NewBuildService(c.Ctx, tLogger)
	subnet, err := s.Subnet(c.Ctx)
	if err != nil {
		response.Error(c.Ctx, http.StatusInternalServerError, err)
		tLogger.WithField(logger.Error, err.Error()).Error("query subnet error")
		return
	}

	response.Success(c.Ctx, http.StatusOK, subnet)
}

func parseTokenToUserId(Ctx *context.Context) string {
	token := Ctx.Input.Header(AuthToken)
	Claims, err := ParseToken(token)
	if err != nil {
		return ""
	}
	return Claims.UserId
}
