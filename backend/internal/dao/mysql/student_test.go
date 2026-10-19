package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"
)

// Called only inside TestAdminManagementIntegration's isolated schema.
func testAdminStudents(t *testing.T, dao *DAO) {
	class := models.Class{College: "学院", Major: "专业", ClassName: "学生测试班", IsEnabled: true}
	apartment := models.Apartment{Name: "测试宿舍", Gender: models.GenderMale, IsEnabled: 1}
	if err := dao.db.Create(&class).Error; err != nil {
		t.Fatal(err)
	}
	if err := dao.db.Create(&apartment).Error; err != nil {
		t.Fatal(err)
	}
	users := []models.User{
		{StudentID: "admin", Role: models.RoleAdmin, Password: "secret", Status: 1},
		{StudentID: "other-admin", Role: models.RoleAdmin, Password: "secret", Status: 1},
		{StudentID: "202600010001", Role: models.RoleStudent, Password: "secret", Status: 1},
		{StudentID: "202600010002", Role: models.RoleStudent, Password: "secret", Status: 0},
		{StudentID: "202600010003", Role: models.RoleStudent, Password: "secret", Status: 1},
	}
	for i := range users {
		if err := dao.db.Create(&users[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	if err := dao.db.Delete(&users[4]).Error; err != nil {
		t.Fatal(err)
	}
	query := dto.StudentQuery{Page: 1, PageSize: 1}
	rows, total, err := dao.GetStudents(query)
	if err != nil || total != 2 || len(rows) != 1 || rows[0].ID != users[2].ID {
		t.Fatalf("list/order/count: %v %d %+v", err, total, rows)
	}
	query.Page = 2
	rows, total, err = dao.GetStudents(query)
	if err != nil || total != 2 || len(rows) != 1 || rows[0].ID != users[3].ID || rows[0].Name != "" || rows[0].ClassID != 0 {
		t.Fatalf("incomplete profile/page: %v %d %+v", err, total, rows)
	}
	input := dto.ProfileRequest{
		Name:            "张三",
		Phone:           "13800138000",
		Gender:          models.GenderMale,
		ClassID:         class.ID,
		ParentName:      "家长",
		ParentPhone:     "13900139000",
		TeacherName:     "老师",
		ApartmentID:     apartment.ID,
		DormitoryNumber: "101",
	}
	for _, id := range []uint{users[0].ID, users[1].ID, users[4].ID, 999999} {
		if _, err := dao.GetStudent(id); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("detail scope %d: %v", id, err)
		}
		if err := saveTestStudent(dao, id, &input); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("edit scope %d: %v", id, err)
		}
		if err := dao.UpdateStudentStatus(id, 0); !errors.Is(err, gorm.ErrRecordNotFound) {
			t.Fatalf("status scope %d: %v", id, err)
		}
	}
	if err := saveTestStudent(dao, users[2].ID, &input); err != nil {
		t.Fatal(err)
	}
	if err := dao.UpdateProfileAvatar(users[2].StudentID, "avatars/test.png"); err != nil {
		t.Fatal(err)
	}
	row, err := dao.GetProfileByStuID(users[2].StudentID)
	if err != nil || row.Name != input.Name || row.ClassID != class.ID || row.ApartmentID != apartment.ID {
		t.Fatalf("profile join: %v %+v", err, row)
	}
	account, err := dao.GetStudent(users[2].ID)
	if err != nil || account.Password != "" {
		t.Fatalf("account query loaded credentials: %v %+v", err, account)
	}
	active := uint8(1)
	query = dto.StudentQuery{Page: 1, PageSize: 10, StudentID: "010001", Name: "张", ClassID: class.ID, Status: &active}
	rows, total, err = dao.GetStudents(query)
	if err != nil || total != 1 || len(rows) != 1 || rows[0].ID != users[2].ID || rows[0].Name != input.Name || rows[0].ClassID != class.ID || rows[0].Gender != input.Gender || rows[0].Phone != input.Phone {
		t.Fatalf("combined filters: %v %d %+v", err, total, rows)
	}
	query.Name = "不存在"
	rows, total, err = dao.GetStudents(query)
	if err != nil || total != 0 || rows == nil || len(rows) != 0 {
		t.Fatalf("empty list: %v %d %+v", err, total, rows)
	}
	if err := dao.db.Model(&class).Update("is_enabled", false).Error; err != nil {
		t.Fatal(err)
	}
	if err := dao.db.Model(&apartment).Update("is_enabled", 0).Error; err != nil {
		t.Fatal(err)
	}
	if err := saveTestStudent(dao, users[2].ID, &input); err != nil {
		t.Fatalf("retain disabled references: %v", err)
	}
	input.Gender = models.GenderMale
	input.ApartmentID = 0
	input.DormitoryNumber = ""
	if err := saveTestStudent(dao, users[2].ID, &input); err != nil {
		t.Fatal(err)
	}
	row, err = dao.GetProfileByStuID(users[2].StudentID)
	if err != nil || row.AvatarURL != "avatars/test.png" || row.ApartmentID != 0 || row.DormitoryNumber != "" {
		t.Fatalf("clear dorm/preserve avatar: %v %+v", err, row)
	}
	classes, err := dao.GetClassOptions(false)
	if err != nil {
		t.Fatal(err)
	}
	apartments, err := dao.GetApartmentOptions("", false)
	if err != nil || len(classes) != 1 || classes[0].IsEnabled || len(apartments) != 1 || apartments[0].IsEnabled != 0 {
		t.Fatalf("disabled options: %v %+v %+v", err, classes, apartments)
	}
	classes, err = dao.GetClassOptions(true)
	if err != nil || len(classes) != 0 {
		t.Fatalf("disabled class excluded from App options: %v %+v", err, classes)
	}
	apartments, err = dao.GetApartmentOptions("", true)
	if err != nil || len(apartments) != 0 {
		t.Fatalf("disabled apartment excluded from App options: %v %+v", err, apartments)
	}
	apartments, err = dao.GetApartmentOptions(models.GenderFemale, false)
	if err != nil || len(apartments) != 0 {
		t.Fatalf("apartment gender filter: %v %+v", err, apartments)
	}
	if err := dao.UpdateStudentStatus(users[2].ID, 0); err != nil {
		t.Fatal(err)
	}
	if err := dao.EnsureUserActive(users[2].StudentID); !errors.Is(err, ErrorUserDisabled) {
		t.Fatalf("disabled authentication: %v", err)
	}
	if err := dao.UpdateStudentStatus(users[2].ID, 1); err != nil {
		t.Fatal(err)
	}
	if err := dao.EnsureUserActive(users[2].StudentID); err != nil {
		t.Fatal(err)
	}
	rollbackErr := errors.New("reject update")
	if err := dao.Transaction(func(tx *gorm.DB) error {
		user, err := dao.GetStudentForUpdate(tx, users[2].ID)
		if err != nil {
			return err
		}
		profile, err := dao.GetStudentProfileForUpdate(tx, user.StudentID)
		if err != nil {
			return err
		}
		if _, err := dao.GetStudentClassForShare(tx, class.ID); err != nil {
			return err
		}
		if _, err := dao.GetStudentApartmentForShare(tx, apartment.ID); err != nil {
			return err
		}
		profile.Name = "rolled back"
		if err := dao.SaveStudentProfile(tx, profile); err != nil {
			return err
		}
		return rollbackErr
	}); !errors.Is(err, rollbackErr) {
		t.Fatalf("transaction error: %v", err)
	}
	row, err = dao.GetProfileByStuID(users[2].StudentID)
	if err != nil || row.Name != input.Name {
		t.Fatalf("transaction rollback: %v %+v", err, row)
	}
	// Deleted profiles are absent in joins, but can be completed again safely.
	if err := dao.db.Model(&models.Profile{}).Where("student_id = ?", users[2].StudentID).Update("deleted_at", time.Now()).Error; err != nil {
		t.Fatal(err)
	}
	row, err = dao.GetProfileByStuID(users[2].StudentID)
	if err != nil || row != nil {
		t.Fatalf("deleted profile lookup: %v %+v", err, row)
	}
	if err := dao.db.Model(&class).Update("is_enabled", true).Error; err != nil {
		t.Fatal(err)
	}
	input.ClassID = class.ID
	if err := saveTestStudent(dao, users[2].ID, &input); err != nil {
		t.Fatal(err)
	}
	row, err = dao.GetProfileByStuID(users[2].StudentID)
	if err != nil || row.Name != input.Name || row.AvatarURL != "avatars/test.png" {
		t.Fatalf("restore profile: %v %+v", err, row)
	}
}

// Exercise persistence separately from service-owned reference validation.
func saveTestStudent(dao *DAO, id uint, input *dto.ProfileRequest) error {
	return dao.Transaction(func(tx *gorm.DB) error {
		user, err := dao.GetStudentForUpdate(tx, id)
		if err != nil {
			return err
		}
		return dao.SaveStudentProfile(tx, &models.Profile{
			StudentID:       user.StudentID,
			Name:            input.Name,
			Phone:           input.Phone,
			Gender:          input.Gender,
			ClassID:         input.ClassID,
			ParentName:      input.ParentName,
			ParentPhone:     input.ParentPhone,
			TeacherName:     input.TeacherName,
			ApartmentID:     input.ApartmentID,
			DormitoryNumber: input.DormitoryNumber,
		})
	})
}
