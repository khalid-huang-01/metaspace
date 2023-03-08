package router

import (
	"fleetmanager/api/controller/lts"

	"github.com/beego/beego/v2/server/web"
)

func initLtsRouter() {
	web.Router("/v1/:project_id/lts-access-config", &lts.LTSController{}, "post:LtsAccessConfig")
	web.Router("/v1/:project_id/lts-access-config", &lts.DeleteLTSController{}, "delete:DeleteLtsAccessConfig")
	web.Router("/v1/:project_id/lts-access-config", &lts.QureyLTSController{}, "get:QueryAccessConfig")
	web.Router("/v1/:project_id/lts-access-config", &lts.UpdateLTSController{}, "put:UpdateAccessConfig")
	web.Router("/v1/:project_id/list-lts-access-config", &lts.QureyLTSController{}, "get:ListAccessConfig")

	web.Router("/v1/:project_id/lts-transfer", &lts.LTSController{}, "post:LtsLogTransfer")
	web.Router("/v1/:project_id/lts-transfer", &lts.QureyLTSController{}, "get:QueryLogTransfer")
	web.Router("/v1/:project_id/list-lts-transfer", &lts.QureyLTSController{}, "get:ListLogTransfer")
	web.Router("/v1/:project_id/lts-transfer", &lts.DeleteLTSController{}, "delete:DeleteLtsTransfer")

	web.Router("/v1/:project_id/list-lts-log-group", &lts.QureyLTSController{}, "get:ListLogGroup")
	web.Router("/v1/:project_id/lts-log-group", &lts.LTSController{}, "post:LtsLogGroup")
}
