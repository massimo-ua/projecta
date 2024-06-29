package web

type LoginDTO struct {
	ID               string `json:"id"`
	IdentityProvider string `json:"identity_provider"`
	Token            string `json:"token"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type ListProjectsResponse struct {
	Projects []ProjectDTO `json:"projects"`
	Limit    int          `json:"limit"`
	Offset   int          `json:"offset"`
}

type ListTypesResponse struct {
	Types  []TypeDTO `json:"types"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}
