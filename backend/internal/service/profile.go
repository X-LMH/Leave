package service

import (
	"backend/internal/dao/mysql"
	"backend/internal/models"
)

func Profile(p *models.ParamProfile, studentID string) (err error) {
	student := &models.Profile{
		StudentID:   studentID,
		Name:        p.Name,
		Phone:       p.Phone,
		Gender:      p.Gender,
		ParentName:  p.ParentName,
		ParentPhone: p.ParentPhone,
		Apartment:   p.Apartment,
		ApartmentID: p.ApartmentID,
		TeacherName: p.TeacherName,
	}
	return mysql.FinishProfile(student)
}

func GetProfile(studentID string) (data *models.Profile, err error) {
	return mysql.GetProfileByStuID(studentID)
}

// CreateRecord 待请假记录接口适配新版 Record 模型后实现。
func CreateRecord(_ string, _ *models.ParamRecord) error {
	return mysql.ErrorSubmitRecord
}

func GetRecord(recordID int) (*models.Record, error) {
	return mysql.GetRecordByID(recordID)
}

// GetRecordsList 待请假记录列表响应适配新版 Record 模型后实现。
func GetRecordsList(_ string) ([]*ResRecord, error) {
	return nil, mysql.ErrorRecordNotExist
}

func DeleteRecord(recordID int) error {
	return mysql.DeleteRecordByID(recordID)
}
