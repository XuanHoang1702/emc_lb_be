package mapping

import (
	"emc_lb/src/internal/db/sqlc"
	"emc_lb/src/pkg/entities"
)

func ToUserEntity(request entities.RegisterUserRequest, passwordHash string) (entities.User, entities.UserProfile) {
	var phone *string
	if request.Phone != "" {
		phone = &request.Phone
	}

	user := entities.User{
		Email:        request.Email,
		PasswordHash: passwordHash,
	}

	profile := entities.UserProfile{
		UserName: request.UserName,
		Phone:    phone,
	}

	return user, profile
}

func ToRegisterUserResponse(user entities.User, profile entities.UserProfile) entities.RegisterUserResponse {
	response := entities.RegisterUserResponse{
		ID:        user.UUID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	}

	if profile.UserName != "" {
		response.UserName = profile.UserName
	}

	if profile.Phone != nil {
		response.Phone = *profile.Phone
	}

	return response
}

func ToUserEntityFromSqlc(row sqlc.CreateUserRow) entities.User {
	return entities.User{
		ID:        row.ID,
		UUID:      row.Uuid,
		Email:     row.Email,
		CreatedAt: row.CreatedAt,
	}
}
