package web

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

// Mocks for web package tests
type mockPeopleService struct {
	user *people.Person
	err  error
}

func (m *mockPeopleService) FindByID(_ context.Context, _ uuid.UUID) (*people.Person, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.user, nil
}
func (m *mockPeopleService) Register(_ context.Context, _ people.RegisterCommand) error {
	return m.err
}

type mockAuthService struct {
	authResp *core.AuthResponse
	err      error
}

func (m *mockAuthService) Login(_ context.Context, _ people.Credentials) (*core.AuthResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.authResp, nil
}
func (m *mockAuthService) Refresh(_ context.Context, _ *core.TokenRing) (*core.AuthResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.authResp, nil
}

type mockTokenProvider struct {
	claims *core.AuthTokenClaims
	err    error
}

func (m *mockTokenProvider) GenerateTokenRing(_ core.AuthTokenPayload) (*core.AuthResponse, error) {
	return nil, nil
}
func (m *mockTokenProvider) ValidateToken(_ string) (*core.AuthTokenClaims, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.claims, nil
}
func (m *mockTokenProvider) DecodeToken(_ string) (*core.AuthTokenClaims, error) {
	return m.claims, nil
}
func (m *mockTokenProvider) ValidateRefreshToken(_ uuid.UUID, _ string) bool {
	return true
}

type mockProjectService struct {
	project *projecta.Project
	err     error
}

func (m *mockProjectService) Find(_ context.Context, _ projecta.ProjectCollectionFilter) ([]*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*projecta.Project{m.project}, nil
}
func (m *mockProjectService) FindOne(_ context.Context, _ projecta.ProjectFilter) (*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
}
func (m *mockProjectService) Create(_ context.Context, _ projecta.CreateProjectCommand) (*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
}
func (m *mockProjectService) Remove(_ context.Context, _ projecta.RemoveProjectCommand) error {
	return m.err
}
func (m *mockProjectService) Update(_ context.Context, _ projecta.UpdateProjectCommand) (*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
}
func (m *mockProjectService) AcceptShare(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*projecta.Project, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.project, nil
}

type mockCategoryService struct {
	cat *projecta.CostCategory
	err error
}

func (m *mockCategoryService) Find(_ context.Context, _ projecta.CategoryCollectionFilter) (*projecta.CostCategoryCollection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return projecta.NewCategoryCollection(1), nil
}
func (m *mockCategoryService) Create(_ context.Context, _ projecta.CreateCategoryCommand) (*projecta.CostCategory, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.cat, nil
}
func (m *mockCategoryService) Update(_ context.Context, _ projecta.UpdateCategoryCommand) error {
	return m.err
}
func (m *mockCategoryService) Remove(_ context.Context, _ projecta.RemoveCategoryCommand) error {
	return m.err
}

type mockTypeService struct {
	costType *projecta.CostType
	err      error
}

func (m *mockTypeService) Find(_ context.Context, _ projecta.TypeCollectionFilter) (*projecta.CostTypeCollection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return projecta.NewCostTypeCollection(1), nil
}
func (m *mockTypeService) FindOne(_ context.Context, _ projecta.TypeFilter) (*projecta.CostType, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.costType, nil
}
func (m *mockTypeService) Create(_ context.Context, _ projecta.CreateTypeCommand) (*projecta.CostType, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.costType, nil
}
func (m *mockTypeService) Remove(_ context.Context, _ projecta.RemoveProjectResourceCommand) error {
	return m.err
}
func (m *mockTypeService) Update(_ context.Context, _ projecta.UpdateTypeCommand) error {
	return m.err
}

type mockPaymentService struct {
	pay *projecta.Payment
	err error
}

