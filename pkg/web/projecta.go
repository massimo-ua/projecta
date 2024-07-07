package web

import (
	"context"
	"encoding/json"
	"github.com/Rhymond/go-money"
	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
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

type CreateExpenseDTO struct {
	ProjectID   string `json:"project_id"`
	CategoryID  string `json:"category_id"`
	TypeID      string `json:"type_id"`
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
	Currency    string `json:"currency"`
	ExpenseDate string `json:"expense_date"`
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

type TypeDTO struct {
	TypeID      string `json:"type_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type ExpenseDTO struct {
	ExpenseID   string      `json:"expense_id"`
	Project     ProjectDTO  `json:"project"`
	Owner       OwnerDTO    `json:"owner"`
	Type        TypeDTO     `json:"type"`
	Category    CategoryDTO `json:"category"`
	Description string      `json:"description"`
	Amount      int64       `json:"amount"`
	Currency    string      `json:"currency"`
	ExpenseDate string      `json:"expense_date"`
}

type ProjectEndpoints struct {
	CreateProject     endpoint.Endpoint
	CreateCategory    endpoint.Endpoint
	CreateType        endpoint.Endpoint
	CreateExpense     endpoint.Endpoint
	ListProjects      endpoint.Endpoint
	ListTypes         endpoint.Endpoint
	ListCategories    endpoint.Endpoint
	ListExpenses      endpoint.Endpoint
	ShowProjectTotals endpoint.Endpoint
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

	var req TypeDTO
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	return projecta.CreateTypeCommand{
		ProjectID:   projectUUID,
		Name:        req.Name,
		Description: req.Description,
	}, err
}

func DecodeCreateExpenseRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateExpenseDTO
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	amount := money.New(req.Amount, req.Currency)

	date, err := time.Parse(time.RFC3339, req.ExpenseDate)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid date", err)
	}

	return projecta.CreateExpenseCommand{
		ProjectID:   projectUUID,
		CategoryID:  uuid.MustParse(req.CategoryID),
		TypeID:      uuid.MustParse(req.TypeID),
		Description: req.Description,
		Amount:      amount,
		ExpenseDate: date,
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

func makeCreateExpenseEndpoint(svc projecta.ExpenseService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(projecta.CreateExpenseCommand)

		expense, err := svc.Create(ctx, command)

		if err != nil {
			return nil, err
		}

		return ExpenseDTO{
			ExpenseID: expense.ID.String(),
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
				CategoryID:  expense.Category.ID.String(),
				Name:        expense.Category.Name,
				Description: expense.Category.Description,
			},
			Description: expense.Description,
			Amount:      expense.Amount.Amount(),
			Currency:    expense.Amount.Currency().Code,
			ExpenseDate: expense.Date.Format(time.RFC3339),
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

		projects, err := svc.Find(ctx, filter)

		var list []TypeDTO = make([]TypeDTO, 0)

		for _, p := range projects {
			list = append(list, TypeDTO{
				TypeID:      p.ID.String(),
				Name:        p.Name,
				Description: p.Description,
			})
		}

		return ListTypesResponse{
			Types: list,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
			},
		}, err
	}
}

func makeListCategoriesEndpoint(svc projecta.CategoryService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.CategoryCollectionFilter)

		categories, err := svc.Find(ctx, filter)

		var list []CategoryDTO = make([]CategoryDTO, 0)

		for _, c := range categories {
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
			},
		}, err
	}
}

func makeListExpensesEndpoint(svc projecta.ExpenseService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(projecta.ExpenseCollectionFilter)

		expenses, err := svc.Find(ctx, filter)

		var list []ExpenseDTO = make([]ExpenseDTO, 0)

		for _, e := range expenses {
			list = append(list, ExpenseDTO{
				ExpenseID: e.ID.String(),
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
					CategoryID:  e.Category.ID.String(),
					Name:        e.Category.Name,
					Description: e.Category.Description,
				},
				Description: e.Description,
				Amount:      e.Amount.Amount(),
				Currency:    e.Amount.Currency().Code,
				ExpenseDate: e.Date.Format(time.RFC3339),
			})
		}

		return ListExpensesResponse{
			Expenses: list,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
			},
		}, err
	}
}

func makeShowProjectTotalsEndpoint(svc projecta.ExpenseService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		projectID := request.(uuid.UUID)

		offset := 0
		limit := 100
		next := true
		var total *money.Money

		for next {
			rows, err := svc.Find(ctx, projecta.ExpenseCollectionFilter{
				ProjectID: projectID,
				Pagination: core.Pagination{
					Limit:  limit,
					Offset: offset,
				},
			})

			if err != nil {
				return nil, err
			}

			for _, e := range rows {
				if total == nil {
					total = e.Amount
				} else {
					if total.Currency() != e.Amount.Currency() {
						return nil, exceptions.NewInternalException("project expenses currencies mismatch", nil)
					}

					total, err = total.Add(e.Amount)

					if err != nil {
						return nil, err
					}
				}
			}

			if len(rows) < limit {
				next = false
			}

			offset += limit
		}

		return ProjectTotalsDTO{
			Totals: []TotalDTO{
				{
					Title:    "Total Expenses",
					Amount:   total.Amount(),
					Currency: total.Currency().Code,
				},
			},
		}, nil
	}
}

func MakeProjectEndpoints(
	projectService projecta.ProjectService,
	categoryService projecta.CategoryService,
	typeService projecta.TypeService,
	expenseService projecta.ExpenseService,
) (ProjectEndpoints, error) {
	return ProjectEndpoints{
		CreateProject:     makeCreateProjectEndpoint(projectService),
		CreateCategory:    makeCreateCategoryEndpoint(categoryService),
		CreateType:        makeCreateTypeEndpoint(typeService),
		CreateExpense:     makeCreateExpenseEndpoint(expenseService),
		ListProjects:      makeListProjectsEndpoint(projectService),
		ListTypes:         makeListProjectTypesEndpoint(typeService),
		ListCategories:    makeListCategoriesEndpoint(categoryService),
		ListExpenses:      makeListExpensesEndpoint(expenseService),
		ShowProjectTotals: makeShowProjectTotalsEndpoint(expenseService),
	}, nil
}
