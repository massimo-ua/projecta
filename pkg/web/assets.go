package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	assetapp "gitlab.com/massimo-ua/projecta/internal/asset/app"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type AssetEndpoints struct {
	CreateAsset             endpoint.Endpoint
	CreateAssetFromPayments endpoint.Endpoint
	RemoveAsset             endpoint.Endpoint
	ListAssets              endpoint.Endpoint
	UpdateAsset             endpoint.Endpoint
	GetAsset                endpoint.Endpoint
	LinkChildAsset          endpoint.Endpoint
	UnlinkChildAsset        endpoint.Endpoint
	AssignInvestments       endpoint.Endpoint
	GroupAssets             endpoint.Endpoint
	LinkChildren            endpoint.Endpoint
	UpdateSchedule          endpoint.Endpoint
	GetCashFlowForecast     endpoint.Endpoint
}

func MakeAssetEndpoints(assetApp *assetapp.Application) AssetEndpoints {
	if assetApp == nil {
		return AssetEndpoints{}
	}
	return AssetEndpoints{
		CreateAsset:             makeCreateAssetEndpoint(assetApp),
		CreateAssetFromPayments: makeCreateAssetFromPaymentsEndpoint(assetApp),
		RemoveAsset:             makeRemoveAssetEndpoint(assetApp),
		ListAssets:              makeListAssetsEndpoint(assetApp),
		UpdateAsset:             makeUpdateAssetEndpoint(assetApp),
		GetAsset:                makeGetAssetEndpoint(assetApp),
		LinkChildAsset:          makeLinkChildAssetEndpoint(assetApp),
		UnlinkChildAsset:        makeUnlinkChildAssetEndpoint(assetApp),
		AssignInvestments:       makeAssignInvestmentsEndpoint(assetApp),
		GroupAssets:             makeGroupAssetsEndpoint(assetApp),
		LinkChildren:            makeLinkChildrenEndpoint(assetApp),
		UpdateSchedule:          makeUpdateScheduleEndpoint(assetApp),
		GetCashFlowForecast:     makeGetCashFlowForecastEndpoint(assetApp),
	}
}

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
	AssetID              string           `json:"asset_id"`
	Name                 string           `json:"name"`
	Description          string           `json:"description"`
	Status               string           `json:"status"`
	StartDate            string           `json:"start_date"`
	CompletedDate        *string          `json:"completed_date,omitempty"`
	TargetCompletionDate *string          `json:"target_completion_date,omitempty"`
	TargetPrice          *int64           `json:"target_price,omitempty"`
	TargetCurrency       string           `json:"target_currency,omitempty"`
	TargetHomeAmount     *int64           `json:"target_home_amount,omitempty"`
	ProgressPercentage   *float64         `json:"progress_percentage,omitempty"`
	DirectCost           int64            `json:"direct_cost"`
	DirectCostHomeAmount int64            `json:"direct_cost_home_amount"`
	TotalCost            int64            `json:"total_cost"`
	Price                int64            `json:"price"` // backward compatibility
	Currency             string           `json:"currency"`
	HomeAmount           int64            `json:"home_amount,omitempty"`
	HomeCurrency         string           `json:"home_currency,omitempty"`
	AcquiredAt           string           `json:"acquired_at"`
	Owner                OwnerDTO         `json:"owner"`
	Project              ProjectDTO       `json:"project"`
	Children             []ChildAssetDTO  `json:"children"`
	Parents              []ParentAssetDTO `json:"parents"`
	Tags                 []string         `json:"tags"`
}

