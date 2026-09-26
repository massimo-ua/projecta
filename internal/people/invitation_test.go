package people_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/people"
)

func TestInvitationModel(t *testing.T) {
	adminID := uuid.New()
	personID := uuid.New()
	codeHash := people.HashInvitationCode("test_code_123")
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	t.Run("NewInvitation valid", func(t *testing.T) {
		inv, err := people.NewInvitation(
			uuid.Nil,
			"user@example.com",
			codeHash,
			expiresAt,
			adminID,
			&personID,
			people.InvitationStatusPending,
			now,
			nil,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if inv.ID() == uuid.Nil {
			t.Errorf("expected generated UUID")
		}
		if inv.Email() != "user@example.com" {
			t.Errorf("expected email user@example.com, got %s", inv.Email())
		}
		if inv.CodeHash() != codeHash {
			t.Errorf("code hash mismatch")
		}
		if inv.CreatedBy() != adminID {
			t.Errorf("created by mismatch")
		}
		if inv.PersonID() == nil || *inv.PersonID() != personID {
			t.Errorf("person ID mismatch")
		}
		if !inv.IsPending() {
			t.Errorf("expected invitation to be pending")
		}
		if inv.IsCompleted() {
			t.Errorf("expected invitation not to be completed")
		}
		if inv.IsExpired() {
			t.Errorf("expected invitation not to be expired")
		}
	})

	t.Run("NewInvitation validation errors", func(t *testing.T) {
		_, err := people.NewInvitation(uuid.Nil, "", codeHash, expiresAt, adminID, nil, "", now, nil)
		if err == nil {
			t.Errorf("expected error for empty email")
		}

		_, err = people.NewInvitation(uuid.Nil, "user@example.com", "", expiresAt, adminID, nil, "", now, nil)
		if err == nil {
			t.Errorf("expected error for empty code hash")
		}

		_, err = people.NewInvitation(uuid.Nil, "user@example.com", codeHash, expiresAt, uuid.Nil, nil, "", now, nil)
		if err == nil {
			t.Errorf("expected error for empty created by")
		}
	})

	t.Run("Complete and Expired", func(t *testing.T) {
		past := now.Add(-1 * time.Hour)
		expiredInv, _ := people.NewInvitation(uuid.Nil, "user@example.com", codeHash, past, adminID, nil, people.InvitationStatusPending, now, nil)
		if !expiredInv.IsExpired() {
			t.Errorf("expected invitation to be expired")
		}
		if expiredInv.IsPending() {
			t.Errorf("expired invitation should not be pending")
		}
		if err := expiredInv.Complete(); err == nil {
			t.Errorf("completing expired invitation should error")
		}

		validInv, _ := people.NewInvitation(uuid.Nil, "user@example.com", codeHash, expiresAt, adminID, nil, people.InvitationStatusPending, now, nil)
		if err := validInv.Complete(); err != nil {
			t.Fatalf("unexpected Complete error: %v", err)
		}
		if !validInv.IsCompleted() {
			t.Errorf("expected completed status")
		}
		if validInv.CompletedAt() == nil {
			t.Errorf("expected non-nil completed_at")
		}
		// Complete again should error
		if err := validInv.Complete(); err == nil {
			t.Errorf("completing already completed invitation should error")
		}
	})
}

func TestInvitationService(t *testing.T) {
	adminID := uuid.New()
	targetEmail := "newuser@example.com"

	t.Run("Create invitation success", func(t *testing.T) {
		invRepo := &mockInvitationRepo{emailExists: false}
		peopleRepo := &mockPeopleRepo{}
		db := &mockDb{}

		svc := people.NewInvitationService(db, invRepo, peopleRepo)

		invWithCode, err := svc.Create(context.Background(), adminID, targetEmail)
		if err != nil {
			t.Fatalf("unexpected Create error: %v", err)
		}
		if invWithCode.Code == "" {
			t.Errorf("expected raw code in return")
		}
		if invWithCode.Invitation == nil {
			t.Fatalf("expected invitation object")
		}
		if invWithCode.Invitation.Email() != targetEmail {
			t.Errorf("email mismatch: %s vs %s", invWithCode.Invitation.Email(), targetEmail)
		}
		if invWithCode.Invitation.CreatedBy() != adminID {
			t.Errorf("createdBy mismatch")
		}
		if invWithCode.Invitation.PersonID() == nil {
			t.Errorf("expected precreated person ID")
		}
		// Verify expiration is set to roughly 7 days
		diff := invWithCode.Invitation.ExpiresAt().Sub(time.Now())
		if diff < 6*24*time.Hour || diff > 8*24*time.Hour {
			t.Errorf("expected expiration around 7 days, got %v", diff)
		}
	})

	t.Run("Create invitation fails when email already exists", func(t *testing.T) {
		invRepo := &mockInvitationRepo{emailExists: true}
		peopleRepo := &mockPeopleRepo{}
		svc := people.NewInvitationService(&mockDb{}, invRepo, peopleRepo)

		_, err := svc.Create(context.Background(), adminID, targetEmail)
		if err == nil {
			t.Errorf("expected error when email already exists")
		}
	})

	t.Run("Create invitation fails with invalid email or nil admin", func(t *testing.T) {
		svc := people.NewInvitationService(&mockDb{}, &mockInvitationRepo{}, &mockPeopleRepo{})

		_, err := svc.Create(context.Background(), uuid.Nil, targetEmail)
		if err == nil {
			t.Errorf("expected error for nil admin ID")
		}

		_, err = svc.Create(context.Background(), adminID, "invalid-email")
		if err == nil {
			t.Errorf("expected error for invalid email")
		}
	})

	t.Run("FindAllByCreator", func(t *testing.T) {
		sampleInv, _ := people.NewInvitation(uuid.New(), targetEmail, "hash", time.Now().Add(time.Hour), adminID, nil, people.InvitationStatusPending, time.Now(), nil)
		invRepo := &mockInvitationRepo{invList: []*people.Invitation{sampleInv}}
		svc := people.NewInvitationService(&mockDb{}, invRepo, &mockPeopleRepo{})

		list, err := svc.FindAllByCreator(context.Background(), adminID)
		if err != nil || len(list) != 1 {
			t.Fatalf("FindAllByCreator failed: %v", err)
		}

		_, err = svc.FindAllByCreator(context.Background(), uuid.Nil)
		if err == nil {
			t.Errorf("expected error for nil admin ID")
		}
	})

	t.Run("Delete invitation removes precreated person if pending", func(t *testing.T) {
		personID := uuid.New()
		invID := uuid.New()
		inv, _ := people.NewInvitation(invID, targetEmail, "hash", time.Now().Add(time.Hour), adminID, &personID, people.InvitationStatusPending, time.Now(), nil)

		invRepo := &mockInvitationRepo{inv: inv}
		peopleRepo := &mockPeopleRepo{}
		svc := people.NewInvitationService(&mockDb{}, invRepo, peopleRepo)

		err := svc.Delete(context.Background(), adminID, invID)
		if err != nil {
			t.Fatalf("unexpected Delete error: %v", err)
		}

		// Unauthorized admin
		otherAdmin := uuid.New()
		err = svc.Delete(context.Background(), otherAdmin, invID)
		if err == nil {
			t.Errorf("expected error when non-creator deletes invitation")
		}
	})

	t.Run("FindByCode checks code hash, expiration, and completion", func(t *testing.T) {
		rawCode := "secret_invite_code"
		codeHash := people.HashInvitationCode(rawCode)

		validInv, _ := people.NewInvitation(uuid.New(), targetEmail, codeHash, time.Now().Add(time.Hour), adminID, nil, people.InvitationStatusPending, time.Now(), nil)
		invRepo := &mockInvitationRepo{inv: validInv}
		svc := people.NewInvitationService(&mockDb{}, invRepo, &mockPeopleRepo{})

		found, err := svc.FindByCode(context.Background(), rawCode)
		if err != nil || found == nil {
			t.Fatalf("FindByCode failed: %v", err)
		}

		// Expired
		expiredInv, _ := people.NewInvitation(uuid.New(), targetEmail, codeHash, time.Now().Add(-1*time.Hour), adminID, nil, people.InvitationStatusPending, time.Now(), nil)
		invRepoExpired := &mockInvitationRepo{inv: expiredInv}
		svcExpired := people.NewInvitationService(&mockDb{}, invRepoExpired, &mockPeopleRepo{})
		_, err = svcExpired.FindByCode(context.Background(), rawCode)
		if err == nil {
			t.Errorf("expected error for expired invitation")
		}

		// Completed
		now := time.Now()
		compInv, _ := people.NewInvitation(uuid.New(), targetEmail, codeHash, time.Now().Add(time.Hour), adminID, nil, people.InvitationStatusCompleted, time.Now(), &now)
		invRepoComp := &mockInvitationRepo{inv: compInv}
		svcComp := people.NewInvitationService(&mockDb{}, invRepoComp, &mockPeopleRepo{})
		_, err = svcComp.FindByCode(context.Background(), rawCode)
		if err == nil {
			t.Errorf("expected error for completed invitation")
		}
	})
}

func TestLoginWithInvitation(t *testing.T) {
	adminID := uuid.New()
	personID := uuid.New()
	targetEmail := "user@gmail.com"
	rawCode := "valid_google_invitation_code_123"
	codeHash := people.HashInvitationCode(rawCode)

	precreatedPerson, _ := people.NewPerson(personID, "Invited", "User", targetEmail, nil, people.RoleUser)
	validInv, _ := people.NewInvitation(uuid.New(), targetEmail, codeHash, time.Now().Add(7*24*time.Hour), adminID, &personID, people.InvitationStatusPending, time.Now(), nil)

	peopleRepo := &mockPeopleRepo{person: precreatedPerson}
	invRepo := &mockInvitationRepo{inv: validInv}
	tokenProvider := &mockTokenProvider{claims: &core.AuthTokenClaims{AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()}}}
	hasher := &mockHasher{compareRes: true}
	googleAuth := &mockThirdPartyAuth{
		claims: &core.AuthTokenClaims{
			AuthTokenPayload: core.AuthTokenPayload{
				Sub:         "google_sub_999",
				Email:       targetEmail,
				DisplayName: "Jane Doe",
				FirstName:   "Jane",
				LastName:    "Doe",
			},
		},
	}

	svc := people.NewAuthService(&mockDb{}, peopleRepo, tokenProvider, hasher, googleAuth, invRepo)

	t.Run("LoginWithInvitation success updates profile and completes invitation", func(t *testing.T) {
		res, err := svc.LoginWithInvitation(context.Background(), "valid_google_token", people.GOOGLE, rawCode)
		if err != nil || res == nil {
			t.Fatalf("unexpected LoginWithInvitation error: %v", err)
		}
		if precreatedPerson.FirstName() != "Jane" || precreatedPerson.LastName() != "Doe" {
			t.Errorf("person profile was not updated from Google claims: got %s %s", precreatedPerson.FirstName(), precreatedPerson.LastName())
		}
	})

	t.Run("LoginWithInvitation fails on email mismatch", func(t *testing.T) {
		mismatchedGoogle := &mockThirdPartyAuth{
			claims: &core.AuthTokenClaims{
				AuthTokenPayload: core.AuthTokenPayload{
					Sub:   "google_sub_888",
					Email: "different@gmail.com",
				},
			},
		}
		svcMismatch := people.NewAuthService(&mockDb{}, peopleRepo, tokenProvider, hasher, mismatchedGoogle, invRepo)
		_, err := svcMismatch.LoginWithInvitation(context.Background(), "token", people.GOOGLE, rawCode)
		if err == nil || !strings.Contains(err.Error(), "does not match") {
			t.Errorf("expected email mismatch error, got: %v", err)
		}
	})

	t.Run("LoginWithInvitation fails on unsupported provider", func(t *testing.T) {
		_, err := svc.LoginWithInvitation(context.Background(), "token", people.LOCAL, rawCode)
		if err == nil {
			t.Errorf("expected error for non-GOOGLE provider")
		}
	})

	t.Run("LoginWithInvitation fails on empty invitation code", func(t *testing.T) {
		if _, err := svc.LoginWithInvitation(context.Background(), "token", people.GOOGLE, ""); err == nil {
			t.Errorf("expected error for empty invitation code")
		}
		if _, err := svc.LoginWithInvitation(context.Background(), "token", people.GOOGLE, "   "); err == nil {
			t.Errorf("expected error for whitespace invitation code")
		}
	})

	t.Run("LoginWithInvitation fails when invitation is completed", func(t *testing.T) {
		now := time.Now()
		completedInv, _ := people.NewInvitation(uuid.New(), targetEmail, codeHash, time.Now().Add(time.Hour), adminID, &personID, people.InvitationStatusCompleted, time.Now(), &now)
		svcCompleted := people.NewAuthService(&mockDb{}, peopleRepo, tokenProvider, hasher, googleAuth, &mockInvitationRepo{inv: completedInv})

		_, err := svcCompleted.LoginWithInvitation(context.Background(), "token", people.GOOGLE, rawCode)
		if err == nil {
			t.Errorf("expected error for completed invitation")
		}
	})

	t.Run("LoginWithInvitation fails when invitation is expired", func(t *testing.T) {
		expiredInv, _ := people.NewInvitation(uuid.New(), targetEmail, codeHash, time.Now().Add(-1*time.Hour), adminID, &personID, people.InvitationStatusPending, time.Now(), nil)
		svcExpired := people.NewAuthService(&mockDb{}, peopleRepo, tokenProvider, hasher, googleAuth, &mockInvitationRepo{inv: expiredInv})

		_, err := svcExpired.LoginWithInvitation(context.Background(), "token", people.GOOGLE, rawCode)
		if err == nil {
			t.Errorf("expected error for expired invitation")
		}
	})

	t.Run("LoginWithInvitation fails on Google validation error", func(t *testing.T) {
		errGoogle := &mockThirdPartyAuth{err: errors.New("bad token")}
		svcBadGoogle := people.NewAuthService(&mockDb{}, peopleRepo, tokenProvider, hasher, errGoogle, invRepo)

		_, err := svcBadGoogle.LoginWithInvitation(context.Background(), "bad_token", people.GOOGLE, rawCode)
		if err == nil {
			t.Errorf("expected error when Google validation fails")
		}
	})
}

