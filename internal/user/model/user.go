package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// User represents the "users" table in Postgres
type User struct {
	bun.BaseModel `bun:"table:users"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"` // primary key
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp"`
	Username  string    `bun:"username,notnull,unique,type:varchar(20)"`
	Password  string    `bun:"password,notnull,type:varchar(255)"`
	Email     string    `bun:"email,notnull,unique,type:varchar(100)"`
}
