package cluster

import (
	"fmt"
	"github.com/cubefs/cubefs/console/cutil"
	"github.com/cubefs/cubefs/console/model"
	cproto "github.com/cubefs/cubefs/console/proto"
	"github.com/cubefs/cubefs/util/log"
	"time"
)

// 因为model不能引用proto(循环引用)  所以放在service中
func GetSourceUsageInfo(cluster, source, zone string, start, end time.Time) ([][]*cproto.SourceUsedInfo, error) {
	db := cutil.CONSOLE_DB.Table(model.VolumeHistoryTableName).
		Select("source, update_time, SUM(used_gb * dp_replica_num) as total_used").
		Where("update_time >= ? AND update_time <= ?", start, end).
		Where("cluster = ?", cluster)
	if zone != "" {
		db = db.Where(fmt.Sprintf("zone like '%%%s%%' ", zone))
	}
	if source != "" {
		db = db.Where("source = ?", source)
	}

	records := make([]*cproto.SourceUsedInfo, 0)
	err := db.Group("source, update_time").Scan(&records).Error
	if err != nil {
		log.LogErrorf("GetSourceUsageInfo failed: source(%v) zone(%v) err(%v)", source, zone, err)
		return nil, err
	}

	result := make(map[string][]*cproto.SourceUsedInfo, 0)
	for _, record := range records {
		sourceList, ok := result[record.Source]
		if !ok {
			sourceList = make([]*cproto.SourceUsedInfo, 0)
		}
		sourceList = append(sourceList, record)
	}

	data := make([][]*cproto.SourceUsedInfo, 0)
	// todo: 返回结构 source单独提出来
	for _, sourceList := range result {
		data = append(data, sourceList)
	}
	return data, nil
}

func GetSourceList() (sourceList []string, err error) {
	err = cutil.MYSQL_DB.Table("apply_token").
		Select("distinct source").
		Find(&sourceList).Error
	if err != nil {
		log.LogErrorf("GetSourceList failed: err(%v)", err)
	}
	return
}