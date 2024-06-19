package users

import (
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

type User struct {
	ID       int32     `bson:"_id"`
	Email    string    `bson:"email"`
	Password string    `bson:"password"`
	Name     string    `bson:"name"`
	Gender   string    `bson:"gender"`
	Age      *Age      `bson:"age"`
	Location *Location `bson:"location"`
	Swipes   []*Swipe  `bson:"swipes"`
}

type Age struct {
	Value int32     `bson:"value"`
	DOB   time.Time `bson:"dob"`
}

type Location struct {
	Type        string       `bson:"type"`
	Coordinates *Coordinates `bson:"coordinates"`
}

func (l *Location) CoordinatesFloat64Slice() []float64 {
	if l == nil || l.Coordinates == nil {
		return []float64{}
	}
	return []float64{l.Coordinates.Longitude, l.Coordinates.Latitude}
}

type Coordinates struct {
	Longitude float64 `bson:"longitude"`
	Latitude  float64 `bson:"latitude"`
}

type Swipe struct {
	ID int32 `bson:"id"`
	OK bool  `bson:"ok"`
}

func (u *User) swipeIDs() []int32 {
	if u.Swipes == nil {
		return nil
	}
	ids := make([]int32, len(u.Swipes))
	for i, swipe := range u.Swipes {
		ids[i] = swipe.ID
	}
	return ids
}

func (u *User) yesSwipeIDs() []int32 {
	if u.Swipes == nil {
		return nil
	}
	ids := make([]int32, 0)
	for _, swipe := range u.Swipes {
		if swipe.OK {
			ids = append(ids, swipe.ID)
		}
	}
	return ids
}

func NewFakeUser(f *gofakeit.Faker) *User {
	return &User{
		Email:    f.Email(),
		Password: f.Password(true, true, true, true, false, 32),
		Name:     f.Name(),
		Gender:   f.Gender(),
		Age:      fakeAge(f),
		Location: &Location{
			Type: "Point",
			Coordinates: &Coordinates{
				Longitude: f.Longitude(),
				Latitude:  f.Latitude(),
			},
		},
		Swipes: []*Swipe{},
	}
}

func fakeAge(f *gofakeit.Faker) *Age {
	now := time.Now()
	dob := f.DateRange(now.AddDate(-100, 0, 0), now.AddDate(-18, 0, 0))
	return &Age{
		Value: calculateAge(now, dob),
		DOB:   dob,
	}
}

func calculateAge(now, dob time.Time) int32 {
	years := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		years--
	}
	return int32(years)
}

type Profile struct {
	ID             int32
	Name           string
	Gender         string
	Age            int32
	DistanceFromMe int32
}

type Rank struct {
	AvgAge           int32
	MostCommonGender string
}
