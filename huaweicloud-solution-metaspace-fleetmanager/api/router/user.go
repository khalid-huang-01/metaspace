package router

import (
	"encoding/json"
	user "fleetmanager/api/controller/user"
	"fleetmanager/api/errors"
	model_user "fleetmanager/api/model/user"
	"fleetmanager/api/response"
	"fleetmanager/db/dao"
	"net/http"

	"github.com/beego/beego/v2/server/web/context"

	"github.com/beego/beego/v2/server/web"
)

func initUserRouter() {
	web.InsertFilter("/v1/user/login", web.BeforeExec, checkActivation)
	web.Router("/v1/user/login", &user.UserController{}, "post:Login")
	web.Router("/v1/user/logout", &user.UserController{}, "post:Logout")
	web.Router("/v1/user/changepass", &user.UserController{}, "put:ChangePass")
	web.Router("/v1/user/resconfig", &user.UserController{}, "get:GetResourceConfig")

	web.Router("/v1/user/userinfo", &user.UserController{}, "get:UserGetInfo")
	web.Router("/v1/user/userinfo", &user.UserController{}, "put:UserReviseInfo")
	web.Router("/v1/user/origininfo", &user.OriginController{}, "get:ListOriginInfo")
	web.Router("/v1/user/support-regions", &user.OriginController{}, "get:ListSupportRegions")

	web.Router("/v1/:project_id/user/bucket", &user.UserController{}, "get:Bucket")
	web.Router("/v1/:project_id/user/image", &user.UserController{}, "get:Image")
	web.Router("/v1/:project_id/user/vpc", &user.UserController{}, "get:Vpc")
	web.Router("/v1/:project_id/user/subnet", &user.UserController{}, "get:Subnet")

	// admin接口过滤器
	web.InsertFilter("/v1/admin/userinfo", web.BeforeExec, checkAdmin)
	web.InsertFilter("/v1/admin/alluser", web.BeforeExec, checkAdmin)
	web.InsertFilter("/v1/admin/resetpw", web.BeforeExec, checkAdmin)
	web.InsertFilter("/v1/admin/resconfig", web.BeforeExec, checkAdmin)

	web.Router("/v1/admin/userinfo", &user.UserController{}, "post:AddUser")
	web.Router("/v1/admin/userinfo", &user.UserController{}, "put:AdminReviseUser")
	web.Router("/v1/admin/userinfo", &user.UserController{}, "delete:AdminDeleteUser")
	web.Router("/v1/admin/alluser", &user.UserController{}, "get:AdminGetAllUser")
	web.Router("/v1/admin/resetpw", &user.UserController{}, "post:AdminResetPassword")
	web.Router("/v1/admin/resconfig", &user.UserController{}, "get:GetAllResourceConfig")
	web.Router("/v1/admin/resconfig", &user.UserController{}, "post:InsertResourceConfig")
	web.Router("/v1/admin/resconfig", &user.UserController{}, "put:UpdateResourceConfig")
	web.Router("/v1/admin/resconfig", &user.UserController{}, "delete:DeleteResConfig")
}

// 验证身份
func checkAdmin(ctx *context.Context) {
	token := ctx.Input.Header("Auth-token")
	claim, err := user.ParseToken(token)
	if err != nil {
		response.Error(ctx.Input.Context, http.StatusUnauthorized, errors.NewError(errors.Unauthorized))
		return
	}
	userinfo, err := dao.GetUser().Get(dao.Filters{"id": claim.UserId})
	if err != nil {
		response.Error(ctx.Input.Context, http.StatusBadRequest, errors.NewError(errors.InvalidUserinfo))
		return
	}
	if userinfo.UserType != model_user.Administrator {
		response.Error(ctx.Input.Context, http.StatusForbidden, errors.NewError(errors.NoPermission))
		return
	}
}

// 登录时根据username查询激活状态
func checkActivation(ctx *context.Context) {
	username := dao.User{}
	if err := json.Unmarshal(ctx.Input.RequestBody, &username); err != nil {
		response.InputError(ctx)
		return
	}
	userinfo, err := dao.GetUser().Get(dao.Filters{"username": username.UserName})
	if err != nil {
		response.Error(ctx.Input.Context, http.StatusBadRequest, errors.NewError(errors.InvalidUserinfo))
		return
	}
	// 冻结状态禁止登录，找管理员解冻并重置密码
	if userinfo.Activation == model_user.STATUS_INACTIVATED {
		response.Error(ctx.Input.Context, http.StatusForbidden, errors.NewError(errors.UserInactivate))
		return
	}

}
