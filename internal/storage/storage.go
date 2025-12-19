package storage

import (
	"context"
)

type Storage interface {
	User() Users
}

type Users interface {
	CreateUser(ctx context.Context)
}