func TestInvitationCode(t *testing.T) {
	t.Run("fails on empty or whitespace code", func(t *testing.T) {
		for _, invalid := range []string{"", " ", "\t\n", "   "} {
			_, err := people.NewInvitationCode(invalid)
			if err == nil {
				t.Errorf("expected error for empty/whitespace code %q, got nil", invalid)
			}
		}
	})

	t.Run("normalizes whitespace and casing", func(t *testing.T) {
		code, err := people.NewInvitationCode("  AbCdEf123  ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if code.String() != "abcdef123" {
			t.Errorf("expected abcdef123, got %s", code.String())
		}
	})

	t.Run("generates valid random invitation code", func(t *testing.T) {
		code1, err := people.GenerateInvitationCode()
		if err != nil {
			t.Fatalf("unexpected error generating code 1: %v", err)
		}
		code2, err := people.GenerateInvitationCode()
		if err != nil {
			t.Fatalf("unexpected error generating code 2: %v", err)
		}

		if len(code1.String()) != 64 {
			t.Errorf("expected 64 hex chars, got %d", len(code1.String()))
		}
		if code1.Equals(code2) {
			t.Errorf("expected two generated codes to be distinct")
		}
		if code1.Hash() == "" {
			t.Errorf("expected non-empty hash")
		}
	})

	t.Run("computes consistent SHA-256 hash", func(t *testing.T) {
		c1, _ := people.NewInvitationCode("MY_SECRET_CODE")
		c2, _ := people.NewInvitationCode("  my_secret_code  ")
		if !c1.Equals(c2) {
			t.Errorf("expected c1 to equal c2 after normalization")
		}
		if c1.Hash() != c2.Hash() {
			t.Errorf("expected hash equality, got %s vs %s", c1.Hash(), c2.Hash())
		}
	})

	t.Run("zero value handling", func(t *testing.T) {
		var zero people.InvitationCode
		if zero.String() != "" {
			t.Errorf("expected empty string for zero value")
		}
		if zero.Hash() != "" {
			t.Errorf("expected empty hash for zero value")
		}
	})
}
