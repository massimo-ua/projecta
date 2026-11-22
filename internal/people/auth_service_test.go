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
	validatedID     uuid.UUID
	validatedToken  string
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
	s.validatedID = tokenID
	s.validatedToken = refreshToken
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
		jti := uuid.New()
		tp := &spyTokenProvider{
			claims: &core.AuthTokenClaims{
				ID:               jti.String(),
				IssuedAt:         time.Now().Add(-6 * 24 * time.Hour).Unix(),
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
		if tp.validatedID != jti || tp.validatedToken != "refresh" {
			t.Errorf("expected ValidateRefreshToken(%s, refresh), got (%s, %q)", jti, tp.validatedID, tp.validatedToken)
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

type spyDb struct {
	txCalls  int
	txErr    error
	skipFunc bool
}

func (d *spyDb) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	d.txCalls++
	if d.txErr != nil {
		return nil, d.txErr
	}
	if d.skipFunc {
		return nil, nil
	}
	return fn(ctx)
}
func (d *spyDb) Close()                        {}
func (d *spyDb) Ping(ctx context.Context) error { return nil }

type spyStrategy struct {
	provider     people.IdentityProvider
	personID     core.PersonID
	authErr      error
	claims       *core.AuthTokenClaims
	resolveErr   error
	resolveCalls int
}

func (s *spyStrategy) Provider() people.IdentityProvider { return s.provider }
func (s *spyStrategy) Authenticate(ctx context.Context, creds people.Credentials) (core.PersonID, error) {
	if s.authErr != nil {
		return core.NilPersonID, s.authErr
	}
	return s.personID, nil
}
func (s *spyStrategy) ResolveInvitationClaims(ctx context.Context, token string) (*core.AuthTokenClaims, error) {
	s.resolveCalls++
	if s.resolveErr != nil {
		return nil, s.resolveErr
	}
	return s.claims, nil
}

type spyInvitationRepo struct {
	inv       *people.Invitation
	findErr   error
	saveErr   error
	savedInv  *people.Invitation
}

func (r *spyInvitationRepo) Save(ctx context.Context, inv *people.Invitation) error {
	r.savedInv = inv
	return r.saveErr
}
func (r *spyInvitationRepo) FindByID(ctx context.Context, id core.InvitationID) (*people.Invitation, error) {
	return r.inv, r.findErr
}
func (r *spyInvitationRepo) FindByCodeHash(ctx context.Context, codeHash string) (*people.Invitation, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.inv, nil
}
func (r *spyInvitationRepo) FindByCreator(ctx context.Context, creatorID core.PersonID) ([]*people.Invitation, error) {
	return nil, nil
}
func (r *spyInvitationRepo) Delete(ctx context.Context, id core.InvitationID) error { return nil }
func (r *spyInvitationRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	return false, nil
}

func TestAuthService_NewAndLogin_Hardening(t *testing.T) {
	personID := core.NewPersonID()
	cred, _ := people.NewCredentials(people.FACEBOOK, "fb_id", "fb_token")
	person, _ := people.NewPerson(personID, "Bob", "Jones", "Bobby", []people.Credentials{cred}, people.RoleUser)

	t.Run("NewAuthService without google auth does not register GOOGLE provider", func(t *testing.T) {
		svc := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: person}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{})
		googleCred, _ := people.NewCredentials(people.GOOGLE, "g_id", "g_tok")

		_, err := svc.Login(context.Background(), googleCred)
		assertExceptionCode(t, err, exceptions.ValidationFailed)

		_, err = svc.LoginWithInvitation(context.Background(), "g_tok", people.GOOGLE, "some_invite_code")
		assertExceptionCode(t, err, exceptions.ValidationFailed)
	})

	t.Run("NewAuthService skips nil strategies without panicking", func(t *testing.T) {
		strat := &spyStrategy{provider: people.FACEBOOK, personID: personID}
		tp := &spyTokenProvider{}
		svc := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: person}, tp, &mockHasher{}, nil, nil, nil, strat)

		resp, err := svc.Login(context.Background(), cred)
		if err != nil || resp == nil {
			t.Fatalf("expected successful login, got resp=%v err=%v", resp, err)
		}
		if len(tp.generated) != 1 || tp.generated[0].DisplayName != "Bobby" || tp.generated[0].Sub != personID.String() {
			t.Errorf("unexpected token payload: %+v", tp.generated)
		}
	})

	t.Run("Login rejects NilPersonID returned by strategy", func(t *testing.T) {
		strat := &spyStrategy{provider: people.FACEBOOK, personID: core.NilPersonID}
		repo := &spyPeopleRepo{person: person}
		svc := people.NewAuthService(&spyDb{}, repo, &spyTokenProvider{}, &mockHasher{}, nil, nil, strat)

		_, err := svc.Login(context.Background(), cred)
		assertExceptionCode(t, err, exceptions.Unauthorized)
		if repo.findIDCalls != 0 {
			t.Errorf("expected no FindByID call for NilPersonID, got %d", repo.findIDCalls)
		}
	})

	t.Run("Login distinguishes NotFound and nil person from DB errors", func(t *testing.T) {
		strat := &spyStrategy{provider: people.FACEBOOK, personID: personID}

		svcNotFound := people.NewAuthService(&spyDb{}, &spyPeopleRepo{
			findIDErr: exceptions.NewNotFoundException("person not found", nil),
		}, &spyTokenProvider{}, &mockHasher{}, nil, nil, strat)
		_, err := svcNotFound.Login(context.Background(), cred)
		assertExceptionCode(t, err, exceptions.Unauthorized)

		svcNil := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: nil}, &spyTokenProvider{}, &mockHasher{}, nil, nil, strat)
		_, err = svcNil.Login(context.Background(), cred)
		assertExceptionCode(t, err, exceptions.Unauthorized)

		svcDBErr := people.NewAuthService(&spyDb{}, &spyPeopleRepo{
			findIDErr: errors.New("db timeout"),
		}, &spyTokenProvider{}, &mockHasher{}, nil, nil, strat)
		_, err = svcDBErr.Login(context.Background(), cred)
		assertExceptionCode(t, err, exceptions.Internal)
	})
}

