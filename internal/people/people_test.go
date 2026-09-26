package people_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/people"
)

type mockDb struct{}

func (m *mockDb) Tx(ctx context.Context, fn func(ctx context.Context) (any, error)) (any, error) {
	return fn(ctx)
}
func (m *mockDb) Close()                        {}
func (m *mockDb) Ping(ctx context.Context) error { return nil }

type mockHasher struct {
	hashErr    error
	compareRes bool
}

func (m *mockHasher) Hash(v string) (string, error) {
	if m.hashErr != nil {
		return "", m.hashErr
	}
	return "hash_" + v, nil
}
func (m *mockHasher) Compare(v, h string) bool { return m.compareRes }

type mockPeopleRepo struct {
	findIDErr   error
	person      *people.Person
	peopleList  []*people.Person
	total       int
	findAllErr  error
	registerErr error
	findCredErr error
	credPersonID uuid.UUID
	credHash    string
	saveRolesErr error
	savedRoles   []people.Role
}

func (m *mockPeopleRepo) FindByID(ctx context.Context, id uuid.UUID) (*people.Person, error) {
	if m.findIDErr != nil {
		return nil, m.findIDErr
	}
	return m.person, nil
}
func (m *mockPeopleRepo) FindAll(ctx context.Context, pagination core.Pagination) ([]*people.Person, int, error) {
	if m.findAllErr != nil {
		return nil, 0, m.findAllErr
	}
	return m.peopleList, m.total, nil
}
func (m *mockPeopleRepo) Register(ctx context.Context, p *people.Person) error {
	return m.registerErr
}
func (m *mockPeopleRepo) FindCredentials(ctx context.Context, provider people.IdentityProvider, regID string) (uuid.UUID, string, error) {
	if m.findCredErr != nil {
		return uuid.Nil, "", m.findCredErr
	}
	return m.credPersonID, m.credHash, nil
}
func (m *mockPeopleRepo) SaveRoles(ctx context.Context, personID uuid.UUID, roles []people.Role) error {
	m.savedRoles = roles
	return m.saveRolesErr
}
func (m *mockPeopleRepo) DeletePerson(ctx context.Context, personID uuid.UUID) error {
	return nil
}
func (m *mockPeopleRepo) UpdateProfile(ctx context.Context, personID uuid.UUID, firstName, lastName, displayName string) error {
	return nil
}
func (m *mockPeopleRepo) SaveCredentials(ctx context.Context, personID uuid.UUID, cred people.Credentials) error {
	return nil
}

type mockInvitationRepo struct {
	createErr   error
	inv         *people.Invitation
	invList     []*people.Invitation
	findByIDErr error
	deleteErr   error
	completeErr error
	emailExists bool
	existsErr   error
}

func (m *mockInvitationRepo) Create(ctx context.Context, inv *people.Invitation) error {
	return m.createErr
}
func (m *mockInvitationRepo) FindByID(ctx context.Context, id uuid.UUID) (*people.Invitation, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.inv, nil
}
func (m *mockInvitationRepo) FindByCodeHash(ctx context.Context, codeHash string) (*people.Invitation, error) {
	if m.findByIDErr != nil {
		return nil, m.findByIDErr
	}
	return m.inv, nil
}
func (m *mockInvitationRepo) FindByCreator(ctx context.Context, creatorID uuid.UUID) ([]*people.Invitation, error) {
	return m.invList, nil
}
func (m *mockInvitationRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return m.deleteErr
}
func (m *mockInvitationRepo) Complete(ctx context.Context, id uuid.UUID) error {
	return m.completeErr
}
func (m *mockInvitationRepo) EmailExists(ctx context.Context, email string) (bool, error) {
	return m.emailExists, m.existsErr
}

type mockTokenProvider struct {
	genErr     error
	decErr     error
	claims     *core.AuthTokenClaims
	valRef     bool
}

func (m *mockTokenProvider) GenerateTokenRing(data core.AuthTokenPayload) (*core.AuthResponse, error) {
	if m.genErr != nil {
		return nil, m.genErr
	}
	return &core.AuthResponse{AccessToken: "acc", RefreshToken: "ref"}, nil
}
func (m *mockTokenProvider) ValidateToken(token string) (*core.AuthTokenClaims, error) {
	return m.claims, nil
}
func (m *mockTokenProvider) DecodeToken(token string) (*core.AuthTokenClaims, error) {
	if m.decErr != nil {
		return nil, m.decErr
	}
	return m.claims, nil
}
func (m *mockTokenProvider) ValidateRefreshToken(tokenID uuid.UUID, refreshToken string) bool {
	return m.valRef
}

