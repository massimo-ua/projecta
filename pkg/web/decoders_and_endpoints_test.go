package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

func TestAuthMiddlewareAndErrorEncoder(t *testing.T) {
	t.Run("jwtMiddleware branches", func(t *testing.T) {
		validID := uuid.New()
		provider := &mockTokenProvider{
			claims: &core.AuthTokenClaims{
				AuthTokenPayload: core.AuthTokenPayload{Sub: validID.String()},
			},
		}

		mw := jwtMiddleware(provider)

		// Missing header
		req1 := httptest.NewRequest("GET", "/", nil)
		ctx1 := mw(context.Background(), req1)
		if _, ok := ctx1.Value(core.RequesterIDContextKey).(uuid.UUID); ok {
			t.Errorf("expected no requesterID for missing header")
		}

		// Malformed header (no space)
		req2 := httptest.NewRequest("GET", "/", nil)
		req2.Header.Set("Authorization", "BearerNoSpace")
		ctx2 := mw(context.Background(), req2)
		if _, ok := ctx2.Value(core.RequesterIDContextKey).(uuid.UUID); ok {
			t.Errorf("expected no requesterID for malformed header")
		}

		// Token provider error
		errProvider := &mockTokenProvider{err: errors.New("invalid token")}
		mwErr := jwtMiddleware(errProvider)
		req3 := httptest.NewRequest("GET", "/", nil)
		req3.Header.Set("Authorization", "Bearer invalid")
		ctx3 := mwErr(context.Background(), req3)
		if _, ok := ctx3.Value(core.RequesterIDContextKey).(uuid.UUID); ok {
			t.Errorf("expected no requesterID when token provider fails")
		}

		// Invalid Sub UUID in claims
		invalidSubProvider := &mockTokenProvider{
			claims: &core.AuthTokenClaims{
				AuthTokenPayload: core.AuthTokenPayload{Sub: "not-a-uuid"},
			},
		}
		mwSub := jwtMiddleware(invalidSubProvider)
		req4 := httptest.NewRequest("GET", "/", nil)
		req4.Header.Set("Authorization", "Bearer valid")
		ctx4 := mwSub(context.Background(), req4)
		if _, ok := ctx4.Value(core.RequesterIDContextKey).(uuid.UUID); ok {
			t.Errorf("expected no requesterID for invalid sub UUID")
		}

		// Success
		req5 := httptest.NewRequest("GET", "/", nil)
		req5.Header.Set("Authorization", "Bearer valid")
		ctx5 := mw(context.Background(), req5)
		gotID, ok := ctx5.Value(core.RequesterIDContextKey).(uuid.UUID)
		if !ok || gotID != validID {
			t.Errorf("expected requesterID %v, got %v", validID, gotID)
		}
	})

	t.Run("encodeErrorResponse with generic error", func(t *testing.T) {
		w := httptest.NewRecorder()
		encodeErrorResponse(context.Background(), errors.New("raw error"), w)
		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500 for generic error, got %d", w.Code)
		}
	})
}

