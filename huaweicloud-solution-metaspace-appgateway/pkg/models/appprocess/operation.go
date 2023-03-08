// Copyright (c) Huawei Technologies Co., Ltd. 2022-2022. All rights reserved.

// 应用进程操作
package appprocess

import (
	"fmt"
	"strconv"

	"github.com/beego/beego/v2/client/orm"

	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/common"
	"codehub-g.huawei.com/videocloud/mediaprocesscenter/application-gateway/pkg/utils/log"
)

type AppProcessDao struct {
	sqlSession orm.Ormer
}

// NewAppProcessDao new app process dao
func NewAppProcessDao(sqlsession orm.Ormer) *AppProcessDao {
	return &AppProcessDao{sqlSession: sqlsession}
}

// CreateAppProcess create app process
func (a *AppProcessDao) CreateAppProcess(ap *AppProcess) (*AppProcess, error) {
	_, err := a.sqlSession.Insert(ap)
	if err != nil {
		log.RunLogger.Errorf("[app process data service] failed to insert app process %v for %v", ap.ID, err)
		return nil, fmt.Errorf("failed to insert app process into database for %v", err)
	}

	return ap, nil
}

// UpdateAppProcess update app process
func (a *AppProcessDao) UpdateAppProcess(ap *AppProcess) (*AppProcess, error) {
	_, err := a.sqlSession.Update(ap)
	if err != nil {
		log.RunLogger.Errorf("[app process data service] failed to update app process %v for %v", ap.ID, err)
		return nil, fmt.Errorf("failed to update app process for %v", err)
	}

	return ap, nil
}

func (a *AppProcessDao) UpdateAppProcessStateAndUpdatedAt(ap *AppProcess) (*AppProcess, error) {
	_, err := a.sqlSession.Update(ap, FieldNameState, FieldNameUpdatedAt)
	if err != nil {
		log.RunLogger.Errorf("[app process data service] failed to update app process state %v for %v", ap.ID, err)
		return nil, fmt.Errorf("failed to update app process for %v", err)
	}

	return ap, nil
}

// DeleteAppProcess delete app process
func (a *AppProcessDao) DeleteAppProcess(ap *AppProcess) error {
	_, err := a.sqlSession.Delete(ap)
	if err != nil {
		log.RunLogger.Errorf("[app process data service] failed to delete app process %v for %v", ap.ID, err)
		return fmt.Errorf("failed to update app process for %v", err)
	}

	return nil
}

// GetAppProcessByID get an app process by id
func (a *AppProcessDao) GetAppProcessByID(id string) (*AppProcess, error) {
	var ap AppProcess

	cond := orm.NewCondition()
	cond = cond.And(FieldNameProcessID, id)
	err := a.sqlSession.QueryTable(&AppProcess{}).SetCond(cond).One(&ap)

	return &ap, err
}

// GetAppProcessByFleetIDAndInstanceID get app process by fleet id and instance id with sort
func (a *AppProcessDao) GetAppProcessByFleetIDAndInstanceID(fleetID, instanceID, sort string,
	offset, limit int) (*[]AppProcess, error) {
	var aps []AppProcess

	cond := orm.NewCondition()

	if fleetID != "" {
		cond = cond.And(FieldNameFleetID, fleetID)
	}
	if instanceID != "" {
		cond = cond.And(FieldNameInstanceID, instanceID)
	}

	_, err := a.sqlSession.QueryTable(&AppProcess{}).SetCond(cond).OrderBy(sort).
		Offset(offset).Limit(limit).All(&aps)

	return &aps, err
}

// GetAllActiveAppProcess get all active processes
func (a *AppProcessDao) GetAllActiveAppProcess() ([]*AppProcess, error) {
	var aps []*AppProcess
	sqlStr := fmt.Sprintf(`select * from %s where STATE="%s"`,
		TableNameAppProcess, common.AppProcessStateActive)
	_, err := a.sqlSession.Raw(sqlStr).QueryRows(&aps)
	return aps, err
}

// GetAllActiveAppProcess get all active processes
func (a *AppProcessDao) GetAllAvailableAppProcess() ([]*AppProcess, error) {
	var aps []*AppProcess
	sqlStr := fmt.Sprintf("select * from %s where STATE=? AND "+
		"SERVER_SESSION_COUNT <MAX_SERVER_SESSION_NUM", TableNameAppProcess)
	_, err := a.sqlSession.Raw(sqlStr, common.AppProcessStateActive).QueryRows(&aps)
	if err == orm.ErrNoRows {
		log.RunLogger.Infof("there is no available app process")
	}
	return aps, err
}

func (a *AppProcessDao) GetAllAvailableAppProcessByFleetID(fleetID string) ([]*AppProcess, error) {
	var aps []*AppProcess
	sqlStr := fmt.Sprintf("select * from %s where FLEET_ID=? AND STATE=? AND "+
		"SERVER_SESSION_COUNT <MAX_SERVER_SESSION_NUM", TableNameAppProcess)
	_, err := a.sqlSession.Raw(sqlStr, fleetID, common.AppProcessStateActive).QueryRows(&aps)
	if err == orm.ErrNoRows {
		log.RunLogger.Infof("there is no available app process")
	}
	return aps, err
}

