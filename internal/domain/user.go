package domain

import "time"

type User struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"type:varchar(255)"`
	Email     string    `json:"email" gorm:"type:varchar(255);uniqueIndex"`
	GithubID  string    `json:"github_id" gorm:"varchar(255);uniqueIndex"`
	AvatarURL string    `json:"avatar_url" gorm:"varchar(255)"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GithubUser struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Login     string `json:"login"`
	AvatarURL string `json:"avatar_url"`
}
