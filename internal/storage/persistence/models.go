package persistence

import (
	"github.com/muzzapp/date-api/internal/users"
	"go.mongodb.org/mongo-driver/bson"
)

type Counter struct {
	ID    string `bson:"_id"`
	Value int32  `bson:"value"`
}

func matchFilter(ID, minAge, maxAge int32, gender string, IDs []int32) map[string]interface{} {
	filters := make(map[string]interface{})
	idsFilter(ID, IDs, filters)
	ageFilter(minAge, maxAge, filters)
	genderFilter(gender, filters)
	return filters
}

func idsFilter(ID int32, IDs []int32, filters map[string]interface{}) {
	IDs = append(IDs, ID)
	filters["_id"] = bson.D{{Key: "$nin", Value: IDs}}
}

func ageFilter(minAge, maxAge int32, filters map[string]interface{}) {
	switch {
	case minAge > 0 && maxAge > 0:
		filters["age.value"] = bson.M{
			"$gte": minAge,
			"$lte": maxAge,
		}
	case minAge > 0:
		filters["age.value"] = bson.M{"$gte": minAge}
	case maxAge > 0:
		filters["age.value"] = bson.M{"$lte": maxAge}
	}
}

func genderFilter(gender string, filters map[string]interface{}) {
	if gender != "" {
		filters["gender"] = gender
	}
}

func rankStages(rank *users.Rank) (bson.D, bson.D) {
	addFields := bson.D{}
	sort := bson.D{}
	if rank.MostCommonGender != "" {
		genderField := bson.E{Key: "genderSort", Value: bson.D{
			{"$cond", bson.A{
				bson.D{{"$eq", bson.A{"$gender", rank.MostCommonGender}}},
				1,
				2,
			}},
		}}
		addFields = append(addFields, genderField)
		sort = append(sort, bson.E{Key: "genderSort", Value: 1})
	}

	ageSortField := bson.E{Key: "ageSort", Value: bson.D{
		{Key: "$abs", Value: bson.D{
			{"$subtract", bson.A{"$age.value", rank.AvgAge}},
		}},
	}}
	addFields = append(addFields, ageSortField)
	sort = append(sort, bson.E{Key: "ageSort", Value: 1})
	sort = append(sort, bson.E{Key: "distanceFromMe", Value: 1})

	return bson.D{{"$addFields", addFields}}, bson.D{{"$sort", sort}}
}
