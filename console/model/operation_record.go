package model

import (
	"fmt"
	"github.com/cubefs/cubefs/console/cutil"
	"github.com/cubefs/cubefs/util/log"
	"time"
)

type NodeOperationRecord struct {
	ClusterName string    `json:"clusterName" gorm:"column:cluster_name"`
	Module      string    `json:"module" gorm:"column:module"`
	NodeAddr    string    `json:"nodeAddr" gorm:"column:node_addr"`
	Operation   string    `json:"operation" gorm:"column:operation"`
	Params      string    `json:"params" gorm:"column:params"`
	Pin         string    `json:"pin" gorm:"column:pin"`
	CreateTime  time.Time `json:"-" gorm:"column:create_time"`
}

func (NodeOperationRecord) TableName() string {
	return "node_operation_record"
}

func NewNodeOperation(cluster, module, nodeAddr, operation, pin string, param ...string) *NodeOperationRecord {
	record := &NodeOperationRecord{
		ClusterName: cluster,
		Module:      module,
		NodeAddr:    nodeAddr,
		Operation:   operation,
		Pin:         pin,
		CreateTime:  time.Now(),
	}
	if param != nil {
		record.Params = param[0]
	}
	return record
}

func (table NodeOperationRecord) InsertRecord(record *NodeOperationRecord) error {
	if err := cutil.CONSOLE_DB.Table(table.TableName()).Create(&record).Error; err != nil {
		log.LogErrorf("InsertRecord failed: err: %v", err)
		return err
	}
	return nil
}

func (table NodeOperationRecord) BatchInsertRecord(records []*NodeOperationRecord) error {
	if err := cutil.CONSOLE_DB.Table(table.TableName()).CreateInBatches(&records, len(records)).Error; err != nil {
		log.LogErrorf("BatchInsertRecord failed: err: %v", err)
		return err
	}
	return nil
}

func (table NodeOperationRecord) GetNodeRecord() []*NodeOperationRecord {
	return nil
}

type VolumeOperationRecord struct {
	ClusterName string    `json:"clusterName" gorm:"column:cluster_name"`
	VolName     string    `json:"volName" gorm:"vol_name"`
	Operation   string    `json:"operation" gorm:"operation"`
	Params      string    `json:"params" gorm:"params"`
	Pin         string    `json:"pin" gorm:"pin"`
	CreateTime  time.Time `json:"-" gorm:"column:create_time"`
}

func (VolumeOperationRecord) TableName() string {
	return "volume_operation_record"
}

func NewVolumeOperation(cluster, volume, operation, pin string, params ...string) *VolumeOperationRecord {
	record := &VolumeOperationRecord{
		ClusterName: cluster,
		VolName:     volume,
		Operation:   operation,
		Pin:         pin,
		CreateTime:  time.Now(),
	}
	if params != nil {
		record.Params = params[0]
	}
	return record
}

func (table VolumeOperationRecord) InsertRecord(record *VolumeOperationRecord) error {
	if err := cutil.CONSOLE_DB.Table(table.TableName()).Create(&record).Error; err != nil {
		log.LogErrorf("InsertRecord failed: err: %v", err)
		return err
	}
	return nil
}

const (
	CliOperationStatusFailed = iota
	CliOperationStatusSuccess
)

// Only insert successful record
type CliOperationRecord struct {
	ClusterName string    `gorm:"column:cluster_name"`
	Module      string    `gorm:"column:module"`
	VolName     string    `gorm:"column:vol_name"`
	Operation   string    `gorm:"column:operation"`
	Params      string    `gorm:"column:params"`
	Pin         string    `gorm:"column:pin"`
	TicketID    uint64    `gorm:"column:ticket_id"`
	Status      int       `gorm:"column:status;default:-1"` //成功-1 失败-0
	ErrMsg      string    `gorm:"column:message"`
	CreateTime  time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
	UpdateTime  time.Time `gorm:"column:update_time"`
}

func (CliOperationRecord) TableName() string {
	return "cli_operation_record"
}

