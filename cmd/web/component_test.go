package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/pkg/crypto"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
	"gitlab.com/massimo-ua/projecta/pkg/dal"
	"gitlab.com/massimo-ua/projecta/pkg/web"
)

var (
	componentMigrateOnce sync.Once
	componentMigrateErr  error
)

type staticComponentRateProvider struct{}

func (staticComponentRateProvider) Convert(a currency.Currency, b currency.Currency) (currency.Currency, error) {
	if a.Code == b.Code {
		return a, nil
	}
	if a.Code == "USD" && b.Code == "UAH" {
		return currency.Currency{Amount: a.Amount * 40, Code: "UAH"}, nil
	}
	if a.Code == "UAH" && b.Code == "USD" {
		return currency.Currency{Amount: a.Amount / 40, Code: "USD"}, nil
	}
	return currency.Currency{Amount: a.Amount, Code: b.Code}, nil
}

type stubComponentGoogleAuth struct{}

func (stubComponentGoogleAuth) ValidateToken(_ string) (*core.AuthTokenClaims, error) {
	return nil, errors.New("google auth disabled in component tests")
}

// ProjectaTestClient wraps HTTP calls in domain-oriented methods for readable component tests.
type ProjectaTestClient struct {
	t       *testing.T
	baseURL string
	client  *http.Client
	token   string
}

func NewProjectaTestClient(t *testing.T, baseURL string) *ProjectaTestClient {
	t.Helper()
	return &ProjectaTestClient{
		t:       t,
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *ProjectaTestClient) do(method, path string, payload any) (int, []byte) {
	c.t.Helper()
	var body io.Reader
	if payload != nil {
		raw, err := json.Marshal(payload)
		if err != nil {
			c.t.Fatalf("marshal payload for %s %s: %v", method, path, err)
		}
		body = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.baseURL+path, body)
	if err != nil {
		c.t.Fatalf("new request %s %s: %v", method, path, err)
	}
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		c.t.Fatalf("execute %s %s: %v", method, path, err)
	}
	defer resp.Body.Close()
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.t.Fatalf("read response %s %s: %v", method, path, err)
	}
	return resp.StatusCode, respBytes
}

