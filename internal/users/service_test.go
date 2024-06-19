package users

import (
	"context"
	"testing"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestService_CreateUser(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	// setUp
	store := NewMockStore(controller)
	faker := gofakeit.New(10)
	userService := NewService(faker, store)
	ctx := context.Background()

	t.Run("when db create fails should return an error", func(t *testing.T) {
		// given
		user := NewFakeUser(faker)
		userService.fakeUserFunc = func(f *gofakeit.Faker) *User {
			return user
		}
		store.EXPECT().CreateUser(ctx, user).Return(nil, ErrInsertUser)

		// when
		_, err := userService.CreateUser(ctx)
		require.Error(t, err)
		// then
		require.ErrorIs(t, err, ErrInsertUser)
	})

	t.Run("successful user creation", func(t *testing.T) {
		// given
		user := NewFakeUser(faker)
		originalPassword := user.Password
		user.Password = hashPassword(user.Password)
		userService.fakeUserFunc = func(f *gofakeit.Faker) *User {
			return user
		}
		store.EXPECT().CreateUser(ctx, user).Return(user, nil)

		// when
		actualUser, err := userService.CreateUser(ctx)
		require.NoError(t, err)

		//  then
		user.Password = originalPassword
		require.Equal(t, user, actualUser)
	})
}

func TestService_Login(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	// setUp
	store := NewMockStore(controller)
	faker := gofakeit.New(10)
	userService := NewService(faker, store)
	ctx := context.Background()

	t.Run("password mismatch", func(t *testing.T) {
		user := NewFakeUser(faker)
		user.Password = hashPassword(user.Password)
		userService.fakeUserFunc = func(f *gofakeit.Faker) *User {
			return user
		}
		store.EXPECT().GetUserByEmail(ctx, user.Email).Return(user, nil)

		loggedUser, err := userService.Login(ctx, user.Email, "diff.password")
		require.ErrorIs(t, err, ErrPasswordMismatch)
		require.Nil(t, loggedUser)
	})

	t.Run("user email not found in db", func(t *testing.T) {
		user := NewFakeUser(faker)
		userService.fakeUserFunc = func(f *gofakeit.Faker) *User {
			return user
		}
		store.EXPECT().GetUserByEmail(ctx, "some.email@gmail.com").Return(nil, ErrUserNotFound)

		loggedUser, err := userService.Login(ctx, "some.email@gmail.com", user.Password)
		require.ErrorIs(t, err, ErrUserNotFound)
		require.Nil(t, loggedUser)
	})

	t.Run("successful user login", func(t *testing.T) {
		user := NewFakeUser(faker)
		originalPassword := user.Password
		user.Password = hashPassword(user.Password)
		userService.fakeUserFunc = func(f *gofakeit.Faker) *User {
			return user
		}
		store.EXPECT().GetUserByEmail(ctx, user.Email).Return(user, nil)

		loggedUser, err := userService.Login(ctx, user.Email, originalPassword)
		require.NoError(t, err)
		require.Equal(t, user, loggedUser)
	})
}

func TestService_Discover(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	// setUp
	store := NewMockStore(controller)
	faker := gofakeit.New(10)
	userService := NewService(faker, store)
	ctx := context.Background()

	t.Run("successful discover all filters", func(t *testing.T) {
		// given
		fiftyUsers := createFiftyUsers(faker)
		user := fiftyUsers[0]
		swipes := []*Swipe{
			{ID: fiftyUsers[3].ID, OK: false},
			{ID: fiftyUsers[5].ID, OK: false},
			{ID: fiftyUsers[8].ID, OK: true},
		}
		user.Swipes = swipes
		rank := &Rank{AvgAge: 40, MostCommonGender: "female"}
		ID := fiftyUsers[0].ID
		discoveredProfiles := usersToProfiles(fiftyUsers[1:])
		store.EXPECT().GetUser(ctx, ID).Return(user, nil)
		store.EXPECT().GetRankByIDs(ctx, user.swipeIDs()).Return(rank, nil)
		store.EXPECT().Discover(ctx, ID, int32(20), int32(40), "female", user.swipeIDs(), user.Location, rank).
			Return(discoveredProfiles, nil)

		// when
		profiles, err := userService.Discover(ctx, ID, 20, 40, "female", true)
		require.NoError(t, err)

		//  then
		require.Equal(t, discoveredProfiles, profiles)
	})

	t.Run("successful discover min and max filter", func(t *testing.T) {
		// given
		fiftyUsers := createFiftyUsers(faker)
		user := fiftyUsers[0]
		ID := fiftyUsers[0].ID
		discoveredProfiles := usersToProfiles(fiftyUsers[1:])
		store.EXPECT().GetUser(ctx, ID).Return(user, nil)
		store.EXPECT().Discover(ctx, ID, int32(20), int32(40), "", []int32{}, user.Location, nil).
			Return(discoveredProfiles, nil)

		// when
		profiles, err := userService.Discover(ctx, ID, 20, 40, "", false)
		require.NoError(t, err)

		//  then
		require.Equal(t, discoveredProfiles, profiles)
	})

	t.Run("successful discover gender filter", func(t *testing.T) {
		// given
		fiftyUsers := createFiftyUsers(faker)
		user := fiftyUsers[0]
		ID := fiftyUsers[0].ID
		discoveredProfiles := usersToProfiles(fiftyUsers[1:])
		store.EXPECT().GetUser(ctx, ID).Return(user, nil)
		store.EXPECT().Discover(ctx, ID, int32(0), int32(0), "male", []int32{}, user.Location, nil).
			Return(discoveredProfiles, nil)

		// when
		profiles, err := userService.Discover(ctx, ID, int32(0), int32(0), "male", false)
		require.NoError(t, err)

		//  then
		require.Equal(t, discoveredProfiles, profiles)
	})

	t.Run("successful discover no filter", func(t *testing.T) {
		// given
		fiftyUsers := createFiftyUsers(faker)
		user := fiftyUsers[0]
		ID := fiftyUsers[0].ID
		discoveredProfiles := usersToProfiles(fiftyUsers[1:])
		store.EXPECT().GetUser(ctx, ID).Return(user, nil)
		store.EXPECT().Discover(ctx, ID, int32(0), int32(0), "", []int32{}, user.Location, nil).
			Return(discoveredProfiles, nil)

		// when
		profiles, err := userService.Discover(ctx, ID, int32(0), int32(0), "", false)
		require.NoError(t, err)

		//  then
		require.Equal(t, discoveredProfiles, profiles)
	})
}

func createFiftyUsers(f *gofakeit.Faker) []*User {
	users := make([]*User, 50)
	for i := range 50 {
		u := NewFakeUser(f)
		u.ID = int32(i)
		users[i] = u
	}
	return users
}

func usersToProfiles(users []*User) []*Profile {
	ps := make([]*Profile, len(users))
	for i, u := range users {
		ps[i] = userToProfile(u)
	}
	return ps
}

func userToProfile(u *User) *Profile {
	return &Profile{
		ID:             u.ID,
		Name:           u.Name,
		Gender:         u.Gender,
		Age:            u.Age.Value,
		DistanceFromMe: int32(u.Location.Coordinates.Longitude),
	}
}

func TestService_Swipe(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	// setUp
	store := NewMockStore(controller)
	faker := gofakeit.New(10)
	userService := NewService(faker, store)
	ctx := context.Background()

	t.Run("successful swipe yes with match", func(t *testing.T) {
		// given
		fiftyUsers := createFiftyUsers(faker)
		ID := fiftyUsers[0].ID
		swipedID := fiftyUsers[10].ID
		swipedUser := fiftyUsers[10]
		swipe := &Swipe{ID: swipedID, OK: true}

		store.EXPECT().GetUser(ctx, ID).Return(swipedUser, nil)
		store.EXPECT().Swipe(ctx, ID, swipe).Return(nil)
		store.EXPECT().Match(ctx, ID, swipedID).Return(true, nil)

		// when
		ok, err := userService.Swipe(ctx, ID, swipedID, true)
		require.NoError(t, err)

		//  then
		require.Equal(t, true, ok)
	})

	t.Run("successful swipe yes with no match", func(t *testing.T) {
		// given
		fiftyUsers := createFiftyUsers(faker)
		ID := fiftyUsers[0].ID
		swipedID := fiftyUsers[10].ID
		swipedUser := fiftyUsers[10]
		swipe := &Swipe{ID: swipedID, OK: true}

		store.EXPECT().GetUser(ctx, ID).Return(swipedUser, nil)
		store.EXPECT().Swipe(ctx, ID, swipe).Return(nil)
		store.EXPECT().Match(ctx, ID, swipedID).Return(false, nil)

		// when
		ok, err := userService.Swipe(ctx, ID, swipedID, true)
		require.NoError(t, err)

		//  then
		require.Equal(t, false, ok)
	})

	t.Run("successful swipe no", func(t *testing.T) {
		// given
		fiftyUsers := createFiftyUsers(faker)
		ID := fiftyUsers[0].ID
		swipedID := fiftyUsers[10].ID
		swipedUser := fiftyUsers[10]
		swipe := &Swipe{ID: swipedID, OK: false}

		store.EXPECT().GetUser(ctx, ID).Return(swipedUser, nil)
		store.EXPECT().Swipe(ctx, ID, swipe).Return(nil)

		// when
		ok, err := userService.Swipe(ctx, ID, swipedID, false)
		require.NoError(t, err)

		//  then
		require.Equal(t, false, ok)
	})
}
