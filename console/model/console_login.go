package model

import (
	"fmt"
	"github.com/cubefs/cubefs/console/cutil"
	"github.com/cubefs/cubefs/util/log"
	"time"
)

type ConsoleUserInfo struct {
	ID         uint64    `gorm:"column:id"`
	User       string    `gorm:"column:user"` // 唯一索引
	Password   string    `gorm:"column:password"`
	Role       int8      `gorm:"column:role"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (ConsoleUserInfo) TableName() string {
	return "console_user"
}

func (table ConsoleUserInfo) InsertConsoleUser(info *ConsoleUserInfo) error {
	if err := cutil.CONSOLE_DB.Table(table.TableName()).Create(&info).Error; err != nil {
		log.LogErrorf("InsertConsoleUser failed: err(%v)", err)
		return err
	}
	return nil
}

func (table ConsoleUserInfo) GetUserInfoByUser(name string) (info *ConsoleUserInfo, err error) {
	if err = cutil.CONSOLE_DB.Table(table.TableName()).Where("user = ?", name).
		Scan(&info).Error; err != nil {
		log.LogErrorf("GetUserInfoByUser failed: user(%v) err(%v)", name, err)
		return nil, err
	}
	return
}

// 获取用户角色 和 校验用户密码 两个方法
func IsAdmin(user string) bool {
	model := &ConsoleUserInfo{}
	userInfo, _ := model.GetUserInfoByUser(user)
	if userInfo != nil {
		return userInfo.Role == 0
	}
	return false
}

func ValidatePassword(user, password string) (passed bool, err error) {
	model := &ConsoleUserInfo{}
	userInfo, err := model.GetUserInfoByUser(user)
	if err != nil {
		return
	}
	// 找不到 和 报错是两回事
	if userInfo != nil {
		return userInfo.Password == password, nil
	}
	err = fmt.Errorf("未识别的用户，请先注册！")
	return
}

type ConsoleAdminOpPassword struct {
	ID         uint64    `gorm:"column:id"`
	User       string    `gorm:"column:user"`
	OpPassword string    `gorm:"column:op_password"`
	CreateTime time.Time `gorm:"column:create_time"`
}

func (ConsoleAdminOpPassword) TableName() string {
	return "admin_operation_password"
}

func (table ConsoleAdminOpPassword) InsertAdminOpPassword(user, password string) error {
	entry := &ConsoleAdminOpPassword{
		User:       user,
		OpPassword: password,
		CreateTime: time.Now(),
	}
	if err := cutil.CONSOLE_DB.Table(table.TableName()).Create(&entry).Error; err != nil {
		log.LogErrorf("InsertAdminOpPassword failed: err(%v)", err)
		return err
	}
	return nil
}

// 要么无 要么有， 要么error  error就返回给用户请重试
func (table ConsoleAdminOpPassword) GetLatestOpPassword(user string) (entry *ConsoleAdminOpPassword, err error) {
	if err = cutil.CONSOLE_DB.Table(table.TableName()).Where("user = ?", user).
		Order("create_time DESC").Limit(1).Scan(&entry).Error; err != nil {
		log.LogErrorf("GetLatestOpPassword failed: user(%v) err(%v)", user, err)
		return nil, err
	}
	return
}

// 是否过期
func IsExpiredOpPassword(key *ConsoleAdminOpPassword) bool {
	if key == nil {
		return true
	}
	if key.CreateTime.Add(time.Hour * 3).Before(time.Now()) {
		return true
	}
	return false
}