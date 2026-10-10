package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

type ChildAssetDTO struct {
	ChildID         string  `json:"child_id"`
	ChildName       string  `json:"child_name"`
	SharePercentage float64 `json:"share_percentage"`
	TotalCost       int64   `json:"total_cost"`
}

type ParentAssetDTO struct {
	ParentID        string  `json:"parent_id"`
	ParentName      string  `json:"parent_name"`
	SharePercentage float64 `json:"share_percentage"`
}

type AssetDTO struct {
	AssetID        string           `json:"asset_id"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	Status         string           `json:"status"`
	StartDate      string           `json:"start_date"`
	CompletedDate  *string          `json:"completed_date,omitempty"`
	TargetPrice    *int64           `json:"target_price,omitempty"`
	TargetCurrency string           `json:"target_currency,omitempty"`
	DirectCost     int64            `json:"direct_cost"`
	TotalCost      int64            `json:"total_cost"`
	Price          int64            `json:"price"` // backward compatibility
	Currency       string           `json:"currency"`
	HomeAmount     int64            `json:"home_amount,omitempty"`
	HomeCurrency   string           `json:"home_currency,omitempty"`
	AcquiredAt     string           `json:"acquired_at"`
	Owner          OwnerDTO         `json:"owner"`
	Project        ProjectDTO       `json:"project"`
	Type           *TypeDTO         `json:"type,omitempty"`
	Children       []ChildAssetDTO  `json:"children"`
	Parents        []ParentAssetDTO `json:"parents"`
	Tags           []string         `json:"tags"`
}

func toAssetDTO(a *asset.Asset, rateProvider currency.CurrencyRateProvider) AssetDTO {
	if a == nil {
		return AssetDTO{}
	}
	projDTO := toProjectDTO(a.Project())
	homeCurrency := projDTO.MainCurrency
	if homeCurrency == "" {
		homeCurrency = "UAH"
	}

	priceAmount := int64(0)
	priceCurrency := homeCurrency
	if a.Price() != nil {
		priceAmount = a.Price().Amount()
		priceCurrency = a.Price().Currency().Code
	}

	homeAmount := priceAmount
	if rateProvider != nil && priceCurrency != homeCurrency && priceAmount > 0 {
		converted, err := rateProvider.Convert(
			currency.NewCurrency(priceAmount, priceCurrency),
			currency.NewCurrency(0, homeCurrency),
		)
		if err == nil {
			homeAmount = converted.Amount
		}
	}

	var directCostVal int64
	if a.DirectCost() != nil {
		directCostVal = a.DirectCost().Amount()
	}

	var totalCostVal int64
	if a.TotalCost() != nil {
		totalCostVal = a.TotalCost().Amount()
	} else {
		totalCostVal = priceAmount
	}

	var targetPriceVal *int64
	var targetCurrVal string
	if a.TargetPrice() != nil {
		v := a.TargetPrice().Amount()
		targetPriceVal = &v
		targetCurrVal = a.TargetPrice().Currency().Code
	}

	var completedDateStr *string
	if a.CompletedDate() != nil {
		s := a.CompletedDate().Format(time.RFC3339)
		completedDateStr = &s
	}

	owner := OwnerDTO{
		PersonID:    a.Owner().PersonID.String(),
		DisplayName: a.Owner().DisplayName,
	}

	var typeDTO *TypeDTO
	if a.Type() != nil {
		typeDTO = &TypeDTO{
			TypeID: a.Type().ID.String(),
			Name:   a.Type().Name,
		}
		if a.Type().Category != nil {
			typeDTO.Category = TypeCategoryDTO{
				CategoryID: a.Type().Category.ID.String(),
				Name:       a.Type().Category.Name,
			}
		}
	}

	childrenDTO := make([]ChildAssetDTO, 0)
	for _, c := range a.Children() {
		var cCost int64
		if c.TotalCost != nil {
			cCost = c.TotalCost.Amount()
		}
		childrenDTO = append(childrenDTO, ChildAssetDTO{
			ChildID:         c.ChildID.String(),
			ChildName:       c.ChildName,
			SharePercentage: c.SharePercentage,
			TotalCost:       cCost,
		})
	}

	parentsDTO := make([]ParentAssetDTO, 0)
	for _, p := range a.Parents() {
		parentsDTO = append(parentsDTO, ParentAssetDTO{
			ParentID:        p.ParentID.String(),
			ParentName:      p.ParentName,
			SharePercentage: p.SharePercentage,
		})
	}

	return AssetDTO{
		AssetID:        a.ID().String(),
		Name:           a.Name(),
		Description:    a.Description(),
		Status:         a.Status().String(),
		StartDate:      a.StartDate().Format(time.RFC3339),
		CompletedDate:  completedDateStr,
		TargetPrice:    targetPriceVal,
		TargetCurrency: targetCurrVal,
		DirectCost:     directCostVal,
		TotalCost:      totalCostVal,
		Price:          totalCostVal,
		Currency:       priceCurrency,
		HomeAmount:     homeAmount,
		HomeCurrency:   homeCurrency,
		AcquiredAt:     a.AcquiredAt().Format(time.RFC3339),
		Owner:          owner,
		Project:        projDTO,
		Type:           typeDTO,
		Children:       childrenDTO,
		Parents:        parentsDTO,
		Tags:           a.Tags(),
	}
}

type CreateAssetDTO struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	TypeID         string   `json:"type_id,omitempty"`
	Price          int64    `json:"price,omitempty"`
	Currency       string   `json:"currency,omitempty"`
	TargetPrice    int64    `json:"target_price,omitempty"`
	TargetCurrency string   `json:"target_currency,omitempty"`
	Status         string   `json:"status,omitempty"`
	StartDate      string   `json:"start_date,omitempty"`
	CompletedDate  string   `json:"completed_date,omitempty"`
	AcquiredAt     string   `json:"acquired_at,omitempty"`
	WithPayment    bool     `json:"with_payment"`
	Tags           []string `json:"tags,omitempty"`
}

type CreateAssetFromPaymentsDTO struct {
	PaymentIDs     []string `json:"payment_ids"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	TypeID         string   `json:"type_id,omitempty"`
	AcquiredAt     string   `json:"acquired_at,omitempty"`
	TargetCurrency string   `json:"target_currency,omitempty"`
}

