package models

// users 表。
const (
	RoleStudent              = "student"
	RoleAdmin                = "admin"
	UserStatusDisabled uint8 = 0
	UserStatusActive   uint8 = 1
)

// profiles 表。
const (
	GenderMale   = "男"
	GenderFemale = "女"
)

// leave_types 表。
const (
	LeaveTypeDisabled = false
	LeaveTypeEnabled  = true
)

// records 表。
const (
	LeaveSchoolNo  = false
	LeaveSchoolYes = true
)