func TestAuthService_LoginWithInvitation_Hardening(t *testing.T) {
	adminID := core.NewPersonID()
	personID := core.NewPersonID()
	email := "invited@example.com"
	rawCode := "invitation_secret_code_42"
	codeHash := people.HashInvitationCode(rawCode)

	newFixture := func() (*people.Person, *people.Invitation, *core.AuthTokenClaims) {
		p, _ := people.NewPerson(personID, "Invited", "User", email, nil, people.RoleUser)
		inv, _ := people.NewInvitation(
			core.NewInvitationID(),
			email,
			codeHash,
			time.Now().Add(24*time.Hour),
			adminID,
			&personID,
			people.InvitationStatusPending,
			time.Now(),
			nil,
		)
		claims := &core.AuthTokenClaims{
			AuthTokenPayload: core.AuthTokenPayload{
				Sub:         "google_sub_42",
				Email:       email,
				FirstName:   "Jane",
				LastName:    "Doe",
				DisplayName: "Jane D.",
			},
		}
		return p, inv, claims
	}

	t.Run("fails fast before ResolveInvitationClaims when invitationRepo or db is nil", func(t *testing.T) {
		_, _, claims := newFixture()

		stratNoRepo := &spyStrategy{provider: people.GOOGLE, claims: claims}
		svcNoRepo := people.NewAuthService(&spyDb{}, &spyPeopleRepo{}, &spyTokenProvider{}, &mockHasher{}, nil, nil, stratNoRepo)
		_, err := svcNoRepo.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Internal)
		if stratNoRepo.resolveCalls != 0 {
			t.Errorf("expected ResolveInvitationClaims not to be called when invitationRepo is nil, got %d calls", stratNoRepo.resolveCalls)
		}

		stratNoDB := &spyStrategy{provider: people.GOOGLE, claims: claims}
		svcNoDB := people.NewAuthService(nil, &spyPeopleRepo{}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{}, stratNoDB)
		_, err = svcNoDB.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Internal)
		if stratNoDB.resolveCalls != 0 {
			t.Errorf("expected ResolveInvitationClaims not to be called when db is nil, got %d calls", stratNoDB.resolveCalls)
		}
	})

	t.Run("nil claims returned by strategy returns Unauthorized without panicking", func(t *testing.T) {
		_, inv, _ := newFixture()
		strat := &spyStrategy{provider: people.GOOGLE, claims: nil}
		svc := people.NewAuthService(&spyDb{}, &spyPeopleRepo{}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{inv: inv}, strat)

		_, err := svc.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Unauthorized)
	})

	t.Run("FindByCodeHash NotFound and nil return NotFound while DB error returns Internal", func(t *testing.T) {
		p, _, claims := newFixture()
		strat := &spyStrategy{provider: people.GOOGLE, claims: claims}

		svcNotFound := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: p}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{
			findErr: exceptions.NewNotFoundException("invitation not found", nil),
		}, strat)
		_, err := svcNotFound.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.NotFound)

		svcNilInv := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: p}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{
			inv: nil,
		}, strat)
		_, err = svcNilInv.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.NotFound)

		svcDBErr := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: p}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{
			findErr: errors.New("deadlock detected"),
		}, strat)
		_, err = svcDBErr.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Internal)
	})

	t.Run("FindByID error, nil person, AcceptInvitation error, and Save errors return expected exceptions", func(t *testing.T) {
		p, inv, claims := newFixture()
		strat := &spyStrategy{provider: people.GOOGLE, claims: claims}

		svcFindPersonErr := people.NewAuthService(&spyDb{}, &spyPeopleRepo{findIDErr: errors.New("db err")}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{inv: inv}, strat)
		_, err := svcFindPersonErr.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Internal)

		svcNilPerson := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: nil}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{inv: inv}, strat)
		_, err = svcNilPerson.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Internal)

		badClaims := &core.AuthTokenClaims{
			AuthTokenPayload: core.AuthTokenPayload{
				Sub:         "google_sub_42",
				Email:       email,
				DisplayName: string(make([]byte, 300)),
			},
		}
		_, inv2, _ := newFixture()
		svcBadProfile := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: p}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{inv: inv2}, &spyStrategy{provider: people.GOOGLE, claims: badClaims})
		_, err = svcBadProfile.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.ValidationFailed)

		p3, inv3, _ := newFixture()
		svcSavePersonErr := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: p3, saveErr: errors.New("save person failed")}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{inv: inv3}, strat)
		_, err = svcSavePersonErr.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Internal)

		p4, inv4, _ := newFixture()
		svcSaveInvErr := people.NewAuthService(&spyDb{}, &spyPeopleRepo{person: p4}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{inv: inv4, saveErr: errors.New("save inv failed")}, strat)
		_, err = svcSaveInvErr.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Internal)

		p5, inv5, _ := newFixture()
		svcSkippedTx := people.NewAuthService(&spyDb{skipFunc: true}, &spyPeopleRepo{person: p5}, &spyTokenProvider{}, &mockHasher{}, nil, &spyInvitationRepo{inv: inv5}, strat)
		_, err = svcSkippedTx.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		assertExceptionCode(t, err, exceptions.Internal)
	})

	t.Run("success uses transaction, avoids redundant FindByID, and issues tokens with updated profile", func(t *testing.T) {
		p, inv, claims := newFixture()
		db := &spyDb{}
		peopleRepo := &spyPeopleRepo{person: p}
		invRepo := &spyInvitationRepo{inv: inv}
		tp := &spyTokenProvider{}
		strat := &spyStrategy{provider: people.GOOGLE, claims: claims}

		svc := people.NewAuthService(db, peopleRepo, tp, &mockHasher{}, nil, invRepo, strat)

		resp, err := svc.LoginWithInvitation(context.Background(), "tok", people.GOOGLE, rawCode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp == nil || resp.AccessToken != "access_"+personID.String() {
			t.Fatalf("unexpected auth response: %+v", resp)
		}
		if db.txCalls != 1 {
			t.Errorf("expected 1 transaction call, got %d", db.txCalls)
		}
		if peopleRepo.findIDCalls != 1 {
			t.Errorf("expected exactly 1 FindByID call (no redundant post-tx query), got %d", peopleRepo.findIDCalls)
		}
		if peopleRepo.saved == nil || invRepo.savedInv == nil || !invRepo.savedInv.IsCompleted() {
			t.Errorf("expected saved person and completed invitation")
		}
		if len(tp.generated) != 1 {
			t.Fatalf("expected 1 GenerateTokenRing call, got %d", len(tp.generated))
		}
		got := tp.generated[0]
		if got.Sub != personID.String() || got.DisplayName != "Jane D." || !reflect.DeepEqual(got.Roles, []string{"User"}) {
			t.Errorf("unexpected token payload: %+v", got)
		}
	})
}
