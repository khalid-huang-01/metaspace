// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// API入口过滤
package entrance

import (
	"fleetmanager/api/common/log"
	user "fleetmanager/api/controller/user"
	"fleetmanager/api/errors"
	model_user "fleetmanager/api/model/user"
	"fleetmanager/api/response"
	"fleetmanager/db/dao"
	"fleetmanager/db/dbm"
	"fleetmanager/logger"
	"fleetmanager/setting"
	"fleetmanager/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/beego/beego/v2/server/web/context"
	"github.com/google/uuid"
)

// 不需要验证会话的有登录和注册
var skipMap map[string]bool = map[string]bool{
	"/v1/user/login": true,
}

const lifttimeMinutes = 30

func generateRequestId() string {
	u, _ := uuid.NewUUID()
	uid := u.String()
	timestamp := time.Now().Unix()
	hostname := utils.GetLocalIP()
	return fmt.Sprintf("%s-%d-%s", uid, timestamp, hostname)
}

// 验证token有效性
func checkSession(ctx *context.Context) error {
	tLogger := log.GetTraceLogger(ctx).WithField(logger.Stage, "check session")
	// token 在header适合api调用
	token := ctx.Input.Header("Auth-token")

	//token是否有效，使用redis中[旧token，新token]的新token验证
	tokenValue, err := dbm.RedisClient.Get(token).Result()
	if err != nil {
		tLogger.Error(err.Error())
		return fmt.Errorf(" get token from redis wrong")
	}
	// 解析token
	claim, err := user.ParseToken(tokenValue)
	if err != nil {
		tLogger.Error(err.Error())
		return fmt.Errorf("Token invalid")
	}

	// token 续期
	if time.Now().Add(lifttimeMinutes * time.Minute).After(time.Unix(claim.ExpiresAt, 0)) {
		lifeTime := setting.JwtTokenLifeTime
		newToken, err := user.GetJWTToken(claim.SessionId, claim.UserId,
			time.Now().Add(time.Second*time.Duration(lifeTime)))
		if err != nil {
			tLogger.Error(err.Error())
			return fmt.Errorf("sessionid or Token invalid")
		}
		dbm.RedisClient.Set(token, newToken, time.Second*time.Duration(lifeTime)).Err()
		ctx.SetCookie("ExpireTime", fmt.Sprint(time.Now().Unix()+int64(setting.JwtTokenLifeTime)))
	}

	tLogger.Info("user validate success")
	return nil
}

// 验证 project_id 是否属于当前用户
func checkProject(ctx *context.Context, project_id string) error {
	token := ctx.Input.Header("Auth-token")
	tokenValue, err := dbm.RedisClient.Get(token).Result()
	if err != nil {
		return fmt.Errorf(" get token from redis wrong")
	}
	claim, err := user.ParseToken(tokenValue)
	if err != nil {
		return fmt.Errorf("token invalid")
	}
	userid := claim.UserId
	userinfo, err := dao.GetUser().Get(dao.Filters{"id": userid})
	if err != nil {
		return fmt.Errorf("DB error")
	}

	// 管理员可以使用所有租户
	if userinfo.UserType == model_user.Administrator {
		return nil
	}
	// 根据参数中的用户id查询是否拥有该租户
	userResConf := dao.UserResConf{}
	err = dao.Filters{"OriginProjectId": project_id,
		"userid": userid}.Filter(dao.UserResConfTable).One(&userResConf)
	if err != nil {
		return err
	}
	return nil
}

// Filter: API入口过滤器
func Filter(ctx *context.Context) {
	requestId := generateRequestId()
	traceLogger := logger.R.WithField(logger.RequestId, requestId)
	URI := ctx.Input.Context.Request.RequestURI
	_, match := skipMap[URI]
	if !match {
		// 除了白名单以外的操作需要验证token
		if err := checkSession(ctx); err != nil {
			response.Error(ctx, http.StatusUnauthorized, errors.NewErrorF(errors.Unauthorized, err.Error()))
			return
		}
	}
	ctx.Input.SetData(logger.StartTime, time.Now())
	ctx.Input.SetData(logger.RequestId, requestId)
	ctx.Input.SetData(logger.TraceLogger, traceLogger)

	project_id := ctx.Input.Param(":project_id")
	if project_id != "" {
		if err := checkProject(ctx, project_id); err != nil {
			response.Error(ctx, http.StatusBadRequest,
				errors.NewErrorF(errors.NoPermission, "check project error"))
			return
		}
	}
}
