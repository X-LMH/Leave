package mysql

import (
	"backend/internal/dto"
	"backend/internal/models"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func studentQuery(tx *gorm.DB) *gorm.DB {
	// Start from accounts so students without profiles remain visible. Apply the
	// same role and deletion scope to counts, details and paginated rows.
	return tx.Model(&models.User{}).Table("users AS u").Joins("LEFT JOIN profiles AS p ON p.student_id = u.student_id AND p.deleted_at IS NULL").
		Where("u.role = ? AND u.deleted_at IS NULL", models.RoleStudent)
}

func GetStudents(query dto.StudentQuery) ([]*dto.StudentListItem, int64, error) {
	q := studentQuery(db)
	if query.StudentID != "" {
		q = q.Where("u.student_id LIKE ?", "%"+query.StudentID+"%")
	}
	if query.Name != "" {
		q = q.Where("p.name LIKE ?", "%"+query.Name+"%")
	}
	if query.ClassID != 0 {
		q = q.Where("p.class_id = ?", query.ClassID)
	}
	if query.Status != nil {
		q = q.Where("u.status = ?", *query.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]*dto.StudentListItem, 0)
	// Keep profile-only detail fields out of both the query and list response.
	err := q.Select(`
		u.id,
		u.student_id,
		u.status,
		u.app_version,
		u.created_at,
		u.last_seen_at,
		COALESCE(p.name, '') AS name,
		COALESCE(p.class_id, 0) AS class_id,
		COALESCE(p.gender, '') AS gender,
		COALESCE(p.phone, '') AS phone
	`).
		Order("u.created_at ASC, u.id ASC").
		Scopes(Paginate(query.Page, query.PageSize)).
		Scan(&rows).Error
	return rows, total, err
}

func GetStudent(id uint) (*models.User, error) {
	var user models.User
	if err := db.Omit("password").Where("role = ?", models.RoleStudent).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Transaction lets the service keep related reads, validation and writes atomic.
func Transaction(fn func(*gorm.DB) error) error {
	return db.Transaction(fn)
}

func GetStudentForUpdate(tx *gorm.DB, id uint) (*models.User, error) {
	var user models.User
	if err := tx.Omit("password").Clauses(clause.Locking{Strength: "UPDATE"}).Where("role = ?", models.RoleStudent).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func GetStudentProfileForUpdate(tx *gorm.DB, studentID string) (*models.Profile, error) {
	var profile models.Profile
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("student_id = ?", studentID).First(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &models.Profile{}, nil
	}
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func GetStudentClassForShare(tx *gorm.DB, id uint) (*models.Class, error) {
	var class models.Class
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).First(&class, id).Error; err != nil {
		return nil, err
	}
	return &class, nil
}

func GetStudentApartmentForShare(tx *gorm.DB, id uint) (*models.Apartment, error) {
	var apartment models.Apartment
	if err := tx.Clauses(clause.Locking{Strength: "SHARE"}).First(&apartment, id).Error; err != nil {
		return nil, err
	}
	return &apartment, nil
}

func SaveStudentProfile(tx *gorm.DB, profile *models.Profile) error {
	// Preserve avatar and restore soft-deleted profiles without violating student_id uniqueness.
	return tx.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "student_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"name", "phone", "gender", "class_id", "parent_name", "parent_phone",
			"teacher_name", "apartment_id", "dormitory_number", "updated_at", "deleted_at",
		}),
	}).Create(profile).Error
}

func UpdateStudentStatus(id uint, status uint8) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("role = ?", models.RoleStudent).First(&user, id).Error; err != nil {
			return err
		}
		return tx.Model(&user).Update("status", status).Error
	})
}
