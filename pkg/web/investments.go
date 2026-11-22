package web

import (
	"context"
	"net/http"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	investmentapp "gitlab.com/massimo-ua/projecta/internal/investment/app"
)

type InvestmentEndpoints struct {
	CreateInvestment       endpoint.Endpoint
	ListInvestments        endpoint.Endpoint
	GetInvestment          endpoint.Endpoint
	UpdateInvestment       endpoint.Endpoint
	RemoveInvestment       endpoint.Endpoint
	CreateBatchInvestments endpoint.Endpoint
	ListTags               endpoint.Endpoint
}

func MakeInvestmentEndpoints(invApp *investmentapp.Application) InvestmentEndpoints {
	if invApp == nil {
		return InvestmentEndpoints{}
	}
	return InvestmentEndpoints{
		CreateInvestment:       makeCreateInvestmentEndpoint(invApp),
		ListInvestments:        makeListInvestmentsEndpoint(invApp),
		GetInvestment:          makeGetInvestmentEndpoint(invApp),
		UpdateInvestment:       makeUpdateInvestmentEndpoint(invApp),
		RemoveInvestment:       makeRemoveInvestmentEndpoint(invApp),
		CreateBatchInvestments: makeCreateBatchInvestmentsEndpoint(invApp),
		ListTags:               makeListTagsEndpoint(invApp),
	}
}

type InvestmentAssetDTO struct {
	AssetID         string  `json:"asset_id"`
	AssetName       string  `json:"asset_name"`
	SharePercentage float64 `json:"share_percentage"`
	Amount          int64   `json:"amount,omitempty"`
}

type InvestmentDTO struct {
	InvestmentID     string               `json:"investment_id"`
	ProjectID        string               `json:"project_id"`
	AssetID          string               `json:"asset_id"`
	AssetName        string               `json:"asset_name"`
	AssetAllocations []InvestmentAssetDTO `json:"asset_allocations"`
	Contributor      OwnerDTO             `json:"contributor"`
	ResourceType     string               `json:"resource_type"`
	Amount           int64                `json:"amount"`
	Currency         string               `json:"currency"`
	HomeAmount       int64                `json:"home_amount,omitempty"`
	HomeCurrency     string               `json:"home_currency,omitempty"`
	TimeHours        *float64             `json:"time_hours,omitempty"`
	TimeHourlyRate   *int64               `json:"time_hourly_rate,omitempty"`
	GoodsQuantity    *float64             `json:"goods_quantity,omitempty"`
	GoodsUnit        string               `json:"goods_unit,omitempty"`
	GoodsItemName    string               `json:"goods_item_name,omitempty"`
	Description      string               `json:"description"`
	Date             string               `json:"date"`
	Tags             []string             `json:"tags"`
}

func toInvestmentDTO(inv *investment.Investment) InvestmentDTO {
	if inv == nil {
		return InvestmentDTO{}
	}

	var contribDTO OwnerDTO
	if !inv.Contributor().PersonID.IsNil() {
		contribDTO = OwnerDTO{
			PersonID:    inv.Contributor().PersonID.String(),
			DisplayName: inv.Contributor().DisplayName,
		}
	}

	var assetName string
	var assetID string
	if !inv.PrimaryAssetID().IsNil() {
		assetID = inv.PrimaryAssetID().String()
		assetName = inv.PrimaryAssetName()
	}

	var projectID string
	if !inv.ProjectID().IsNil() {
		projectID = inv.ProjectID().String()
	}

	var hourlyRateVal *int64
	if inv.TimeHourlyRate() != nil {
		v := inv.TimeHourlyRate().Amount()
		hourlyRateVal = &v
	}

	allocationsDTO := make([]InvestmentAssetDTO, 0, len(inv.AssetAllocations()))
	for _, a := range inv.AssetAllocations() {
		allocationsDTO = append(allocationsDTO, InvestmentAssetDTO{
			AssetID:         a.AssetID.String(),
			AssetName:       a.AssetName,
			SharePercentage: a.SharePercentage,
			Amount:          a.AllocatedAmount(inv.Amount()),
		})
	}

	var amountVal int64
	var currVal string
	if inv.Amount() != nil {
		amountVal = inv.Amount().Amount()
		if inv.Amount().Currency() != nil {
			currVal = inv.Amount().Currency().Code
		}
	}

	return InvestmentDTO{
		InvestmentID:     inv.ID().String(),
		ProjectID:        projectID,
		AssetID:          assetID,
		AssetName:        assetName,
		AssetAllocations: allocationsDTO,
		Contributor:      contribDTO,
		ResourceType:     inv.ResourceType().String(),
		Amount:           amountVal,
		Currency:         currVal,
		HomeAmount:       inv.HomeAmount(),
		HomeCurrency:     inv.HomeCurrency(),
		TimeHours:        inv.TimeHours(),
		TimeHourlyRate:   hourlyRateVal,
		GoodsQuantity:    inv.GoodsQuantity(),
		GoodsUnit:        inv.GoodsUnit(),
		GoodsItemName:    inv.GoodsItemName(),
		Description:      inv.Description(),
		Date:             inv.Date().Format(time.RFC3339),
		Tags:             inv.Tags(),
	}
}

