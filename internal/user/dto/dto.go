package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/skamranahmed/go-bank/internal/user/model"
)

type CreateUserDto struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
}

func TransformToCreateUserDto(user *model.User) *CreateUserDto {
	return &CreateUserDto{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Username:  user.Username,
	}
}
