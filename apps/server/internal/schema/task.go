package schema

import "time"

type CreateTaskRequest struct {
	Title string `json:"title" binding:"required,min=3,max=200"`
}
type UpdateTaskRequest struct {
	Title  *string `json:"title" binding:"omitempty,min=3,max=200"`
	Status *string `json:"status" binding:"omitempty,oneof=pending completed"`
}
type TaskResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
