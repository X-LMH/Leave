package service

import (
	"backend/dao/mysql"
	"backend/models"
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

func CreateRecord(studentID string, p *models.ParamRecord) (err error) {
	// 获取姓名
	profile, err := mysql.GetProfileByStuID(studentID)
	if err != nil {
		return err
	}
	student := &models.Record{
		StudentID:   studentID,
		Name:        profile.Name,
		LeaveType:   p.LeaveType,
		LeaveReason: p.LeaveReason,
		LeaveSchool: p.LeaveSchool,
		PlaceDetail: p.PlaceDetail,
		LeaveWay:    p.LeaveWay,
		BackSchool:  p.BackSchool,
		StartTime:   p.StartTime,
		EndTime:     p.EndTime,
		Nigao:       p.Nigao,
	}
	return mysql.InsertRecord(student)
}

func GetRecord(recordID int) (*models.Record, error) {
	return mysql.GetRecordByID(recordID)
}

func GetRecordsList(studentID string) ([]*ResRecord, error) {
	fullRecords, err := mysql.GetRecordsByStuID(studentID)
	if err != nil {
		return nil, err
	}
	responses := make([]*ResRecord, 0, len(fullRecords))

	for _, record := range fullRecords {
		respRecord := &ResRecord{
			ID:    record.ID,
			Name:  record.Name,
			Nigao: record.Nigao,
		}
		responses = append(responses, respRecord)
	}
	return responses, nil
}

func DeleteRecord(recordID int) error {
	return mysql.DeleteRecordByID(recordID)
}
