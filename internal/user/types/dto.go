package types

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

type GetMeDto struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
}

type GetMeResponse struct {
	Data GetMeDto `json:"data"`
}

func TransformToGetMeDto(user *model.User) *GetMeDto {
	return &GetMeDto{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Username:  user.Username,
	}
}

type UpdateUserRequest struct {
	Data UpdateUserData `json:"data" binding:"required"`
}

type UpdateUserData struct {
	Username string `json:"username" binding:"required,min=1"`
}

type UpdateUserDto struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
}

type UpdateUserResponse struct {
	Data UpdateUserDto `json:"data"`
}

func TransformToUpdateUserDto(user *model.User) *UpdateUserDto {
	return &UpdateUserDto{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Username:  user.Username,
	}
}

type UpdatePasswordRequest struct {
	Data UpdatePasswordData `json:"data" binding:"required"`
}

type UpdatePasswordData struct {
	CurrentPassword string `json:"current_password" binding:"required,min=8"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

type UpdatePasswordResponse struct {
	Success bool `json:"success"`
}
