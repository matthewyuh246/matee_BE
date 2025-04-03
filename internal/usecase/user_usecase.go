package usecase

import (
	"errors"
	"log"

	"github.com/matthewyuh246/Matee/internal/domain"
	"github.com/matthewyuh246/Matee/internal/repository"
)

type IUserUsecase interface {
	FindOrCreateUserByGitHub(userInfo *domain.User) (*domain.User, error)
}

type userUsecase struct {
	ur repository.IUserRepository
}

func NewUserUsecase(ur repository.IUserRepository) IUserUsecase {
	return &userUsecase{ur}
}

func (uu *userUsecase) FindOrCreateUserByGitHub(userInfo *domain.User) (*domain.User, error) {
	if userInfo.GithubID == "" {
		return nil, errors.New("github id is empty")
	}

	existing, err := uu.ur.FindByGitHubID(userInfo.GithubID)
	if err != nil {
		log.Fatalln("RecordNotFound")
	}

	if existing != nil {
		existing.Name = userInfo.Name
		existing.Email = userInfo.Email
		existing.AvatarURL = userInfo.AvatarURL
		err = uu.ur.Update(existing)
		if err != nil {
			return nil, err
		}
		return existing, nil
	} else {
		err = uu.ur.Create(userInfo)
		if err != nil {
			return nil, err
		}
		return userInfo, nil
	}
}
