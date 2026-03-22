package model

import "time"

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Timezone  string    `json:"timezone"`
	CreatedAt time.Time `json:"created_at"`
}