func toAssetDTO(a *asset.Asset) AssetDTO {
	if a == nil {
		return AssetDTO{}
	}
	homeCurrency := a.HomeCurrency()
	projDTO := ProjectDTO{
		ProjectID:    a.ProjectID().String(),
		MainCurrency: homeCurrency,
	}

	priceAmount := int64(0)
	priceCurrency := homeCurrency
	if a.Price() != nil {
		priceAmount = a.Price().Amount()
		priceCurrency = a.Price().Currency().Code
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

	if a.TargetPrice() != nil && a.TargetPrice().Amount() > 0 {
		v := a.TargetPrice().Amount()
		targetPriceVal = &v
		targetCurrVal = a.TargetPrice().Currency().Code
	}

	var completedDateStr *string
	if a.CompletedDate() != nil {
		s := a.CompletedDate().Format(time.RFC3339)
		completedDateStr = &s
	}

	var targetCompletionDateStr *string
	if a.TargetCompletionDate() != nil {
		s := a.TargetCompletionDate().Format(time.RFC3339)
		targetCompletionDateStr = &s
	}

	owner := OwnerDTO{
		PersonID:    a.Owner().PersonID.String(),
		DisplayName: a.Owner().DisplayName,
	}

	childrenDTO := make([]ChildAssetDTO, 0, len(a.Children()))
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

	parentsDTO := make([]ParentAssetDTO, 0, len(a.Parents()))
	for _, p := range a.Parents() {
		parentsDTO = append(parentsDTO, ParentAssetDTO{
			ParentID:        p.ParentID.String(),
			ParentName:      p.ParentName,
			SharePercentage: p.SharePercentage,
		})
	}

	return AssetDTO{
		AssetID:              a.ID().String(),
		Name:                 a.Name(),
		Description:          a.Description(),
		Status:               a.Status().String(),
		StartDate:            a.StartDate().Format(time.RFC3339),
		CompletedDate:        completedDateStr,
		TargetCompletionDate: targetCompletionDateStr,
		TargetPrice:          targetPriceVal,
		TargetCurrency:       targetCurrVal,
		TargetHomeAmount:     a.TargetHomeAmount(),
		ProgressPercentage:   a.ProgressPercentage(),
		DirectCost:           directCostVal,
		DirectCostHomeAmount: a.DirectCostHomeAmount(),
		TotalCost:            totalCostVal,
		Price:                totalCostVal,
		Currency:             priceCurrency,
		HomeAmount:           a.HomeAmount(),
		HomeCurrency:         homeCurrency,
		AcquiredAt:           a.AcquiredAt().Format(time.RFC3339),
		Owner:                owner,
		Project:              projDTO,
		Children:             childrenDTO,
		Parents:              parentsDTO,
		Tags:                 a.Tags(),
	}
}

type CreateAssetDTO struct {
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	TypeID               string   `json:"type_id,omitempty"`
	Price                int64    `json:"price,omitempty"`
	Currency             string   `json:"currency,omitempty"`
	TargetPrice          int64    `json:"target_price,omitempty"`
	TargetCurrency       string   `json:"target_currency,omitempty"`
	Status               string   `json:"status,omitempty"`
	StartDate            string   `json:"start_date,omitempty"`
	CompletedDate        string   `json:"completed_date,omitempty"`
	TargetCompletionDate string   `json:"target_completion_date,omitempty"`
	AcquiredAt           string   `json:"acquired_at,omitempty"`
	WithPayment          bool     `json:"with_payment"`
	Tags                 []string `json:"tags,omitempty"`
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
	Name                 string   `json:"name"`
	Description          string   `json:"description"`
	TypeID               string   `json:"type_id,omitempty"`
	Price                int64    `json:"price,omitempty"`
	Currency             string   `json:"currency,omitempty"`
	TargetPrice          int64    `json:"target_price,omitempty"`
	TargetCurrency       string   `json:"target_currency,omitempty"`
	Status               string   `json:"status,omitempty"`
	StartDate            string   `json:"start_date,omitempty"`
	CompletedDate        string   `json:"completed_date,omitempty"`
	TargetCompletionDate string   `json:"target_completion_date,omitempty"`
	ClearTargetDate      bool     `json:"clear_target_date,omitempty"`
	AcquiredAt           string   `json:"acquired_at,omitempty"`
	Tags                 []string `json:"tags,omitempty"`
}

type LinkChildAssetDTO struct {
	ChildAssetID    string   `json:"child_asset_id"`
	SharePercentage *float64 `json:"share_percentage,omitempty"`
}

type AssignInvestmentsDTO struct {
	InvestmentIDs []string `json:"investment_ids"`
}

type ListAssetsResponse struct {
	Assets []AssetDTO `json:"assets"`
	PaginationDTO
}

func decodeCreateAssetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateAssetDTO
	err = decodeJSONBody(r, &req)
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

	price, targetPrice, err := decodeAssetPriceInputs(req.Price, req.TargetPrice, req.Currency, req.TargetCurrency, req.WithPayment)
	if err != nil {
		return nil, err
	}

	var date time.Time
	if req.AcquiredAt != "" {
		parsed, err := parseFlexibleDate(req.AcquiredAt)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid acquired at date", err)
		}
		date = parsed
	} else if req.StartDate != "" {
		parsed, err := parseFlexibleDate(req.StartDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid acquired at date", err)
		}
		date = parsed
	}
	if date.IsZero() {
		date = time.Now()
	}

	var completedDate *time.Time
	if req.CompletedDate != "" {
		cd, err := parseFlexibleDate(req.CompletedDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid completed date", err)
		}
		completedDate = &cd
	}

	var targetCompletionDate *time.Time
	if req.TargetCompletionDate != "" {
		td, err := parseFlexibleDate(req.TargetCompletionDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid target completion date", err)
		}
		targetCompletionDate = &td
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.CreateAsset{
		AssetID:              core.NewAssetID(),
		RequesterID:          requesterID,
		Name:                 req.Name,
		Description:          req.Description,
		ProjectID:            projectUUID,
		TypeID:               typeUUID,
		Price:                price,
		TargetPrice:          targetPrice,
		Status:               asset.ToAssetStatus(req.Status),
		StartDate:            date,
		CompletedDate:        completedDate,
		TargetCompletionDate: targetCompletionDate,
		AcquiredAt:           date,
		WithPayment:          req.WithPayment,
		Tags:                 req.Tags,
	}, nil
}

func decodeCreateAssetFromPaymentsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateAssetFromPaymentsDTO
	err = decodeJSONBody(r, &req)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if req.Name == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	if len(req.PaymentIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one payment must be selected", nil)
	}

	var paymentUUIDs []core.InvestmentID
	for _, pid := range req.PaymentIDs {
		pUUID, err := core.ParseInvestmentID(pid)
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

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.CreateAssetFromInvestments{
		AssetID:        core.NewAssetID(),
		RequesterID:    requesterID,
		ProjectID:      projectUUID,
		InvestmentIDs:  paymentUUIDs,
		Name:           req.Name,
		Description:    req.Description,
		TypeID:         typeUUID,
		AcquiredAt:     date,
		TargetCurrency: req.TargetCurrency,
	}, nil
}

func decodeGetAssetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetID, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}

	assetUUID, err := core.ParseAssetID(assetID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.GetAsset{
		ID:          assetUUID,
		ProjectID:   projectUUID,
		RequesterID: requesterID,
	}, nil
}

func decodeUpdateAssetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetID, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}

	assetUUID, err := core.ParseAssetID(assetID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var req UpdateAssetDTO
	err = decodeJSONBody(r, &req)
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

	price, targetPrice, err := decodeAssetPriceInputs(req.Price, req.TargetPrice, req.Currency, req.TargetCurrency, false)
	if err != nil {
		return nil, err
	}

	var date time.Time
	if req.AcquiredAt != "" {
		parsed, err := parseFlexibleDate(req.AcquiredAt)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid acquired at date", err)
		}
		date = parsed
	} else if req.StartDate != "" {
		parsed, err := parseFlexibleDate(req.StartDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid acquired at date", err)
		}
		date = parsed
	}

	var completedDate *time.Time
	if req.CompletedDate != "" {
		cd, err := parseFlexibleDate(req.CompletedDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid completed date", err)
		}
		completedDate = &cd
	}

	var targetCompletionDate *time.Time
	if req.TargetCompletionDate != "" {
		td, err := parseFlexibleDate(req.TargetCompletionDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid target completion date", err)
		}
		targetCompletionDate = &td
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.UpdateAsset{
		AssetID:              assetUUID,
		RequesterID:          requesterID,
		Name:                 req.Name,
		Description:          req.Description,
		ProjectID:            projectUUID,
		TypeID:               typeUUID,
		Price:                price,
		TargetPrice:          targetPrice,
		Status:               asset.ToAssetStatus(req.Status),
		StartDate:            date,
		CompletedDate:        completedDate,
		TargetCompletionDate: targetCompletionDate,
		ClearTargetDate:      req.ClearTargetDate,
		AcquiredAt:           date,
		Tags:                 req.Tags,
	}, nil
}

func decodeListAssetsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	pagination, err := parsePagination(r, core.DefaultLimit)
	if err != nil {
		return nil, err
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

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.ListAssets{
		Pagination:  pagination,
		RequesterID: requesterID,
		ProjectID:   projectUUID,
		TypeID:      typeID,
		Name:        name,
	}, nil
}

func decodeLinkChildAssetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetIDStr, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}
	parentID, err := core.ParseAssetID(assetIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var req LinkChildAssetDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	childID, err := core.ParseAssetID(req.ChildAssetID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid child asset id", err)
	}

	share := 100.0
	if req.SharePercentage != nil {
		share = *req.SharePercentage
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.LinkChildAsset{
		RequesterID:     requesterID,
		ParentID:        parentID,
		ChildID:         childID,
		ProjectID:       projectID,
		SharePercentage: share,
	}, nil
}

func decodeUnlinkChildAssetRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	parentIDStr, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid parent asset id", nil)
	}
	parentID, err := core.ParseAssetID(parentIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid parent asset id", err)
	}

	childIDStr, ok := vars["child_asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid child asset id", nil)
	}
	childID, err := core.ParseAssetID(childIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid child asset id", err)
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.UnlinkChildAsset{
		RequesterID: requesterID,
		ParentID:    parentID,
		ChildID:     childID,
		ProjectID:   projectID,
	}, nil
}

func makeGetAssetEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(assetapp.GetAsset)
		a, err := app.Queries.GetAsset.Handle(ctx, query)
		if err != nil {
			return nil, err
		}
		return toAssetDTO(a), nil
	}
}

func makeCreateAssetEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cmd := request.(assetapp.CreateAsset)
		if cmd.AssetID.IsNil() {
			cmd.AssetID = core.NewAssetID()
		}
		if err := app.Commands.CreateAsset.Handle(ctx, cmd); err != nil {
			return nil, err
		}
		a, err := app.Queries.GetAsset.Handle(ctx, assetapp.GetAsset{
			ID:          cmd.AssetID,
			ProjectID:   cmd.ProjectID,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			return nil, err
		}
		return toAssetDTO(a), nil
	}
}

func makeCreateAssetFromPaymentsEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cmd := request.(assetapp.CreateAssetFromInvestments)
		if cmd.AssetID.IsNil() {
			cmd.AssetID = core.NewAssetID()
		}
		if err := app.Commands.CreateAssetFromInvestments.Handle(ctx, cmd); err != nil {
			return nil, err
		}
		a, err := app.Queries.GetAsset.Handle(ctx, assetapp.GetAsset{
			ID:          cmd.AssetID,
			ProjectID:   cmd.ProjectID,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			return nil, err
		}
		return toAssetDTO(a), nil
	}
}

func makeUpdateAssetEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(assetapp.UpdateAsset)
		return nil, app.Commands.UpdateAsset.Handle(ctx, cmd)
	}
}

func decodeRemoveAssetRequest(ctx context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("missing project_id", nil)
	}
	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project_id", err)
	}
	assetID, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("missing asset_id", nil)
	}
	assetUUID, err := core.ParseAssetID(assetID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset_id", err)
	}
	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}
	return assetapp.RemoveAsset{
		ProjectID:   projectUUID,
		AssetID:     assetUUID,
		RequesterID: requesterID,
	}, nil
}

func makeRemoveAssetEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd, ok := request.(assetapp.RemoveAsset)
		if !ok {
			return nil, exceptions.NewValidationException("invalid request", nil)
		}
		return nil, app.Commands.RemoveAsset.Handle(ctx, cmd)
	}
}

func makeListAssetsEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(assetapp.ListAssets)
		collection, err := app.Queries.ListAssets.Handle(ctx, query)
		if err != nil {
			return nil, err
		}

		list := make([]AssetDTO, 0, len(collection.Elements()))
		for _, e := range collection.Elements() {
			list = append(list, toAssetDTO(e))
		}

		return ListAssetsResponse{
			Assets: list,
			PaginationDTO: PaginationDTO{
				Limit:  query.Limit,
				Offset: query.Offset,
				Total:  collection.Total(),
			},
		}, nil
	}
}

func makeLinkChildAssetEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(assetapp.LinkChildAsset)
		return nil, app.Commands.LinkChildAsset.Handle(ctx, cmd)
	}
}

func makeUnlinkChildAssetEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(assetapp.UnlinkChildAsset)
		return nil, app.Commands.UnlinkChildAsset.Handle(ctx, cmd)
	}
}

func decodeAssignInvestmentsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetIDStr, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}
	assetID, err := core.ParseAssetID(assetIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var req AssignInvestmentsDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if len(req.InvestmentIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one investment must be selected", nil)
	}

	invUUIDs := make([]core.InvestmentID, 0, len(req.InvestmentIDs))
	for _, idStr := range req.InvestmentIDs {
		u, err := core.ParseInvestmentID(idStr)
		if err != nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("invalid investment id: %s", idStr), err)
		}
		invUUIDs = append(invUUIDs, u)
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.AssignInvestments{
		RequesterID:   requesterID,
		AssetID:       assetID,
		ProjectID:     projectID,
		InvestmentIDs: invUUIDs,
	}, nil
}

func makeAssignInvestmentsEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(assetapp.AssignInvestments)
		return nil, app.Commands.AssignInvestments.Handle(ctx, cmd)
	}
}

type GroupAssetsDTO struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	ChildAssetIDs []string `json:"child_asset_ids"`
	Tags          []string `json:"tags"`
}

func decodeGroupAssetsRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req GroupAssetsDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if req.Name == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}
	if len(req.ChildAssetIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one asset must be selected to group", nil)
	}

	childUUIDs := make([]core.AssetID, 0, len(req.ChildAssetIDs))
	for _, idStr := range req.ChildAssetIDs {
		u, err := core.ParseAssetID(idStr)
		if err != nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("invalid child asset id: %s", idStr), err)
		}
		childUUIDs = append(childUUIDs, u)
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.GroupAssets{
		GroupAssetID:  core.NewAssetID(),
		RequesterID:   requesterID,
		ProjectID:     projectID,
		Name:          req.Name,
		Description:   req.Description,
		ChildAssetIDs: childUUIDs,
		Tags:          req.Tags,
	}, nil
}

func makeGroupAssetsEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(assetapp.GroupAssets)
		if cmd.GroupAssetID.IsNil() {
			cmd.GroupAssetID = core.NewAssetID()
		}
		if err := app.Commands.GroupAssets.Handle(ctx, cmd); err != nil {
			return nil, err
		}
		a, err := app.Queries.GetAsset.Handle(ctx, assetapp.GetAsset{
			ID:          cmd.GroupAssetID,
			ProjectID:   cmd.ProjectID,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			return nil, err
		}
		return toAssetDTO(a), nil
	}
}

type LinkChildrenDTO struct {
	ChildAssetIDs []string `json:"child_asset_ids"`
}

func decodeLinkChildrenRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectIDStr, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}
	projectID, err := core.ParseProjectID(projectIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	assetIDStr, ok := vars["asset_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid asset id", nil)
	}
	parentID, err := core.ParseAssetID(assetIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var req LinkChildrenDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if len(req.ChildAssetIDs) == 0 {
		return nil, exceptions.NewValidationException("child asset ids are required", nil)
	}

	childUUIDs := make([]core.AssetID, 0, len(req.ChildAssetIDs))
	for _, idStr := range req.ChildAssetIDs {
		u, err := core.ParseAssetID(idStr)
		if err != nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("invalid child asset id: %s", idStr), err)
		}
		childUUIDs = append(childUUIDs, u)
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.LinkChildrenAssets{
		RequesterID:   requesterID,
		ProjectID:     projectID,
		ParentID:      parentID,
		ChildAssetIDs: childUUIDs,
	}, nil
}

func makeLinkChildrenEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(assetapp.LinkChildrenAssets)
		return nil, app.Commands.LinkChildrenAssets.Handle(ctx, cmd)
	}
}

func parseFlexibleDate(raw string) (time.Time, error) {
	if t, err := time.Parse(time.RFC3339, raw); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", raw)
}

