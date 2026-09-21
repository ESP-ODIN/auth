package model

import "time"

type Role struct {
	ID        int       `db:"id"`
	Code      string    `db:"code"`
	Label     string    `db:"label"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
