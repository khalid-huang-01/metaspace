package lts

import (
	"encoding/json"
	"fleetmanager/api/common/log"
	"fleetmanager/api/model/lts"
	"fleetmanager/api/response"
	service "fleetmanager/api/service/lts"
	"fleetmanager/api/validator"
	"fleetmanager/logger"
	"net/http"

	"github.com/beego/beego/v2/server/web"
)

type LTSController struct {
	web.Controller
}

// 新建日志接入
func (c *LTSController) LtsAccessConfig() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_access_config")

	createReq := lts.CreateAccessConfigReq{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &createReq); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read request body error")
		return
	}
	if err := validator.Validate(&createReq); err != nil {
		tLogger.Error("validate request body err:%+v", err.Error())
		response.ParamsError(c.Ctx, err)
		return
	}
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")

	respMsg, errNew := s.CreateLTSAccessConfig(projectId, createReq)
	if errNew != nil {
		tLogger.Error("create access config err:%+v", errNew)
		response.Error(c.Ctx, http.StatusInternalServerError, errNew.ErrD)
		return
	}
	tLogger.Info("create access config success")
	response.Success(c.Ctx, http.StatusOK, respMsg)
}

// create log transfer
func (c *LTSController) LtsLogTransfer() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_log_transfer")
	createReqTemple := lts.CreateTransferTemplate{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &createReqTemple); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read request body error")
		return
	}
	if err := validator.Validate(&createReqTemple); err != nil {
		tLogger.Error("validate request body err:%+v", err.Error())
		response.ParamsError(c.Ctx, err)
		return
	}
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	createReq := lts.LogTransferReq{
		LogStreamId: createReqTemple.LogStreamId,
		LogGroupId:  createReqTemple.LogGroupId,
		TransferInfo: lts.TransferInfo{
			ObsPeriodUnit:   getPeriodUnit(createReqTemple.TransferDetail.ObsPeriod),
			ObsPeriod:       getPeriodNum(createReqTemple.TransferDetail.ObsPeriod),
			ObsBucketName:   createReqTemple.TransferDetail.ObsBucketName,
			ObsTransferPath: createReqTemple.TransferDetail.ObsTransferPath,
		},
	}
	respMsg, errNew := s.CreateTransfer(projectId, createReq)
	if errNew != nil {
		tLogger.Error("create log transfer err:%s", errNew.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errNew.ErrD)
		return
	}
	tLogger.Info("create log transfer success")
	response.Success(c.Ctx, http.StatusOK, respMsg)

}

// create log group
func (c *LTSController) LtsLogGroup() {
	tLogger := log.GetTraceLogger(c.Ctx).WithField(logger.Stage, "create_log_group")
	createReq := lts.CreateLogGroup{}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &createReq); err != nil {
		response.InputError(c.Ctx)
		tLogger.WithField(logger.Error, err.Error()).Error("read request body error")
		return
	}
	if err := validator.Validate(&createReq); err != nil {
		tLogger.Error("validate request body err:%+v", err.Error())
		response.ParamsError(c.Ctx, err)
		return
	}
	s := service.NewLtsService(c.Ctx, tLogger)
	projectId := c.Ctx.Input.Param(":project_id")
	respMsg, errNew := s.CreateLogGroup(projectId, createReq)
	if errNew != nil {
		tLogger.Error("create log groupr err:%s", errNew.Error())
		response.Error(c.Ctx, http.StatusInternalServerError, errNew.ErrD)
		return
	}
	tLogger.Info("create log group success")
	response.Success(c.Ctx, http.StatusOK, respMsg)

}

func getPeriodUnit(num string) string {
	switch num {
	case "2min":
		return "min"
	case "5min":
		return "min"
	case "30min":
		return "min"
	case "1hour":
		return "hour"
	case "3hour":
		return "hour"
	case "6hour":
		return "hour"
	case "12hour":
		return "hour"
	default:
		return "min"
	}
}

func getPeriodNum(num string) int {
	switch num {
	case "2min":
		return 2
	case "5min":
		return 5
	case "30min":
		return 30
	case "1hour":
		return 1
	case "3hour":
		return 3
	case "6hour":
		return 6
	case "12hour":
		return 12
	default:
		return 2
	}
}
