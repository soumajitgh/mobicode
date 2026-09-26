package resolver

import (
	"encoding/base64"
	"strconv"

	"github.com/vektah/gqlparser/v2/gqlerror"

	"github.com/soumajitgh/mobicode/internal/graphql/model"
	"github.com/soumajitgh/mobicode/internal/store/repository"
)

func graphUser(user repository.MobileIdentity) *model.User {
	id := base64.RawURLEncoding.EncodeToString([]byte("User:" + strconv.FormatUint(uint64(user.UserID), 10)))
	return &model.User{ID: id, Email: user.Email}
}

func coded(code, message string) error {
	return &gqlerror.Error{Message: message, Extensions: map[string]any{"code": code}}
}