func (m *mockPaymentService) FindOne(_ context.Context, _ projecta.PaymentFilter) (*projecta.Payment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pay, nil
}
func (m *mockPaymentService) Find(_ context.Context, _ projecta.PaymentCollectionFilter) (*projecta.PaymentCollection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return projecta.NewPaymentCollection(1), nil
}
func (m *mockPaymentService) Create(_ context.Context, _ projecta.CreatePaymentCommand) (*projecta.Payment, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.pay, nil
}
func (m *mockPaymentService) Update(_ context.Context, _ projecta.UpdatePaymentCommand) error {
	return m.err
}
func (m *mockPaymentService) Remove(_ context.Context, _ projecta.RemovePaymentCommand) error {
	return m.err
}

type mockAssetService struct {
	asset *asset.Asset
	err   error
}

func (m *mockAssetService) Find(_ context.Context, _ asset.CollectionFilter) (*asset.Collection, error) {
	if m.err != nil {
		return nil, m.err
	}
	return asset.NewCollection(1), nil
}
func (m *mockAssetService) FindOne(_ context.Context, _ asset.Filter) (*asset.Asset, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.asset, nil
}
func (m *mockAssetService) Create(_ context.Context, _ asset.CreateAssetCommand) (*asset.Asset, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.asset, nil
}
func (m *mockAssetService) Remove(_ context.Context, _ asset.RemoveAssetCommand) error {
	return m.err
}
func (m *mockAssetService) Update(_ context.Context, _ asset.UpdateAssetCommand) error {
	return m.err
}

