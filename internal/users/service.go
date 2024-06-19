package users

import (
	"context"
	"errors"
	"log/slog"

	"github.com/brianvoe/gofakeit/v7"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	store Store
	faker *gofakeit.Faker

	fakeUserFunc func(faker *gofakeit.Faker) *User
}

func NewService(faker *gofakeit.Faker, store Store) *Service {
	return &Service{
		store: store,
		faker: faker,
	}
}

func (s *Service) CreateUser(ctx context.Context) (*User, error) {
	var password string
	user := s.newFakeUser()
	password, user.Password = user.Password, hashPassword(user.Password)

	createdUser, err := s.store.CreateUser(ctx, user)
	if err != nil {
		slog.Error("create user", "err", err)
		return nil, err
	}
	createdUser.Password = password
	return createdUser, nil
}

func (s *Service) newFakeUser() *User {
	if s.fakeUserFunc == nil {
		return NewFakeUser(s.faker)
	}
	return s.fakeUserFunc(s.faker)
}

func hashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		slog.Warn("hash password", "err", err)
		return ""
	}
	return string(bytes)
}

func (s *Service) Login(ctx context.Context, email, password string) (*User, error) {
	foundUser, err := s.store.GetUserByEmail(ctx, email)
	if err != nil {
		slog.Error("login GetUserByEmail", "email", email, "err", err)
		return nil, err
	}
	if !verifyPassword(foundUser.Password, password) {
		return nil, ErrPasswordMismatch
	}
	return foundUser, nil
}

func verifyPassword(hashedPassword, password string) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return false
	}
	return true
}

func (s *Service) Discover(ctx context.Context, ID, minAge, maxAge int32, gender string, ranked bool) ([]*Profile, error) {
	user, err := s.store.GetUser(ctx, ID)
	if err != nil {
		slog.Error("Discover GetUser", "ID", ID, "err", err)
		return nil, err
	}
	swipeIDs := user.swipeIDs()
	rank, err := s.rankedDiscover(ctx, ranked, swipeIDs)
	if err != nil {
		slog.Error("Discover rankedDiscover", "ranked", ranked, "swipeIDs", swipeIDs, "err", err)
		return nil, err
	}

	profiles, err := s.store.Discover(ctx, ID, minAge, maxAge, gender, swipeIDs, user.Location, rank)
	if err != nil {
		slog.Error("Discover",
			"ID", ID, "minAge", minAge, "maxAge", maxAge, "gender", gender, "swipeIDs", swipeIDs,
			"Coordinates", user.Location.CoordinatesFloat64Slice(), "ranked", ranked, "err", err)
		return nil, err
	}
	return profiles, nil
}

func (s *Service) rankedDiscover(ctx context.Context, ranked bool, swipeIDs []int32) (*Rank, error) {
	if !ranked || len(swipeIDs) == 0 {
		return nil, nil
	}
	return s.store.GetRankByIDs(ctx, swipeIDs)
}

func (s *Service) Swipe(ctx context.Context, ID, swipedID int32, ok bool) (bool, error) {
	_, err := s.store.GetUser(ctx, ID)
	if err != nil {
		if !errors.Is(err, ErrUserNotFound) {
			slog.Error("Swipe GetUser", "ID", ID, "err", err)
		}
		return false, err
	}
	if err = s.store.Swipe(ctx, ID, &Swipe{ID: swipedID, OK: ok}); err != nil {
		slog.Error("Swipe", "ID", ID, "swipedID", swipedID, "err", err)
		return false, err
	}
	if ok {
		ok, err = s.store.Match(ctx, ID, swipedID)
		if err != nil {
			slog.Error("Swipe Match", "ID", ID, "swipedID", swipedID, "err", err)
			return false, err
		}
	}
	return ok, nil
}
