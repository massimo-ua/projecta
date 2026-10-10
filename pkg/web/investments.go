package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

type InvestmentDTO struct {
	InvestmentID   string   `json:"investment_id"`
	ProjectID      string   `json:"project_id"`
	AssetID        string   `json:"asset_id"`
	AssetName      string   `json:"asset_name"`
	Contributor    OwnerDTO `json:"contributor"`
	ResourceType   string   `json:"resource_type"`
	Amount         int64    `json:"amount"`
	Currency       string   `json:"currency"`
	HomeAmount     int64    `json:"home_amount,omitempty"`
	HomeCurrency   string   `json:"home_currency,omitempty"`
	TimeHours      *float64 `json:"time_hours,omitempty"`
	TimeHourlyRate *int64   `json:"time_hourly_rate,omitempty"`
	GoodsQuantity  *float64 `json:"goods_quantity,omitempty"`
	GoodsUnit      string   `json:"goods_unit,omitempty"`
	GoodsItemName  string   `json:"goods_item_name,omitempty"`
	Description    string   `json:"description"`
	Date           string   `json:"date"`
	Tags           []string `json:"tags"`
}

func toInvestmentDTO(inv *investment.Investment, rateProvider currency.CurrencyRateProvider) InvestmentDTO {
	if inv == nil {
		return InvestmentDTO{}
	}

	homeCurrency := "UAH"
	if inv.Project != nil && inv.Project.MainCurrency != "" {
		homeCurrency = inv.Project.MainCurrency
	}

	homeAmount := inv.Amount.Amount()
	if rateProvider != nil && inv.Amount.Currency().Code != homeCurrency {
		converted, err := rateProvider.Convert(
			currency.NewCurrency(inv.Amount.Amount(), inv.Amount.Currency().Code),
			currency.NewCurrency(0, homeCurrency),
		)
		if err == nil {
			homeAmount = converted.Amount
		}
	}

	var contribDTO OwnerDTO
	if inv.Contributor != nil {
		contribDTO = OwnerDTO{
			PersonID:    inv.Contributor.PersonID.String(),
			DisplayName: inv.Contributor.DisplayName,
		}
	}

	var assetName string
	var assetID string
	if inv.Asset != nil {
		assetID = inv.Asset.ID().String()
		assetName = inv.Asset.Name()
	}

	var projectID string
	if inv.Project != nil {
		projectID = inv.Project.ProjectID.String()
	}

	var hourlyRateVal *int64
	if inv.TimeHourlyRate != nil {
		v := inv.TimeHourlyRate.Amount()
		hourlyRateVal = &v
	}

	return InvestmentDTO{
		InvestmentID:   inv.ID.String(),
		ProjectID:      projectID,
		AssetID:        assetID,
		AssetName:      assetName,
		Contributor:    contribDTO,
		ResourceType:   inv.ResourceType.String(),
		Amount:         inv.Amount.Amount(),
		Currency:       inv.Amount.Currency().Code,
		HomeAmount:     homeAmount,
		HomeCurrency:   homeCurrency,
		TimeHours:      inv.TimeHours,
		TimeHourlyRate: hourlyRateVal,
		GoodsQuantity:  inv.GoodsQuantity,
		GoodsUnit:      inv.GoodsUnit,
		GoodsItemName:  inv.GoodsItemName,
		Description:    inv.Description,
		Date:           inv.Date.Format(time.RFC3339),
		Tags:           inv.Tags,
	}
}

type CreateInvestmentDTO struct {
	AssetID        string   `json:"asset_id"`
	ResourceType   string   `json:"resource_type"`
	Amount         int64    `json:"amount"`
	Currency       string   `json:"currency"`
	TimeHours      *float64 `json:"time_hours,omitempty"`
	TimeHourlyRate *int64   `json:"time_hourly_rate,omitempty"`
	GoodsQuantity  *float64 `json:"goods_quantity,omitempty"`
	GoodsUnit      string   `json:"goods_unit,omitempty"`
	GoodsItemName  string   `json:"goods_item_name,omitempty"`
	Description    string   `json:"description"`
	Date           string   `json:"date,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type UpdateInvestmentDTO struct {
	AssetID        string   `json:"asset_id,omitempty"`
	ResourceType   string   `json:"resource_type,omitempty"`
	Amount         int64    `json:"amount,omitempty"`
	Currency       string   `json:"currency,omitempty"`
	TimeHours      *float64 `json:"time_hours,omitempty"`
	TimeHourlyRate *int64   `json:"time_hourly_rate,omitempty"`
	GoodsQuantity  *float64 `json:"goods_quantity,omitempty"`
	GoodsUnit      string   `json:"goods_unit,omitempty"`
	GoodsItemName  string   `json:"goods_item_name,omitempty"`
	Description    string   `json:"description,omitempty"`
	Date           string   `json:"date,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type BatchCreateInvestmentItemDTO struct {
	AssetID        string   `json:"asset_id,omitempty"`
	ResourceType   string   `json:"resource_type,omitempty"`
	Amount         int64    `json:"amount"`
	Currency       string   `json:"currency"`
	TimeHours      *float64 `json:"time_hours,omitempty"`
	TimeHourlyRate *int64   `json:"time_hourly_rate,omitempty"`
	GoodsQuantity  *float64 `json:"goods_quantity,omitempty"`
	GoodsUnit      string   `json:"goods_unit,omitempty"`
	GoodsItemName  string   `json:"goods_item_name,omitempty"`
	Description    string   `json:"description"`
	Date           string   `json:"date,omitempty"`
	Tags           []string `json:"tags,omitempty"`
}

