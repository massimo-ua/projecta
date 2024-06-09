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
