package handler

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Result *User `json:"result"`
}

type User struct {
	ID       int32  `json:"id"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Gender   string `json:"gender"`
	Age      int32  `json:"age"`
}

type DiscoverRequest struct {
	Gender string `query:"gender"`
	MinAge int32  `query:"min-age"`
	MaxAge int32  `query:"max-age"`
	Ranked bool   `query:"ranked"`
}

func (d *DiscoverRequest) validate() {
	if d.MinAge > 0 && d.MinAge < 18 {
		d.MinAge = 18
	}
	if d.MaxAge > 0 && d.MaxAge > 100 {
		d.MinAge = 100
	}
	switch d.Gender {
	case "", "male", "female":
	default:
		d.Gender = ""
	}
}

type Profile struct {
	ID             int32  `json:"id"`
	Name           string `json:"name"`
	Gender         string `json:"gender"`
	Age            int32  `json:"age"`
	DistanceFromMe int32  `json:"distanceFromMe"`
}

type DiscoverResponse struct {
	Results []*Profile `json:"results"`
}

type SwipeRequest struct {
	SwipedID int32 `json:"id"`
	Ok       bool  `json:"ok"`
}

type SwipeResponse struct {
	Swipe Swipe `json:"result"`
}

type Swipe struct {
	Matched   bool  `json:"matched"`
	MatchedID int32 `json:"matchedID,omitempty"`
}
