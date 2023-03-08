package db

import "time"

const tableNameLtsTransfer = "log_transfer"

type LogTransfer struct {
	Id              string    `orm:"column(id);size(128);pk"`
	ProjectId       string    `orm:"column(project_id);size(128)"`
	LogGroupId      string    `orm:"column(log_group_id);size(128)"`
	LogStreamId     string    `orm:"column(log_stream_id);size(128)"`
	LogTransferId   string    `orm:"column(log_transfer_id);size(128)"`
	ObsPeriodUnit   string    `orm:"column(obs_period_unit);size(128)"`
	ObsPeriod       int       `orm:"column(obs_period)"`
	ObsTransferPath string    `orm:"column(obs_transfer_path);size(128)"`
	ObsBucketName   string    `orm:"column(obs_bucket_name);size(128)"`
	CreateTime      time.Time `orm:"column(create_time);auto_now;type(datetime)"`
	Description     string    `orm:"column(description);size(128)"`
}

type logTransferTable struct{}

var transfertable = logTransferTable{}

func TransferTable() *logTransferTable {
	return &transfertable
}

func (s *logTransferTable) Insert(logTransfer *LogTransfer) error {
	_, err := ormer.Insert(logTransfer)
	return err
}

func (s *logTransferTable) Get(f Filters) (LogTransfer, error) {
	var logTransfer LogTransfer
	err := f.Filter(tableNameLtsTransfer).One(&logTransfer)
	return logTransfer, err
}

func (s *logTransferTable) Update(lts *LogTransfer, cols ...string) error {
	_, err := ormer.Update(lts, cols...)
	return err
}

func (s *logTransferTable) Delete(lts *LogTransfer) error {
	_, err := ormer.Delete(lts)
	return err
}

func (s *logTransferTable) DeleteByTransferId(transferId string) error {
	_, err := Filters{"LogTransferId": transferId}.Filter(tableNameLtsTransfer).Delete()
	return err
}

func (s *logTransferTable) ListAll() ([]LogTransfer, error) {
	var list []LogTransfer
	qs := ormer.QueryTable(tableNameLtsTransfer)
	_, err := qs.All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *logTransferTable) ListbyProjectAndLogStreamId(offset int, limit int, projectId string, LogStreamId string) ([]LogTransfer, error) {
	var list []LogTransfer
	qs := ormer.QueryTable(tableNameLtsTransfer)
	qsf := qs.Filter("ProjectId", projectId)
	if LogStreamId != "" {
		qsf = qsf.Filter("LogStreamId", LogStreamId)
	}
	_, err := qsf.Offset(offset).Limit(limit).All(&list)
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *logTransferTable) Total(projectId string) (int, error) {
	qs := ormer.QueryTable(tableNameLtsTransfer)
	total, err := qs.Filter("ProjectId", projectId).Count()
	if err != nil {
		return 0, err
	}
	return int(total), nil
}
