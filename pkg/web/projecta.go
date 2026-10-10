package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

type CreateProjectDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateCategoryDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreatePaymentDTO struct {
	ProjectID   string `json:"project_id"`
	TypeID      string `json:"type_id"`
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	PaymentDate string `json:"payment_date"`
	Kind        string `json:"kind,omitempty"`
}

type OwnerDTO struct {
	PersonID    string `json:"person_id"`
	DisplayName string `json:"display_name"`
}

type ProjectDTO struct {
	ProjectID    string                 `json:"project_id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description"`
	Owner        OwnerDTO               `json:"owner"`
	ShareToken   string                 `json:"share_token,omitempty"`
	IsShared     bool                   `json:"is_shared,omitempty"`
	MainCurrency string                 `json:"mainCurrency,omitempty"`
	Participants []projecta.Participant `json:"participants"`
}

type UpdateProjectDTO struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	MainCurrency string `json:"mainCurrency,omitempty"`
}

func toProjectDTO(project *projecta.Project) ProjectDTO {
	if project == nil {
		return ProjectDTO{}
	}
	mainCurrency := project.MainCurrency
	if mainCurrency == "" {
		mainCurrency = "UAH"
	}
	participants := project.Participants
	if participants == nil {
		participants = make([]projecta.Participant, 0)
	}
	dto := ProjectDTO{
		ProjectID:    project.ProjectID.String(),
		Name:         project.Name,
		Description:  project.Description,
		ShareToken:   project.ShareToken.String(),
		IsShared:     project.IsShared,
		MainCurrency: mainCurrency,
		Participants: participants,
	}
	if project.Owner != nil {
		dto.Owner = OwnerDTO{
			PersonID:    project.Owner.PersonID.String(),
			DisplayName: project.Owner.DisplayName,
		}
	}
	return dto
}

