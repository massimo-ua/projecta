package people_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/people"
)

type customMockStrategy struct {
	provider people.IdentityProvider
	personID uuid.UUID
	claims   *core.AuthTokenClaims
	err      error
}

func (s *customMockStrategy) Provider() people.IdentityProvider {
	return s.provider
}

func (s *customMockStrategy) Authenticate(ctx context.Context, creds people.Credentials) (uuid.UUID, error) {
	if s.err != nil {
		return uuid.Nil, s.err
	}
	return s.personID, nil
}

func (s *customMockStrategy) ResolveInvitationClaims(ctx context.Context, token string) (*core.AuthTokenClaims, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.claims, nil
}

func TestLocalAuthStrategy(t *testing.T) {
	personID := uuid.New()
	cred, err := people.NewCredentials("LOCAL", "user@example.com", "mysecret")
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}

	t.Run("successful local authentication", func(t *testing.T) {
		repo := &mockPeopleRepo{credPersonID: personID, credHash: "hash_mysecret"}
		hasher := &mockHasher{compareRes: true}
		strat := people.NewLocalAuthStrategy(repo, hasher)

		if strat.Provider() != people.LOCAL {
			t.Errorf("expected provider LOCAL, got %s", strat.Provider())
		}

		id, err := strat.Authenticate(context.Background(), cred)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != personID {
			t.Errorf("expected personID %v, got %v", personID, id)
		}
	})

	t.Run("FindCredentials fails", func(t *testing.T) {
		repo := &mockPeopleRepo{findCredErr: errors.New("db error")}
		hasher := &mockHasher{compareRes: true}
		strat := people.NewLocalAuthStrategy(repo, hasher)

		_, err := strat.Authenticate(context.Background(), cred)
		if err == nil {
			t.Errorf("expected error when FindCredentials fails")
		}
	})

	t.Run("password hash comparison fails", func(t *testing.T) {
		repo := &mockPeopleRepo{credPersonID: personID, credHash: "hash_mysecret"}
		hasher := &mockHasher{compareRes: false}
		strat := people.NewLocalAuthStrategy(repo, hasher)

		_, err := strat.Authenticate(context.Background(), cred)
		if err == nil {
			t.Errorf("expected error when password compare fails")
		}
	})
}

func TestGoogleAuthStrategy(t *testing.T) {
	personID := uuid.New()
	cred, err := people.NewCredentials("GOOGLE", "google_id", "google_token")
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}

	t.Run("successful google authentication", func(t *testing.T) {
		repo := &mockPeopleRepo{credPersonID: personID, credHash: "google_id"}
		google := &mockThirdPartyAuth{
			claims: &core.AuthTokenClaims{
				AuthTokenPayload: core.AuthTokenPayload{
					Sub:         "google_id",
					DisplayName: "Google User",
				},
			},
		}
		strat := people.NewGoogleAuthStrategy(repo, google)

		if strat.Provider() != people.GOOGLE {
			t.Errorf("expected provider GOOGLE, got %s", strat.Provider())
		}

		id, err := strat.Authenticate(context.Background(), cred)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if id != personID {
			t.Errorf("expected personID %v, got %v", personID, id)
		}
	})

	t.Run("google token validation fails", func(t *testing.T) {
		repo := &mockPeopleRepo{credPersonID: personID}
		google := &mockThirdPartyAuth{err: errors.New("invalid google token")}
		strat := people.NewGoogleAuthStrategy(repo, google)

		_, err := strat.Authenticate(context.Background(), cred)
		if err == nil {
			t.Errorf("expected error when google token is invalid")
		}
	})

	t.Run("user not found for google sub", func(t *testing.T) {
		repo := &mockPeopleRepo{findCredErr: errors.New("user not found")}
		google := &mockThirdPartyAuth{
			claims: &core.AuthTokenClaims{
				AuthTokenPayload: core.AuthTokenPayload{Sub: "unmapped_sub"},
			},
		}
		strat := people.NewGoogleAuthStrategy(repo, google)

		_, err := strat.Authenticate(context.Background(), cred)
		if err == nil {
			t.Errorf("expected error when user credentials are not found")
		}
	})
}

func TestAuthService_CustomStrategyInjection(t *testing.T) {
	personID := uuid.New()
	cred, err := people.NewCredentials("FACEBOOK", "fb_id", "fb_token")
	if err != nil {
		t.Fatalf("failed to create credentials: %v", err)
	}
	p, _ := people.NewPerson(personID, "Facebook", "User", "FB User", []people.Credentials{cred})

	repo := &mockPeopleRepo{person: p}
	tokenProvider := &mockTokenProvider{
		claims: &core.AuthTokenClaims{
			ID:               uuid.New().String(),
			AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()},
		},
	}
	hasher := &mockHasher{}
	customStrat := &customMockStrategy{
		provider: people.FACEBOOK,
		personID: personID,
	}

	svc := people.NewAuthService(
		&mockDb{},
		repo,
		tokenProvider,
		hasher,
		nil,
		nil,
		customStrat,
	)

	res, err := svc.Login(context.Background(), cred)
	if err != nil || res == nil {
		t.Fatalf("unexpected error logging in with custom strategy: %v", err)
	}

	// Strategy returning error
	customStratErr := &customMockStrategy{
		provider: people.FACEBOOK,
		err:      errors.New("custom auth failed"),
	}
	svcErr := people.NewAuthService(
		&mockDb{},
		repo,
		tokenProvider,
		hasher,
		nil,
		nil,
		customStratErr,
	)
	_, err = svcErr.Login(context.Background(), cred)
	if err == nil {
		t.Errorf("expected error when custom strategy fails")
	}
}
