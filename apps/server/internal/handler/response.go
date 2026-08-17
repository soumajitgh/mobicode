package handler

import (
	"mobicode/apps/server/internal/apperror"
	"mobicode/apps/server/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
)

type envelope struct {
	Data      any    `json:"data"`
	Error     any    `json:"error"`
	RequestID string `json:"request_id"`
}
type errorBody struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Fields  []string `json:"fields"`
}

// Success writes a successful response envelope.
func Success(c *gin.Context, status int, data any) {
	c.JSON(status, envelope{Data: data, Error: nil, RequestID: middleware.RequestID(c)})
}

// Respond maps an application result or error to HTTP.
func Respond(c *gin.Context, err error, data any) {
	if err == nil {
		Success(c, http.StatusOK, data)
		return
	}
	status := http.StatusInternalServerError
	code := apperror.Code(err)
	message := "internal server error"
	switch code {
	case "validation":
		status = http.StatusUnprocessableEntity
		message = err.Error()
	case "not_found":
		status = http.StatusNotFound
		message = err.Error()
	case "conflict":
		status = http.StatusConflict
		message = err.Error()
	}
	c.JSON(status, envelope{Data: nil, Error: errorBody{Code: code, Message: message, Fields: []string{}}, RequestID: middleware.RequestID(c)})
}