func TestDecodersValidationErrors(t *testing.T) {
	t.Run("decodeListProjectsRequest validation errors", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/projects?limit=invalid", nil)
		_, err := decodeListProjectsRequest(context.Background(), req)
		if err == nil {
			t.Errorf("expected limit validation error")
		}

		reqOff, _ := http.NewRequest("GET", "/projects?offset=invalid", nil)
		_, err = decodeListProjectsRequest(context.Background(), reqOff)
		if err == nil {
			t.Errorf("expected offset validation error")
		}
	})

	t.Run("decodeListTypesRequest validation errors", func(t *testing.T) {
		// Missing project_id
		reqNoVars, _ := http.NewRequest("GET", "/types", nil)
		_, err := decodeListTypesRequest(context.Background(), reqNoVars)
		if err == nil {
			t.Errorf("expected missing project_id error")
		}

		// Invalid project_id
		reqBadUUID, _ := http.NewRequest("GET", "/types", nil)
		reqBadUUID = mux.SetURLVars(reqBadUUID, map[string]string{"project_id": "bad-uuid"})
		_, err = decodeListTypesRequest(context.Background(), reqBadUUID)
		if err == nil {
			t.Errorf("expected invalid project_id error")
		}

		// Invalid limit & offset
		validUUID := uuid.New().String()
		reqBadLimit, _ := http.NewRequest("GET", "/types?limit=bad", nil)
		reqBadLimit = mux.SetURLVars(reqBadLimit, map[string]string{"project_id": validUUID})
		_, err = decodeListTypesRequest(context.Background(), reqBadLimit)
		if err == nil {
			t.Errorf("expected limit error")
		}

		reqBadOffset, _ := http.NewRequest("GET", "/types?offset=bad", nil)
		reqBadOffset = mux.SetURLVars(reqBadOffset, map[string]string{"project_id": validUUID})
		_, err = decodeListTypesRequest(context.Background(), reqBadOffset)
		if err == nil {
			t.Errorf("expected offset error")
		}
	})

	t.Run("decodeListCategoriesRequest validation errors", func(t *testing.T) {
		reqNoVars, _ := http.NewRequest("GET", "/categories", nil)
		_, err := decodeListCategoriesRequest(context.Background(), reqNoVars)
		if err == nil {
			t.Errorf("expected missing project_id error")
		}

		reqBadUUID, _ := http.NewRequest("GET", "/categories", nil)
		reqBadUUID = mux.SetURLVars(reqBadUUID, map[string]string{"project_id": "bad-uuid"})
		_, err = decodeListCategoriesRequest(context.Background(), reqBadUUID)
		if err == nil {
			t.Errorf("expected invalid project_id error")
		}

		validUUID := uuid.New().String()
		reqBadLimit, _ := http.NewRequest("GET", "/categories?limit=bad", nil)
		reqBadLimit = mux.SetURLVars(reqBadLimit, map[string]string{"project_id": validUUID})
		_, err = decodeListCategoriesRequest(context.Background(), reqBadLimit)
		if err == nil {
			t.Errorf("expected limit error")
		}

		reqBadOffset, _ := http.NewRequest("GET", "/categories?offset=bad", nil)
		reqBadOffset = mux.SetURLVars(reqBadOffset, map[string]string{"project_id": validUUID})
		_, err = decodeListCategoriesRequest(context.Background(), reqBadOffset)
		if err == nil {
			t.Errorf("expected offset error")
		}
	})

	t.Run("decodeListPaymentsRequest validation errors", func(t *testing.T) {
		validUUID := uuid.New().String()

		reqNoVars, _ := http.NewRequest("GET", "/payments", nil)
		_, err := decodeListPaymentsRequest(context.Background(), reqNoVars)
		if err == nil {
			t.Errorf("expected missing project_id error")
		}

		reqBadUUID, _ := http.NewRequest("GET", "/payments", nil)
		reqBadUUID = mux.SetURLVars(reqBadUUID, map[string]string{"project_id": "bad"})
		_, err = decodeListPaymentsRequest(context.Background(), reqBadUUID)
		if err == nil {
			t.Errorf("expected invalid project_id error")
		}

		reqBadCat, _ := http.NewRequest("GET", "/payments?category_id=bad", nil)
		reqBadCat = mux.SetURLVars(reqBadCat, map[string]string{"project_id": validUUID})
		_, err = decodeListPaymentsRequest(context.Background(), reqBadCat)
		if err == nil {
			t.Errorf("expected invalid category_id error")
		}

		reqBadType, _ := http.NewRequest("GET", "/payments?type_id=bad", nil)
		reqBadType = mux.SetURLVars(reqBadType, map[string]string{"project_id": validUUID})
		_, err = decodeListPaymentsRequest(context.Background(), reqBadType)
		if err == nil {
			t.Errorf("expected invalid type_id error")
		}

		reqBadLimit, _ := http.NewRequest("GET", "/payments?limit=bad", nil)
		reqBadLimit = mux.SetURLVars(reqBadLimit, map[string]string{"project_id": validUUID})
		_, err = decodeListPaymentsRequest(context.Background(), reqBadLimit)
		if err == nil {
			t.Errorf("expected invalid limit error")
		}

		reqBadOffset, _ := http.NewRequest("GET", "/payments?offset=bad", nil)
		reqBadOffset = mux.SetURLVars(reqBadOffset, map[string]string{"project_id": validUUID})
		_, err = decodeListPaymentsRequest(context.Background(), reqBadOffset)
		if err == nil {
			t.Errorf("expected invalid offset error")
		}
	})

	t.Run("decodeProjectTotalsRequest validation errors", func(t *testing.T) {
		reqNoVars, _ := http.NewRequest("GET", "/totals", nil)
		_, err := decodeProjectTotalsRequest(context.Background(), reqNoVars)
		if err == nil {
			t.Errorf("expected missing project_id error")
		}

		reqBadUUID, _ := http.NewRequest("GET", "/totals", nil)
		reqBadUUID = mux.SetURLVars(reqBadUUID, map[string]string{"project_id": "bad"})
		_, err = decodeProjectTotalsRequest(context.Background(), reqBadUUID)
		if err == nil {
			t.Errorf("expected invalid project_id error")
		}
	})

	t.Run("decodeProjectResourceRemoveCommand validation errors", func(t *testing.T) {
		fn := decodeProjectResourceRemoveCommand("project_id", "resource_id")
		validUUID := uuid.New().String()

		reqNoProj, _ := http.NewRequest("DELETE", "/", nil)
		_, err := fn(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("DELETE", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = fn(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqNoRes, _ := http.NewRequest("DELETE", "/", nil)
		reqNoRes = mux.SetURLVars(reqNoRes, map[string]string{"project_id": validUUID})
		_, err = fn(context.Background(), reqNoRes)
		if err == nil {
			t.Errorf("expected missing resource_id")
		}

		reqBadRes, _ := http.NewRequest("DELETE", "/", nil)
		reqBadRes = mux.SetURLVars(reqBadRes, map[string]string{"project_id": validUUID, "resource_id": "bad"})
		_, err = fn(context.Background(), reqBadRes)
		if err == nil {
			t.Errorf("expected invalid resource_id")
		}
	})
}

func TestAssetDecodersValidationErrors(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("decodeCreateAssetRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("POST", "/", nil)
		_, err := decodeCreateAssetRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("POST", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = decodeCreateAssetRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		// Invalid JSON
		reqBadJSON, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte("not json")))
		reqBadJSON = mux.SetURLVars(reqBadJSON, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetRequest(context.Background(), reqBadJSON)
		if err == nil {
			t.Errorf("expected invalid json error")
		}

		// Empty name
		bodyEmptyName, _ := json.Marshal(CreateAssetDTO{Name: ""})
		reqEmptyName, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyEmptyName))
		reqEmptyName = mux.SetURLVars(reqEmptyName, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetRequest(context.Background(), reqEmptyName)
		if err == nil {
			t.Errorf("expected empty name error")
		}

		// Bad type_id
		bodyBadType, _ := json.Marshal(CreateAssetDTO{Name: "Laptop", TypeID: "bad"})
		reqBadType, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadType))
		reqBadType = mux.SetURLVars(reqBadType, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetRequest(context.Background(), reqBadType)
		if err == nil {
			t.Errorf("expected bad type_id error")
		}

		// Price <= 0
		bodyZeroPrice, _ := json.Marshal(CreateAssetDTO{Name: "Laptop", TypeID: validUUID, Price: 0})
		reqZeroPrice, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyZeroPrice))
		reqZeroPrice = mux.SetURLVars(reqZeroPrice, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetRequest(context.Background(), reqZeroPrice)
		if err == nil {
			t.Errorf("expected price <= 0 error")
		}

		// Empty currency
		bodyEmptyCurr, _ := json.Marshal(CreateAssetDTO{Name: "Laptop", TypeID: validUUID, Price: 100, Currency: ""})
		reqEmptyCurr, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyEmptyCurr))
		reqEmptyCurr = mux.SetURLVars(reqEmptyCurr, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetRequest(context.Background(), reqEmptyCurr)
		if err == nil {
			t.Errorf("expected empty currency error")
		}

		// Bad acquired_at date
		bodyBadDate, _ := json.Marshal(CreateAssetDTO{Name: "Laptop", TypeID: validUUID, Price: 100, Currency: "USD", AcquiredAt: "bad-date"})
		reqBadDate, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadDate))
		reqBadDate = mux.SetURLVars(reqBadDate, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetRequest(context.Background(), reqBadDate)
		if err == nil {
			t.Errorf("expected bad acquired_at date error")
		}
	})

	t.Run("decodeUpdateAssetRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("PUT", "/", nil)
		_, err := decodeUpdateAssetRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("PUT", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = decodeUpdateAssetRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqNoAsset, _ := http.NewRequest("PUT", "/", nil)
		reqNoAsset = mux.SetURLVars(reqNoAsset, map[string]string{"project_id": validUUID})
		_, err = decodeUpdateAssetRequest(context.Background(), reqNoAsset)
		if err == nil {
			t.Errorf("expected missing asset_id")
		}

		reqBadAsset, _ := http.NewRequest("PUT", "/", nil)
		reqBadAsset = mux.SetURLVars(reqBadAsset, map[string]string{"project_id": validUUID, "asset_id": "bad"})
		_, err = decodeUpdateAssetRequest(context.Background(), reqBadAsset)
		if err == nil {
			t.Errorf("expected invalid asset_id")
		}

		reqBadJSON, _ := http.NewRequest("PUT", "/", bytes.NewReader([]byte("not json")))
		reqBadJSON = mux.SetURLVars(reqBadJSON, map[string]string{"project_id": validUUID, "asset_id": validUUID})
		_, err = decodeUpdateAssetRequest(context.Background(), reqBadJSON)
		if err == nil {
			t.Errorf("expected invalid json error")
		}

		bodyEmptyName, _ := json.Marshal(UpdateAssetDTO{Name: ""})
		reqEmptyName, _ := http.NewRequest("PUT", "/", bytes.NewReader(bodyEmptyName))
		reqEmptyName = mux.SetURLVars(reqEmptyName, map[string]string{"project_id": validUUID, "asset_id": validUUID})
		_, err = decodeUpdateAssetRequest(context.Background(), reqEmptyName)
		if err == nil {
			t.Errorf("expected empty name error")
		}

		bodyBadType, _ := json.Marshal(UpdateAssetDTO{Name: "Laptop", TypeID: "bad"})
		reqBadType, _ := http.NewRequest("PUT", "/", bytes.NewReader(bodyBadType))
		reqBadType = mux.SetURLVars(reqBadType, map[string]string{"project_id": validUUID, "asset_id": validUUID})
		_, err = decodeUpdateAssetRequest(context.Background(), reqBadType)
		if err == nil {
			t.Errorf("expected bad type_id error")
		}

		bodyZeroPrice, _ := json.Marshal(UpdateAssetDTO{Name: "Laptop", TypeID: validUUID, Price: 0})
		reqZeroPrice, _ := http.NewRequest("PUT", "/", bytes.NewReader(bodyZeroPrice))
		reqZeroPrice = mux.SetURLVars(reqZeroPrice, map[string]string{"project_id": validUUID, "asset_id": validUUID})
		_, err = decodeUpdateAssetRequest(context.Background(), reqZeroPrice)
		if err == nil {
			t.Errorf("expected price <= 0 error")
		}

		bodyEmptyCurr, _ := json.Marshal(UpdateAssetDTO{Name: "Laptop", TypeID: validUUID, Price: 100, Currency: ""})
		reqEmptyCurr, _ := http.NewRequest("PUT", "/", bytes.NewReader(bodyEmptyCurr))
		reqEmptyCurr = mux.SetURLVars(reqEmptyCurr, map[string]string{"project_id": validUUID, "asset_id": validUUID})
		_, err = decodeUpdateAssetRequest(context.Background(), reqEmptyCurr)
		if err == nil {
			t.Errorf("expected empty currency error")
		}

		bodyBadDate, _ := json.Marshal(UpdateAssetDTO{Name: "Laptop", TypeID: validUUID, Price: 100, Currency: "USD", AcquiredAt: "bad-date"})
		reqBadDate, _ := http.NewRequest("PUT", "/", bytes.NewReader(bodyBadDate))
		reqBadDate = mux.SetURLVars(reqBadDate, map[string]string{"project_id": validUUID, "asset_id": validUUID})
		_, err = decodeUpdateAssetRequest(context.Background(), reqBadDate)
		if err == nil {
			t.Errorf("expected bad acquired_at date error")
		}
	})

	t.Run("decodeGetAssetRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("GET", "/", nil)
		_, err := decodeGetAssetRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("GET", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = decodeGetAssetRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqNoAsset, _ := http.NewRequest("GET", "/", nil)
		reqNoAsset = mux.SetURLVars(reqNoAsset, map[string]string{"project_id": validUUID})
		_, err = decodeGetAssetRequest(context.Background(), reqNoAsset)
		if err == nil {
			t.Errorf("expected missing asset_id")
		}

		reqBadAsset, _ := http.NewRequest("GET", "/", nil)
		reqBadAsset = mux.SetURLVars(reqBadAsset, map[string]string{"project_id": validUUID, "asset_id": "bad"})
		_, err = decodeGetAssetRequest(context.Background(), reqBadAsset)
		if err == nil {
			t.Errorf("expected invalid asset_id")
		}
	})

	t.Run("decodeListAssetsRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("GET", "/", nil)
		_, err := decodeListAssetsRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("GET", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = decodeListAssetsRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqBadType, _ := http.NewRequest("GET", "/assets?type_id=bad", nil)
		reqBadType = mux.SetURLVars(reqBadType, map[string]string{"project_id": validUUID})
		_, err = decodeListAssetsRequest(context.Background(), reqBadType)
		if err == nil {
			t.Errorf("expected invalid type_id")
		}

		reqBadLimit, _ := http.NewRequest("GET", "/assets?limit=bad", nil)
		reqBadLimit = mux.SetURLVars(reqBadLimit, map[string]string{"project_id": validUUID})
		_, err = decodeListAssetsRequest(context.Background(), reqBadLimit)
		if err == nil {
			t.Errorf("expected invalid limit")
		}

		reqBadOffset, _ := http.NewRequest("GET", "/assets?offset=bad", nil)
		reqBadOffset = mux.SetURLVars(reqBadOffset, map[string]string{"project_id": validUUID})
		_, err = decodeListAssetsRequest(context.Background(), reqBadOffset)
		if err == nil {
			t.Errorf("expected invalid offset")
		}
	})
}