type CreateBatchInvestmentsDTO struct {
	DefaultAsset string                         `json:"default_asset,omitempty"`
	Items        []BatchCreateInvestmentItemDTO `json:"items"`
}

type ListInvestmentsResponse struct {
	Investments []InvestmentDTO `json:"investments"`
	PaginationDTO
}

func decodeCreateInvestmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateInvestmentDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	assetID, err := uuid.Parse(req.AssetID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var date time.Time
	if req.Date != "" {
		date, _ = time.Parse(time.RFC3339, req.Date)
		if date.IsZero() {
			date, _ = time.Parse("2006-01-02", req.Date)
		}
	}

	resType, err := investment.ToResourceType(req.ResourceType)
	if err != nil && req.ResourceType != "" {
		return nil, err
	}

	var amount *money.Money
	if req.Amount > 0 && req.Currency != "" {
		amount = money.New(req.Amount, req.Currency)
	}

	var hourlyRate *money.Money
	if req.TimeHourlyRate != nil && req.Currency != "" {
		hourlyRate = money.New(*req.TimeHourlyRate, req.Currency)
	}

	return investment.CreateInvestmentCommand{
		ProjectID:      projectID,
		AssetID:        assetID,
		ResourceType:   resType,
		Amount:         amount,
		TimeHours:      req.TimeHours,
		TimeHourlyRate: hourlyRate,
		GoodsQuantity:  req.GoodsQuantity,
		GoodsUnit:      req.GoodsUnit,
		GoodsItemName:  req.GoodsItemName,
		Description:    req.Description,
		Date:           date,
		Tags:           req.Tags,
	}, nil
}

func decodeUpdateInvestmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	invIDStr, ok := vars["investment_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid investment id", nil)
	}
	invID, err := uuid.Parse(invIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid investment id", err)
	}

	var req UpdateInvestmentDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	var assetID uuid.UUID
	if req.AssetID != "" {
		assetID, _ = uuid.Parse(req.AssetID)
	}

	var date time.Time
	if req.Date != "" {
		date, _ = time.Parse(time.RFC3339, req.Date)
		if date.IsZero() {
			date, _ = time.Parse("2006-01-02", req.Date)
		}
	}

	var resType investment.ResourceType
	if req.ResourceType != "" {
		resType, _ = investment.ToResourceType(req.ResourceType)
	}

	var amount *money.Money
	if req.Amount > 0 && req.Currency != "" {
		amount = money.New(req.Amount, req.Currency)
	}

	var hourlyRate *money.Money
	if req.TimeHourlyRate != nil && req.Currency != "" {
		hourlyRate = money.New(*req.TimeHourlyRate, req.Currency)
	}

	return investment.UpdateInvestmentCommand{
		ID:             invID,
		ProjectID:      projectID,
		AssetID:        assetID,
		ResourceType:   resType,
		Amount:         amount,
		TimeHours:      req.TimeHours,
		TimeHourlyRate: hourlyRate,
		GoodsQuantity:  req.GoodsQuantity,
		GoodsUnit:      req.GoodsUnit,
		GoodsItemName:  req.GoodsItemName,
		Description:    req.Description,
		Date:           date,
		Tags:           req.Tags,
	}, nil
}

func decodeListInvestmentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	limit := 50
	offset := 0
	if l := r.URL.Query().Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	var assetID uuid.UUID
	if a := r.URL.Query().Get("asset_id"); a != "" {
		assetID, _ = uuid.Parse(a)
	}

	var resType investment.ResourceType
	if rt := r.URL.Query().Get("resource_type"); rt != "" {
		resType, _ = investment.ToResourceType(rt)
	}

	tag := r.URL.Query().Get("tag")

	return investment.CollectionFilter{
		Pagination: core.Pagination{
			Limit:  limit,
			Offset: offset,
		},
		ProjectID:    projectID,
		AssetID:      assetID,
		ResourceType: resType,
		Tag:          tag,
	}, nil
}

func decodeGetInvestmentRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	invIDStr, ok := vars["investment_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid investment id", nil)
	}
	invID, err := uuid.Parse(invIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid investment id", err)
	}

	return investment.Filter{
		ID:        invID,
		ProjectID: projectID,
	}, nil
}

func decodeCreateBatchInvestmentsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateBatchInvestmentsDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	var defaultAsset uuid.UUID
	if req.DefaultAsset != "" {
		defaultAsset, _ = uuid.Parse(req.DefaultAsset)
	}

	var items []investment.BatchCreateInvestmentItem
	for _, it := range req.Items {
		var aID uuid.UUID
		if it.AssetID != "" {
			aID, _ = uuid.Parse(it.AssetID)
		}

		var date time.Time
		if it.Date != "" {
			date, _ = time.Parse(time.RFC3339, it.Date)
			if date.IsZero() {
				date, _ = time.Parse("2006-01-02", it.Date)
			}
		}

		resType, _ := investment.ToResourceType(it.ResourceType)
		var amount *money.Money
		if it.Amount > 0 && it.Currency != "" {
			amount = money.New(it.Amount, it.Currency)
		}
		var hourlyRate *money.Money
		if it.TimeHourlyRate != nil && it.Currency != "" {
			hourlyRate = money.New(*it.TimeHourlyRate, it.Currency)
		}

		items = append(items, investment.BatchCreateInvestmentItem{
			AssetID:        aID,
			ResourceType:   resType,
			Amount:         amount,
			TimeHours:      it.TimeHours,
			TimeHourlyRate: hourlyRate,
			GoodsQuantity:  it.GoodsQuantity,
			GoodsUnit:      it.GoodsUnit,
			GoodsItemName:  it.GoodsItemName,
			Description:    it.Description,
			Date:           date,
			Tags:           it.Tags,
		})
	}

	return investment.CreateBatchInvestmentsCommand{
		ProjectID:    projectID,
		DefaultAsset: defaultAsset,
		Items:        items,
	}, nil
}

func decodeListTagsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}
	return projectID, nil
}

func makeCreateInvestmentEndpoint(svc investment.Service, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(investment.CreateInvestmentCommand)
		inv, err := svc.Create(ctx, cmd)
		if err != nil {
			return nil, err
		}
		return toInvestmentDTO(inv, rateProvider), nil
	}
}

func makeUpdateInvestmentEndpoint(svc investment.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(investment.UpdateInvestmentCommand)
		return nil, svc.Update(ctx, cmd)
	}
}

func makeRemoveInvestmentEndpoint(svc investment.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd, ok := request.(projecta.RemoveProjectResourceCommand)
		if !ok {
			return nil, exceptions.NewValidationException("invalid request", nil)
		}
		return nil, svc.Remove(ctx, investment.RemoveInvestmentCommand{
			ID:        cmd.ResourceID,
			ProjectID: cmd.ProjectID,
		})
	}
}

func makeGetInvestmentEndpoint(svc investment.Service, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(investment.Filter)
		inv, err := svc.FindOne(ctx, filter)
		if err != nil {
			return nil, err
		}
		return toInvestmentDTO(inv, rateProvider), nil
	}
}

func makeListInvestmentsEndpoint(svc investment.Service, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(investment.CollectionFilter)
		col, err := svc.Find(ctx, filter)
		if err != nil {
			return nil, err
		}

		dtos := make([]InvestmentDTO, 0)
		for _, e := range col.Elements() {
			dtos = append(dtos, toInvestmentDTO(e, rateProvider))
		}

		return ListInvestmentsResponse{
			Investments: dtos,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
				Total:  col.Total(),
			},
		}, nil
	}
}

func makeCreateBatchInvestmentsEndpoint(svc investment.Service, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(investment.CreateBatchInvestmentsCommand)
		created, err := svc.CreateBatch(ctx, cmd)
		if err != nil {
			return nil, err
		}
		dtos := make([]InvestmentDTO, 0)
		for _, inv := range created {
			dtos = append(dtos, toInvestmentDTO(inv, rateProvider))
		}
		return dtos, nil
	}
}

func makeListTagsEndpoint(svc investment.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		projectID := request.(uuid.UUID)
		tags, err := svc.FindTags(ctx, projectID)
		if err != nil {
			return nil, err
		}
		if tags == nil {
			tags = make([]string, 0)
		}
		return tags, nil
	}
}
