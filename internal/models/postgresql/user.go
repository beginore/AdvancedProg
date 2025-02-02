package postgresql

import (
	"forum/internal/models"
	"gorm.io/gorm"
)

type UserModel struct {
	Db *gorm.DB
}

func (m *UserModel) CreateUser(user *models.User) error {
	return m.Db.Create(user).Error
}

func (m *UserModel) GetUserByID(userID int) (*models.User, error) {
	var user models.User
	if err := m.Db.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (m *UserModel) UpdateUser(user *models.User) error {
	return m.Db.Save(user).Error
}

func (m *UserModel) DeleteUser(userID int) error {
	return m.Db.Delete(&models.User{}, "id = ?", userID).Error
}
