package people_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
)

type spyTokenProvider struct {
	claims          *core.AuthTokenClaims
	decodeErr       error
	validateRefresh bool
	genErr          error
	generated       []core.AuthTokenPayload
}

func (s *spyTokenProvider) GenerateTokenRing(data core.AuthTokenPayload) (*core.AuthResponse, error) {
	s.generated = append(s.generated, data)
	if s.genErr != nil {
		return nil, s.genErr
	}
	return &core.AuthResponse{
		AccessToken:  "access_" + data.Sub,
		RefreshToken: "refresh_" + data.Sub,
	}, nil
}

func (s *spyTokenProvider) ValidateToken(token string) (*core.AuthTokenClaims, error) {
	return s.claims, nil
}

func (s *spyTokenProvider) DecodeToken(token string) (*core.AuthTokenClaims, error) {
	if s.decodeErr != nil {
		return nil, s.decodeErr
	}
	return s.claims, nil
}

func (s *spyTokenProvider) ValidateRefreshToken(tokenID uuid.UUID, refreshToken string) bool {
	return s.validateRefresh
}

type spyPeopleRepo struct {
	person      *people.Person
	findIDErr   error
	findIDCalls int
	saveErr     error
	saved       *people.Person
}

func (s *spyPeopleRepo) FindByID(ctx context.Context, id core.PersonID) (*people.Person, error) {
	s.findIDCalls++
	if s.findIDErr != nil {
		return nil, s.findIDErr
	}
	return s.person, nil
}

func (s *spyPeopleRepo) FindAll(ctx context.Context, pagination core.Pagination) ([]*people.Person, int, error) {
	return nil, 0, nil
}

func (s *spyPeopleRepo) Save(ctx context.Context, p *people.Person) error {
	s.saved = p
	return s.saveErr
}

func (s *spyPeopleRepo) FindCredentials(ctx context.Context, provider people.IdentityProvider, regID string) (core.PersonID, string, error) {
	return core.NilPersonID, "", nil
}

func (s *spyPeopleRepo) DeletePerson(ctx context.Context, personID core.PersonID) error {
	return nil
}

func assertExceptionCode(t *testing.T, err error, wantCode exceptions.ErrorCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error with code %q, got nil", wantCode)
	}
	var exc exceptions.Exception
	if !errors.As(err, &exc) {
		t.Fatalf("expected exceptions.Exception with code %q, got %T: %v", wantCode, err, err)
	}
	if exc.Code != wantCode {
		t.Errorf("expected exception code %q, got %q (message: %q)", wantCode, exc.Code, exc.Message)
	}
}