func TestPaymentsDecodersValidationErrors(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("decodeUpdatePaymentRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("PUT", "/", nil)
		_, err := decodeUpdatePaymentRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("PUT", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = decodeUpdatePaymentRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqNoPay, _ := http.NewRequest("PUT", "/", nil)
		reqNoPay = mux.SetURLVars(reqNoPay, map[string]string{"project_id": validUUID})
		_, err = decodeUpdatePaymentRequest(context.Background(), reqNoPay)
		if err == nil {
			t.Errorf("expected missing payment_id")
		}

		reqBadPay, _ := http.NewRequest("PUT", "/", nil)
		reqBadPay = mux.SetURLVars(reqBadPay, map[string]string{"project_id": validUUID, "payment_id": "bad"})
		_, err = decodeUpdatePaymentRequest(context.Background(), reqBadPay)
		if err == nil {
			t.Errorf("expected invalid payment_id")
		}

		reqBadJSON, _ := http.NewRequest("PUT", "/", bytes.NewReader([]byte("not json")))
		reqBadJSON = mux.SetURLVars(reqBadJSON, map[string]string{"project_id": validUUID, "payment_id": validUUID})
		_, err = decodeUpdatePaymentRequest(context.Background(), reqBadJSON)
		if err == nil {
			t.Errorf("expected invalid json")
		}

		bodyBadDate, _ := json.Marshal(UpdatePaymentDTO{PaymentDate: "bad-date"})
		reqBadDate, _ := http.NewRequest("PUT", "/", bytes.NewReader(bodyBadDate))
		reqBadDate = mux.SetURLVars(reqBadDate, map[string]string{"project_id": validUUID, "payment_id": validUUID})
		_, err = decodeUpdatePaymentRequest(context.Background(), reqBadDate)
		if err == nil {
			t.Errorf("expected invalid date")
		}

		nowStr := time.Now().Format(time.RFC3339)
		bodyBadType, _ := json.Marshal(UpdatePaymentDTO{PaymentDate: nowStr, TypeID: "bad"})
		reqBadType, _ := http.NewRequest("PUT", "/", bytes.NewReader(bodyBadType))
		reqBadType = mux.SetURLVars(reqBadType, map[string]string{"project_id": validUUID, "payment_id": validUUID})
		_, err = decodeUpdatePaymentRequest(context.Background(), reqBadType)
		if err == nil {
			t.Errorf("expected invalid type_id")
		}

		bodyBadKind, _ := json.Marshal(UpdatePaymentDTO{PaymentDate: nowStr, TypeID: validUUID, Kind: "INVALID_KIND"})
		reqBadKind, _ := http.NewRequest("PUT", "/", bytes.NewReader(bodyBadKind))
		reqBadKind = mux.SetURLVars(reqBadKind, map[string]string{"project_id": validUUID, "payment_id": validUUID})
		_, err = decodeUpdatePaymentRequest(context.Background(), reqBadKind)
		if err == nil {
			t.Errorf("expected invalid payment kind")
		}
	})

	t.Run("decodeGetPaymentRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("GET", "/", nil)
		_, err := decodeGetPaymentRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("GET", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = decodeGetPaymentRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqNoPay, _ := http.NewRequest("GET", "/", nil)
		reqNoPay = mux.SetURLVars(reqNoPay, map[string]string{"project_id": validUUID})
		_, err = decodeGetPaymentRequest(context.Background(), reqNoPay)
		if err == nil {
			t.Errorf("expected missing payment_id")
		}

		reqBadPay, _ := http.NewRequest("GET", "/", nil)
		reqBadPay = mux.SetURLVars(reqBadPay, map[string]string{"project_id": validUUID, "payment_id": "bad"})
		_, err = decodeGetPaymentRequest(context.Background(), reqBadPay)
		if err == nil {
			t.Errorf("expected invalid payment_id")
		}
	})
}

