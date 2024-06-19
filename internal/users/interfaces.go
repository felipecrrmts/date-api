package users

import "context"

//go:generate mockgen -source=interfaces.go -destination=interfaces_mock.go -package=users

type Store interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUser(ctx context.Context, ID int32) (*User, error)
	GetRankByIDs(ctx context.Context, IDs []int32) (*Rank, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	Discover(ctx context.Context, ID, minAge, maxAge int32, gender string, IDs []int32, location *Location, rank *Rank) ([]*Profile, error)
	Swipe(ctx context.Context, ID int32, swipe *Swipe) error
	Match(ctx context.Context, ID, swipedID int32) (bool, error)
}