func (c *ProjectaTestClient) Login(email, password string) {
	c.t.Helper()
	code, raw := c.do(http.MethodPost, "/login", web.LoginDTO{
		ID:               email,
		Token:            password,
		IdentityProvider: string(people.LOCAL),
	})
	if code != http.StatusOK {
		c.t.Fatalf("Login(%s) failed with %d: %s", email, code, string(raw))
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || out.AccessToken == "" {
		c.t.Fatalf("Login(%s) invalid response: %s (err=%v)", email, string(raw), err)
	}
	c.token = out.AccessToken
}

func (c *ProjectaTestClient) CreateProject(dto web.CreateProjectDTO) web.ProjectDTO {
	c.t.Helper()
	code, raw := c.do(http.MethodPost, "/projects", dto)
	if code != http.StatusCreated {
		c.t.Fatalf("CreateProject failed with %d: %s", code, string(raw))
	}
	var out web.ProjectDTO
	if err := json.Unmarshal(raw, &out); err != nil {
		c.t.Fatalf("unmarshal ProjectDTO: %v", err)
	}
	return out
}

func (c *ProjectaTestClient) GetProjectRaw(projectID string) (int, []byte) {
	c.t.Helper()
	return c.do(http.MethodGet, "/projects/"+projectID, nil)
}

func (c *ProjectaTestClient) DeleteProjectRaw(projectID string) (int, []byte) {
	c.t.Helper()
	return c.do(http.MethodDelete, "/projects/"+projectID, nil)
}

func (c *ProjectaTestClient) UpdateBudget(projectID string, dto web.UpdateProjectBudgetDTO) web.ProjectDTO {
	c.t.Helper()
	code, raw := c.do(http.MethodPut, "/projects/"+projectID+"/budget", dto)
	if code != http.StatusOK {
		c.t.Fatalf("UpdateBudget failed with %d: %s", code, string(raw))
	}
	var out web.ProjectDTO
	if err := json.Unmarshal(raw, &out); err != nil {
		c.t.Fatalf("unmarshal ProjectDTO: %v", err)
	}
	return out
}

func (c *ProjectaTestClient) CreateAsset(projectID string, dto web.CreateAssetDTO) web.AssetDTO {
	c.t.Helper()
	code, raw := c.do(http.MethodPost, "/projects/"+projectID+"/assets", dto)
	if code != http.StatusCreated {
		c.t.Fatalf("CreateAsset failed with %d: %s", code, string(raw))
	}
	var out web.AssetDTO
	if err := json.Unmarshal(raw, &out); err != nil {
		c.t.Fatalf("unmarshal AssetDTO: %v", err)
	}
	return out
}

func (c *ProjectaTestClient) CreateInvestment(projectID string, dto web.CreateInvestmentDTO) web.InvestmentDTO {
	c.t.Helper()
	code, raw := c.do(http.MethodPost, "/projects/"+projectID+"/investments", dto)
	if code != http.StatusCreated {
		c.t.Fatalf("CreateInvestment failed with %d: %s", code, string(raw))
	}
	var out web.InvestmentDTO
	if err := json.Unmarshal(raw, &out); err != nil {
		c.t.Fatalf("unmarshal InvestmentDTO: %v", err)
	}
	return out
}

func (c *ProjectaTestClient) GetTotals(projectID string) web.ProjectTotalsDTO {
	c.t.Helper()
	code, raw := c.do(http.MethodGet, "/projects/"+projectID+"/totals", nil)
	if code != http.StatusOK {
		c.t.Fatalf("GetTotals failed with %d: %s", code, string(raw))
	}
	var out web.ProjectTotalsDTO
	if err := json.Unmarshal(raw, &out); err != nil {
		c.t.Fatalf("unmarshal ProjectTotalsDTO: %v", err)
	}
	return out
}

func (c *ProjectaTestClient) GetCashFlowForecast(projectID, query string) web.CashFlowForecastDTO {
	c.t.Helper()
	code, raw := c.do(http.MethodGet, "/projects/"+projectID+"/cashflow-forecast"+query, nil)
	if code != http.StatusOK {
		c.t.Fatalf("GetCashFlowForecast failed with %d: %s", code, string(raw))
	}
	var out web.CashFlowForecastDTO
	if err := json.Unmarshal(raw, &out); err != nil {
		c.t.Fatalf("unmarshal CashFlowForecastDTO: %v", err)
	}
	return out
}

func (c *ProjectaTestClient) AcceptShare(shareToken string) web.ProjectDTO {
	c.t.Helper()
	code, raw := c.do(http.MethodPost, "/projects/share/"+shareToken, nil)
	if code != http.StatusOK {
		c.t.Fatalf("AcceptShare failed with %d: %s", code, string(raw))
	}
	var out web.ProjectDTO
	if err := json.Unmarshal(raw, &out); err != nil {
		c.t.Fatalf("unmarshal ProjectDTO: %v", err)
	}
	return out
}

func openComponentTestEnv(t *testing.T) (*dal.PgDbConnection, *httptest.Server) {
	t.Helper()

	dbURI := os.Getenv("TEST_DB_URI")
	if dbURI == "" {
		t.Skip("TEST_DB_URI is not set; skipping HTTP component test")
	}

	componentMigrateOnce.Do(func() {
		migrationsDir, err := filepath.Abs(filepath.Join("..", "..", "migrations"))
		if err != nil {
			componentMigrateErr = err
			return
		}
		m, err := migrate.New("file://"+migrationsDir, dbURI)
		if err != nil {
			componentMigrateErr = err
			return
		}
		defer func() { _, _ = m.Close() }()
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			componentMigrateErr = err
		}
	})
	if componentMigrateErr != nil {
		t.Fatalf("failed to apply migrations: %v", componentMigrateErr)
	}

	db, err := dal.NewPgDbConnection(dbURI)
	if err != nil {
		t.Fatalf("failed to connect to component test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	cfg := &core.AppConfig{
		DbUri:              dbURI,
		HttpUri:            ":0",
		GoogleClientID:     "component_test_client",
		GoogleClientSecret: "component_test_secret",
		JwtSecret:          "component_test_jwt_secret_1234567890",
		TokenTTL:           3600,
		GoogleCertTTL:      3600,
		HttpReadTimeout:    10 * time.Second,
		HttpWriteTimeout:   10 * time.Second,
		ShutdownTimeout:    5 * time.Second,
	}

	handler, err := setupAppHandlerWithDeps(cfg, db, appHandlerDeps{
		rateProvider: staticComponentRateProvider{},
		googleAuth:   stubComponentGoogleAuth{},
	})
	if err != nil {
		t.Fatalf("setupAppHandlerWithDeps failed: %v", err)
	}

	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return db, srv
}

func seedComponentUser(t *testing.T, db *dal.PgDbConnection, first, last, password string) string {
	t.Helper()
	hasher := crypto.NewBcryptHasher(4)
	hashed, err := hasher.Hash(password)
	if err != nil {
		t.Fatalf("hash password failed: %v", err)
	}
	id := core.NewPersonID()
	email := fmt.Sprintf("%s@component.test", id.String())
	cred, err := people.NewCredentials(people.LOCAL, email, hashed)
	if err != nil {
		t.Fatalf("NewCredentials failed: %v", err)
	}
	person, err := people.NewPerson(id, first, last, first+" "+last, []people.Credentials{cred}, people.RoleUser)
	if err != nil {
		t.Fatalf("NewPerson failed: %v", err)
	}
	repo := dal.NewPgPeopleRepository(db)
	if err := repo.Save(context.Background(), person); err != nil {
		t.Fatalf("save person failed: %v", err)
	}
	return email
}

func TestHTTPComponent_FullProjectLifecycleAndTenantIsolation(t *testing.T) {
	t.Parallel()
	db, srv := openComponentTestEnv(t)

	ownerEmail := seedComponentUser(t, db, "Alice", "Owner", "SecretPass123!")
	guestEmail := seedComponentUser(t, db, "Bob", "Guest", "SecretPass123!")

	ownerClient := NewProjectaTestClient(t, srv.URL)
	ownerClient.Login(ownerEmail, "SecretPass123!")

	guestClient := NewProjectaTestClient(t, srv.URL)
	guestClient.Login(guestEmail, "SecretPass123!")

	// 1. Create Project
	createdProj := ownerClient.CreateProject(web.CreateProjectDTO{
		Name:         "Component Project " + core.NewProjectID().String()[:8],
		Description:  "End-to-end HTTP + Postgres test",
		MainCurrency: "UAH",
	})
	if createdProj.ProjectID == "" || createdProj.ShareToken == "" {
		t.Fatalf("expected non-empty ProjectID and ShareToken, got %+v", createdProj)
	}

	// 2. Tenant Isolation: Guest cannot view project before accepting share
	statusBeforeShare, _ := guestClient.GetProjectRaw(createdProj.ProjectID)
	if statusBeforeShare != http.StatusNotFound && statusBeforeShare != http.StatusForbidden {
		t.Errorf("expected 404 or 403 for unshared guest GET /projects/{id}, got %d", statusBeforeShare)
	}

	// 3. Configure Tag Budget
	budgetedProj := ownerClient.UpdateBudget(createdProj.ProjectID, web.UpdateProjectBudgetDTO{
		TagBudgets: []web.TagBudgetDTO{
			{
				Tag:          "home",
				Frequency:    "MONTHLY",
				TargetMonth:  1,
				ResourceType: "MONEY",
				Amount:       100000,
				Currency:     "UAH",
				Description:  "Monthly home ceiling",
			},
		},
	})
	if len(budgetedProj.TagBudgets) != 1 {
		t.Fatalf("expected 1 tag budget, got %+v", budgetedProj.TagBudgets)
	}

	// 4. Create Asset & Investment
	createdAsset := ownerClient.CreateAsset(createdProj.ProjectID, web.CreateAssetDTO{
		Name:                 "Heat Pump",
		Description:          "Main heating unit",
		TargetPrice:          60000,
		TargetCurrency:       "UAH",
		StartDate:            "2026-05-01",
		TargetCompletionDate: "2026-07-01",
		Tags:                 []string{"home"},
	})
	if createdAsset.AssetID == "" {
		t.Fatalf("expected created AssetID")
	}

	createdInv := ownerClient.CreateInvestment(createdProj.ProjectID, web.CreateInvestmentDTO{
		AssetID:      createdAsset.AssetID,
		ResourceType: "MONEY",
		Amount:       25000,
		Currency:     "UAH",
		Description:  "Initial deposit",
		Date:         "2026-05-05",
		Tags:         []string{"home"},
	})
	if createdInv.InvestmentID == "" {
		t.Fatalf("expected created InvestmentID")
	}

	// 5. Totals & Cash Flow Forecast queries
	totals := ownerClient.GetTotals(createdProj.ProjectID)
	if len(totals.Totals) != 2 {
		t.Fatalf("expected 2 totals entries, got %+v", totals)
	}
	if totals.Totals[0].Amount != 25000 {
		t.Errorf("expected Total Invested 25000, got %+v", totals.Totals[0])
	}

	forecast := ownerClient.GetCashFlowForecast(createdProj.ProjectID, "?from_month=2026-05&to_month=2026-07&as_of_date=2026-05-10")
	if len(forecast.Months) != 3 {
		t.Fatalf("expected 3 forecast months, got %d", len(forecast.Months))
	}

	// 6. Share Acceptance: Guest can now view project, but still cannot delete it
	sharedProj := guestClient.AcceptShare(createdProj.ShareToken)
	if sharedProj.ProjectID != createdProj.ProjectID || !sharedProj.IsShared {
		t.Errorf("expected guest AcceptShare to return shared project, got %+v", sharedProj)
	}

	statusAfterShare, _ := guestClient.GetProjectRaw(createdProj.ProjectID)
	if statusAfterShare != http.StatusOK {
		t.Errorf("expected 200 OK for shared guest GET /projects/{id}, got %d", statusAfterShare)
	}

	deleteByGuestStatus, _ := guestClient.DeleteProjectRaw(createdProj.ProjectID)
	if deleteByGuestStatus != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden when shared guest tries to delete project, got %d", deleteByGuestStatus)
	}

	// 7. Owner can cleanly delete the project
	deleteByOwnerStatus, _ := ownerClient.DeleteProjectRaw(createdProj.ProjectID)
	if deleteByOwnerStatus != http.StatusNoContent {
		t.Errorf("expected 204 No Content when owner deletes project, got %d", deleteByOwnerStatus)
	}
}
