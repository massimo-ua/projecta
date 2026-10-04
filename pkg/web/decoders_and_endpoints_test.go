package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
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

		reqBadFromDate, _ := http.NewRequest("GET", "/payments?from_date=bad", nil)
		reqBadFromDate = mux.SetURLVars(reqBadFromDate, map[string]string{"project_id": validUUID})
		_, err = decodeListPaymentsRequest(context.Background(), reqBadFromDate)
		if err == nil {
			t.Errorf("expected invalid from_date error")
		}

		reqBadToDate, _ := http.NewRequest("GET", "/payments?to_date=bad", nil)
		reqBadToDate = mux.SetURLVars(reqBadToDate, map[string]string{"project_id": validUUID})
		_, err = decodeListPaymentsRequest(context.Background(), reqBadToDate)
		if err == nil {
			t.Errorf("expected invalid to_date error")
		}

		reqValidDates, _ := http.NewRequest("GET", "/payments?from_date=2026-01-01&to_date=2026-01-31", nil)
		reqValidDates = mux.SetURLVars(reqValidDates, map[string]string{"project_id": validUUID})
		res, err := decodeListPaymentsRequest(context.Background(), reqValidDates)
		if err != nil {
			t.Fatalf("unexpected error decoding valid dates: %v", err)
		}
		filter := res.(projecta.PaymentCollectionFilter)
		if filter.FromDate.IsZero() || filter.ToDate.IsZero() {
			t.Errorf("expected non-zero FromDate and ToDate")
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

	t.Run("decodeCreateAssetFromPaymentsRequest validation and success", func(t *testing.T) {
		reqNoProj, _ := http.NewRequest("POST", "/", nil)
		_, err := decodeCreateAssetFromPaymentsRequest(context.Background(), reqNoProj)
		if err == nil {
			t.Errorf("expected missing project_id")
		}

		reqBadProj, _ := http.NewRequest("POST", "/", nil)
		reqBadProj = mux.SetURLVars(reqBadProj, map[string]string{"project_id": "bad"})
		_, err = decodeCreateAssetFromPaymentsRequest(context.Background(), reqBadProj)
		if err == nil {
			t.Errorf("expected invalid project_id")
		}

		reqBadJSON, _ := http.NewRequest("POST", "/", bytes.NewReader([]byte("not json")))
		reqBadJSON = mux.SetURLVars(reqBadJSON, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetFromPaymentsRequest(context.Background(), reqBadJSON)
		if err == nil {
			t.Errorf("expected invalid json error")
		}

		// Empty name
		bodyEmptyName, _ := json.Marshal(CreateAssetFromPaymentsDTO{Name: "", PaymentIDs: []string{validUUID}})
		reqEmptyName, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyEmptyName))
		reqEmptyName = mux.SetURLVars(reqEmptyName, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetFromPaymentsRequest(context.Background(), reqEmptyName)
		if err == nil {
			t.Errorf("expected empty name error")
		}

		// Empty payment_ids
		bodyEmptyPays, _ := json.Marshal(CreateAssetFromPaymentsDTO{Name: "Asset", PaymentIDs: []string{}})
		reqEmptyPays, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyEmptyPays))
		reqEmptyPays = mux.SetURLVars(reqEmptyPays, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetFromPaymentsRequest(context.Background(), reqEmptyPays)
		if err == nil {
			t.Errorf("expected empty payment_ids error")
		}

		// Invalid payment id
		bodyBadPayID, _ := json.Marshal(CreateAssetFromPaymentsDTO{Name: "Asset", PaymentIDs: []string{"bad-id"}})
		reqBadPayID, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadPayID))
		reqBadPayID = mux.SetURLVars(reqBadPayID, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetFromPaymentsRequest(context.Background(), reqBadPayID)
		if err == nil {
			t.Errorf("expected invalid payment id error")
		}

		// Bad type_id
		bodyBadType, _ := json.Marshal(CreateAssetFromPaymentsDTO{Name: "Asset", PaymentIDs: []string{validUUID}, TypeID: "bad"})
		reqBadType, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadType))
		reqBadType = mux.SetURLVars(reqBadType, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetFromPaymentsRequest(context.Background(), reqBadType)
		if err == nil {
			t.Errorf("expected bad type_id error")
		}

		// Bad acquired_at
		bodyBadDate, _ := json.Marshal(CreateAssetFromPaymentsDTO{Name: "Asset", PaymentIDs: []string{validUUID}, AcquiredAt: "bad-date"})
		reqBadDate, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyBadDate))
		reqBadDate = mux.SetURLVars(reqBadDate, map[string]string{"project_id": validUUID})
		_, err = decodeCreateAssetFromPaymentsRequest(context.Background(), reqBadDate)
		if err == nil {
			t.Errorf("expected bad acquired_at error")
		}

		// Valid YYYY-MM-DD
		bodyValidDay, _ := json.Marshal(CreateAssetFromPaymentsDTO{Name: "Asset", PaymentIDs: []string{validUUID}, AcquiredAt: "2026-05-15", TargetCurrency: "USD"})
		reqValidDay, _ := http.NewRequest("POST", "/", bytes.NewReader(bodyValidDay))
		reqValidDay = mux.SetURLVars(reqValidDay, map[string]string{"project_id": validUUID})
		res, err := decodeCreateAssetFromPaymentsRequest(context.Background(), reqValidDay)
		if err != nil {
			t.Fatalf("unexpected error decoding valid day request: %v", err)
		}
		cmd := res.(asset.CreateAssetFromPaymentsCommand)
		if cmd.Name != "Asset" || len(cmd.PaymentIDs) != 1 || cmd.TargetCurrency != "USD" {
			t.Errorf("unexpected decoded command: %+v", cmd)
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

	t.Run("makeUpdateProjectEndpoint error", func(t *testing.T) {
		ep := makeUpdateProjectEndpoint(projSvcErr)
		_, err := ep(context.Background(), projecta.UpdateProjectCommand{})
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

func (m *mockPaymentServiceWithCol) Find(_ context.Context, _ projecta.PaymentCollectionFilter) (*projecta.PaymentCollection, error) {
	return m.col, nil
}

type mockAssetServiceWithCol struct {
	mockAssetService
	col *asset.Collection
}

func (m *mockAssetServiceWithCol) Find(_ context.Context, _ asset.CollectionFilter) (*asset.Collection, error) {
	return m.col, nil
}

type mockRateProvider struct {
	err error
}

func (m *mockRateProvider) Convert(a currency.Currency, b currency.Currency) (currency.Currency, error) {
	if m.err != nil {
		return currency.Currency{}, m.err
	}
	return currency.Currency{Amount: a.Amount * 40, Code: b.Code}, nil
}

func TestAcceptShareAndGetProjectDecodersAndEndpoints(t *testing.T) {
	validUUID := uuid.New().String()

	t.Run("DecodeAcceptShareRequest", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, "/projects/share/", nil)
		_, err := DecodeAcceptShareRequest(context.Background(), req)
		if err == nil {
			t.Error("expected error for missing token")
		}

		req = mux.SetURLVars(req, map[string]string{"share_token": "invalid-uuid"})
		_, err = DecodeAcceptShareRequest(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid token UUID")
		}

		req = mux.SetURLVars(req, map[string]string{"share_token": validUUID})
		_, err = DecodeAcceptShareRequest(context.Background(), req)
		if err == nil {
			t.Error("expected error for missing requester ID")
		}

		ctx := context.WithValue(context.Background(), core.RequesterIDContextKey, uuid.New())
		res, err := DecodeAcceptShareRequest(ctx, req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := res.(AcceptShareCommand); !ok {
			t.Errorf("expected AcceptShareCommand result")
		}
	})

	t.Run("DecodeGetProjectRequest", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/projects/", nil)
		_, err := DecodeGetProjectRequest(context.Background(), req)
		if err == nil {
			t.Error("expected error for missing project_id")
		}

		req = mux.SetURLVars(req, map[string]string{"project_id": "invalid"})
		_, err = DecodeGetProjectRequest(context.Background(), req)
		if err == nil {
			t.Error("expected error for invalid project_id")
		}

		req = mux.SetURLVars(req, map[string]string{"project_id": validUUID})
		res, err := DecodeGetProjectRequest(context.Background(), req)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if _, ok := res.(projecta.ProjectFilter); !ok {
			t.Errorf("expected ProjectFilter result")
		}
	})

	t.Run("makeGetProjectEndpoint and makeAcceptShareEndpoint", func(t *testing.T) {
		owner := &projecta.Owner{PersonID: uuid.New(), DisplayName: "Owner"}
		proj, _ := projecta.NewProject(uuid.New(), "Project One", "Desc", owner, time.Now(), time.Now())
		projSvc := &mockProjectService{project: proj}

		epGet := makeGetProjectEndpoint(projSvc)
		res, err := epGet(context.Background(), projecta.ProjectFilter{ProjectID: proj.ProjectID})
		if err != nil || res == nil {
			t.Fatalf("makeGetProjectEndpoint error: %v", err)
		}

		epAccept := makeAcceptShareEndpoint(projSvc)
		res, err = epAccept(context.Background(), AcceptShareCommand{ShareToken: proj.ShareToken, PersonID: owner.PersonID})
		if err != nil || res == nil {
			t.Fatalf("makeAcceptShareEndpoint error: %v", err)
		}

		projErrSvc := &mockProjectService{err: errors.New("svc error")}
		epGetErr := makeGetProjectEndpoint(projErrSvc)
		if _, err := epGetErr(context.Background(), projecta.ProjectFilter{}); err == nil {
			t.Error("expected error from makeGetProjectEndpoint")
		}

		epAcceptErr := makeAcceptShareEndpoint(projErrSvc)
		if _, err := epAcceptErr(context.Background(), AcceptShareCommand{}); err == nil {
			t.Error("expected error from makeAcceptShareEndpoint")
		}
	})

	t.Run("toProjectDTO and ProjectDTO JSON marshaling with participants", func(t *testing.T) {
		owner := &projecta.Owner{PersonID: uuid.New(), DisplayName: "Owner Name"}
		proj, _ := projecta.NewProject(uuid.New(), "Project Alpha", "Desc", owner, time.Now(), time.Now())
		part1, _ := projecta.NewParticipant("Participant Two")
		proj.AddParticipant(part1)

		dto := toProjectDTO(proj)
		if len(dto.Participants) != 2 {
			t.Fatalf("expected 2 participants, got %d", len(dto.Participants))
		}

		data, err := json.Marshal(dto)
		if err != nil {
			t.Fatalf("failed to marshal ProjectDTO: %v", err)
		}

		var raw map[string]any
		if err := json.Unmarshal(data, &raw); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		partsRaw, ok := raw["participants"].([]any)
		if !ok {
			t.Fatalf("expected 'participants' array in JSON, got %T", raw["participants"])
		}

		if len(partsRaw) != 2 || partsRaw[0] != "Owner Name" || partsRaw[1] != "Participant Two" {
			t.Errorf("unexpected participants JSON: %v", partsRaw)
		}
	})

	t.Run("decodeUpdateProjectRequest error branches", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPatch, "/projects/123", nil)
		_, err := decodeUpdateProjectRequest(context.Background(), req)
		if err == nil {
			t.Error("expected error for missing requester ID")
		}

		ctx := context.WithValue(context.Background(), core.RequesterIDContextKey, uuid.New())
		_, err = decodeUpdateProjectRequest(ctx, req)
		if err == nil {
			t.Error("expected error for missing project_id")
		}

		req = mux.SetURLVars(req, map[string]string{"project_id": "bad-id"})
		_, err = decodeUpdateProjectRequest(ctx, req)
		if err == nil {
			t.Error("expected error for invalid project_id")
		}
	})

	t.Run("currency conversion endpoints with rateProvider", func(t *testing.T) {
		owner := &projecta.Owner{PersonID: uuid.New(), DisplayName: "Owner"}
		proj, _ := projecta.NewProject(uuid.New(), "Project One", "Desc", owner, time.Now(), time.Now())
		proj.MainCurrency = "UAH"
		cat, _ := projecta.NewCostCategory(uuid.New(), proj.ProjectID, "Cat1", "Desc")
		costType, _ := projecta.NewCostType(proj.ProjectID, cat, "Type1", "Desc")

		payUSD := projecta.NewPayment(uuid.New(), proj, owner, costType, "Pay USD", money.New(100, money.USD), time.Now(), projecta.DownPayment)
		astUSD := asset.NewAsset(uuid.New(), "Ast USD", "Desc", proj, costType, money.New(200, money.USD), time.Now(), owner)

		rateProv := &mockRateProvider{}

		pDto := toPaymentDTO(payUSD, rateProv)
		if pDto.HomeAmount != 4000 {
			t.Errorf("expected converted home amount 4000, got %d", pDto.HomeAmount)
		}

		aDto := toAssetDTO(astUSD, rateProv)
		if aDto.HomeAmount != 8000 {
			t.Errorf("expected converted home amount 8000, got %d", aDto.HomeAmount)
		}

		paySvc := &mockPaymentService{pay: payUSD}
		epCreatePay := makeCreatePaymentEndpoint(paySvc, rateProv)
		if _, err := epCreatePay(context.Background(), projecta.CreatePaymentCommand{}); err != nil {
			t.Errorf("makeCreatePaymentEndpoint error: %v", err)
		}

		epGetPay := makeGetPaymentEndpoint(paySvc, rateProv)
		if _, err := epGetPay(context.Background(), projecta.PaymentFilter{}); err != nil {
			t.Errorf("makeGetPaymentEndpoint error: %v", err)
		}

		col := projecta.NewPaymentCollection(1)
		col.Add(payUSD)
		mPayColSvc := &mockPaymentServiceWithCol{col: col}
		epListPay := makeListPaymentsEndpoint(mPayColSvc, rateProv)
		if _, err := epListPay(context.Background(), projecta.PaymentCollectionFilter{}); err != nil {
			t.Errorf("makeListPaymentsEndpoint error: %v", err)
		}

		astSvc := &mockAssetService{asset: astUSD}
		epCreateAst := makeCreateAssetEndpoint(astSvc, rateProv)
		if _, err := epCreateAst(context.Background(), asset.CreateAssetCommand{}); err != nil {
			t.Errorf("makeCreateAssetEndpoint error: %v", err)
		}

		epCreateFromPays := makeCreateAssetFromPaymentsEndpoint(astSvc, rateProv)
		if _, err := epCreateFromPays(context.Background(), asset.CreateAssetFromPaymentsCommand{}); err != nil {
			t.Errorf("makeCreateAssetFromPaymentsEndpoint error: %v", err)
		}

		epGetAst := makeGetAssetEndpoint(astSvc, rateProv)
		if _, err := epGetAst(context.Background(), asset.Filter{}); err != nil {
			t.Errorf("makeGetAssetEndpoint error: %v", err)
		}

		astCol := asset.NewCollection(1)
		astCol.Add(astUSD)
		mAstColSvc := &mockAssetServiceWithCol{col: astCol}
		epListAst := makeListAssetsEndpoint(mAstColSvc, rateProv)
		if _, err := epListAst(context.Background(), asset.CollectionFilter{}); err != nil {
			t.Errorf("makeListAssetsEndpoint error: %v", err)
		}

		mProjSvc := &mockProjectService{project: proj}
		epTotals := makeShowProjectTotalsEndpoint(mProjSvc, mPayColSvc, mAstColSvc, rateProv)
		resTotals, err := epTotals(context.Background(), proj.ProjectID)
		if err != nil || resTotals == nil {
			t.Fatalf("makeShowProjectTotalsEndpoint error: %v", err)
		}

		errRateProv := &mockRateProvider{err: errors.New("rate error")}
		epTotalsErr := makeShowProjectTotalsEndpoint(mProjSvc, mPayColSvc, mAstColSvc, errRateProv)
		if _, err := epTotalsErr(context.Background(), proj.ProjectID); err == nil {
			t.Error("expected rate error in makeShowProjectTotalsEndpoint")
		}
	})

	t.Run("User roles decoders and endpoints", func(t *testing.T) {
		// decodeListUsersRequest default and params
		req1, _ := http.NewRequest(http.MethodGet, "/users", nil)
		res1, err := decodeListUsersRequest(context.Background(), req1)
		if err != nil || res1.(core.Pagination).Limit != core.DefaultLimit {
			t.Errorf("decodeListUsersRequest default failed: %v", err)
		}

		req2, _ := http.NewRequest(http.MethodGet, "/users?limit=5&offset=10", nil)
		res2, err := decodeListUsersRequest(context.Background(), req2)
		if err != nil || res2.(core.Pagination).Limit != 5 || res2.(core.Pagination).Offset != 10 {
			t.Errorf("decodeListUsersRequest custom failed: %v", err)
		}

		reqBadLimit, _ := http.NewRequest(http.MethodGet, "/users?limit=bad", nil)
		_, err = decodeListUsersRequest(context.Background(), reqBadLimit)
		if err == nil {
			t.Error("expected error for bad limit")
		}

		reqBadOffset, _ := http.NewRequest(http.MethodGet, "/users?offset=bad", nil)
		_, err = decodeListUsersRequest(context.Background(), reqBadOffset)
		if err == nil {
			t.Error("expected error for bad offset")
		}

		// decodeAssignRolesRequest
		targetID := uuid.New()
		bodyBytes, _ := json.Marshal(AssignRolesDTO{Roles: []string{"User", "Administrator"}})
		reqAssign, _ := http.NewRequest(http.MethodPut, "/users/"+targetID.String()+"/roles", bytes.NewReader(bodyBytes))
		reqAssign = mux.SetURLVars(reqAssign, map[string]string{"user_id": targetID.String()})
		resAssign, err := decodeAssignRolesRequest(context.Background(), reqAssign)
		if err != nil || resAssign.(AssignRolesRequest).UserID != targetID || len(resAssign.(AssignRolesRequest).Roles) != 2 {
			t.Fatalf("decodeAssignRolesRequest failed: %v", err)
		}

		// decodeAssignRolesRequest errors
		reqNoVar, _ := http.NewRequest(http.MethodPut, "/users//roles", bytes.NewReader(bodyBytes))
		_, err = decodeAssignRolesRequest(context.Background(), reqNoVar)
		if err == nil {
			t.Error("expected missing user_id error")
		}

		reqBadID, _ := http.NewRequest(http.MethodPut, "/users/not-uuid/roles", bytes.NewReader(bodyBytes))
		reqBadID = mux.SetURLVars(reqBadID, map[string]string{"user_id": "not-uuid"})
		_, err = decodeAssignRolesRequest(context.Background(), reqBadID)
		if err == nil {
			t.Error("expected invalid user_id error")
		}

		reqBadJSON, _ := http.NewRequest(http.MethodPut, "/users/"+targetID.String()+"/roles", bytes.NewReader([]byte("{bad json")))
		reqBadJSON = mux.SetURLVars(reqBadJSON, map[string]string{"user_id": targetID.String()})
		_, err = decodeAssignRolesRequest(context.Background(), reqBadJSON)
		if err == nil {
			t.Error("expected bad JSON error")
		}

		// makeListUsersEndpoint
		cred, _ := people.NewCredentials(people.LOCAL, "user@test.com", "pass")
		testUser, _ := people.NewPerson(targetID, "Test", "User", "Tester", []people.Credentials{cred}, people.RoleUser)
		mPeopleSvc := &mockPeopleService{user: testUser}
		epListUsers := makeListUsersEndpoint(mPeopleSvc)
		resListUsers, err := epListUsers(context.Background(), core.Pagination{Limit: 10, Offset: 0})
		if err != nil || len(resListUsers.(ListUsersResponse).Users) != 1 {
			t.Fatalf("makeListUsersEndpoint failed: %v", err)
		}

		mPeopleSvcErr := &mockPeopleService{err: errors.New("people err")}
		epListUsersErr := makeListUsersEndpoint(mPeopleSvcErr)
		if _, err := epListUsersErr(context.Background(), core.Pagination{}); err == nil {
			t.Error("expected error from makeListUsersEndpoint")
		}

		// makeAssignRolesEndpoint
		epAssignRoles := makeAssignRolesEndpoint(mPeopleSvc)
		resAssigned, err := epAssignRoles(context.Background(), AssignRolesRequest{
			UserID: targetID,
			Roles:  []string{"Administrator"},
		})
		if err != nil || resAssigned.(UserDTO).CustomerID != targetID.String() {
			t.Fatalf("makeAssignRolesEndpoint failed: %v", err)
		}

		// makeAssignRolesEndpoint invalid role
		_, err = epAssignRoles(context.Background(), AssignRolesRequest{
			UserID: targetID,
			Roles:  []string{"InvalidRole"},
		})
		if err == nil {
			t.Error("expected invalid role error")
		}

		// makeAssignRolesEndpoint service error
		epAssignRolesErr := makeAssignRolesEndpoint(mPeopleSvcErr)
		_, err = epAssignRolesErr(context.Background(), AssignRolesRequest{
			UserID: targetID,
			Roles:  []string{"User"},
		})
		if err == nil {
			t.Error("expected service error")
		}
	})

	t.Run("Invitation Endpoints and Decoders", func(t *testing.T) {
		adminID := uuid.New()
		invID := uuid.New()
		inv, _ := people.NewInvitation(invID, "invited@example.com", "hash123", time.Now().Add(7*24*time.Hour), adminID, nil, people.InvitationStatusPending, time.Now(), nil)
		invWithCode := &people.InvitationWithCode{Invitation: inv, Code: "raw_code_123"}

		invSvc := &mockInvitationService{
			invWithCode: invWithCode,
			invList:     []*people.Invitation{inv},
			inv:         inv,
		}

		eps := MakeInvitationEndpoints(invSvc)

		// Create
		resCreate, err := eps.CreateInvitation(context.Background(), struct {
			AdminID uuid.UUID
			Email   string
		}{AdminID: adminID, Email: "invited@example.com"})
		if err != nil {
			t.Fatalf("CreateInvitation failed: %v", err)
		}
		createResp := resCreate.(InvitationCreatedResponseDTO)
		if createResp.Code != "raw_code_123" || createResp.Email != "invited@example.com" {
			t.Errorf("CreateInvitation response mismatch: %+v", createResp)
		}

		// List
		resList, err := eps.ListInvitations(context.Background(), adminID)
		if err != nil {
			t.Fatalf("ListInvitations failed: %v", err)
		}
		listResp := resList.(ListInvitationsResponse)
		if len(listResp.Invitations) != 1 {
			t.Errorf("expected 1 invitation in list")
		}

		// Delete
		_, err = eps.DeleteInvitation(context.Background(), DeleteInvitationRequest{AdminID: adminID, InvitationID: invID})
		if err != nil {
			t.Fatalf("DeleteInvitation failed: %v", err)
		}

		// Validate
		resVal, err := eps.ValidateInvitation(context.Background(), "raw_code_123")
		if err != nil {
			t.Fatalf("ValidateInvitation failed: %v", err)
		}
		valResp := resVal.(ValidateInvitationResponseDTO)
		if valResp.Email != "invited@example.com" {
			t.Errorf("ValidateInvitation response mismatch: %+v", valResp)
		}

		// Login with Invitation Code
		loginEp := makeLoginEndpoint(&mockAuthService{authResp: &core.AuthResponse{AccessToken: "acc"}})
		resLogin, err := loginEp(context.Background(), LoginDTO{
			IdentityProvider: "GOOGLE",
			Token:            "g_token",
			InvitationCode:   "raw_code_123",
		})
		if err != nil || resLogin == nil {
			t.Fatalf("Login with invitation failed: %v", err)
		}

		// Login with invalid provider
		_, err = loginEp(context.Background(), LoginDTO{
			IdentityProvider: "INVALID_PROV",
			Token:            "g_token",
			InvitationCode:   "raw_code_123",
		})
		if err == nil {
			t.Errorf("expected error for invalid provider")
		}

		// Decoders tests
		t.Run("Invitation Decoders", func(t *testing.T) {
			body := strings.NewReader(`{"email":"test@example.com"}`)
			req := httptest.NewRequest(http.MethodPost, "/invitations", body)
			_, err := decodeCreateInvitationRequest(context.Background(), req)
			if err == nil {
				t.Errorf("expected unauthorized error")
			}

			authCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, adminID)
			body = strings.NewReader(`{"email":"test@example.com"}`)
			req = httptest.NewRequest(http.MethodPost, "/invitations", body)
			decCreate, err := decodeCreateInvitationRequest(authCtx, req)
			if err != nil {
				t.Fatalf("unexpected error decoding create invitation: %v", err)
			}
			createReq := decCreate.(struct {
				AdminID uuid.UUID
				Email   string
			})
			if createReq.Email != "test@example.com" || createReq.AdminID != adminID {
				t.Errorf("mismatch create invitation request: %+v", createReq)
			}

			badBody := strings.NewReader(`invalid json`)
			req = httptest.NewRequest(http.MethodPost, "/invitations", badBody)
			_, err = decodeCreateInvitationRequest(authCtx, req)
			if err == nil {
				t.Errorf("expected validation error for bad json")
			}

			req = httptest.NewRequest(http.MethodGet, "/invitations", nil)
			_, err = decodeListInvitationsRequest(context.Background(), req)
			if err == nil {
				t.Errorf("expected unauthorized error")
			}
			decList, err := decodeListInvitationsRequest(authCtx, req)
			if err != nil || decList.(uuid.UUID) != adminID {
				t.Errorf("unexpected error decoding list invitations: %v", err)
			}

			req = httptest.NewRequest(http.MethodDelete, "/invitations/"+invID.String(), nil)
			req = mux.SetURLVars(req, map[string]string{"invitation_id": invID.String()})
			_, err = decodeDeleteInvitationRequest(context.Background(), req)
			if err == nil {
				t.Errorf("expected unauthorized error")
			}
			decDel, err := decodeDeleteInvitationRequest(authCtx, req)
			if err != nil {
				t.Fatalf("unexpected error decoding delete invitation: %v", err)
			}
			delReq := decDel.(DeleteInvitationRequest)
			if delReq.InvitationID != invID || delReq.AdminID != adminID {
				t.Errorf("mismatch delete invitation request: %+v", delReq)
			}

			badDelReq := httptest.NewRequest(http.MethodDelete, "/invitations/bad-id", nil)
			badDelReq = mux.SetURLVars(badDelReq, map[string]string{"invitation_id": "bad-id"})
			_, err = decodeDeleteInvitationRequest(authCtx, badDelReq)
			if err == nil {
				t.Errorf("expected validation error for bad invitation id")
			}

			req = httptest.NewRequest(http.MethodGet, "/invitations/code/raw_code_123", nil)
			req = mux.SetURLVars(req, map[string]string{"code": "raw_code_123"})
			decCode, err := decodeValidateInvitationRequest(context.Background(), req)
			if err != nil || decCode.(string) != "raw_code_123" {
				t.Errorf("unexpected error decoding validate invitation request: %v", err)
			}

			badCodeReq := httptest.NewRequest(http.MethodGet, "/invitations/code/", nil)
			_, err = decodeValidateInvitationRequest(context.Background(), badCodeReq)
			if err == nil {
				t.Errorf("expected validation error for empty code")
			}

			whitespaceCodeReq := httptest.NewRequest(http.MethodGet, "/invitations/code/%20", nil)
			whitespaceCodeReq = mux.SetURLVars(whitespaceCodeReq, map[string]string{"code": "   "})
			_, err = decodeValidateInvitationRequest(context.Background(), whitespaceCodeReq)
			if err == nil {
				t.Errorf("expected validation error for whitespace code")
			}
		})
	})
}