type InvestmentAssetInputDTO struct {
	AssetID         string  `json:"asset_id"`
	SharePercentage float64 `json:"share_percentage"`
}

type CreateInvestmentDTO struct {
	AssetID          string                    `json:"asset_id,omitempty"`
	AssetAllocations []InvestmentAssetInputDTO `json:"asset_allocations,omitempty"`
	ResourceType     string                    `json:"resource_type"`
	Amount           int64                     `json:"amount"`
	Currency         string                    `json:"currency"`
	TimeHours        *float64                  `json:"time_hours,omitempty"`
	TimeHourlyRate   *int64                    `json:"time_hourly_rate,omitempty"`
	GoodsQuantity    *float64                  `json:"goods_quantity,omitempty"`
	GoodsUnit        string                    `json:"goods_unit,omitempty"`
	GoodsItemName    string                    `json:"goods_item_name,omitempty"`
	Description      string                    `json:"description"`
	Date             string                    `json:"date,omitempty"`
	Tags             []string                  `json:"tags,omitempty"`
}

type UpdateInvestmentDTO struct {
	AssetID          string                    `json:"asset_id,omitempty"`
	AssetAllocations []InvestmentAssetInputDTO `json:"asset_allocations,omitempty"`
	ResourceType     string                    `json:"resource_type,omitempty"`
	Amount           int64                     `json:"amount,omitempty"`
	Currency         string                    `json:"currency,omitempty"`
	TimeHours        *float64                  `json:"time_hours,omitempty"`
	TimeHourlyRate   *int64                    `json:"time_hourly_rate,omitempty"`
	GoodsQuantity    *float64                  `json:"goods_quantity,omitempty"`
	GoodsUnit        string                    `json:"goods_unit,omitempty"`
	GoodsItemName    string                    `json:"goods_item_name,omitempty"`
	Description      string                    `json:"description,omitempty"`
	Date             string                    `json:"date,omitempty"`
	Tags             []string                  `json:"tags,omitempty"`
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

func decodeCreateInvestmentRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateInvestmentDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	var assetID core.AssetID
	if req.AssetID != "" {
		var err error
		assetID, err = core.ParseAssetID(req.AssetID)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid asset id", err)
		}
	} else if len(req.AssetAllocations) == 0 {
		return nil, exceptions.NewValidationException("asset_id or asset_allocations required", nil)
	}

	var assetAllocations []investmentapp.AssetAllocationInput
	for _, a := range req.AssetAllocations {
		aID, err := core.ParseAssetID(a.AssetID)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid allocation asset id", err)
		}
		assetAllocations = append(assetAllocations, investmentapp.AssetAllocationInput{
			AssetID:         aID,
			SharePercentage: a.SharePercentage,
		})
	}

	var date time.Time
	if req.Date != "" {
		date, err = parseFlexibleDate(req.Date)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid date", err)
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

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return investmentapp.CreateInvestment{
		ID:               core.NewInvestmentID(),
		RequesterID:      requesterID,
		ProjectID:        projectID,
		AssetID:          assetID,
		AssetAllocations: assetAllocations,
		ResourceType:     resType,
		Amount:           amount,
		TimeHours:        req.TimeHours,
		TimeHourlyRate:   hourlyRate,
		GoodsQuantity:    req.GoodsQuantity,
		GoodsUnit:        req.GoodsUnit,
		GoodsItemName:    req.GoodsItemName,
		Description:      req.Description,
		Date:             date,
		Tags:             req.Tags,
	}, nil
}

func decodeUpdateInvestmentRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	invIDStr, ok := vars["investment_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid investment id", nil)
	}
	invID, err := core.ParseInvestmentID(invIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid investment id", err)
	}

	var req UpdateInvestmentDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	var assetID core.AssetID
	if req.AssetID != "" {
		assetID, err = core.ParseAssetID(req.AssetID)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid asset id", err)
		}
	}

	var assetAllocations []investmentapp.AssetAllocationInput
	for _, a := range req.AssetAllocations {
		aID, err := core.ParseAssetID(a.AssetID)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid allocation asset id", err)
		}
		assetAllocations = append(assetAllocations, investmentapp.AssetAllocationInput{
			AssetID:         aID,
			SharePercentage: a.SharePercentage,
		})
	}

	var date time.Time
	if req.Date != "" {
		date, err = parseFlexibleDate(req.Date)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid date", err)
		}
	}

	var resType investment.ResourceType
	if req.ResourceType != "" {
		resType, err = investment.ToResourceType(req.ResourceType)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid resource type", err)
		}
	}

	var amount *money.Money
	if req.Amount > 0 && req.Currency != "" {
		amount = money.New(req.Amount, req.Currency)
	}

	var hourlyRate *money.Money
	if req.TimeHourlyRate != nil && req.Currency != "" {
		hourlyRate = money.New(*req.TimeHourlyRate, req.Currency)
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return investmentapp.UpdateInvestment{
		ID:               invID,
		RequesterID:      requesterID,
		ProjectID:        projectID,
		AssetID:          assetID,
		AssetAllocations: assetAllocations,
		ResourceType:     resType,
		Amount:           amount,
		TimeHours:        req.TimeHours,
		TimeHourlyRate:   hourlyRate,
		GoodsQuantity:    req.GoodsQuantity,
		GoodsUnit:        req.GoodsUnit,
		GoodsItemName:    req.GoodsItemName,
		Description:      req.Description,
		Date:             date,
		Tags:             req.Tags,
	}, nil
}

func decodeListInvestmentsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	pagination, err := parsePagination(r, 50)
	if err != nil {
		return nil, err
	}

	var assetID core.AssetID
	if a := r.URL.Query().Get("asset_id"); a != "" {
		assetID, err = core.ParseAssetID(a)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid asset id", err)
		}
	}

	var resType investment.ResourceType
	if rt := r.URL.Query().Get("resource_type"); rt != "" {
		resType, err = investment.ToResourceType(rt)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid resource type", err)
		}
	}

	tag := r.URL.Query().Get("tag")

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return investmentapp.ListInvestments{
		Pagination:   pagination,
		RequesterID:  requesterID,
		ProjectID:    projectID,
		AssetID:      assetID,
		ResourceType: resType,
		Tag:          tag,
	}, nil
}

func decodeGetInvestmentRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	invIDStr, ok := vars["investment_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid investment id", nil)
	}
	invID, err := core.ParseInvestmentID(invIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid investment id", err)
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return investmentapp.GetInvestment{
		ID:          invID,
		ProjectID:   projectID,
		RequesterID: requesterID,
	}, nil
}

func decodeCreateBatchInvestmentsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateBatchInvestmentsDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	var defaultAsset core.AssetID
	if req.DefaultAsset != "" {
		defaultAsset, err = core.ParseAssetID(req.DefaultAsset)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid default asset id", err)
		}
	}

	var items []investmentapp.BatchCreateInvestmentItem
	for _, it := range req.Items {
		var aID core.AssetID
		if it.AssetID != "" {
			aID, err = core.ParseAssetID(it.AssetID)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid asset id in item", err)
			}
		}

		var date time.Time
		if it.Date != "" {
			date, err = parseFlexibleDate(it.Date)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid date in item", err)
			}
		}

		var resType investment.ResourceType
		if it.ResourceType != "" {
			resType, err = investment.ToResourceType(it.ResourceType)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid resource type in item", err)
			}
		}
		var amount *money.Money
		if it.Amount > 0 && it.Currency != "" {
			amount = money.New(it.Amount, it.Currency)
		}
		var hourlyRate *money.Money
		if it.TimeHourlyRate != nil && it.Currency != "" {
			hourlyRate = money.New(*it.TimeHourlyRate, it.Currency)
		}

		items = append(items, investmentapp.BatchCreateInvestmentItem{
			ID:             core.NewInvestmentID(),
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

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return investmentapp.CreateBatchInvestments{
		RequesterID:  requesterID,
		ProjectID:    projectID,
		DefaultAsset: defaultAsset,
		Items:        items,
	}, nil
}

func decodeListTagsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}
	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}
	return investmentapp.ListProjectTags{
		ProjectID:   projectID,
		RequesterID: requesterID,
	}, nil
}