type mockThirdPartyAuth struct {
	claims *core.AuthTokenClaims
	err    error
}

func (m *mockThirdPartyAuth) ValidateToken(token string) (*core.AuthTokenClaims, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.claims, nil
}

func TestEmailAddress(t *testing.T) {
	email, err := people.NewEmailAddress("user@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if email.String() != "user@example.com" {
		t.Errorf("unexpected string: %s", email.String())
	}

	other, _ := people.NewEmailAddress("user@example.com")
	if !email.Equals(other) {
		t.Errorf("expected emails to be equal")
	}

	different, _ := people.NewEmailAddress("other@example.com")
	if email.Equals(different) {
		t.Errorf("expected emails not equal")
	}

	_, err = people.NewEmailAddress("invalid-email")
	if err == nil {
		t.Errorf("expected error for invalid email")
	}
}

func TestCredentials(t *testing.T) {
	cred, err := people.NewCredentials("LOCAL", "user@example.com", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cred.Provider() != people.LOCAL {
		t.Errorf("expected LOCAL provider")
	}
	if cred.RegistrationID() != "user@example.com" {
		t.Errorf("unexpected registration ID")
	}
	if cred.Identifier() != "secret" {
		t.Errorf("unexpected identifier")
	}

	otherCred, _ := people.NewCredentials("LOCAL", "user@example.com", "secret")
	if !cred.Equals(otherCred) {
		t.Errorf("expected credentials equal")
	}

	updatedCred := cred.SetIdentifier("new_secret")
	if updatedCred.Identifier() != "new_secret" {
		t.Errorf("unexpected updated identifier")
	}

	// Invalid provider
	_, err = people.NewCredentials("UNKNOWN", "id", "ident")
	if err == nil {
		t.Errorf("expected error for unknown provider")
	}

	// Empty ID for LOCAL
	_, err = people.NewCredentials("LOCAL", "", "ident")
	if err == nil {
		t.Errorf("expected error for empty id on LOCAL")
	}

	// Empty identity
	_, err = people.NewCredentials("GOOGLE", "id", "")
	if err == nil {
		t.Errorf("expected error for empty identity")
	}
}

func TestPerson(t *testing.T) {
	cred, _ := people.NewCredentials("LOCAL", "user@example.com", "secret")
	pID := uuid.New()
	p, err := people.NewPerson(pID, "John", "Doe", "J.D.", []people.Credentials{cred})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if p.ID() != pID {
		t.Errorf("ID mismatch")
	}
	if p.FirstName() != "John" || p.LastName() != "Doe" {
		t.Errorf("Name mismatch")
	}
	if p.FullName() != "John Doe" {
		t.Errorf("FullName mismatch: %s", p.FullName())
	}
	if p.DisplayName() != "J.D." {
		t.Errorf("DisplayName mismatch")
	}
	if len(p.Identities()) != 1 {
		t.Errorf("Identities count mismatch")
	}

	// DisplayName fallback to FullName
	pNoDisp, _ := people.NewPerson(pID, "John", "Doe", "", []people.Credentials{cred})
	if pNoDisp.DisplayName() != "John Doe" {
		t.Errorf("expected fallback to FullName")
	}

	// Identify success & failure
	ok, err := p.Identify(cred)
	if !ok || err != nil {
		t.Errorf("expected identify success")
	}

	wrongCred, _ := people.NewCredentials("GOOGLE", "other@example.com", "token")
	ok, err = p.Identify(wrongCred)
	if ok || err == nil {
		t.Errorf("expected identify failure for wrong creds")
	}

	// AddOrReplaceIdentity
	googleCred, _ := people.NewCredentials("GOOGLE", "google_id", "google_token")
	err = p.AddOrReplaceIdentity(googleCred)
	if err != nil || len(p.Identities()) != 2 {
		t.Errorf("expected 2 identities after add")
	}

	// Replace existing LOCAL identity
	newLocalCred, _ := people.NewCredentials("LOCAL", "user@example.com", "new_hash")
	err = p.AddOrReplaceIdentity(newLocalCred)
	if err != nil || len(p.Identities()) != 2 {
		t.Errorf("expected identity count remains 2 after replace")
	}

	// NewPerson validation errors
	_, err = people.NewPerson(uuid.Nil, "J", "Doe", "", []people.Credentials{cred}) // short first name
	if err == nil {
		t.Errorf("expected error for short first name")
	}

	_, err = people.NewPerson(uuid.Nil, "John", "D", "", []people.Credentials{cred}) // short last name
	if err == nil {
		t.Errorf("expected error for short last name")
	}

	_, err = people.NewPerson(uuid.Nil, "John", "Doe", "", []people.Credentials{}) // empty identities slice
	if err == nil {
		t.Errorf("expected error for empty identities")
	}
}

func TestUserService(t *testing.T) {
	personID := uuid.New()
	cred, _ := people.NewCredentials("LOCAL", "user@example.com", "secret")
	p, _ := people.NewPerson(personID, "John", "Doe", "J.D.", []people.Credentials{cred})

	repo := &mockPeopleRepo{person: p}
	hasher := &mockHasher{compareRes: true}

	svc := people.NewCustomerService(&mockDb{}, repo, hasher)

	// FindByID
	found, err := svc.FindByID(context.Background(), personID)
	if err != nil || found != p {
		t.Errorf("FindByID error: %v", err)
	}

	// Register LOCAL success
	regCmdLocal := people.RegisterCommand{
		IdentityProvider: people.LOCAL,
		FirstName:        "Jane",
		LastName:         "Doe",
		Login:            "jane@example.com",
		Token:            "secret123",
	}
	err = svc.Register(context.Background(), regCmdLocal)
	if err != nil {
		t.Fatalf("unexpected error registering local user: %v", err)
	}

	// Register GOOGLE success
	regCmdGoogle := people.RegisterCommand{
		IdentityProvider: people.GOOGLE,
		FirstName:        "Jane",
		LastName:         "Doe",
		Login:            "google_id",
		Token:            "google_token",
	}
	err = svc.Register(context.Background(), regCmdGoogle)
	if err != nil {
		t.Fatalf("unexpected error registering google user: %v", err)
	}

	// Register error cases
	// Invalid login email for LOCAL
	regInvalidEmail := regCmdLocal
	regInvalidEmail.Login = "invalid-email"
	err = svc.Register(context.Background(), regInvalidEmail)
	if err == nil {
		t.Errorf("expected error for invalid email")
	}

	// Hasher error
	svcHashErr := people.NewCustomerService(&mockDb{}, repo, &mockHasher{hashErr: errors.New("hash err")})
	err = svcHashErr.Register(context.Background(), regCmdLocal)
	if err == nil {
		t.Errorf("expected error when hasher fails")
	}

	// Invalid credentials (e.g. empty token)
	regEmptyToken := regCmdGoogle
	regEmptyToken.Token = ""
	err = svc.Register(context.Background(), regEmptyToken)
	if err == nil {
		t.Errorf("expected error for empty token")
	}

	// Person creation error (e.g. short name)
	regShortName := regCmdGoogle
	regShortName.FirstName = "A"
	err = svc.Register(context.Background(), regShortName)
	if err == nil {
		t.Errorf("expected error for short name")
	}

	// Repo register error
	svcRepoErr := people.NewCustomerService(&mockDb{}, &mockPeopleRepo{registerErr: errors.New("reg err")}, hasher)
	err = svcRepoErr.Register(context.Background(), regCmdLocal)
	if err == nil {
		t.Errorf("expected repo error")
	}
}

func TestAuthService(t *testing.T) {
	personID := uuid.New()
	cred, _ := people.NewCredentials("LOCAL", "user@example.com", "secret")
	p, _ := people.NewPerson(personID, "John", "Doe", "J.D.", []people.Credentials{cred})

	repo := &mockPeopleRepo{person: p, credPersonID: personID, credHash: "hash_secret"}
	tokenProvider := &mockTokenProvider{
		claims: &core.AuthTokenClaims{
			ID:               uuid.New().String(),
			AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()},
		},
		valRef: true,
	}
	hasher := &mockHasher{compareRes: true}
	googleAuth := &mockThirdPartyAuth{
		claims: &core.AuthTokenClaims{AuthTokenPayload: core.AuthTokenPayload{Sub: "google_sub_123"}},
	}

	newAuth := func(r people.Repository, tp core.AuthTokenProvider, h core.Hasher, g core.ThirdPartyAuth) people.AuthService {
		return people.NewAuthService(&mockDb{}, r, tp, h, g, &mockInvitationRepo{})
	}

	svc := newAuth(repo, tokenProvider, hasher, googleAuth)

	t.Run("Login LOCAL success and errors", func(t *testing.T) {
		res, err := svc.Login(context.Background(), cred)
		if err != nil || res == nil {
			t.Fatalf("unexpected login error: %v", err)
		}

		// FindCredentials error
		svcErr := newAuth(&mockPeopleRepo{findCredErr: errors.New("err")}, tokenProvider, hasher, googleAuth)
		_, err = svcErr.Login(context.Background(), cred)
		if err == nil {
			t.Errorf("expected error when FindCredentials fails")
		}

		// Password compare failure
		svcHashMismatch := newAuth(repo, tokenProvider, &mockHasher{compareRes: false}, googleAuth)
		_, err = svcHashMismatch.Login(context.Background(), cred)
		if err == nil {
			t.Errorf("expected error when password compare fails")
		}

		// FindByID error in authorizePerson
		svcFindErr := newAuth(&mockPeopleRepo{findIDErr: errors.New("err")}, tokenProvider, hasher, googleAuth)
		_, err = svcFindErr.Login(context.Background(), cred)
		if err == nil {
			t.Errorf("expected error when FindByID fails")
		}

		// TokenProvider error
		svcTokErr := newAuth(repo, &mockTokenProvider{genErr: errors.New("gen err")}, hasher, googleAuth)
		_, err = svcTokErr.Login(context.Background(), cred)
		if err == nil {
			t.Errorf("expected error when token generation fails")
		}
	})

	t.Run("Login GOOGLE success and errors", func(t *testing.T) {
		googleCred, _ := people.NewCredentials("GOOGLE", "google_id", "google_token")

		res, err := svc.Login(context.Background(), googleCred)
		if err != nil || res == nil {
			t.Fatalf("unexpected google login error: %v", err)
		}

		// Google validate token failure
		svcGoogleErr := newAuth(repo, tokenProvider, hasher, &mockThirdPartyAuth{err: errors.New("google err")})
		_, err = svcGoogleErr.Login(context.Background(), googleCred)
		if err == nil {
			t.Errorf("expected error when google validate token fails")
		}

		// FindCredentials error
		svcCredErr := newAuth(&mockPeopleRepo{findCredErr: errors.New("err")}, tokenProvider, hasher, googleAuth)
		_, err = svcCredErr.Login(context.Background(), googleCred)
		if err == nil {
			t.Errorf("expected error when FindCredentials fails for google")
		}
	})

	t.Run("Login Unsupported provider", func(t *testing.T) {
		unsupportedCred, _ := people.NewCredentials("FACEBOOK", "id", "token")
		_, err := svc.Login(context.Background(), unsupportedCred)
		if err == nil {
			t.Errorf("expected error for unsupported provider")
		}
	})

	t.Run("Refresh success and error branches", func(t *testing.T) {
		ring, _ := core.NewTokenRing("access", "refresh")

		res, err := svc.Refresh(context.Background(), ring)
		if err != nil || res == nil {
			t.Fatalf("unexpected refresh error: %v", err)
		}

		// DecodeToken failure
		svcDecErr := newAuth(repo, &mockTokenProvider{decErr: errors.New("dec err")}, hasher, googleAuth)
		_, err = svcDecErr.Refresh(context.Background(), ring)
		if err == nil {
			t.Errorf("expected error when DecodeToken fails")
		}

		// Invalid claims.ID (not a UUID)
		svcInvalidID := newAuth(repo, &mockTokenProvider{claims: &core.AuthTokenClaims{ID: "invalid-uuid"}}, hasher, googleAuth)
		_, err = svcInvalidID.Refresh(context.Background(), ring)
		if err == nil {
			t.Errorf("expected error for invalid claim ID")
		}

		// ValidateRefreshToken fails
		svcInvalidRef := newAuth(repo, &mockTokenProvider{claims: tokenProvider.claims, valRef: false}, hasher, googleAuth)
		_, err = svcInvalidRef.Refresh(context.Background(), ring)
		if err == nil {
			t.Errorf("expected error when ValidateRefreshToken fails")
		}

		// Invalid claims.Sub (not a UUID)
		svcInvalidSub := newAuth(repo, &mockTokenProvider{claims: &core.AuthTokenClaims{ID: uuid.New().String(), AuthTokenPayload: core.AuthTokenPayload{Sub: "invalid-uuid"}}, valRef: true}, hasher, googleAuth)
		_, err = svcInvalidSub.Refresh(context.Background(), ring)
		if err == nil {
			t.Errorf("expected error for invalid claim Sub")
		}

		// FindByID fails
		svcFindErr := newAuth(&mockPeopleRepo{findIDErr: errors.New("find err")}, tokenProvider, hasher, googleAuth)
		_, err = svcFindErr.Refresh(context.Background(), ring)
		if err == nil {
			t.Errorf("expected error when FindByID fails")
		}

		// GenerateTokenRing fails
		svcGenErr := newAuth(repo, &mockTokenProvider{claims: tokenProvider.claims, valRef: true, genErr: errors.New("gen err")}, hasher, googleAuth)
		_, err = svcGenErr.Refresh(context.Background(), ring)
		if err == nil {
			t.Errorf("expected error when GenerateTokenRing fails")
		}
	})
}