type CategoryDTO struct {
	CategoryID  string `json:"category_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TypeCategoryDTO struct {
	CategoryID string `json:"category_id"`
	Name       string `json:"name"`
}

type TypeDTO struct {
	TypeID   string          `json:"type_id"`
	Name     string          `json:"name"`
	Category TypeCategoryDTO `json:"category"`
}

type PaymentDTO struct {
	PaymentID    string      `json:"payment_id"`
	Project      ProjectDTO  `json:"project"`
	Owner        OwnerDTO    `json:"owner"`
	Type         TypeDTO     `json:"type"`
	Category     CategoryDTO `json:"category"`
	Description  string      `json:"description"`
	Amount       int64       `json:"amount"`
	Currency     string      `json:"currency"`
	HomeAmount   int64       `json:"home_amount,omitempty"`
	HomeCurrency string      `json:"home_currency,omitempty"`
	PaymentDate  string      `json:"payment_date"`
	Kind         string      `json:"kind,omitempty"`
}

func toPaymentDTO(p *projecta.Payment, rateProvider currency.CurrencyRateProvider) PaymentDTO {
	if p == nil {
		return PaymentDTO{}
	}
	projDTO := toProjectDTO(p.Project)
	homeCurrency := projDTO.MainCurrency
	if homeCurrency == "" {
		homeCurrency = "UAH"
	}

	homeAmount := p.Amount.Amount()
	if rateProvider != nil && p.Amount.Currency().Code != homeCurrency {
		converted, err := rateProvider.Convert(
			currency.NewCurrency(p.Amount.Amount(), p.Amount.Currency().Code),
			currency.NewCurrency(0, homeCurrency),
		)
		if err == nil {
			homeAmount = converted.Amount
		}
	}

	return PaymentDTO{
		PaymentID: p.ID.String(),
		Project:   projDTO,
		Owner: OwnerDTO{
			PersonID:    p.Owner.PersonID.String(),
			DisplayName: p.Owner.DisplayName,
		},
		Type: TypeDTO{
			TypeID: p.Type.ID.String(),
			Name:   p.Type.Name,
			Category: TypeCategoryDTO{
				CategoryID: p.Type.Category.ID.String(),
				Name:       p.Type.Category.Name,
			},
		},
		Description:  p.Description,
		Amount:       p.Amount.Amount(),
		Currency:     p.Amount.Currency().Code,
		HomeAmount:   homeAmount,
		HomeCurrency: homeCurrency,
		PaymentDate:  p.Date.Format(time.RFC3339),
		Kind:         p.Kind.String(),
	}
}

type ProjectEndpoints struct {
	CreateProject     endpoint.Endpoint
	GetProject        endpoint.Endpoint
	AcceptShare       endpoint.Endpoint
	CreateCategory    endpoint.Endpoint
	CreateType        endpoint.Endpoint
	CreatePayment     endpoint.Endpoint
	ListProjects      endpoint.Endpoint
	ListTypes         endpoint.Endpoint
	ListCategories    endpoint.Endpoint
	ListPayments      endpoint.Endpoint
	ShowProjectTotals endpoint.Endpoint
	RemoveType        endpoint.Endpoint
	RemovePayment     endpoint.Endpoint
	CreateAsset               endpoint.Endpoint
	CreateAssetFromPayments   endpoint.Endpoint
	RemoveAsset               endpoint.Endpoint
	ListAssets        endpoint.Endpoint
	UpdateAsset       endpoint.Endpoint
	GetAsset          endpoint.Endpoint
	LinkChildAsset          endpoint.Endpoint
	UnlinkChildAsset        endpoint.Endpoint
	AssignInvestments       endpoint.Endpoint
	GroupAssets             endpoint.Endpoint
	LinkChildren            endpoint.Endpoint
	CreateInvestment        endpoint.Endpoint
	ListInvestments         endpoint.Endpoint
	GetInvestment           endpoint.Endpoint
	UpdateInvestment        endpoint.Endpoint
	RemoveInvestment        endpoint.Endpoint
	CreateBatchInvestments  endpoint.Endpoint
	ListTags                endpoint.Endpoint
	UpdatePayment       endpoint.Endpoint
	GetPayment          endpoint.Endpoint
	UpdateProject       endpoint.Endpoint
	ParseStatement      endpoint.Endpoint
	CreateBatchPayments endpoint.Endpoint
}

func DecodeCreateProjectRequest(ctx context.Context, r *http.Request) (any, error) {
	personID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if !ok {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	var req CreateProjectDTO
	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	return projecta.CreateProjectCommand{
		PersonID:    personID,
		Name:        req.Name,
		Description: req.Description,
	}, err
}

func DecodeCreateCategoryRequest(ctx context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("project id not found", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	personID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)

	if !ok {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	var req CreateCategoryDTO
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	return projecta.CreateCategoryCommand{
		ProjectID:   projectUUID,
		PersonID:    personID,
		Name:        req.Name,
		Description: req.Description,
	}, err
}

func DecodeCreateTypeRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateTypeDTO
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if req.Name == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	categoryUUID, err := uuid.Parse(req.CategoryID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid category id", err)
	}

	return projecta.CreateTypeCommand{
		ProjectID:   projectUUID,
		Name:        req.Name,
		Description: req.Description,
		CategoryID:  categoryUUID,
	}, err
}

func DecodeCreatePaymentRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreatePaymentDTO
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	amount := money.New(req.Amount, req.Currency)

	date, err := time.Parse(time.RFC3339, req.PaymentDate)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid date", err)
	}

	typeUUID, err := uuid.Parse(req.TypeID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid type id", err)
	}

	var paymentKind projecta.PaymentKind

	if req.Kind == "" {
		paymentKind = projecta.UponCompletionPayment
	} else {
		paymentKind, err = projecta.ToPaymentKind(req.Kind)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid payment kind", err)
		}
	}

	return projecta.CreatePaymentCommand{
		ProjectID:   projectUUID,
		TypeID:      typeUUID,
		Description: req.Description,
		Amount:      amount,
		PaymentDate: date,
		Kind:        paymentKind,
	}, err
}

func makeCreateProjectEndpoint(svc projecta.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.CreateProjectCommand)

		project, err := svc.Create(ctx, command)

		if err != nil {
			return nil, err
		}

		return toProjectDTO(project), nil
	}
}

func makeCreateCategoryEndpoint(svc projecta.CategoryService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.CreateCategoryCommand)

		category, err := svc.Create(ctx, command)

		if err != nil {
			return nil, err
		}

		return CategoryDTO{
			CategoryID:  category.ID.String(),
			Name:        category.Name,
			Description: category.Description,
		}, nil
	}
}

func makeCreateTypeEndpoint(svc projecta.TypeService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.CreateTypeCommand)

		costType, err := svc.Create(ctx, command)

		if err != nil {
			return nil, err
		}

		return TypeDTO{
			TypeID: costType.ID.String(),
			Name:   costType.Name,
		}, nil
	}
}

func makeCreatePaymentEndpoint(svc projecta.PaymentService, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.CreatePaymentCommand)

		expense, err := svc.Create(ctx, command)

		if err != nil {
			return nil, err
		}

		return toPaymentDTO(expense, rateProvider), nil
	}
}

func makeListProjectsEndpoint(svc projecta.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.ProjectCollectionFilter)

		projects, err := svc.Find(ctx, filter)

		var list []ProjectDTO = make([]ProjectDTO, 0)

		for _, p := range projects {
			list = append(list, toProjectDTO(p))
		}

		return ListProjectsResponse{
			Projects: list,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
			},
		}, err
	}
}

func makeListProjectTypesEndpoint(svc projecta.TypeService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.TypeCollectionFilter)

		collection, err := svc.Find(ctx, filter)

		var list []TypeDTO = make([]TypeDTO, 0)

		for _, p := range collection.Elements() {
			list = append(list, TypeDTO{
				TypeID: p.ID.String(),
				Name:   p.Name,
				Category: TypeCategoryDTO{
					CategoryID: p.Category.ID.String(),
					Name:       p.Category.Name,
				},
			})
		}

		return ListTypesResponse{
			Types: list,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
				Total:  collection.Total(),
			},
		}, err
	}
}

func makeListCategoriesEndpoint(svc projecta.CategoryService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.CategoryCollectionFilter)

		collection, err := svc.Find(ctx, filter)

		var list []CategoryDTO = make([]CategoryDTO, 0)

		for _, c := range collection.Elements() {
			list = append(list, CategoryDTO{
				CategoryID:  c.ID.String(),
				Name:        c.Name,
				Description: c.Description,
			})
		}

		return ListCategoriesResponse{
			Categories: list,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
				Total:  collection.Total(),
			},
		}, err
	}
}

func makeListPaymentsEndpoint(svc projecta.PaymentService, rateProvider currency.CurrencyRateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.PaymentCollectionFilter)

		collection, err := svc.Find(ctx, filter)

		if err != nil {
			return nil, err
		}

		var list []PaymentDTO = make([]PaymentDTO, 0)

		for _, e := range collection.Elements() {
			list = append(list, toPaymentDTO(e, rateProvider))
		}

		return ListPaymentsResponse{
			Payments: list,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
				Total:  collection.Total(),
			},
		}, err
	}
}

func makeShowProjectTotalsEndpoint(
	projectSvc projecta.ProjectService,
	payments projecta.PaymentService,
	assets asset.Service,
	rateProvider currency.CurrencyRateProvider,
	investments ...investment.Service,
) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		projectID := request.(uuid.UUID)

		proj, err := projectSvc.FindOne(ctx, projecta.ProjectFilter{ProjectID: projectID})
		if err != nil {
			return nil, err
		}

		homeCurrency := proj.MainCurrency
		if homeCurrency == "" {
			homeCurrency = "UAH"
		}

		var invSvc investment.Service
		if len(investments) > 0 {
			invSvc = investments[0]
		}

		offset := 0
		limit := 100
		next := true
		var totalInvestedAmount int64
		var totalAssetsAmount int64
		hasInvestments := false
		hasAssets := false

		if invSvc != nil {
			for next {
				page, err := invSvc.Find(ctx, investment.CollectionFilter{
					ProjectID: projectID,
					Pagination: core.Pagination{
						Limit:  limit,
						Offset: offset,
					},
				})
				if err != nil {
					return nil, err
				}

				if page.Total() == 0 {
					break
				}

				for _, e := range page.Elements() {
					hasInvestments = true
					amount := e.Amount.Amount()
					if rateProvider != nil && e.Amount.Currency().Code != homeCurrency {
						converted, err := rateProvider.Convert(
							currency.NewCurrency(e.Amount.Amount(), e.Amount.Currency().Code),
							currency.NewCurrency(0, homeCurrency),
						)
						if err != nil {
							return nil, err
						}
						amount = converted.Amount
					}
					totalInvestedAmount += amount
				}

				if len(page.Elements()) < limit {
					next = false
				}

				offset += limit
			}
		} else if payments != nil {
			for next {
				page, err := payments.Find(ctx, projecta.PaymentCollectionFilter{
					ProjectID: projectID,
					Pagination: core.Pagination{
						Limit:  limit,
						Offset: offset,
					},
				})

				if err != nil {
					return nil, err
				}

				if page.Total() == 0 {
					break
				}

				for _, e := range page.Elements() {
					hasInvestments = true
					amount := e.Amount.Amount()
					if rateProvider != nil && e.Amount.Currency().Code != homeCurrency {
						converted, err := rateProvider.Convert(
							currency.NewCurrency(e.Amount.Amount(), e.Amount.Currency().Code),
							currency.NewCurrency(0, homeCurrency),
						)
						if err != nil {
							return nil, err
						}
						amount = converted.Amount
					}
					totalInvestedAmount += amount
				}

				if len(page.Elements()) < limit {
					next = false
				}

				offset += limit
			}
		}

		next = true
		offset = 0
		limit = 100
		for next {
			page, err := assets.Find(ctx, asset.CollectionFilter{
				ProjectID: projectID,
				Pagination: core.Pagination{
					Limit:  limit,
					Offset: offset,
				},
			})

			if err != nil {
				return nil, err
			}

			if page.Total() == 0 {
				break
			}

			for _, e := range page.Elements() {
				// Don't count child assets if they have parents to avoid double-counting in project valuation
				if len(e.Parents()) > 0 {
					continue
				}

				var price int64
				var priceCurrency string
				if e.TotalCost() != nil {
					price = e.TotalCost().Amount()
					priceCurrency = e.TotalCost().Currency().Code
				} else if e.Price() != nil {
					price = e.Price().Amount()
					priceCurrency = e.Price().Currency().Code
				}

				if price > 0 {
					hasAssets = true
					if rateProvider != nil && priceCurrency != "" && priceCurrency != homeCurrency {
						converted, err := rateProvider.Convert(
							currency.NewCurrency(price, priceCurrency),
							currency.NewCurrency(0, homeCurrency),
						)
						if err != nil {
							return nil, err
						}
						price = converted.Amount
					}
					totalAssetsAmount += price
				}
			}

			if len(page.Elements()) < limit {
				next = false
			}

			offset += limit
		}

		totals := make([]TotalDTO, 0)

		totals = append(totals, TotalDTO{
			Title:    "Total Invested",
			Amount:   totalInvestedAmount,
			Currency: homeCurrency,
		})

		if hasAssets || hasInvestments {
			totals = append(totals, TotalDTO{
				Title:    "Project Balance",
				Amount:   totalInvestedAmount - totalAssetsAmount,
				Currency: homeCurrency,
			})
		}

		return ProjectTotalsDTO{
			Totals: totals,
		}, nil
	}
}

func makeRemoveTypeEndpoint(svc projecta.TypeService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.RemoveProjectResourceCommand)

		err := svc.Remove(ctx, command)

		return nil, err
	}
}

func makeRemovePaymentEndpoint(svc projecta.PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command, ok := request.(projecta.RemoveProjectResourceCommand)

		if !ok {
			return nil, exceptions.NewValidationException("invalid request", nil)
		}

		err := svc.Remove(ctx, projecta.RemovePaymentCommand{
			ID:        command.ResourceID,
			ProjectID: command.ProjectID,
		})

		return nil, err
	}
}

type AcceptShareCommand struct {
	ShareToken uuid.UUID
	PersonID   uuid.UUID
}

func DecodeAcceptShareRequest(ctx context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	tokenStr, ok := vars["share_token"]
	if !ok {
		return nil, exceptions.NewValidationException("share token is required", nil)
	}

	shareToken, err := uuid.Parse(tokenStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid share token", err)
	}

	personID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)
	if !ok {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	return AcceptShareCommand{
		ShareToken: shareToken,
		PersonID:   personID,
	}, nil
}

func DecodeGetProjectRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	return projecta.ProjectFilter{
		ProjectID: projectUUID,
	}, nil
}

func makeGetProjectEndpoint(svc projecta.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.ProjectFilter)
		project, err := svc.FindOne(ctx, filter)
		if err != nil {
			return nil, err
		}

		return toProjectDTO(project), nil
	}
}

func makeAcceptShareEndpoint(svc projecta.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(AcceptShareCommand)
		project, err := svc.AcceptShare(ctx, cmd.ShareToken, cmd.PersonID)
		if err != nil {
			return nil, err
		}

		return toProjectDTO(project), nil
	}
}

func decodeUpdateProjectRequest(ctx context.Context, r *http.Request) (any, error) {
	personID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)
	if !ok {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}

	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req UpdateProjectDTO
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, exceptions.NewValidationException("invalid json", err)
	}

	return projecta.UpdateProjectCommand{
		ProjectID:    projectUUID,
		PersonID:     personID,
		Name:         req.Name,
		Description:  req.Description,
		MainCurrency: req.MainCurrency,
	}, nil
}

func makeUpdateProjectEndpoint(svc projecta.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.UpdateProjectCommand)

		project, err := svc.Update(ctx, command)
		if err != nil {
			return nil, err
		}

		return toProjectDTO(project), nil
	}
}

func MakeProjectEndpoints(
	projectService projecta.ProjectService,
	categoryService projecta.CategoryService,
	typeService projecta.TypeService,
	expenseService projecta.PaymentService,
	assetService asset.Service,
	rateProvider currency.CurrencyRateProvider,
	investmentService ...investment.Service,
) (ProjectEndpoints, error) {
	var invSvc investment.Service
	if len(investmentService) > 0 {
		invSvc = investmentService[0]
	}

	var epCreateInv endpoint.Endpoint
	var epListInv endpoint.Endpoint
	var epGetInv endpoint.Endpoint
	var epUpdInv endpoint.Endpoint
	var epRemInv endpoint.Endpoint
	var epBatchInv endpoint.Endpoint
	var epListTags endpoint.Endpoint
	if invSvc != nil {
		epCreateInv = makeCreateInvestmentEndpoint(invSvc, rateProvider)
		epListInv = makeListInvestmentsEndpoint(invSvc, rateProvider)
		epGetInv = makeGetInvestmentEndpoint(invSvc, rateProvider)
		epUpdInv = makeUpdateInvestmentEndpoint(invSvc)
		epRemInv = makeRemoveInvestmentEndpoint(invSvc)
		epBatchInv = makeCreateBatchInvestmentsEndpoint(invSvc, rateProvider)
		epListTags = makeListTagsEndpoint(invSvc)
	}

	return ProjectEndpoints{
		CreateProject:     makeCreateProjectEndpoint(projectService),
		GetProject:        makeGetProjectEndpoint(projectService),
		AcceptShare:       makeAcceptShareEndpoint(projectService),
		CreateCategory:    makeCreateCategoryEndpoint(categoryService),
		CreateType:        makeCreateTypeEndpoint(typeService),
		CreatePayment:     makeCreatePaymentEndpoint(expenseService, rateProvider),
		ListProjects:      makeListProjectsEndpoint(projectService),
		ListTypes:         makeListProjectTypesEndpoint(typeService),
		ListCategories:    makeListCategoriesEndpoint(categoryService),
		ListPayments:      makeListPaymentsEndpoint(expenseService, rateProvider),
		ShowProjectTotals: makeShowProjectTotalsEndpoint(projectService, expenseService, assetService, rateProvider, invSvc),
		RemoveType:        makeRemoveTypeEndpoint(typeService),
		RemovePayment:     makeRemovePaymentEndpoint(expenseService),
		CreateAsset:             makeCreateAssetEndpoint(assetService, rateProvider),
		CreateAssetFromPayments: makeCreateAssetFromPaymentsEndpoint(assetService, rateProvider),
		RemoveAsset:             makeRemoveAssetEndpoint(assetService),
		ListAssets:        makeListAssetsEndpoint(assetService, rateProvider),
		UpdateAsset:       makeUpdateAssetEndpoint(assetService),
		GetAsset:          makeGetAssetEndpoint(assetService, rateProvider),
		LinkChildAsset:    makeLinkChildAssetEndpoint(assetService),
		UnlinkChildAsset:  makeUnlinkChildAssetEndpoint(assetService),
		AssignInvestments: makeAssignInvestmentsEndpoint(assetService),
		GroupAssets:       makeGroupAssetsEndpoint(assetService, rateProvider),
		LinkChildren:      makeLinkChildrenEndpoint(assetService),
		CreateInvestment:  epCreateInv,
		ListInvestments:   epListInv,
		GetInvestment:     epGetInv,
		UpdateInvestment:  epUpdInv,
		RemoveInvestment:  epRemInv,
		CreateBatchInvestments: epBatchInv,
		ListTags:          epListTags,
		UpdatePayment:       makeUpdatePaymentEndpoint(expenseService),
		GetPayment:          makeGetPaymentEndpoint(expenseService, rateProvider),
		UpdateProject:       makeUpdateProjectEndpoint(projectService),
		ParseStatement:      makeParseStatementEndpoint(expenseService),
		CreateBatchPayments: makeCreateBatchPaymentsEndpoint(expenseService, rateProvider),
	}, nil
}