func TestProjectaDecodersValidationErrors(t *testing.T) {
	validUUID := uuid.New().String()
	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, uuid.New())

	t.Run("DecodeCreateProjectRequest errors", func(t *testing.T) {
		reqUnauth, _ := http.NewRequest("POST", "/", nil)
		_, err := DecodeCreateProjectRequest(context.Background(), reqUnauth)
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		reqBadJSON, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte("not json")))
		_, err = DecodeCreateProjectRequest(authedCtx, reqBadJSON)
		if err == nil {
			t.Errorf("expected invalid json")
		}
	})

	t.Run("DecodeCreateCategoryRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("POST", "/", nil)
		_, err := DecodeCreateCategoryRequest(authedCtx, reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("POST", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = DecodeCreateCategoryRequest(authedCtx, reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqUnauth, _ := http.NewRequest("POST", "/", nil)
		reqUnauth = mux.SetURLVars(reqUnauth, map[string]string{"project_id": validUUID})
		_, err = DecodeCreateCategoryRequest(context.Background(), reqUnauth)
		if err == nil {
			t.Errorf("expected unauthorized")
		}

		reqBadJSON, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte("not json")))
		reqBadJSON = mux.SetURLVars(reqBadJSON, map[string]string{"project_id": validUUID})
		_, err = DecodeCreateCategoryRequest(authedCtx, reqBadJSON)
		if err == nil {
			t.Errorf("expected invalid json")
		}
	})

	t.Run("DecodeCreateTypeRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("POST", "/", nil)
		_, err := DecodeCreateTypeRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("POST", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = DecodeCreateTypeRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqBadJSON, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte("not json")))
		reqBadJSON = mux.SetURLVars(reqBadJSON, map[string]string{"project_id": validUUID})
		_, err = DecodeCreateTypeRequest(context.Background(), reqBadJSON)
		if err == nil {
			t.Errorf("expected invalid json")
		}

		bodyEmptyName, _ := json.Marshal(CreateTypeDTO{Name: ""})
		reqEmptyName, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyEmptyName))
		reqEmptyName = mux.SetURLVars(reqEmptyName, map[string]string{"project_id": validUUID})
		_, err = DecodeCreateTypeRequest(context.Background(), reqEmptyName)
		if err == nil {
			t.Errorf("expected empty name error")
		}

		bodyBadCat, _ := json.Marshal(CreateTypeDTO{Name: "Type", CategoryID: "bad"})
		reqBadCat, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadCat))
		reqBadCat = mux.SetURLVars(reqBadCat, map[string]string{"project_id": validUUID})
		_, err = DecodeCreateTypeRequest(context.Background(), reqBadCat)
		if err == nil {
			t.Errorf("expected invalid category_id")
		}
	})

	t.Run("DecodeCreatePaymentRequest errors", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("POST", "/", nil)
		_, err := DecodeCreatePaymentRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("POST", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = DecodeCreatePaymentRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqBadJSON, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte("not json")))
		reqBadJSON = mux.SetURLVars(reqBadJSON, map[string]string{"project_id": validUUID})
		_, err = DecodeCreatePaymentRequest(context.Background(), reqBadJSON)
		if err == nil {
			t.Errorf("expected invalid json")
		}

		bodyBadDate, _ := json.Marshal(CreatePaymentDTO{PaymentDate: "bad-date"})
		reqBadDate, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadDate))
		reqBadDate = mux.SetURLVars(reqBadDate, map[string]string{"project_id": validUUID})
		_, err = DecodeCreatePaymentRequest(context.Background(), reqBadDate)
		if err == nil {
			t.Errorf("expected invalid date")
		}

		nowStr := time.Now().Format(time.RFC3339)
		bodyBadType, _ := json.Marshal(CreatePaymentDTO{PaymentDate: nowStr, TypeID: "bad"})
		reqBadType, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadType))
		reqBadType = mux.SetURLVars(reqBadType, map[string]string{"project_id": validUUID})
		_, err = DecodeCreatePaymentRequest(context.Background(), reqBadType)
		if err == nil {
			t.Errorf("expected invalid type_id")
		}

		bodyBadKind, _ := json.Marshal(CreatePaymentDTO{PaymentDate: nowStr, TypeID: validUUID, Kind: "BAD_KIND"})
		reqBadKind, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadKind))
		reqBadKind = mux.SetURLVars(reqBadKind, map[string]string{"project_id": validUUID})
		_, err = DecodeCreatePaymentRequest(context.Background(), reqBadKind)
		if err == nil {
			t.Errorf("expected invalid payment kind")
		}
	})
}