func TestPersonRoles(t *testing.T) {
	// Role value object validation
	if !people.RoleUser.IsValid() {
		t.Errorf("expected RoleUser to be valid")
	}
	if !people.RoleAdministrator.IsValid() {
		t.Errorf("expected RoleAdministrator to be valid")
	}
	if people.Role("InvalidRole").IsValid() {
		t.Errorf("expected InvalidRole to be invalid")
	}
	if people.RoleUser.String() != "User" {
		t.Errorf("unexpected string representation: %s", people.RoleUser.String())
	}

	rUser, err := people.ToRole("User")
	if err != nil || rUser != people.RoleUser {
		t.Errorf("ToRole(User) failed: %v", err)
	}
	rAdmin, err := people.ToRole("Administrator")
	if err != nil || rAdmin != people.RoleAdministrator {
		t.Errorf("ToRole(Administrator) failed: %v", err)
	}
	_, err = people.ToRole("Superuser")
	if err == nil {
		t.Errorf("expected error for invalid role string")
	}

	// Person entity encapsulation
	cred, _ := people.NewCredentials(people.LOCAL, "alice@example.com", "pass")
	p, err := people.NewPerson(uuid.New(), "Alice", "Smith", "Alice", []people.Credentials{cred}, people.RoleUser)
	if err != nil {
		t.Fatalf("unexpected NewPerson error: %v", err)
	}

	if !p.IsUser() || !p.CanHaveProjects() {
		t.Errorf("expected user to be User and CanHaveProjects")
	}
	if p.IsAdministrator() || p.CanManageUsers() {
		t.Errorf("expected user not to be Administrator or CanManageUsers")
	}
	if len(p.Roles()) != 1 || p.Roles()[0] != people.RoleUser {
		t.Errorf("unexpected roles: %v", p.Roles())
	}

	// Assign roles: User + Administrator
	err = p.AssignRoles([]people.Role{people.RoleUser, people.RoleAdministrator})
	if err != nil {
		t.Fatalf("AssignRoles error: %v", err)
	}
	if !p.IsUser() || !p.IsAdministrator() {
		t.Errorf("expected person to have both User and Administrator roles")
	}
	if !p.CanHaveProjects() || !p.CanManageUsers() {
		t.Errorf("expected person with dual roles to have both capabilities")
	}

	// Assign strictly Administrator
	err = p.AssignRoles([]people.Role{people.RoleAdministrator})
	if err != nil {
		t.Fatalf("AssignRoles error: %v", err)
	}
	if p.CanHaveProjects() {
		t.Errorf("administrator without User role should not be able to have projects")
	}
	if !p.CanManageUsers() {
		t.Errorf("administrator should be able to manage users")
	}

	// Assign invalid role
	err = p.AssignRoles([]people.Role{people.Role("Hacker")})
	if err == nil {
		t.Errorf("expected error assigning invalid role")
	}

	// AddRole
	err = p.AddRole(people.RoleUser)
	if err != nil || !p.HasRole(people.RoleUser) {
		t.Errorf("AddRole failed: %v", err)
	}
	// Add existing role is idempotent
	err = p.AddRole(people.RoleUser)
	if err != nil {
		t.Errorf("AddRole duplicate failed: %v", err)
	}

	// Add invalid role
	err = p.AddRole(people.Role("BadRole"))
	if err == nil {
		t.Errorf("expected error adding invalid role")
	}
}