type UpdateAssetDTO struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	TypeID         string   `json:"type_id,omitempty"`
	Price          int64    `json:"price,omitempty"`
	Currency       string   `json:"currency,omitempty"`
	TargetPrice    int64    `json:"target_price,omitempty"`
	TargetCurrency string   `json:"target_currency,omitempty"`
	Status         string   `json:"status,omitempty"`
	StartDate      string   `json:"start_date,omitempty"`
	CompletedDate  string   `json:"completed_date,omitempty"`
	AcquiredAt     string   `json:"acquired_at,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type LinkChildAssetDTO struct {
	ChildAssetID    string  `json:"child_asset_id"`
	SharePercentage float64 `json:"share_percentage"`
}

type AssignInvestmentsDTO struct {
	InvestmentIDs []string `json:"investment_ids"`
}


type ListAssetsResponse struct {
	Assets []AssetDTO `json:"assets"`
	PaginationDTO
}

func decodeCreateAssetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateAssetDTO
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if req.Name == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	var typeUUID uuid.UUID
	if req.TypeID != "" {
		var err error
		typeUUID, err = uuid.Parse(req.TypeID)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid type id", err)
		}
	}

	if req.Price <= 0 && req.TargetPrice <= 0 {
		return nil, exceptions.NewValidationException("price must be greater than 0", nil)
	}

	if req.Currency == "" && req.TargetCurrency == "" {
		return nil, exceptions.NewValidationException("currency is required", nil)
	}

	var price *money.Money
	if req.Price > 0 && req.Currency != "" {
		price = money.New(req.Price, req.Currency)
	}

	var targetPrice *money.Money
	if req.TargetPrice > 0 && req.TargetCurrency != "" {
		targetPrice = money.New(req.TargetPrice, req.TargetCurrency)
	}

	var date time.Time
	if req.AcquiredAt != "" {
		var err error
		date, err = time.Parse(time.RFC3339, req.AcquiredAt)
		if err != nil {
			date, err = time.Parse("2006-01-02", req.AcquiredAt)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid acquired at date", err)
			}
		}
	} else if req.StartDate != "" {
		var err error
		date, err = time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			date, err = time.Parse("2006-01-02", req.StartDate)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid acquired at date", err)
			}
		}
	}
	if date.IsZero() {
		date = time.Now()
	}

	var completedDate *time.Time
	if req.CompletedDate != "" {
		cd, err := time.Parse(time.RFC3339, req.CompletedDate)
		if err == nil {
			completedDate = &cd
		} else {
			cd, err = time.Parse("2006-01-02", req.CompletedDate)
			if err == nil {
				completedDate = &cd
			}
		}
	}

	return asset.CreateAssetCommand{
		Name:          req.Name,
		Description:   req.Description,
		ProjectID:     projectUUID,
		TypeID:        typeUUID,
		Price:         price,
		TargetPrice:   targetPrice,
		Status:        asset.ToAssetStatus(req.Status),
		StartDate:     date,
		CompletedDate: completedDate,
		AcquiredAt:    date,
		WithPayment:   req.WithPayment,
		Tags:          req.Tags,
	}, nil
}

func decodeCreateAssetFromPaymentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateAssetFromPaymentsDTO
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if req.Name == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	if len(req.PaymentIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one payment must be selected", nil)
	}

	var paymentUUIDs []uuid.UUID
	for _, pid := range req.PaymentIDs {
		pUUID, err := uuid.Parse(pid)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid payment id in list", err)
		}
		paymentUUIDs = append(paymentUUIDs, pUUID)
	}

	var typeUUID uuid.UUID
	if req.TypeID != "" {
		var err error
		typeUUID, err = uuid.Parse(req.TypeID)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid type id", err)
		}
	}

	var date time.Time
	if req.AcquiredAt != "" {
		date, err = time.Parse(time.RFC3339, req.AcquiredAt)
		if err != nil {
			date, err = time.Parse("2006-01-02", req.AcquiredAt)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid acquired at date", err)
			}
		}
	}

	return asset.CreateAssetFromPaymentsCommand{
		ProjectID:      projectUUID,
		PaymentIDs:     paymentUUIDs,
		Name:           req.Name,
		Description:    req.Description,
		TypeID:         typeUUID,
		AcquiredAt:     date,
		TargetCurrency: req.TargetCurrency,
	}, nil
}

func decodeGetAssetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetID, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}

	assetUUID, err := uuid.Parse(assetID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	return asset.Filter{
		ID:        assetUUID,
		ProjectID: projectUUID,
	}, nil
}

func decodeUpdateAssetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetID, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}

	assetUUID, err := uuid.Parse(assetID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var req UpdateAssetDTO
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if req.Name == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	var typeUUID uuid.UUID
	if req.TypeID != "" {
		var err error
		typeUUID, err = uuid.Parse(req.TypeID)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid type id", err)
		}
	}

	if req.Price <= 0 && req.TargetPrice <= 0 {
		return nil, exceptions.NewValidationException("price must be greater than 0", nil)
	}

	if req.Currency == "" && req.TargetCurrency == "" {
		return nil, exceptions.NewValidationException("currency is required", nil)
	}

	var price *money.Money
	if req.Price > 0 && req.Currency != "" {
		price = money.New(req.Price, req.Currency)
	}

	var targetPrice *money.Money
	if req.TargetPrice > 0 && req.TargetCurrency != "" {
		targetPrice = money.New(req.TargetPrice, req.TargetCurrency)
	}

	var date time.Time
	if req.AcquiredAt != "" {
		var err error
		date, err = time.Parse(time.RFC3339, req.AcquiredAt)
		if err != nil {
			date, err = time.Parse("2006-01-02", req.AcquiredAt)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid acquired at date", err)
			}
		}
	} else if req.StartDate != "" {
		var err error
		date, err = time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			date, err = time.Parse("2006-01-02", req.StartDate)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid acquired at date", err)
			}
		}
	}

	var completedDate *time.Time
	if req.CompletedDate != "" {
		cd, err := time.Parse(time.RFC3339, req.CompletedDate)
		if err == nil {
			completedDate = &cd
		} else {
			cd, err = time.Parse("2006-01-02", req.CompletedDate)
			if err == nil {
				completedDate = &cd
			}
		}
	}

	return asset.UpdateAssetCommand{
		AssetID:       assetUUID,
		Name:          req.Name,
		Description:   req.Description,
		ProjectID:     projectUUID,
		TypeID:        typeUUID,
		Price:         price,
		TargetPrice:   targetPrice,
		Status:        asset.ToAssetStatus(req.Status),
		StartDate:     date,
		CompletedDate: completedDate,
		AcquiredAt:    date,
		Tags:          req.Tags,
	}, nil
}

func decodeListAssetsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	limit := 10
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		val, err := strconv.Atoi(l)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid limit", err)
		}
		if val > 0 {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		val, err := strconv.Atoi(o)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid offset", err)
		}
		if val >= 0 {
			offset = val
		}
	}

	var typeID uuid.UUID
	if t := r.URL.Query().Get("type_id"); t != "" {
		var err error
		typeID, err = uuid.Parse(t)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid type id", err)
		}
	}

	name := r.URL.Query().Get("name")

	return asset.CollectionFilter{
		Pagination: core.Pagination{
			Limit:  limit,
			Offset: offset,
		},
		ProjectID: projectUUID,
		TypeID:    typeID,
		Name:      name,
	}, nil
}

func decodeLinkChildAssetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetIDStr, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}
	parentID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var req LinkChildAssetDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	childID, err := uuid.Parse(req.ChildAssetID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid child asset id", err)
	}

	share := req.SharePercentage
	if share <= 0 {
		share = 100.0
	}

	return asset.LinkChildCommand{
		ParentID:        parentID,
		ChildID:         childID,
		ProjectID:       projectID,
		SharePercentage: share,
	}, nil
}

func decodeUnlinkChildAssetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	parentIDStr, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid parent asset id", nil)
	}
	parentID, err := uuid.Parse(parentIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid parent asset id", err)
	}

	childIDStr, ok := vars["child_asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid child asset id", nil)
	}
	childID, err := uuid.Parse(childIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid child asset id", err)
	}

	return asset.UnlinkChildCommand{
		ParentID:  parentID,
		ChildID:   childID,
		ProjectID: projectID,
	}, nil
}

func makeGetAssetEndpoint(s asset.Service, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(asset.Filter)
		a, err := s.FindOne(ctx, filter)
		if err != nil {
			return nil, err
		}
		return toAssetDTO(a, rateProvider), nil
	}
}

func makeCreateAssetEndpoint(s asset.Service, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cmd := request.(asset.CreateAssetCommand)
		a, err := s.Create(ctx, cmd)
		if err != nil {
			return nil, err
		}
		return toAssetDTO(a, rateProvider), nil
	}
}

func makeCreateAssetFromPaymentsEndpoint(s asset.Service, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cmd := request.(asset.CreateAssetFromPaymentsCommand)
		a, err := s.CreateFromPayments(ctx, cmd)
		if err != nil {
			return nil, err
		}
		return toAssetDTO(a, rateProvider), nil
	}
}

func makeUpdateAssetEndpoint(s asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(asset.UpdateAssetCommand)
		err := s.Update(ctx, cmd)
		return nil, err
	}
}

func makeRemoveAssetEndpoint(svc asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command, ok := request.(projecta.RemoveProjectResourceCommand)
		if !ok {
			return nil, exceptions.NewValidationException("invalid request", nil)
		}
		err := svc.Remove(ctx, asset.RemoveAssetCommand{
			AssetID:   command.ResourceID,
			ProjectID: command.ProjectID,
		})
		return nil, err
	}
}

func makeListAssetsEndpoint(svc asset.Service, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(asset.CollectionFilter)
		collection, err := svc.Find(ctx, filter)
		if err != nil {
			return nil, err
		}

		var list []AssetDTO = make([]AssetDTO, 0)
		for _, e := range collection.Elements() {
			list = append(list, toAssetDTO(e, rateProvider))
		}

		return ListAssetsResponse{
			Assets: list,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
				Total:  collection.Total(),
			},
		}, err
	}
}

func makeLinkChildAssetEndpoint(svc asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(asset.LinkChildCommand)
		return nil, svc.LinkChild(ctx, cmd)
	}
}

func makeUnlinkChildAssetEndpoint(svc asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(asset.UnlinkChildCommand)
		return nil, svc.UnlinkChild(ctx, cmd)
	}
}

func decodeAssignInvestmentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetIDStr, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}
	assetID, err := uuid.Parse(assetIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var req AssignInvestmentsDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if len(req.InvestmentIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one investment must be selected", nil)
	}

	invUUIDs := make([]uuid.UUID, 0, len(req.InvestmentIDs))
	for _, idStr := range req.InvestmentIDs {
		u, err := uuid.Parse(idStr)
		if err != nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("invalid investment id: %s", idStr), err)
		}
		invUUIDs = append(invUUIDs, u)
	}

	return asset.AssignInvestmentsCommand{
		AssetID:       assetID,
		ProjectID:     projectID,
		InvestmentIDs: invUUIDs,
	}, nil
}

func makeAssignInvestmentsEndpoint(s asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(asset.AssignInvestmentsCommand)
		return nil, s.AssignInvestments(ctx, cmd)
	}
}