func makeCreateInvestmentEndpoint(app *investmentapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(investmentapp.CreateInvestment)
		if cmd.ID.IsNil() {
			cmd.ID = core.NewInvestmentID()
		}
		if err := app.Commands.CreateInvestment.Handle(ctx, cmd); err != nil {
			return nil, err
		}
		inv, err := app.Queries.GetInvestment.Handle(ctx, investmentapp.GetInvestment{
			ID:          cmd.ID,
			ProjectID:   cmd.ProjectID,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			return nil, err
		}
		return toInvestmentDTO(inv), nil
	}
}

func makeUpdateInvestmentEndpoint(app *investmentapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(investmentapp.UpdateInvestment)
		return nil, app.Commands.UpdateInvestment.Handle(ctx, cmd)
	}
}

func decodeRemoveInvestmentRequest(ctx context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("missing project_id", nil)
	}
	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project_id", err)
	}
	investmentID, ok := vars["investment_id"]
	if !ok {
		return nil, exceptions.NewValidationException("missing investment_id", nil)
	}
	invUUID, err := core.ParseInvestmentID(investmentID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid investment_id", err)
	}
	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}
	return investmentapp.RemoveInvestment{
		ProjectID:   projectUUID,
		ID:          invUUID,
		RequesterID: requesterID,
	}, nil
}

func makeRemoveInvestmentEndpoint(app *investmentapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd, ok := request.(investmentapp.RemoveInvestment)
		if !ok {
			return nil, exceptions.NewValidationException("invalid request", nil)
		}
		return nil, app.Commands.RemoveInvestment.Handle(ctx, cmd)
	}
}

func makeGetInvestmentEndpoint(app *investmentapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(investmentapp.GetInvestment)
		inv, err := app.Queries.GetInvestment.Handle(ctx, query)
		if err != nil {
			return nil, err
		}
		return toInvestmentDTO(inv), nil
	}
}

func makeListInvestmentsEndpoint(app *investmentapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(investmentapp.ListInvestments)
		col, err := app.Queries.ListInvestments.Handle(ctx, query)
		if err != nil {
			return nil, err
		}

		dtos := make([]InvestmentDTO, 0)
		for _, e := range col.Elements() {
			dtos = append(dtos, toInvestmentDTO(e))
		}

		return ListInvestmentsResponse{
			Investments: dtos,
			PaginationDTO: PaginationDTO{
				Limit:  query.Limit,
				Offset: query.Offset,
				Total:  col.Total(),
			},
		}, nil
	}
}

func makeCreateBatchInvestmentsEndpoint(app *investmentapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(investmentapp.CreateBatchInvestments)
		ids := make([]core.InvestmentID, len(cmd.Items))
		for i := range cmd.Items {
			if cmd.Items[i].ID.IsNil() {
				cmd.Items[i].ID = core.NewInvestmentID()
			}
			ids[i] = cmd.Items[i].ID
		}

		if err := app.Commands.CreateBatchInvestments.Handle(ctx, cmd); err != nil {
			return nil, err
		}

		col, err := app.Queries.ListInvestments.Handle(ctx, investmentapp.ListInvestments{
			RequesterID: cmd.RequesterID,
			ProjectID:   cmd.ProjectID,
			IDs:         ids,
			Pagination:  core.Pagination{Limit: len(ids)},
		})
		if err != nil {
			return nil, err
		}

		byID := make(map[core.InvestmentID]*investment.Investment, len(col.Elements()))
		for _, inv := range col.Elements() {
			byID[inv.ID()] = inv
		}

		dtos := make([]InvestmentDTO, 0, len(ids))
		for _, id := range ids {
			if inv, ok := byID[id]; ok {
				dtos = append(dtos, toInvestmentDTO(inv))
			}
		}
		return dtos, nil
	}
}

func makeListTagsEndpoint(app *investmentapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(investmentapp.ListProjectTags)
		tags, err := app.Queries.ListProjectTags.Handle(ctx, query)
		if err != nil {
			return nil, err
		}
		if tags == nil {
			tags = make([]string, 0)
		}
		return tags, nil
	}
}
