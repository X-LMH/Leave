package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID          int            `gorm:"primaryKey;autoIncrement;comment:学生记录主键"`
	StudentID   string         `gorm:"uniqueIndex;not null;size:22;comment:学号（唯一标识）"`
	Password    string         `gorm:"not null;size:100;comment:登录密码"`
	Role        string         `gorm:"column:role;type:varchar(10);not null;default:'user';comment:角色"`
	CreatedTime time.Time      `gorm:"autoCreateTime;comment:记录创建时间"`
	DeletedAt   gorm.DeletedAt `gorm:"index;comment:软删除标记"`
}

type Profile struct {
	ID          int            `json:"id" gorm:"primaryKey;autoIncrement;comment:个人信息主键"`
	StudentID   string         `json:"student_id" gorm:"index;not null;comment:关联的学号"`
	Name        string         `json:"name" gorm:"not null;comment:学生姓名"`
	Phone       string         `json:"phone" gorm:"not null;comment:学生电话"`
	Gender      string         `json:"gender" gorm:"not null;comment:性别（男/女）"`
	ParentName  string         `json:"parent_name" gorm:"not null;comment:家长姓名"`
	ParentPhone string         `json:"parent_phone" gorm:"not null;comment:家长电话"`
	Apartment   string         `json:"apartment" gorm:"not null;comment:公寓名称"`
	ApartmentID int            `json:"apartment_id" gorm:"not null;comment:公寓ID"`
	TeacherName string         `json:"teacher_name" gorm:"comment:辅导员姓名"`
	CreateAt    time.Time      `json:"-" gorm:"autoCreateTime;comment:记录创建时间"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index;comment:软删除标记"` // 新增软删除支持
}

type Record struct {
	ID          int            `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement;comment:主键ID" json:"id"`
	StudentID   string         `gorm:"column:student_id;type:varchar(20);not null;index;comment:学号（普通索引）" json:"student_id"`
	Name        string         `gorm:"column:name;type:varchar(10);not null;comment:姓名" json:"name"`
	LeaveType   string         `gorm:"column:leave_type;type:varchar(10);not null;default:'';comment:请假类型" json:"leave_type"`
	LeaveReason string         `gorm:"column:leave_reason;type:text;comment:请假原因" json:"leave_reason"`
	LeaveSchool string         `gorm:"column:leave_school;type:varchar(1);not null;default:'';comment:请假期间是否离校" json:"leave_school"`
	PlaceDetail string         `gorm:"column:place_detail;type:varchar(10);not null;default:'';comment:详细地点" json:"place_detail"`
	LeaveWay    string         `gorm:"column:leave_way;type:varchar(10);not null;default:'';comment:离开方式" json:"leave_way"`
	BackSchool  string         `gorm:"column:back_school;type:varchar(1);not null;default:'';comment:请假期间是否回校住宿" json:"back_school"`
	StartTime   time.Time      `gorm:"column:start_time;type:datetime;not null;comment:请假开始时间（格式：YYYY-MM-DD HH:mm:ss）" json:"start_time"`
	EndTime     time.Time      `gorm:"column:end_time;type:datetime;not null;comment:请假结束时间（格式：YYYY-MM-DD HH:mm:ss）" json:"end_time"`
	Nigao       time.Time      `gorm:"column:nigao;type:datetime;not null;comment:拟稿时间（格式：YYYY-MM-DD HH:mm:ss）" json:"nigao"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:datetime;not null;default:CURRENT_TIMESTAMP;comment:记录创建时间" json:"-"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;type:datetime;index;comment:软删除时间（NULL表示未删除）" json:"-"`
}