func TestUpdateProfileEndpointAndDecoder(t *testing.T) {
	validUserID := uuid.New()
	authCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, validUserID)

	t.Run("decodeUpdateProfileRequest success", func(t *testing.T) {
		body := bytes.NewBufferString(`{"display_name": "New Display Name"}`)
		req := httptest.NewRequest(http.MethodPut, "/profile", body)
		res, err := decodeUpdateProfileRequest(authCtx, req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		updateReq, ok := res.(UpdateProfileRequest)
		if !ok {
			t.Fatalf("expected UpdateProfileRequest type")
		}
		if updateReq.PersonID != validUserID {
			t.Errorf("expected PersonID %s, got %s", validUserID, updateReq.PersonID)
		}
		if updateReq.DisplayName.String() != "New Display Name" {
			t.Errorf("expected 'New Display Name', got '%s'", updateReq.DisplayName.String())
		}
	})

	t.Run("decodeUpdateProfileRequest unauthenticated", func(t *testing.T) {
		body := bytes.NewBufferString(`{"display_name": "New Display Name"}`)
		req := httptest.NewRequest(http.MethodPut, "/profile", body)
		_, err := decodeUpdateProfileRequest(context.Background(), req)
		if err == nil {
			t.Errorf("expected unauthorized error")
		}
	})

	t.Run("decodeUpdateProfileRequest invalid JSON", func(t *testing.T) {
		body := bytes.NewBufferString(`{invalid json}`)
		req := httptest.NewRequest(http.MethodPut, "/profile", body)
		_, err := decodeUpdateProfileRequest(authCtx, req)
		if err == nil {
			t.Errorf("expected validation error on invalid JSON")
		}
	})

	t.Run("decodeUpdateProfileRequest invalid display name length via json.Unmarshaler", func(t *testing.T) {
		longName := strings.Repeat("x", 256)
		payload, _ := json.Marshal(map[string]string{"display_name": longName})
		req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewReader(payload))
		_, err := decodeUpdateProfileRequest(authCtx, req)
		if err == nil {
			t.Errorf("expected error decoding too long display name into DisplayName VO")
		}
	})

	t.Run("makeUpdateProfileEndpoint success and error", func(t *testing.T) {
		dn, _ := people.NewDisplayName("Updated Name")
		p, _ := people.NewPerson(validUserID, "First", "Last", "Old", nil, people.RoleUser)

		// Success
		svcOk := &mockPeopleService{user: p}
		epOk := makeUpdateProfileEndpoint(svcOk)
		res, err := epOk(context.Background(), UpdateProfileRequest{
			PersonID:    validUserID,
			DisplayName: dn,
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		dto := res.(UserDTO)
		if dto.DisplayName != "Updated Name" {
			t.Errorf("expected display name 'Updated Name', got '%s'", dto.DisplayName)
		}

		// Service error
		svcErr := &mockPeopleService{err: errors.New("svc error")}
		epErr := makeUpdateProfileEndpoint(svcErr)
		_, err = epErr(context.Background(), UpdateProfileRequest{
			PersonID:    validUserID,
			DisplayName: dn,
		})
		if err == nil {
			t.Errorf("expected service error")
		}
	})
}