func TestWebHandlersAndEndpoints(t *testing.T) {
	personID := uuid.New()
	owner := &projecta.Owner{PersonID: personID, DisplayName: "John Doe"}
	proj, _ := projecta.NewProject(uuid.New(), "Project 1", "Desc", owner, time.Now(), time.Now())
	cat, _ := projecta.NewCostCategory(uuid.New(), proj.ProjectID, "Category 1", "Desc")
	costType, _ := projecta.NewCostType(proj.ProjectID, cat, "Type 1", "Desc")
	pay := projecta.NewPayment(uuid.New(), proj, owner, costType, "Pay", money.New(100, money.USD), time.Now(), projecta.DownPayment)
	ast := asset.NewAsset(uuid.New(), "Laptop", "Desc", proj, costType, money.New(1000, money.USD), time.Now(), owner)
	cred, _ := people.NewCredentials("LOCAL", "john@example.com", "secret")
	person, _ := people.NewPerson(personID, "John", "Doe", "J.D.", []people.Credentials{cred})

	peopleSvc := &mockPeopleService{user: person}
	authSvc := &mockAuthService{authResp: &core.AuthResponse{AccessToken: "token_123", RefreshToken: "ref_123"}}
	tokenProv := &mockTokenProvider{claims: &core.AuthTokenClaims{ID: uuid.New().String(), AuthTokenPayload: core.AuthTokenPayload{Sub: personID.String()}}}
	projSvc := &mockProjectService{project: proj}
	catSvc := &mockCategoryService{cat: cat}
	typeSvc := &mockTypeService{costType: costType}
	paySvc := &mockPaymentService{pay: pay}
	astSvc := &mockAssetService{asset: ast}

	handler, err := MakeHTTPHandler(peopleSvc, tokenProv, authSvc, projSvc, catSvc, typeSvc, paySvc, astSvc, nil)
	if err != nil || handler == nil {
		t.Fatalf("failed to create http handler: %v", err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	t.Run("POST /register success and error", func(t *testing.T) {
		body, _ := json.Marshal(RegisterUserDTO{
			FirstName:        "John",
			LastName:         "Doe",
			Login:            "john@example.com",
			Token:            "secret",
			IdentityProvider: "LOCAL",
		})
		resp, err := http.Post(server.URL+"/register", "application/json", bytes.NewReader(body))
		if err != nil || resp.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 Created for register, got %v", resp.StatusCode)
		}
	})

	t.Run("POST /login success", func(t *testing.T) {
		body, _ := json.Marshal(LoginDTO{
			ID:               "john@example.com",
			Token:            "secret",
			IdentityProvider: "LOCAL",
		})
		resp, err := http.Post(server.URL+"/login", "application/json", bytes.NewReader(body))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for login, got %v", resp.StatusCode)
		}
	})

	t.Run("GET /profile unauthorized and authorized", func(t *testing.T) {
		// Unauthorized
		resp, _ := http.Get(server.URL + "/profile")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized for profile without token, got %v", resp.StatusCode)
		}

		// Authorized
		req, _ := http.NewRequest("GET", server.URL+"/profile", nil)
		req.Header.Set("Authorization", "Bearer valid_token")
		client := &http.Client{}
		respAuth, err := client.Do(req)
		if err != nil || respAuth.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for profile with token, got %v", respAuth.StatusCode)
		}
	})

	t.Run("POST /refresh success", func(t *testing.T) {
		body, _ := json.Marshal(RefreshTokenDTO{
			AccessToken:  "acc",
			RefreshToken: "ref",
		})
		resp, err := http.Post(server.URL+"/refresh", "application/json", bytes.NewReader(body))
		if err != nil || resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 OK for refresh, got %v", resp.StatusCode)
		}
	})

	t.Run("Projects endpoints /projects", func(t *testing.T) {
		client := &http.Client{}

		// GET /projects
		reqList, _ := http.NewRequest("GET", server.URL+"/projects?limit=10&offset=0&name=alpha", nil)
		reqList.Header.Set("Authorization", "Bearer token")
		respList, err := client.Do(reqList)
		if err != nil || respList.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GET /projects, got %v", respList.StatusCode)
		}

		// POST /projects
		createProjBody, _ := json.Marshal(CreateProjectDTO{Name: "New Project", Description: "Desc"})
		reqCreate, _ := http.NewRequest("POST", server.URL+"/projects", bytes.NewReader(createProjBody))
		reqCreate.Header.Set("Authorization", "Bearer token")
		respCreate, err := client.Do(reqCreate)
		if err != nil || respCreate.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 for POST /projects, got %v", respCreate.StatusCode)
		}

		// PATCH /projects/{id}
		patchProjBody, _ := json.Marshal(UpdateProjectDTO{MainCurrency: "USD"})
		reqPatch, _ := http.NewRequest("PATCH", server.URL+"/projects/"+proj.ProjectID.String(), bytes.NewReader(patchProjBody))
		reqPatch.Header.Set("Authorization", "Bearer token")
		respPatch, err := client.Do(reqPatch)
		if err != nil || respPatch.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for PATCH /projects/{id}, got %v", respPatch.StatusCode)
		}
	})

	t.Run("Categories endpoints", func(t *testing.T) {
		client := &http.Client{}
		pID := proj.ProjectID.String()

		// GET /projects/{id}/categories
		reqList, _ := http.NewRequest("GET", server.URL+"/projects/"+pID+"/categories?limit=5&offset=0&name=cat", nil)
		reqList.Header.Set("Authorization", "Bearer token")
		respList, err := client.Do(reqList)
		if err != nil || respList.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GET categories, got %v", respList.StatusCode)
		}

		// POST /projects/{id}/categories
		body, _ := json.Marshal(CreateCategoryDTO{Name: "New Cat", Description: "Desc"})
		reqCreate, _ := http.NewRequest("POST", server.URL+"/projects/"+pID+"/categories", bytes.NewReader(body))
		reqCreate.Header.Set("Authorization", "Bearer token")
		respCreate, err := client.Do(reqCreate)
		if err != nil || respCreate.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 for POST category, got %v", respCreate.StatusCode)
		}
	})

	t.Run("Types endpoints", func(t *testing.T) {
		client := &http.Client{}
		pID := proj.ProjectID.String()
		tID := costType.ID.String()

		// GET /projects/{id}/types
		reqList, _ := http.NewRequest("GET", server.URL+"/projects/"+pID+"/types?limit=5&offset=0", nil)
		reqList.Header.Set("Authorization", "Bearer token")
		respList, _ := client.Do(reqList)
		if respList.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GET types")
		}

		// POST /projects/{id}/types
		body, _ := json.Marshal(CreateTypeDTO{CategoryID: cat.ID.String(), Name: "Type", Description: "Desc"})
		reqCreate, _ := http.NewRequest("POST", server.URL+"/projects/"+pID+"/types", bytes.NewReader(body))
		reqCreate.Header.Set("Authorization", "Bearer token")
		respCreate, _ := client.Do(reqCreate)
		if respCreate.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 for POST type")
		}

		// DELETE /projects/{id}/types/{type_id}
		reqDel, _ := http.NewRequest("DELETE", server.URL+"/projects/"+pID+"/types/"+tID, nil)
		reqDel.Header.Set("Authorization", "Bearer token")
		respDel, _ := client.Do(reqDel)
		if respDel.StatusCode != http.StatusNoContent {
			t.Errorf("expected 244 No Content for DELETE type, got %v", respDel.StatusCode)
		}
	})

	t.Run("Payments endpoints", func(t *testing.T) {
		client := &http.Client{}
		pID := proj.ProjectID.String()
		payID := pay.ID.String()
		nowStr := time.Now().Format(time.RFC3339)

		// GET /projects/{id}/payments
		reqList, _ := http.NewRequest("GET", server.URL+"/projects/"+pID+"/payments?category_id="+cat.ID.String()+"&type_id="+costType.ID.String()+"&order_by=date&order=DESC", nil)
		reqList.Header.Set("Authorization", "Bearer token")
		respList, _ := client.Do(reqList)
		if respList.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GET payments")
		}

		// GET /projects/{id}/totals
		reqTotals, _ := http.NewRequest("GET", server.URL+"/projects/"+pID+"/totals", nil)
		reqTotals.Header.Set("Authorization", "Bearer token")
		respTotals, _ := client.Do(reqTotals)
		if respTotals.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GET totals")
		}

		// POST /projects/{id}/payments
		payBody, _ := json.Marshal(CreatePaymentDTO{TypeID: costType.ID.String(), Description: "Pay", Amount: 50, Currency: "USD", PaymentDate: nowStr, Kind: "DOWN_PAYMENT"})
		reqCreate, _ := http.NewRequest("POST", server.URL+"/projects/"+pID+"/payments", bytes.NewReader(payBody))
		reqCreate.Header.Set("Authorization", "Bearer token")
		respCreate, _ := client.Do(reqCreate)
		if respCreate.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 for POST payment, got %v", respCreate.StatusCode)
		}

		// PUT /projects/{id}/payments/{payment_id}
		updPayBody, _ := json.Marshal(UpdatePaymentDTO{TypeID: costType.ID.String(), Description: "Upd Pay", Amount: 60, Currency: "USD", PaymentDate: nowStr, Kind: "DOWN_PAYMENT"})
		reqPut, _ := http.NewRequest("PUT", server.URL+"/projects/"+pID+"/payments/"+payID, bytes.NewReader(updPayBody))
		reqPut.Header.Set("Authorization", "Bearer token")
		respPut, _ := client.Do(reqPut)
		if respPut.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204 for PUT payment, got %v", respPut.StatusCode)
		}

		// GET /projects/{id}/payments/{payment_id}
		reqGetOne, _ := http.NewRequest("GET", server.URL+"/projects/"+pID+"/payments/"+payID, nil)
		reqGetOne.Header.Set("Authorization", "Bearer token")
		respGetOne, _ := client.Do(reqGetOne)
		if respGetOne.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GET payment one")
		}

		// DELETE /projects/{id}/payments/{payment_id}
		reqDel, _ := http.NewRequest("DELETE", server.URL+"/projects/"+pID+"/payments/"+payID, nil)
		reqDel.Header.Set("Authorization", "Bearer token")
		respDel, _ := client.Do(reqDel)
		if respDel.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204 for DELETE payment")
		}
	})

	t.Run("Assets endpoints", func(t *testing.T) {
		client := &http.Client{}
		pID := proj.ProjectID.String()
		astID := ast.ID().String()
		nowStr := time.Now().Format(time.RFC3339)

		// GET /projects/{id}/assets
		reqList, _ := http.NewRequest("GET", server.URL+"/projects/"+pID+"/assets?limit=5&offset=0", nil)
		reqList.Header.Set("Authorization", "Bearer token")
		respList, _ := client.Do(reqList)
		if respList.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GET assets")
		}

		// POST /projects/{id}/assets
		assetBody, _ := json.Marshal(CreateAssetDTO{TypeID: costType.ID.String(), Name: "Laptop", Description: "Desc", Price: 1000, Currency: "USD", AcquiredAt: nowStr, WithPayment: true})
		reqCreate, _ := http.NewRequest("POST", server.URL+"/projects/"+pID+"/assets", bytes.NewReader(assetBody))
		reqCreate.Header.Set("Authorization", "Bearer token")
		respCreate, _ := client.Do(reqCreate)
		if respCreate.StatusCode != http.StatusCreated {
			t.Errorf("expected 201 for POST asset, got %v", respCreate.StatusCode)
		}

		// GET /projects/{id}/assets/{asset_id}
		reqGetOne, _ := http.NewRequest("GET", server.URL+"/projects/"+pID+"/assets/"+astID, nil)
		reqGetOne.Header.Set("Authorization", "Bearer token")
		respGetOne, _ := client.Do(reqGetOne)
		if respGetOne.StatusCode != http.StatusOK {
			t.Errorf("expected 200 for GET asset one")
		}

		// PUT /projects/{id}/assets/{asset_id}
		updAssetBody, _ := json.Marshal(UpdateAssetDTO{TypeID: costType.ID.String(), Name: "Desktop", Description: "Desc", Price: 1200, Currency: "USD", AcquiredAt: nowStr})
		reqPut, _ := http.NewRequest("PUT", server.URL+"/projects/"+pID+"/assets/"+astID, bytes.NewReader(updAssetBody))
		reqPut.Header.Set("Authorization", "Bearer token")
		respPut, _ := client.Do(reqPut)
		if respPut.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204 for PUT asset, got %v", respPut.StatusCode)
		}

		// DELETE /projects/{id}/assets/{asset_id}
		reqDel, _ := http.NewRequest("DELETE", server.URL+"/projects/"+pID+"/assets/"+astID, nil)
		reqDel.Header.Set("Authorization", "Bearer token")
		respDel, _ := client.Do(reqDel)
		if respDel.StatusCode != http.StatusNoContent {
			t.Errorf("expected 204 for DELETE asset")
		}
	})

	t.Run("Swagger UI handler", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/swagger/")
		if err != nil {
			t.Fatalf("unexpected error fetching swagger: %v", err)
		}
		_ = resp
	})
}

