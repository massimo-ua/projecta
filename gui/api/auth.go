package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type LoginDTO struct {
	Login    string `json:"id"`
	Password string `json:"token"`
	Provider string `json:"identity_provider"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

func newLoginDTO(login, password string) *LoginDTO {
	return &LoginDTO{Login: login, Password: password, Provider: "LOCAL"}
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
}

type HttpAuthProvider struct {
	url          string
	accessToken  string
	refreshToken string
	expiresAt    time.Time
}

func NewHttpAuthProvider(url string) (*HttpAuthProvider, error) {
	if url == "" {
		return nil, errors.New("url cannot be empty")
	}

	return &HttpAuthProvider{url: url}, nil
}

func (p *HttpAuthProvider) Login(username, password string) error {
	credentials := newLoginDTO(username, password)
	jsonBody, err := json.Marshal(credentials)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/login", p.url), bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("authentication failed")
	}

	var tokenResp AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return err
	}

	p.accessToken = tokenResp.AccessToken
	p.refreshToken = tokenResp.RefreshToken
	p.expiresAt = time.Unix(tokenResp.ExpiresAt, 0)
	return nil
}

func (p *HttpAuthProvider) refreshAuth() error {
	if p.refreshToken == "" {
		return errors.New("no refreshAuth token available")
	}

	jsonBody, err := json.Marshal(RefreshTokenDTO{
		RefreshToken: p.refreshToken,
		AccessToken:  p.accessToken,
	})

	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/refresh", p.url), bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("refreshAuth failed")
	}

	var tokenResp AuthResponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResp)
	if err != nil {
		return err
	}

	p.accessToken = tokenResp.AccessToken
	p.refreshToken = tokenResp.RefreshToken
	p.expiresAt = time.Unix(tokenResp.ExpiresAt, 0)
	return nil
}

func (p *HttpAuthProvider) GetAccessToken() (string, error) {
	if p.IsAuthorised() {
		return p.accessToken, nil
	}

	err := p.refreshAuth()

	if err != nil {
		return "", err
	}

	return p.accessToken, nil
}

func (p *HttpAuthProvider) IsAuthorised() bool {
	return p.accessToken != "" && time.Now().Before(p.expiresAt)
}