func decodeAssetPriceInputs(priceVal, targetPriceVal int64, curr, targetCurr string, requirePositive bool) (*money.Money, *money.Money, error) {
	if priceVal < 0 || targetPriceVal < 0 {
		return nil, nil, exceptions.NewValidationException("price must be greater than 0", nil)
	}
	if requirePositive && priceVal <= 0 && targetPriceVal <= 0 {
		return nil, nil, exceptions.NewValidationException("price must be greater than 0", nil)
	}
	if (priceVal > 0 && curr == "") || (targetPriceVal > 0 && targetCurr == "") || (requirePositive && curr == "" && targetCurr == "") {
		return nil, nil, exceptions.NewValidationException("currency is required", nil)
	}

	var price *money.Money
	if priceVal > 0 && curr != "" {
		price = money.New(priceVal, curr)
	}
	var targetPrice *money.Money
	if targetPriceVal > 0 && targetCurr != "" {
		targetPrice = money.New(targetPriceVal, targetCurr)
	}
	return price, targetPrice, nil
}

type RescheduleAssetDTO struct {
	StartDate            string  `json:"start_date,omitempty"`
	TargetCompletionDate *string `json:"target_completion_date,omitempty"`
	ClearTargetDate      bool    `json:"clear_target_date,omitempty"`
}

type RescheduleAssetResponseDTO struct {
	Asset    AssetDTO `json:"asset"`
	Warnings []string `json:"warnings"`
}

func decodeUpdateScheduleRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, err := core.ParseProjectID(vars["project_id"])
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}
	assetID, err := core.ParseAssetID(vars["asset_id"])
	if err != nil {
		return nil, exceptions.NewValidationException("invalid asset id", err)
	}

	var req RescheduleAssetDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	var startDate *time.Time
	if req.StartDate != "" {
		parsed, err := parseFlexibleDate(req.StartDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid start date", err)
		}
		startDate = &parsed
	}

	clearTarget := req.ClearTargetDate
	var targetDate *time.Time
	if req.TargetCompletionDate != nil {
		if *req.TargetCompletionDate == "" {
			clearTarget = true
		} else {
			parsed, err := parseFlexibleDate(*req.TargetCompletionDate)
			if err != nil {
				return nil, exceptions.NewValidationException("invalid target completion date", err)
			}
			targetDate = &parsed
		}
	}

	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return assetapp.RescheduleAsset{
		RequesterID:          requesterID,
		ProjectID:            projectID,
		AssetID:              assetID,
		StartDate:            startDate,
		TargetCompletionDate: targetDate,
		ClearTargetDate:      clearTarget,
	}, nil
}

func makeUpdateScheduleEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(assetapp.RescheduleAsset)
		if err := app.Commands.RescheduleAsset.Handle(ctx, cmd); err != nil {
			return nil, err
		}
		view, err := app.Queries.GetAssetScheduleView.Handle(ctx, assetapp.GetAssetScheduleView{
			AssetID:     cmd.AssetID,
			ProjectID:   cmd.ProjectID,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			return nil, err
		}
		warnings := view.Warnings
		if warnings == nil {
			warnings = make([]string, 0)
		}
		return RescheduleAssetResponseDTO{
			Asset:    toAssetDTO(view.Asset),
			Warnings: warnings,
		}, nil
	}
}

type ResourceBreakdownDTO struct {
	CashCents  int64   `json:"cash_cents"`
	TimeHours  float64 `json:"time_hours"`
	GoodsCents int64   `json:"goods_cents"`
}

type MonthlyBurnRateDTO struct {
	CashCentsPerMonth  int64   `json:"cash_cents_per_month"`
	TimeHoursPerMonth  float64 `json:"time_hours_per_month"`
	GoodsCentsPerMonth int64   `json:"goods_cents_per_month"`
}

type AssetDemandItemDTO struct {
	AssetID     string `json:"asset_id"`
	AssetName   string `json:"asset_name"`
	DemandCents int64  `json:"demand_cents"`
	IsOverdue   bool   `json:"is_overdue"`
}