func NewCliOperation(cluster, module, operation, pin, params string, ticketID uint64, status int) *CliOperationRecord {
	record := &CliOperationRecord{
		ClusterName: cluster,
		Module:      module,
		Operation:   operation,
		Pin:         pin,
		Params:      params,
		TicketID:    ticketID,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	return record
}

func (table CliOperationRecord) SetCliRecordVolume(volume string) {
	if volume != "" {
		table.VolName = volume
	}
}

func (table CliOperationRecord) SetErrMsg(err error) {
	if err != nil {
		table.ErrMsg = err.Error()
	}
}

func (table CliOperationRecord) InsertRecord(record *CliOperationRecord) error {
	if err := cutil.CONSOLE_DB.Table(table.TableName()).Create(&record).Error; err != nil {
		log.LogErrorf("InsertRecord failed: err: %v", err)
		return err
	}
	return nil
}

func (table CliOperationRecord) UpdateRecordStatus(ticketID uint64, status int) {
	if err := cutil.CONSOLE_DB.Table(table.TableName()).Where("ticket_id = ?", ticketID).
		Updates(map[string]interface{}{
			"status":      status,
			"update_time": time.Now(),
		}).Error; err != nil {
		log.LogErrorf("UpdateRecordStatus failed: ticketID[%v] err: %v", ticketID, err)
	}
}

func (table CliOperationRecord) LoadRecordByTicketID(ticketID uint64) *CliOperationRecord {
	record := new(CliOperationRecord)
	res := cutil.CONSOLE_DB.Table(table.TableName()).Where("ticket_id = ?", ticketID).
		Limit(1).Scan(&record)
	if res.Error != nil {
		log.LogErrorf("LoadRecordByTicketID failed: ticketID[%v], err(%v)", ticketID, res.Error)
		return nil
	}
	if res.RowsAffected == 0 {
		log.LogErrorf("LoadRecordByTicketID: can't find apply record by ticketID[%v]", ticketID)
		return nil
	}
	return record
}

func (table CliOperationRecord) GetOperationRecordCount(cluster, module, operation string) (count int64, err error) {
	err = cutil.CONSOLE_DB.Table(table.TableName()).
		Where("cluster_name = ? AND module = ? AND operation = ?", cluster, module, operation).
		Count(&count).Error
	if err != nil {
		log.LogErrorf("GetRecordCount failed: cluster(%v) module(%v) operation(%v) err(%v)", cluster, module, operation, err)
	}
	return
}

func (table CliOperationRecord) GetOperationRecord(cluster, module, operation string, pageNum, pageSize int) (results []*CliOperationRecord, err error) {
	err = cutil.CONSOLE_DB.Table(table.TableName()).
		Where("cluster_name = ? AND module = ? AND operation = ?", cluster, module, operation).
		Order("create_time DESC").
		Offset((pageNum - 1) * pageSize).
		Limit(pageSize).
		Find(&results).Error
	if err != nil {
		log.LogErrorf("GetOperationRecord failed: cluster(%v) module(%v) operation(%v) err(%v)", cluster, module, operation, err)
	}
	return
}

// todo: 索引
type KeyValueOperation struct {
	ID             uint64    `gorm:"column:id"`
	Module         string    `gorm:"column:module"`
	URI            string    `gorm:"column:url_path"`
	SparkSupport   int       `gorm:"column:spark_support"`
	ReleaseSupport int       `gorm:"column:release_support"`
	CreateTime     time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
}

func (KeyValueOperation) TableName() string {
	return "cli_keyvalue_operation_path"
}

func (table KeyValueOperation) String() string {
	return fmt.Sprintf("%s: supportSpark:%v supportRelease:%v\n", table.URI, table.SparkSupport, table.ReleaseSupport)
}

func (table KeyValueOperation) GetOperationByCluster(cluster string, isRelease bool) ([]*KeyValueOperation, error) {
	operations := make([]*KeyValueOperation, 0)
	db := cutil.CONSOLE_DB.Table(table.TableName())
	if isRelease {
		db.Where("release_support = ?", 1)
	} else {
		db.Where("spark_support = ?", 1)
	}
	err := db.Scan(&operations).Error
	if err != nil {
		return nil, err
	}
	return operations, nil
}

func (table KeyValueOperation) GetOperation(id int) (*KeyValueOperation, error) {
	operation := new(KeyValueOperation)
	result := cutil.CONSOLE_DB.Table(table.TableName()).Where("id = ?", id).First(&operation)
	if result.RowsAffected > 0 {
		return operation, nil
	}
	return nil, result.Error

}

type KeyValuePathParams struct {
	Id         uint64    `gorm:"column:id"`
	PathID     uint64    `gorm:"column:path_id"`
	ValueName  string    `gorm:"column:value_name"`
	ValueType  string    `gorm:"column:value_type"`
	CreateTime time.Time `gorm:"column:create_time;default:CURRENT_TIMESTAMP"`
}

func (KeyValuePathParams) TableName() string {
	return "cli_keyvalue_path_params"
}

func (table KeyValuePathParams) GetPathParams(pathId uint64, isRelease bool) ([]*KeyValuePathParams, error) {
	params := make([]*KeyValuePathParams, 0)
	db := cutil.CONSOLE_DB.Table(table.TableName()).Where("path_id = ?", pathId)
	if isRelease {
		db.Where("release_support = ?", 1)
	} else {
		db.Where("spark_support = ?", 1)
	}
	err := db.Find(&params).Error
	if err != nil {
		return nil, err
	}
	return params, nil
}