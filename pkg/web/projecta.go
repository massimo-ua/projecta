package web

import (
	"context"
	"encoding/json"
	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"net/http"
	"time"
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
	ProjectID   string   `json:"project_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Owner       OwnerDTO `json:"owner"`
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
	TypeID      string          `json:"type_id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Category    TypeCategoryDTO `json:"category"`
}

type PaymentDTO struct {
	PaymentID   string      `json:"payment_id"`
	Project     ProjectDTO  `json:"project"`
	Owner       OwnerDTO    `json:"owner"`
	Type        TypeDTO     `json:"type"`
	Category    CategoryDTO `json:"category"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
	Currency    string      `json:"currency"`
	PaymentDate string      `json:"payment_date"`
}

type ProjectEndpoints struct {
	CreateProject     endpoint.Endpoint
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
	CreateAsset       endpoint.Endpoint
	RemoveAsset       endpoint.Endpoint
	ListAssets        endpoint.Endpoint
	UpdateAsset       endpoint.Endpoint
	GetAsset          endpoint.Endpoint
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

		return ProjectDTO{
			ProjectID:   project.ProjectID.String(),
			Name:        project.Name,
			Description: project.Description,
			Owner: OwnerDTO{
				PersonID:    project.Owner.PersonID.String(),
				DisplayName: project.Owner.DisplayName,
			},
		}, nil
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
			TypeID:      costType.ID.String(),
			Name:        costType.Name,
			Description: costType.Description,
		}, nil
	}
}

func makeCreatePaymentEndpoint(svc projecta.PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.CreatePaymentCommand)

		expense, err := svc.Create(ctx, command)

		if err != nil {
			return nil, err
		}

		return PaymentDTO{
			PaymentID: expense.ID.String(),
			Project: ProjectDTO{
				ProjectID:   expense.Project.ProjectID.String(),
				Name:        expense.Project.Name,
				Description: expense.Project.Description,
				Owner: OwnerDTO{
					PersonID:    expense.Owner.PersonID.String(),
					DisplayName: expense.Owner.DisplayName,
				},
			},
			Owner: OwnerDTO{
				PersonID:    expense.Owner.PersonID.String(),
				DisplayName: expense.Owner.DisplayName,
			},
			Type: TypeDTO{
				TypeID:      expense.Type.ID.String(),
				Name:        expense.Type.Name,
				Description: expense.Type.Description,
			},
			Category: CategoryDTO{
				CategoryID:  expense.Type.Category.ID.String(),
				Name:        expense.Type.Category.Name,
				Description: expense.Type.Category.Description,
			},
			Description: expense.Description,
			Amount:      expense.Amount.Amount(),
			Currency:    expense.Amount.Currency().Code,
			PaymentDate: expense.Date.Format(time.RFC3339),
		}, nil
	}
}

func makeListProjectsEndpoint(svc projecta.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.ProjectCollectionFilter)

		projects, err := svc.Find(ctx, filter)

		var list []ProjectDTO = make([]ProjectDTO, 0)

		for _, p := range projects {
			list = append(list, ProjectDTO{
				ProjectID:   p.ProjectID.String(),
				Name:        p.Name,
				Description: p.Description,
				Owner: OwnerDTO{
					PersonID:    p.Owner.PersonID.String(),
					DisplayName: p.Owner.DisplayName,
				},
			})
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
				TypeID:      p.ID.String(),
				Name:        p.Name,
				Description: p.Description,
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

func makeListPaymentsEndpoint(svc projecta.PaymentService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.PaymentCollectionFilter)

		collection, err := svc.Find(ctx, filter)

		if err != nil {
			return nil, err
		}

		var list []PaymentDTO = make([]PaymentDTO, 0)

		for _, e := range collection.Elements() {
			list = append(list, PaymentDTO{
				PaymentID: e.ID.String(),
				Project: ProjectDTO{
					ProjectID:   e.Project.ProjectID.String(),
					Name:        e.Project.Name,
					Description: e.Project.Description,
					Owner: OwnerDTO{
						PersonID:    e.Owner.PersonID.String(),
						DisplayName: e.Owner.DisplayName,
					},
				},
				Owner: OwnerDTO{
					PersonID:    e.Owner.PersonID.String(),
					DisplayName: e.Owner.DisplayName,
				},
				Type: TypeDTO{
					TypeID:      e.Type.ID.String(),
					Name:        e.Type.Name,
					Description: e.Type.Description,
				},
				Category: CategoryDTO{
					CategoryID:  e.Type.Category.ID.String(),
					Name:        e.Type.Category.Name,
					Description: e.Type.Category.Description,
				},
				Description: e.Description,
				Amount:      e.Amount.Amount(),
				Currency:    e.Amount.Currency().Code,
				PaymentDate: e.Date.Format(time.RFC3339),
			})
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

func makeShowProjectTotalsEndpoint(payments projecta.PaymentService, assets asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		projectID := request.(uuid.UUID)

		offset := 0
		limit := 100
		next := true
		var totalPayments *money.Money
		var totalAssets *money.Money

		for next {
			page, err := payments.Find(ctx, projecta.PaymentCollectionFilter{
				ProjectID: projectID,
				Kind:      projecta.DownPayment,
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
				if totalPayments == nil {
					totalPayments = e.Amount
				} else {
					if totalPayments.Currency() != e.Amount.Currency() {
						return nil, exceptions.NewInternalException("project payments currencies mismatch", nil)
					}

					totalPayments, err = totalPayments.Add(e.Amount)

					if err != nil {
						return nil, err
					}
				}
			}

			if len(page.Elements()) < limit {
				next = false
			}

			offset += limit
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
				if totalAssets == nil {
					totalAssets = e.Price
				} else {
					if totalAssets.Currency() != e.Price.Currency() {
						return nil, exceptions.NewInternalException("project assets currencies mismatch", nil)
					}

					totalAssets, err = totalAssets.Add(e.Price)

					if err != nil {
						return nil, err
					}
				}
			}

			if len(page.Elements()) < limit {
				next = false
			}

			offset += limit
		}

		totals := make([]TotalDTO, 0)

		if totalPayments != nil {
			totals = append(totals, TotalDTO{
				Title:    "Total Payments",
				Amount:   totalPayments.Amount(),
				Currency: totalPayments.Currency().Code,
			})
		}

		if totalAssets != nil {
			var toalPaymentsAmount int64

			if totalPayments != nil {
				toalPaymentsAmount = totalPayments.Amount()
			}

			totals = append(totals, TotalDTO{
				Title:    "Project Balance",
				Amount:   toalPaymentsAmount - totalAssets.Amount(),
				Currency: totalAssets.Currency().Code,
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

func MakeProjectEndpoints(
	projectService projecta.ProjectService,
	categoryService projecta.CategoryService,
	typeService projecta.TypeService,
	expenseService projecta.PaymentService,
	assetService asset.Service,
) (ProjectEndpoints, error) {
	return ProjectEndpoints{
		CreateProject:     makeCreateProjectEndpoint(projectService),
		CreateCategory:    makeCreateCategoryEndpoint(categoryService),
		CreateType:        makeCreateTypeEndpoint(typeService),
		CreatePayment:     makeCreatePaymentEndpoint(expenseService),
		ListProjects:      makeListProjectsEndpoint(projectService),
		ListTypes:         makeListProjectTypesEndpoint(typeService),
		ListCategories:    makeListCategoriesEndpoint(categoryService),
		ListPayments:      makeListPaymentsEndpoint(expenseService),
		ShowProjectTotals: makeShowProjectTotalsEndpoint(expenseService, assetService),
		RemoveType:        makeRemoveTypeEndpoint(typeService),
		RemovePayment:     makeRemovePaymentEndpoint(expenseService),
		CreateAsset:       makeCreateAssetEndpoint(assetService),
		RemoveAsset:       makeRemoveAssetEndpoint(assetService),
		ListAssets:        makeListAssetsEndpoint(assetService),
		UpdateAsset:       makeUpdateAssetEndpoint(assetService),
		GetAsset:          makeGetAssetEndpoint(assetService),
	}, nil
}
