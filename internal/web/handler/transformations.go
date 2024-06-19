package handler

import "github.com/muzzapp/date-api/internal/users"

func toCreateUserResponse(u *users.User) *CreateUserResponse {
	if u == nil {
		return nil
	}
	return &CreateUserResponse{
		Result: &User{
			ID:       u.ID,
			Email:    u.Email,
			Password: u.Password,
			Name:     u.Name,
			Gender:   u.Gender,
			Age:      u.Age.Value,
		},
	}
}

func toDiscoverResponse(ps []*users.Profile) *DiscoverResponse {
	return &DiscoverResponse{
		Results: toProfiles(ps),
	}
}

func toProfiles(ps []*users.Profile) []*Profile {
	profiles := make([]*Profile, len(ps))
	for i, p := range ps {
		profiles[i] = toProfile(p)
	}
	return profiles
}

func toProfile(p *users.Profile) *Profile {
	if p == nil {
		return nil
	}
	return &Profile{
		ID:             p.ID,
		Name:           p.Name,
		Gender:         p.Gender,
		Age:            p.Age,
		DistanceFromMe: p.DistanceFromMe,
	}
}