// GetAvailableAppProcessByFleetID get available app processes by fleet id
func (a *AppProcessDao) GetAvailableAppProcessByFleetID(fleetID string) (AppProcess, error) {
	var ap AppProcess
	sqlStr := fmt.Sprintf("select * from %s where FLEET_ID=? AND STATE=? AND "+
		"SERVER_SESSION_COUNT <MAX_SERVER_SESSION_NUM", TableNameAppProcess)
	err := a.sqlSession.Raw(sqlStr, fleetID, common.AppProcessStateActive).QueryRow(&ap)
	if err == orm.ErrNoRows {
		log.RunLogger.Infof("there is no available app process")
	}
	return ap, err
}

// GetBusiestAndAvailableAppProcessByFleetID 用于获取可以的process
func (a *AppProcessDao) GetBusiestAndAvailableAppProcessByFleetID(fleetID string) (AppProcess, error) {
	var ap AppProcess
	sqlStr := fmt.Sprintf("select * from %s where FLEET_ID=? AND STATE=? AND SERVER_SESSION_COUNT < "+
		"MAX_SERVER_SESSION_NUM ORDER BY SERVER_SESSION_COUNT/MAX_SERVER_SESSION_NUM DESC LIMIT 1", TableNameAppProcess)

	err := a.sqlSession.Raw(sqlStr, fleetID, common.AppProcessStateActive).QueryRow(&ap)
	if err == orm.ErrNoRows {
		log.RunLogger.Infof("there is no available app process")
	}
	return ap, err
}

// GetProcessOfBusiestAndAvailableInstanceByFleetID 在有空闲server session的情况下，选择instance最繁忙的放置，
// process选择instance里面server session 最少的
func (a *AppProcessDao) GetProcessOfBusiestAndAvailableInstanceByFleetID(fleetID string) (AppProcess, error) {
	var ap AppProcess
	sqlStr := fmt.Sprintf("SELECT * from %s where FLEET_ID=? and STATE=?"+
		"INSTANCE_ID=(SELECT INSTANCE_ID From %s where FLEET_ID=? and STATE=? GROUP BY INSTANCE_ID "+
		"HAVING SUM(SERVER_SESSION_COUNT) < SUM(MAX_SERVER_SESSION_NUM) "+
		"ORDER BY SUM(SERVER_SESSION_COUNT) DESC LIMIT 1) ORDER BY SERVER_SESSION_COUNT/MAX_SERVER_SESSION_NUM "+
		"ASC limit 1", TableNameAppProcess, TableNameAppProcess)

	err := a.sqlSession.Raw(sqlStr, fleetID, common.AppProcessStateActive, fleetID,
		common.AppProcessStateActive).QueryRow(&ap)
	if err == orm.ErrNoRows {
		log.RunLogger.Infof("there is no available app process")
	}
	return ap, err
}

// GetAllFleet get all fleet
func (a *AppProcessDao) GetAllFleet() ([]string, error) {
	var fleets []string

	var rows []orm.Params
	queryStr := fmt.Sprintf("SELECT %s FROM %s GROUP BY %s", FieldNameFleetID, TableNameAppProcess, FieldNameFleetID)
	_, err := a.sqlSession.Raw(queryStr).Values(&rows)
	if err != nil {
		return nil, err
	}

	for i, _ := range rows {
		fleetID, _ := rows[i][FieldNameFleetID].(string)

		fleets = append(fleets, fleetID)
	}

	return fleets, err
}

// ProcessCounts process counts
func (a *AppProcessDao) ProcessCounts(fleetID string) (map[string]int, error) {
	processCounts := map[string]int{}

	var rows []orm.Params
	sqlStr := fmt.Sprintf(`SELECT %s, COUNT(*) as COUNT FROM %s WHERE %s="%s" GROUP BY %s`,
		FieldNameState, TableNameAppProcess, FieldNameFleetID, fleetID, FieldNameState)
	log.RunLogger.Infof(sqlStr)
	_, err := a.sqlSession.Raw(sqlStr).Values(&rows)
	if err != nil {
		return nil, err
	}

	for i, _ := range rows {
		state, _ := rows[i]["STATE"].(string)
		countStr, _ := rows[i]["COUNT"].(string)

		count, err := strconv.Atoi(countStr)
		if err != nil {
			continue
		}
		processCounts[state] = count
	}

	return processCounts, err
}

// VerifyZombieProcess 识别僵尸进程，判断条件是超过90s没有更新状态
func (a *AppProcessDao) VerifyZombieProcess() error {
	sqlStr := fmt.Sprintf("update APP_PROCESS set STATE=? WHERE STATE=? and 90 < TIMESTAMPDIFF(SECOND, " +
		"UPDATED_AT, NOW())")
	rsl, err := a.sqlSession.Raw(sqlStr, common.AppProcessStateError,
		common.AppProcessStateActive).Exec()
	if err != nil {
		return err
	}
	num, err := rsl.RowsAffected()
	if err != nil {
		return err
	}
	log.RunLogger.Infof("[app process dao] verify zombie process, affected %v", num)
	return nil
}

// CleanAppProcess clean app process
func (a *AppProcessDao) CleanAppProcess() error {
	sqlStr := fmt.Sprintf(`update APP_PROCESS SET IS_DELETE = 1 WHERE DATEDIFF(NOW(),CREATED_AT) > 14 and STATE = "TERMINATED"`)
	_, err := a.sqlSession.Raw(sqlStr).Exec()
	if err != nil {
		return err
	}
	return nil

}
