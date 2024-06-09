package web

import (
    "context"
    "encoding/json"
    "net/http"
)

func decodeRegisterUser(_ context.Context, r *http.Request) (any, error) {
    var req RegisterUserDTO
    err := json.NewDecoder(r.Body).Decode(&req)
    return req, err
}

func decodeLoginUser(_ context.Context, r *http.Request) (any, error) {
    var req LoginDTO
    err := json.NewDecoder(r.Body).Decode(&req)
    return req, err
}

func decodeRefreshUserToken(_ context.Context, r *http.Request) (any, error) {
    var req RefreshTokenDTO
    err := json.NewDecoder(r.Body).Decode(&req)
    return req, err
}
