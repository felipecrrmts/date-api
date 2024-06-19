package handler

import (
	"context"

	"github.com/muzzapp/date-api/internal/users"
)

type Users interface {
	CreateUser(ctx context.Context) (*users.User, error)
	Login(ctx context.Context, email, password string) (*users.User, error)
	Discover(ctx context.Context, ID, minAge, maxAge int32, gender string, ranked bool) ([]*users.Profile, error)
	Swipe(ctx context.Context, ID, swipedID int32, ok bool) (bool, error)
}