func TestRegisterDefaultUserRole(t *testing.T) {
	var savedPerson *people.Person
	repo := &mockPeopleRepo{
		person: nil,
	}
	// override Register to capture saved person
	hasher := &mockHasher{}
	db := &mockDb{}
	svc := people.NewCustomerService(db, repo, hasher)

	err := svc.Register(context.Background(), people.RegisterCommand{
		Login:            "test@example.com",
		FirstName:        "John",
		LastName:         "Doe",
		IdentityProvider: people.LOCAL,
		Token:            "secret",
	})
	if err != nil {
		t.Fatalf("unexpected Register error: %v", err)
	}
	_ = savedPerson
}

func TestUserService_FindAll_And_AssignRoles(t *testing.T) {
	pID := uuid.New()
	cred, _ := people.NewCredentials(people.LOCAL, "bob@example.com", "pass")
	p, _ := people.NewPerson(pID, "Bob", "Smith", "Bob", []people.Credentials{cred}, people.RoleUser)

	repo := &mockPeopleRepo{
		person:     p,
		peopleList: []*people.Person{p},
		total:      1,
	}
	db := &mockDb{}
	svc := people.NewCustomerService(db, repo, &mockHasher{})

	// FindAll
	list, total, err := svc.FindAll(context.Background(), core.Pagination{Limit: 10, Offset: 0})
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("FindAll failed: %v", err)
	}

	// AssignRoles
	err = svc.AssignRoles(context.Background(), people.AssignRolesCommand{
		PersonID: pID,
		Roles:    []people.Role{people.RoleUser, people.RoleAdministrator},
	})
	if err != nil {
		t.Fatalf("AssignRoles failed: %v", err)
	}
	if len(repo.savedRoles) != 2 {
		t.Errorf("expected 2 saved roles, got %v", repo.savedRoles)
	}

	// AssignRoles invalid role
	err = svc.AssignRoles(context.Background(), people.AssignRolesCommand{
		PersonID: pID,
		Roles:    []people.Role{people.Role("Unknown")},
	})
	if err == nil {
		t.Errorf("expected error when assigning invalid role")
	}

	// AssignRoles person not found
	svcNotFound := people.NewCustomerService(db, &mockPeopleRepo{findIDErr: errors.New("not found")}, &mockHasher{})
	err = svcNotFound.AssignRoles(context.Background(), people.AssignRolesCommand{
		PersonID: uuid.New(),
		Roles:    []people.Role{people.RoleUser},
	})
	if err == nil {
		t.Errorf("expected error when person not found")
	}
}