func TestEndpointsErrorBranches(t *testing.T) {
	errSvc := errors.New("service failure")
	owner := &projecta.Owner{PersonID: uuid.New()}
	proj, _ := projecta.NewProject(uuid.New(), "Project", "Desc", owner, time.Now(), time.Now())
	cat, _ := projecta.NewCostCategory(uuid.New(), proj.ProjectID, "Cat", "Desc")
	costType, _ := projecta.NewCostType(proj.ProjectID, cat, "Type", "Desc")
	pay := projecta.NewPayment(uuid.New(), proj, owner, costType, "Pay", money.New(100, money.USD), time.Now(), projecta.DownPayment)
	ast := asset.NewAsset(uuid.New(), "Asset", "Desc", proj, costType, money.New(1000, money.USD), time.Now(), owner)

	peopleSvcErr := &mockPeopleService{err: errSvc}
	authSvcErr := &mockAuthService{err: errSvc}
	projSvcErr := &mockProjectService{err: errSvc}
	catSvcErr := &mockCategoryService{err: errSvc}
	typeSvcErr := &mockTypeService{err: errSvc}
	paySvcErr := &mockPaymentService{err: errSvc}
	astSvcErr := &mockAssetService{err: errSvc}

	t.Run("makeRegisterEndpoint errors", func(t *testing.T) {
		ep := makeRegisterEndpoint(peopleSvcErr)
		// Unknown identity provider
		_, err := ep(context.Background(), RegisterUserDTO{IdentityProvider: "UNKNOWN"})
		if err == nil {
			t.Errorf("expected unknown provider error")
		}

		// Service error
		_, err = ep(context.Background(), RegisterUserDTO{IdentityProvider: "LOCAL", Login: "john@example.com", Token: "sec"})
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeLoginEndpoint errors", func(t *testing.T) {
		ep := makeLoginEndpoint(authSvcErr)
		// Invalid credentials (empty token)
		_, err := ep(context.Background(), LoginDTO{IdentityProvider: "LOCAL", ID: "john@example.com", Token: ""})
		if err == nil {
			t.Errorf("expected invalid credentials error")
		}

		// Service error
		_, err = ep(context.Background(), LoginDTO{IdentityProvider: "LOCAL", ID: "john@example.com", Token: "sec"})
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeProfileEndpoint error", func(t *testing.T) {
		ep := makeProfileEndpoint(peopleSvcErr)
		_, err := ep(context.Background(), uuid.New())
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeRefreshTokenEndpoint errors", func(t *testing.T) {
		ep := makeRefreshTokenEndpoint(authSvcErr)
		// Empty tokens
		_, err := ep(context.Background(), RefreshTokenDTO{})
		if err == nil {
			t.Errorf("expected empty tokens error")
		}

		// Service error
		_, err = ep(context.Background(), RefreshTokenDTO{AccessToken: "acc", RefreshToken: "ref"})
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeCreateProjectEndpoint error", func(t *testing.T) {
		ep := makeCreateProjectEndpoint(projSvcErr)
		_, err := ep(context.Background(), projecta.CreateProjectCommand{})
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeCreateCategoryEndpoint error", func(t *testing.T) {
		ep := makeCreateCategoryEndpoint(catSvcErr)
		_, err := ep(context.Background(), projecta.CreateCategoryCommand{})
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeCreateTypeEndpoint error", func(t *testing.T) {
		ep := makeCreateTypeEndpoint(typeSvcErr)
		_, err := ep(context.Background(), projecta.CreateTypeCommand{})
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeCreatePaymentEndpoint error", func(t *testing.T) {
		ep := makeCreatePaymentEndpoint(paySvcErr, nil)
		_, err := ep(context.Background(), projecta.CreatePaymentCommand{})
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeListPaymentsEndpoint error", func(t *testing.T) {
		ep := makeListPaymentsEndpoint(paySvcErr, nil)
		_, err := ep(context.Background(), projecta.PaymentCollectionFilter{})
		if err == nil {
			t.Errorf("expected service error")
		}
	})

	t.Run("makeRemovePaymentEndpoint error", func(t *testing.T) {
		ep := makeRemovePaymentEndpoint(paySvcErr)
		_, err := ep(context.Background(), "invalid request type")
		if err == nil {
			t.Errorf("expected type assertion error")
		}
	})

	t.Run("makeRemoveAssetEndpoint error", func(t *testing.T) {
		ep := makeRemoveAssetEndpoint(astSvcErr)
		_, err := ep(context.Background(), "invalid request type")
		if err == nil {
			t.Errorf("expected type assertion error")
		}
	})

	t.Run("makeShowProjectTotalsEndpoint currency mismatch and service error", func(t *testing.T) {
		mProjSvc := &mockProjectService{project: proj}
		// Payments service error
		epErr := makeShowProjectTotalsEndpoint(mProjSvc, paySvcErr, astSvcErr, nil)
		_, err := epErr(context.Background(), uuid.New())
		if err == nil {
			t.Errorf("expected payments service error")
		}

		// Assets service error
		paySvcOk := &mockPaymentService{pay: pay}
		epAssetErr := makeShowProjectTotalsEndpoint(mProjSvc, paySvcOk, astSvcErr, nil)
		_, err = epAssetErr(context.Background(), uuid.New())
		if err == nil {
			t.Errorf("expected assets service error")
		}

		// Payments conversion
		payEUR := projecta.NewPayment(uuid.New(), proj, owner, costType, "Pay EUR", money.New(50, money.EUR), time.Now(), projecta.DownPayment)
		colMismatch := projecta.NewPaymentCollection(2)
		colMismatch.Add(pay, payEUR)
		mPaySvc := &mockPaymentServiceWithCol{col: colMismatch}
		mAstSvc := &mockAssetService{asset: ast}

		epMismatch := makeShowProjectTotalsEndpoint(mProjSvc, mPaySvc, mAstSvc, nil)
		_, err = epMismatch(context.Background(), uuid.New())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Assets conversion
		astEUR := asset.NewAsset(uuid.New(), "Asset EUR", "Desc", proj, costType, money.New(500, money.EUR), time.Now(), owner)
		astColMismatch := asset.NewCollection(2)
		astColMismatch.Add(ast, astEUR)

		mPaySvcSingle := &mockPaymentServiceWithCol{col: projecta.NewPaymentCollection(1)}
		mPaySvcSingle.col.Add(pay)

		mAstSvcMismatch := &mockAssetServiceWithCol{col: astColMismatch}

		epAstMismatch := makeShowProjectTotalsEndpoint(mProjSvc, mPaySvcSingle, mAstSvcMismatch, nil)
		_, err = epAstMismatch(context.Background(), uuid.New())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

type mockPaymentServiceWithCol struct {
	mockPaymentService
	col *projecta.PaymentCollection
}

func (m *mockPaymentServiceWithCol) Find(ctx context.Context, filter projecta.PaymentCollectionFilter) (*projecta.PaymentCollection, error) {
	return m.col, nil
}

type mockAssetServiceWithCol struct {
	mockAssetService
	col *asset.Collection
}

func (m *mockAssetServiceWithCol) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	return m.col, nil
}
