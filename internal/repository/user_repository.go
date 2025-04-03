package repository

import (
	"github.com/matthewyuh246/Matee/internal/domain"
	"gorm.io/gorm"
)

type IUserRepository interface {
	FindByGitHubID(githubID string) (*domain.User, error)
	Create(user *domain.User) error
	Update(user *domain.User) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &userRepository{db}
}

func (ur *userRepository) FindByGitHubID(githubID string) (*domain.User, error) {
	var user domain.User
	result := ur.db.Where("github_id = ?", githubID).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}

func (ur *userRepository) Create(user *domain.User) error {
	return ur.db.Create(user).Error
}

func (ur *userRepository) Update(user *domain.User) error {
	return ur.db.Save(user).Error
}