func TestAuthService_TokenRolesPropagation(t *testing.T) {
	pID := uuid.New()
	cred, _ := people.NewCredentials(people.LOCAL, "carol@example.com", "pass")
	p, _ := people.NewPerson(pID, "Carol", "Danvers", "Captain", []people.Credentials{cred}, people.RoleAdministrator, people.RoleUser)

	var capturedPayload core.AuthTokenPayload
	captureTokenProvider := &mockTokenProvider{
		claims: &core.AuthTokenClaims{
			ID: uuid.New().String(),
			AuthTokenPayload: core.AuthTokenPayload{
				Sub:   pID.String(),
				Roles: p.RoleStrings(),
			},
		},
		valRef: true,
	}

	repo := &mockPeopleRepo{
		person:       p,
		credPersonID: pID,
		credHash:     "hash_pass",
	}
	hasher := &mockHasher{compareRes: true}
	svc := people.NewAuthService(&mockDb{}, repo, captureTokenProvider, hasher, &mockThirdPartyAuth{}, &mockInvitationRepo{})

	// Login
	resp, err := svc.Login(context.Background(), cred)
	if err != nil || resp == nil {
		t.Fatalf("Login error: %v", err)
	}

	// Refresh
	ring, _ := core.NewTokenRing("acc", "ref")
	respRef, err := svc.Refresh(context.Background(), ring)
	if err != nil || respRef == nil {
		t.Fatalf("Refresh error: %v", err)
	}
	_ = capturedPayload
}

