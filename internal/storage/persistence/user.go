package persistence

import (
	"context"
	"errors"

	"github.com/muzzapp/date-api/internal/users"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

type User struct {
	coll          *mongo.Collection
	collSecondary *mongo.Collection
	collCounters  *mongo.Collection
}

const (
	usersColl    = "users"
	countersColl = "counters"
)

var (
	_ users.Store = (*User)(nil)
)

func NewItemPersistence(db *mongo.Database) *User {
	return &User{
		coll: db.Collection(usersColl,
			options.Collection().SetReadPreference(readpref.PrimaryPreferred()),
		),
		collSecondary: db.Collection(usersColl,
			options.Collection().SetReadPreference(readpref.SecondaryPreferred()),
		),
		collCounters: db.Collection(countersColl,
			options.Collection().SetReadPreference(readpref.PrimaryPreferred()),
		),
	}
}

func (u *User) nextID(ctx context.Context) (int32, error) {
	var counter Counter
	filter := bson.M{"_id": usersColl}
	update := bson.M{"$inc": bson.M{"value": 1}}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	if err := u.collCounters.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter); err != nil {
		return 0, err
	}
	return counter.Value, nil
}

func (u *User) CreateUser(ctx context.Context, user *users.User) (*users.User, error) {
	nextID, err := u.nextID(ctx)
	if err != nil {
		return nil, err
	}
	user.ID = nextID
	res, err := u.coll.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = res.InsertedID.(int32)
	return user, nil
}

func (u *User) GetUser(ctx context.Context, ID int32) (*users.User, error) {
	user := new(users.User)
	if err := u.collSecondary.FindOne(ctx, bson.M{"_id": ID}).Decode(user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = users.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (u *User) GetRankByIDs(ctx context.Context, IDs []int32) (*users.Rank, error) {
	matchStage := bson.M{"$match": bson.M{"_id": bson.M{"$in": IDs}}}
	facetStage := bson.M{
		"$facet": bson.M{
			"avgAge": bson.A{
				bson.M{"$group": bson.M{
					"_id":    nil,
					"avgAge": bson.M{"$avg": "$age.value"},
				},
				},
			},
			"genderCount": bson.A{
				bson.M{"$group": bson.M{
					"_id":   "$gender",
					"count": bson.M{"$sum": 1},
				}},
				bson.M{"$sort": bson.M{"count": -1}},
			},
		},
	}
	projectStage := bson.M{
		"$project": bson.M{
			"avgAge":           bson.M{"$arrayElemAt": bson.A{"$avgAge.avgAge", 0}},
			"mostCommonGender": bson.M{"$arrayElemAt": bson.A{"$genderCount._id", 0}},
			"firstCount":       bson.M{"$arrayElemAt": bson.A{"$genderCount.count", 0}},
			"secondCount":      bson.M{"$arrayElemAt": bson.A{"$genderCount.count", 1}},
		},
	}
	pipeline := []bson.M{matchStage, facetStage, projectStage}

	cursor, err := u.collSecondary.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err = cursor.All(ctx, &result); err != nil {
		return nil, err
	}
	if len(result) == 0 {
		return nil, nil
	}
	if result[0]["secondCount"] != nil {
		firstCount := result[0]["firstCount"].(int32)
		secondCount := result[0]["secondCount"].(int32)
		if firstCount == secondCount {
			return &users.Rank{
				AvgAge: int32(result[0]["avgAge"].(float64)),
			}, nil
		}
	}
	return &users.Rank{
		AvgAge:           int32(result[0]["avgAge"].(float64)),
		MostCommonGender: result[0]["mostCommonGender"].(string),
	}, nil
}

func (u *User) GetUserByEmail(ctx context.Context, email string) (*users.User, error) {
	user := new(users.User)
	if err := u.collSecondary.FindOne(ctx, bson.M{"email": email}).Decode(user); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			err = users.ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (u *User) Discover(ctx context.Context, ID, minAge, maxAge int32, gender string, IDs []int32,
	location *users.Location, rank *users.Rank) ([]*users.Profile, error) {

	pipeline := discoverPipeline(ID, minAge, maxAge, gender, IDs, location, rank)
	cursor, err := u.collSecondary.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	profiles := make([]*users.Profile, 0)
	for cursor.Next(ctx) {
		var result bson.M
		if err = cursor.Decode(&result); err != nil {
			return nil, err
		}
		profiles = append(profiles, &users.Profile{
			ID:             result["_id"].(int32),
			Name:           result["name"].(string),
			Gender:         result["gender"].(string),
			Age:            result["age"].(bson.M)["value"].(int32),
			DistanceFromMe: int32(result["distanceFromMe"].(float64)),
		})
	}
	return profiles, nil
}

func discoverPipeline(ID, minAge, maxAge int32, gender string, IDs []int32, location *users.Location, rank *users.Rank) mongo.Pipeline {
	nearCoordinates := bson.D{
		{"type", "Point"},
		{"coordinates", location.CoordinatesFloat64Slice()},
	}
	geoNearStage := bson.D{
		{"$geoNear", bson.D{
			{"near", nearCoordinates},
			{"key", "location"},
			{"distanceField", "distanceFromMe"},
			{"query", matchFilter(ID, minAge, maxAge, gender, IDs)},
		}},
	}
	projectStage := bson.D{
		{"$project", bson.D{
			{"_id", 1},
			{"name", 1},
			{"gender", 1},
			{"age.value", 1},
			{"distanceFromMe", 1},
		}},
	}
	if rank == nil {
		return mongo.Pipeline{geoNearStage, projectStage}
	}
	addFieldsStage, sortStage := rankStages(rank)
	return mongo.Pipeline{geoNearStage, addFieldsStage, sortStage, projectStage}
}

func (u *User) Swipe(ctx context.Context, ID int32, swipe *users.Swipe) error {
	filter := bson.M{"_id": ID}
	update := bson.M{"$addToSet": bson.M{"swipes": swipe}}
	if _, err := u.coll.UpdateOne(ctx, filter, update); err != nil {
		return err
	}
	return nil
}

func (u *User) Match(ctx context.Context, ID, swipedID int32) (bool, error) {
	filter := bson.M{
		"_id": swipedID,
		"swipes": bson.M{
			"$elemMatch": bson.M{
				"id": ID,
				"ok": true,
			},
		},
	}
	count, err := u.collSecondary.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