type MonthCashFlowDTO struct {
	Month                     string               `json:"month"`
	TagBudgetCeilingCents     int64                `json:"tag_budget_ceiling_cents"`
	Actual                    ResourceBreakdownDTO `json:"actual"`
	ScheduledDemandCents      int64                `json:"scheduled_demand_cents"`
	EstimatedOngoingCashCents int64                `json:"estimated_ongoing_cash_cents"`
	CommittedCashCents        int64                `json:"committed_cash_cents"`
	CashUtilizationPct        float64              `json:"cash_utilization_pct"`
	AlertStatus               string               `json:"alert_status"`
	AssetDemands              []AssetDemandItemDTO `json:"asset_demands"`
}

type AssetForecastItemDTO struct {
	AssetID                string               `json:"asset_id"`
	AssetName              string               `json:"asset_name"`
	Status                 string               `json:"status"`
	StartDate              string               `json:"start_date"`
	TargetCompletionDate   *string              `json:"target_completion_date,omitempty"`
	CompletedDate          *string              `json:"completed_date,omitempty"`
	HasTargetBudget        bool                 `json:"has_target_budget"`
	TargetHomeCents        int64                `json:"target_home_cents"`
	ActualToDate           ResourceBreakdownDTO `json:"actual_to_date"`
	RolledUpTotalCashCents int64                `json:"rolled_up_total_cash_cents"`
	ResidualTargetCents    int64                `json:"residual_target_cents"`
	BurnRateMonthly        MonthlyBurnRateDTO   `json:"burn_rate_monthly"`
	ForecastMode           string               `json:"forecast_mode"`
	EstimatedCompletion    *string              `json:"estimated_completion,omitempty"`
}

type CashFlowSummaryDTO struct {
	TotalTargetCents                 int64   `json:"total_target_cents"`
	TotalActualCashCents             int64   `json:"total_actual_cash_cents"`
	TotalRemainingScheduledCents     int64   `json:"total_remaining_scheduled_cents"`
	TotalTimeHours                   float64 `json:"total_time_hours"`
	TotalGoodsCents                  int64   `json:"total_goods_cents"`
	EstimatedOngoingMonthlyCashCents int64   `json:"estimated_ongoing_monthly_cash_cents"`
	MonthlyBudgetCeilingAvgCents     int64   `json:"monthly_budget_ceiling_avg_cents"`
	PeakMonth                        string  `json:"peak_month"`
	PeakCommittedCashCents           int64   `json:"peak_committed_cash_cents"`
	AlertMonthsCount                 int     `json:"alert_months_count"`
}

type CashFlowForecastDTO struct {
	ProjectID            string                 `json:"project_id"`
	Currency             string                 `json:"currency"`
	FromMonth            string                 `json:"from_month"`
	ToMonth              string                 `json:"to_month"`
	TrailingWindowMonths int                    `json:"trailing_window_months"`
	Summary              CashFlowSummaryDTO     `json:"summary"`
	Months               []MonthCashFlowDTO     `json:"months"`
	Assets               []AssetForecastItemDTO `json:"assets"`
}