func TestErrorCodeToHttpStatus(t *testing.T) {
	tests := []struct {
		code exceptions.ErrorCode
		want int
	}{
		{exceptions.NotFound, http.StatusNotFound},
		{exceptions.ValidationFailed, http.StatusBadRequest},
		{exceptions.Internal, http.StatusInternalServerError},
		{exceptions.Unauthorized, http.StatusUnauthorized},
		{"UNKNOWN_CODE", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		ex := exceptions.NewApplicationError("msg", tt.code, nil)
		got := errorCodeToHttpStatus(ex)
		if got != tt.want {
			t.Errorf("errorCodeToHttpStatus(%s) = %d; want %d", tt.code, got, tt.want)
		}
	}
}

func TestEncodeJSON(t *testing.T) {
	encoder := encodeJSON(0)
	w := httptest.NewRecorder()
	err := encoder(context.Background(), w, map[string]string{"foo": "bar"})
	if err != nil || w.Code != http.StatusOK {
		t.Errorf("expected StatusOK for 0 status code")
	}

	encoderNoContent := encodeJSON(http.StatusNoContent)
	wNC := httptest.NewRecorder()
	err = encoderNoContent(context.Background(), wNC, nil)
	if err != nil || wNC.Code != http.StatusNoContent {
		t.Errorf("expected StatusNoContent")
	}
}
