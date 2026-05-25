package mapping

import (
	"emc_lb/src/pkg/entities"

	"github.com/google/uuid"
)

func ToUserEntity(request entities.RegisterUserRequest, passwordHash string) entities.User {
	var userName *string
	if request.UserName != "" {
		userName = &request.UserName
	}

	var phone *string
	if request.Phone != "" {
		phone = &request.Phone
	}

	return entities.User{
		ID:           uuid.New(),
		Email:        request.Email,
		PasswordHash: passwordHash,
		UserName:     *userName,
		Phone:        phone,
	}
}

func ToRegisterUserResponse(user entities.User) entities.RegisterUserResponse {
	response := entities.RegisterUserResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	if user.UserName != "" {
		response.UserName = user.UserName
	}

	if user.Phone != nil {
		response.Phone = *user.Phone
	}

	return response
}