func toCashFlowForecastDTO(fc *asset.CashFlowForecast) CashFlowForecastDTO {
	if fc == nil {
		return CashFlowForecastDTO{
			Months: make([]MonthCashFlowDTO, 0),
			Assets: make([]AssetForecastItemDTO, 0),
		}
	}

	months := make([]MonthCashFlowDTO, 0, len(fc.Months))
	for _, m := range fc.Months {
		demands := make([]AssetDemandItemDTO, 0, len(m.AssetDemands))
		for _, d := range m.AssetDemands {
			demands = append(demands, AssetDemandItemDTO{
				AssetID:     d.AssetID.String(),
				AssetName:   d.AssetName,
				DemandCents: d.DemandCents,
				IsOverdue:   d.IsOverdue,
			})
		}
		months = append(months, MonthCashFlowDTO{
			Month:                 m.Month,
			TagBudgetCeilingCents: m.TagBudgetCeilingCents,
			Actual: ResourceBreakdownDTO{
				CashCents:  m.Actual.CashCents,
				TimeHours:  m.Actual.TimeHours,
				GoodsCents: m.Actual.GoodsCents,
			},
			ScheduledDemandCents:      m.ScheduledDemandCents,
			EstimatedOngoingCashCents: m.EstimatedOngoingCashCents,
			CommittedCashCents:        m.CommittedCashCents,
			CashUtilizationPct:        m.CashUtilizationPct,
			AlertStatus:               m.AlertStatus,
			AssetDemands:              demands,
		})
	}

	assets := make([]AssetForecastItemDTO, 0, len(fc.Assets))
	for _, a := range fc.Assets {
		var tcdStr *string
		if a.TargetCompletionDate != nil {
			s := a.TargetCompletionDate.Format(time.RFC3339)
			tcdStr = &s
		}
		var cdStr *string
		if a.CompletedDate != nil {
			s := a.CompletedDate.Format(time.RFC3339)
			cdStr = &s
		}
		var estStr *string
		if a.EstimatedCompletion != nil {
			s := a.EstimatedCompletion.Format("2006-01")
			estStr = &s
		}
		assets = append(assets, AssetForecastItemDTO{
			AssetID:              a.AssetID.String(),
			AssetName:            a.AssetName,
			Status:               a.Status.String(),
			StartDate:            a.StartDate.Format(time.RFC3339),
			TargetCompletionDate: tcdStr,
			CompletedDate:        cdStr,
			HasTargetBudget:      a.HasTargetBudget,
			TargetHomeCents:      a.TargetHomeCents,
			ActualToDate: ResourceBreakdownDTO{
				CashCents:  a.ActualToDate.CashCents,
				TimeHours:  a.ActualToDate.TimeHours,
				GoodsCents: a.ActualToDate.GoodsCents,
			},
			RolledUpTotalCashCents: a.RolledUpTotalCashCents,
			ResidualTargetCents:    a.ResidualTargetCents,
			BurnRateMonthly: MonthlyBurnRateDTO{
				CashCentsPerMonth:  a.BurnRateMonthly.CashCentsPerMonth,
				TimeHoursPerMonth:  a.BurnRateMonthly.TimeHoursPerMonth,
				GoodsCentsPerMonth: a.BurnRateMonthly.GoodsCentsPerMonth,
			},
			ForecastMode:        a.ForecastMode,
			EstimatedCompletion: estStr,
		})
	}

	return CashFlowForecastDTO{
		ProjectID:            fc.ProjectID.String(),
		Currency:             fc.Currency,
		FromMonth:            fc.FromMonth,
		ToMonth:              fc.ToMonth,
		TrailingWindowMonths: fc.TrailingWindowMonths,
		Summary: CashFlowSummaryDTO{
			TotalTargetCents:                 fc.Summary.TotalTargetCents,
			TotalActualCashCents:             fc.Summary.TotalActualCashCents,
			TotalRemainingScheduledCents:     fc.Summary.TotalRemainingScheduledCents,
			TotalTimeHours:                   fc.Summary.TotalTimeHours,
			TotalGoodsCents:                  fc.Summary.TotalGoodsCents,
			EstimatedOngoingMonthlyCashCents: fc.Summary.EstimatedOngoingMonthlyCashCents,
			MonthlyBudgetCeilingAvgCents:     fc.Summary.MonthlyBudgetCeilingAvgCents,
			PeakMonth:                        fc.Summary.PeakMonth,
			PeakCommittedCashCents:           fc.Summary.PeakCommittedCashCents,
			AlertMonthsCount:                 fc.Summary.AlertMonthsCount,
		},
		Months: months,
		Assets: assets,
	}
}

func decodeGetCashFlowForecastRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)
	projectID, err := core.ParseProjectID(vars["project_id"])
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}
	q := r.URL.Query()
	var asOf time.Time
	if rawAsOf := q.Get("as_of_date"); rawAsOf != "" {
		parsed, err := parseFlexibleDate(rawAsOf)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid as_of_date", err)
		}
		asOf = parsed
	}
	requesterID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}
	return assetapp.GetCashFlowForecast{
		RequesterID: requesterID,
		ProjectID:   projectID,
		FromMonth:   q.Get("from_month"),
		ToMonth:     q.Get("to_month"),
		Tag:         q.Get("tag"),
		AsOfDate:    asOf,
	}, nil
}

func makeGetCashFlowForecastEndpoint(app *assetapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(assetapp.GetCashFlowForecast)
		fc, err := app.Queries.GetCashFlowForecast.Handle(ctx, query)
		if err != nil {
			return nil, err
		}
		return toCashFlowForecastDTO(fc), nil
	}
}
