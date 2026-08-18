package model

import "time"

type Session struct {
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	RoleName  string    `json:"role_name"`
	CreatedAt time.Time `json:"created_at"`
}