func TestAuthService_Refresh_Hardening(t *testing.T) {
	personID := core.NewPersonID()
	cred, _ := people.NewCredentials(people.LOCAL, "alice@example.com", "secret")
	person, _ := people.NewPerson(personID, "Alice", "Smith", "Ally", []people.Credentials{cred}, people.RoleUser, people.RoleAdministrator)
	ring, _ := core.NewTokenRing("access", "refresh")

	t.Run("nil token ring returns Unauthorized without panicking", func(t *testing.T) {
		svc := people.NewAuthService(&mockDb{}, &spyPeopleRepo{person: person}, &spyTokenProvider{}, &mockHasher{}, nil, nil)

		_, err := svc.Refresh(context.Background(), nil)
		assertExceptionCode(t, err, exceptions.Unauthorized)
		if !errors.Is(err, core.RefreshTokenIsInvalid) {
			t.Errorf("expected RefreshTokenIsInvalid, got %v", err)
		}
	})

	t.Run("nil claims from DecodeToken returns Unauthorized without panicking", func(t *testing.T) {
		tp := &spyTokenProvider{claims: nil, validateRefresh: true}
		svc := people.NewAuthService(&mockDb{}, &spyPeopleRepo{person: person}, tp, &mockHasher{}, nil, nil)

		_, err := svc.Refresh(context.Background(), ring)
		assertExceptionCode(t, err, exceptions.Unauthorized)
		if !errors.Is(err, core.RefreshTokenIsInvalid) {
			t.Errorf("expected RefreshTokenIsInvalid, got %v", err)
		}
	})

	t.Run("missing or negative IssuedAt is rejected instead of bypassing max age", func(t *testing.T) {
		for _, tc := range []struct {
			name     string
			issuedAt int64
		}{
			{name: "zero iat", issuedAt: 0},
			{name: "negative iat", issuedAt: -100},
		} {
			t.Run(tc.name, func(t *testing.T) {
				tp := &spyTokenProvider{
					claims: &core.AuthTokenClaims{
						ID:               uuid.New().String(),
						IssuedAt:         tc.issuedAt,
						AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()},
					},
					validateRefresh: true,
				}
				svc := people.NewAuthService(&mockDb{}, &spyPeopleRepo{person: person}, tp, &mockHasher{}, nil, nil)

				_, err := svc.Refresh(context.Background(), ring)
				assertExceptionCode(t, err, exceptions.Unauthorized)
				if !errors.Is(err, core.RefreshTokenIsInvalid) {
					t.Errorf("expected RefreshTokenIsInvalid, got %v", err)
				}
			})
		}
	})

	t.Run("nil UUID token ID and nil PersonID subject are rejected", func(t *testing.T) {
		tpNilJTI := &spyTokenProvider{
			claims: &core.AuthTokenClaims{
				ID:               uuid.Nil.String(),
				IssuedAt:         time.Now().Unix(),
				AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()},
			},
			validateRefresh: true,
		}
		svcNilJTI := people.NewAuthService(&mockDb{}, &spyPeopleRepo{person: person}, tpNilJTI, &mockHasher{}, nil, nil)
		_, err := svcNilJTI.Refresh(context.Background(), ring)
		assertExceptionCode(t, err, exceptions.Unauthorized)

		tpNilSub := &spyTokenProvider{
			claims: &core.AuthTokenClaims{
				ID:               uuid.New().String(),
				IssuedAt:         time.Now().Unix(),
				AuthTokenPayload: core.AuthTokenPayload{Sub: core.NilPersonID.String()},
			},
			validateRefresh: true,
		}
		svcNilSub := people.NewAuthService(&mockDb{}, &spyPeopleRepo{person: person}, tpNilSub, &mockHasher{}, nil, nil)
		_, err = svcNilSub.Refresh(context.Background(), ring)
		assertExceptionCode(t, err, exceptions.Unauthorized)
	})

	t.Run("FindByID not found vs nil person vs database error", func(t *testing.T) {
		validTP := func() *spyTokenProvider {
			return &spyTokenProvider{
				claims: &core.AuthTokenClaims{
					ID:               uuid.New().String(),
					IssuedAt:         time.Now().Unix(),
					AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()},
				},
				validateRefresh: true,
			}
		}

		svcNotFound := people.NewAuthService(&mockDb{}, &spyPeopleRepo{
			findIDErr: exceptions.NewNotFoundException("person not found", nil),
		}, validTP(), &mockHasher{}, nil, nil)
		_, err := svcNotFound.Refresh(context.Background(), ring)
		assertExceptionCode(t, err, exceptions.Unauthorized)
		if !errors.Is(err, core.RefreshTokenIsInvalid) {
			t.Errorf("expected RefreshTokenIsInvalid on not found, got %v", err)
		}

		svcNilPerson := people.NewAuthService(&mockDb{}, &spyPeopleRepo{person: nil}, validTP(), &mockHasher{}, nil, nil)
		_, err = svcNilPerson.Refresh(context.Background(), ring)
		assertExceptionCode(t, err, exceptions.Unauthorized)
		if !errors.Is(err, core.RefreshTokenIsInvalid) {
			t.Errorf("expected RefreshTokenIsInvalid on nil person, got %v", err)
		}

		dbErr := errors.New("connection reset")
		svcDBErr := people.NewAuthService(&mockDb{}, &spyPeopleRepo{findIDErr: dbErr}, validTP(), &mockHasher{}, nil, nil)
		_, err = svcDBErr.Refresh(context.Background(), ring)
		assertExceptionCode(t, err, exceptions.Internal)
		if errors.Is(err, core.RefreshTokenIsInvalid) {
			t.Errorf("database error during refresh must not be classified as RefreshTokenIsInvalid")
		}
	})

	t.Run("GenerateTokenRing failure returns Internal without RefreshTokenIsInvalid", func(t *testing.T) {
		tp := &spyTokenProvider{
			claims: &core.AuthTokenClaims{
				ID:               uuid.New().String(),
				IssuedAt:         time.Now().Unix(),
				AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()},
			},
			validateRefresh: true,
			genErr:          errors.New("signing failure"),
		}
		svc := people.NewAuthService(&mockDb{}, &spyPeopleRepo{person: person}, tp, &mockHasher{}, nil, nil)

		_, err := svc.Refresh(context.Background(), ring)
		assertExceptionCode(t, err, exceptions.Internal)
		if !errors.Is(err, core.AuthTokenGenerationFailed) {
			t.Errorf("expected AuthTokenGenerationFailed, got %v", err)
		}
		if errors.Is(err, core.RefreshTokenIsInvalid) {
			t.Errorf("token generation failure must not be wrapped with RefreshTokenIsInvalid")
		}
	})

	t.Run("Refresh propagates Sub, DisplayName, and Roles to token provider", func(t *testing.T) {
		tp := &spyTokenProvider{
			claims: &core.AuthTokenClaims{
				ID:               uuid.New().String(),
				IssuedAt:         time.Now().Unix(),
				AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()},
			},
			validateRefresh: true,
		}
		svc := people.NewAuthService(&mockDb{}, &spyPeopleRepo{person: person}, tp, &mockHasher{}, nil, nil)

		resp, err := svc.Refresh(context.Background(), ring)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.AccessToken != "access_"+personID.String() {
			t.Errorf("unexpected access token: %q", resp.AccessToken)
		}
		if len(tp.generated) != 1 {
			t.Fatalf("expected 1 GenerateTokenRing call, got %d", len(tp.generated))
		}
		got := tp.generated[0]
		if got.Sub != personID.String() || got.DisplayName != "Ally" || !reflect.DeepEqual(got.Roles, []string{"User", "Administrator"}) {
			t.Errorf("unexpected token payload: %+v", got)
		}
	})
}
