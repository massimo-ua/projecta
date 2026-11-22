package web

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
	projectapp "gitlab.com/massimo-ua/projecta/internal/project/app"
)

type TagBudgetDTO struct {
	Tag             string   `json:"tag"`
	Frequency       string   `json:"frequency"`
	TargetMonth     int      `json:"target_month,omitempty"`
	ResourceType    string   `json:"resource_type"`
	Amount          int64    `json:"amount"`
	Currency        string   `json:"currency,omitempty"`
	HomeAmount      int64    `json:"home_amount,omitempty"`
	TimeHours       *float64 `json:"time_hours,omitempty"`
	TimeHourlyRate  *int64   `json:"time_hourly_rate,omitempty"`
	GoodsQuantity   *float64 `json:"goods_quantity,omitempty"`
	GoodsUnit       string   `json:"goods_unit,omitempty"`
	GoodsItemName   string   `json:"goods_item_name,omitempty"`
	Description     string   `json:"description,omitempty"`
	AnnualTotal     int64    `json:"annual_total"`
	AnnualHomeTotal int64    `json:"annual_home_total,omitempty"`
}

type CreateProjectDTO struct {
	Name         string         `json:"name"`
	Description  string         `json:"description"`
	MainCurrency string         `json:"mainCurrency,omitempty"`
	StartDate    string         `json:"start_date,omitempty"`
	EndDate      string         `json:"end_date,omitempty"`
	TagBudgets   []TagBudgetDTO `json:"tag_budgets,omitempty"`
}

type OwnerDTO struct {
	PersonID    string `json:"person_id"`
	DisplayName string `json:"display_name"`
}

type ProjectDTO struct {
	ProjectID           string         `json:"project_id"`
	Name                string         `json:"name"`
	Description         string         `json:"description"`
	Owner               OwnerDTO       `json:"owner"`
	StartDate           string         `json:"start_date,omitempty"`
	EndDate             string         `json:"end_date,omitempty"`
	ShareToken          string         `json:"share_token,omitempty"`
	IsShared            bool           `json:"is_shared,omitempty"`
	MainCurrency        string         `json:"mainCurrency,omitempty"`
	Participants        []string       `json:"participants"`
	TagBudgets          []TagBudgetDTO `json:"tag_budgets"`
	AnnualPlannedBudget int64          `json:"annual_planned_budget"`
}

type UpdateProjectDTO struct {
	Name         string          `json:"name,omitempty"`
	Description  string          `json:"description,omitempty"`
	MainCurrency string          `json:"mainCurrency,omitempty"`
	StartDate    string          `json:"start_date,omitempty"`
	EndDate      string          `json:"end_date,omitempty"`
	TagBudgets   *[]TagBudgetDTO `json:"tag_budgets,omitempty"`
}

type UpdateProjectBudgetDTO struct {
	TagBudgets []TagBudgetDTO `json:"tag_budgets"`
}

func parseOptionalDate(str string) (time.Time, error) {
	str = strings.TrimSpace(str)
	if str == "" {
		return time.Time{}, nil
	}
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		return t, nil
	}
	return time.Parse("2006-01-02", str)
}

func toTagBudgetInputs(dtos []TagBudgetDTO) []project.TagBudgetInput {
	if dtos == nil {
		return nil
	}
	res := make([]project.TagBudgetInput, 0, len(dtos))
	for _, d := range dtos {
		res = append(res, project.TagBudgetInput{
			Tag:            d.Tag,
			Frequency:      d.Frequency,
			TargetMonth:    d.TargetMonth,
			ResourceType:   d.ResourceType,
			Amount:         d.Amount,
			Currency:       d.Currency,
			TimeHours:      d.TimeHours,
			TimeHourlyRate: d.TimeHourlyRate,
			GoodsQuantity:  d.GoodsQuantity,
			GoodsUnit:      d.GoodsUnit,
			GoodsItemName:  d.GoodsItemName,
			Description:    d.Description,
		})
	}
	return res
}

func toProjectDTO(p *project.Project) ProjectDTO {
	if p == nil {
		return ProjectDTO{}
	}
	mainCurrency := p.MainCurrency()
	if mainCurrency == "" {
		mainCurrency = "UAH"
	}
	rawParticipants := p.Participants()
	participants := make([]string, 0, len(rawParticipants))
	for _, part := range rawParticipants {
		participants = append(participants, part.DisplayName())
	}

	rawBudgets := p.TagBudgets()
	tagBudgets := make([]TagBudgetDTO, 0, len(rawBudgets))
	for _, b := range rawBudgets {
		tagBudgets = append(tagBudgets, TagBudgetDTO{
			Tag:             b.Tag(),
			Frequency:       b.Frequency().String(),
			TargetMonth:     b.TargetMonth(),
			ResourceType:    b.ResourceType().String(),
			Amount:          b.Amount(),
			Currency:        b.Currency(),
			HomeAmount:      b.HomeAmount(),
			TimeHours:       b.TimeHours(),
			TimeHourlyRate:  b.TimeHourlyRate(),
			GoodsQuantity:   b.GoodsQuantity(),
			GoodsUnit:       b.GoodsUnit(),
			GoodsItemName:   b.GoodsItemName(),
			Description:     b.Description(),
			AnnualTotal:     b.AnnualTotal(p.StartDate(), p.EndDate()),
			AnnualHomeTotal: b.AnnualHomeTotal(p.StartDate(), p.EndDate()),
		})
	}

	startDateStr := ""
	if !p.StartDate().IsZero() {
		startDateStr = p.StartDate().Format(time.RFC3339)
	}
	endDateStr := ""
	if !p.EndDate().IsZero() {
		endDateStr = p.EndDate().Format(time.RFC3339)
	}

	shareToken := ""
	if !p.IsShared() {
		shareToken = p.ShareToken().String()
	}

	dto := ProjectDTO{
		ProjectID:           p.ID().String(),
		Name:                p.Name(),
		Description:         p.Description(),
		StartDate:           startDateStr,
		EndDate:             endDateStr,
		ShareToken:          shareToken,
		IsShared:            p.IsShared(),
		MainCurrency:        mainCurrency,
		Participants:        participants,
		TagBudgets:          tagBudgets,
		AnnualPlannedBudget: p.AnnualPlannedBudget(),
	}
	if p.Owner() != nil {
		dto.Owner = OwnerDTO{
			PersonID:    p.Owner().PersonID.String(),
			DisplayName: p.Owner().DisplayName,
		}
	}
	return dto
}

type ProjectEndpoints struct {
	CreateProject       endpoint.Endpoint
	GetProject          endpoint.Endpoint
	AcceptShare         endpoint.Endpoint
	ListProjects        endpoint.Endpoint
	ShowProjectTotals   endpoint.Endpoint
	UpdateProject       endpoint.Endpoint
	UpdateProjectBudget endpoint.Endpoint
	RemoveProject       endpoint.Endpoint
}

func DecodeCreateProjectRequest(ctx context.Context, r *http.Request) (any, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	var req CreateProjectDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	startDate, err := parseOptionalDate(req.StartDate)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid start date", err)
	}

	endDate, err := parseOptionalDate(req.EndDate)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid end date", err)
	}
	var endPtr *time.Time
	if !endDate.IsZero() {
		endPtr = &endDate
	}

	return projectapp.CreateProject{
		ProjectID:    core.NewProjectID(),
		RequesterID:  personID,
		Name:         req.Name,
		Description:  req.Description,
		MainCurrency: req.MainCurrency,
		StartDate:    startDate,
		EndDate:      endPtr,
		TagBudgets:   toTagBudgetInputs(req.TagBudgets),
	}, nil
}

func makeCreateProjectEndpoint(app *projectapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(projectapp.CreateProject)

		if err := app.Commands.CreateProject.Handle(ctx, cmd); err != nil {
			return nil, err
		}

		proj, err := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
			ProjectID:   cmd.ProjectID,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			// Idempotent create by duplicate name: resolve existing owned project.
			proj, err = app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
				Name:        strings.TrimSpace(cmd.Name),
				RequesterID: cmd.RequesterID,
			})
			if err != nil {
				return nil, err
			}
		}

		return toProjectDTO(proj), nil
	}
}

func makeListProjectsEndpoint(app *projectapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(projectapp.ListProjects)

		projects, err := app.Queries.ListProjects.Handle(ctx, query)

		list := make([]ProjectDTO, 0, len(projects))
		for _, p := range projects {
			list = append(list, toProjectDTO(p))
		}

		return ListProjectsResponse{
			Projects: list,
			PaginationDTO: PaginationDTO{
				Limit:  query.Pagination.Limit,
				Offset: query.Pagination.Offset,
			},
		}, err
	}
}

func makeShowProjectTotalsEndpoint(app *projectapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(projectapp.GetProjectTotals)

		calculatedTotals, err := app.Queries.GetProjectTotals.Handle(ctx, query)
		if err != nil {
			return nil, err
		}

		totals := []TotalDTO{
			{
				Title:    "Total Invested",
				Amount:   calculatedTotals.TotalInvested,
				Currency: calculatedTotals.Currency,
			},
		}

		if calculatedTotals.HasActivity {
			totals = append(totals, TotalDTO{
				Title:    "Project Balance",
				Amount:   calculatedTotals.Balance,
				Currency: calculatedTotals.Currency,
			})
		}

		return ProjectTotalsDTO{
			Totals: totals,
		}, nil
	}
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

	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return projectapp.AcceptProjectShare{
		ShareToken:  shareToken,
		RequesterID: personID,
	}, nil
}

func DecodeGetProjectRequest(ctx context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return projectapp.GetProject{
		ProjectID:   projectUUID,
		RequesterID: personID,
	}, nil
}

func makeGetProjectEndpoint(app *projectapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		query := request.(projectapp.GetProject)
		proj, err := app.Queries.GetProject.Handle(ctx, query)
		if err != nil {
			return nil, err
		}

		return toProjectDTO(proj), nil
	}
}

func makeAcceptShareEndpoint(app *projectapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(projectapp.AcceptProjectShare)
		if err := app.Commands.AcceptProjectShare.Handle(ctx, cmd); err != nil {
			return nil, err
		}

		proj, err := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
			ShareToken:  cmd.ShareToken,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			return nil, err
		}

		return toProjectDTO(proj), nil
	}
}

func decodeUpdateProjectRequest(ctx context.Context, r *http.Request) (any, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req UpdateProjectDTO
	if err := decodeJSONBody(r, &req); err != nil && err != io.EOF {
		return nil, exceptions.NewValidationException("invalid json", err)
	}

	var startPtr *time.Time
	if req.StartDate != "" {
		st, err := parseOptionalDate(req.StartDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid start date", err)
		}
		startPtr = &st
	}

	var endPtr *time.Time
	if req.EndDate != "" {
		et, err := parseOptionalDate(req.EndDate)
		if err != nil {
			return nil, exceptions.NewValidationException("invalid end date", err)
		}
		endPtr = &et
	}

	var tagBudgetsPtr *[]project.TagBudgetInput
	if req.TagBudgets != nil {
		tb := toTagBudgetInputs(*req.TagBudgets)
		tagBudgetsPtr = &tb
	}

	return projectapp.UpdateProject{
		ProjectID:    projectUUID,
		RequesterID:  personID,
		Name:         req.Name,
		Description:  req.Description,
		MainCurrency: req.MainCurrency,
		StartDate:    startPtr,
		EndDate:      endPtr,
		TagBudgets:   tagBudgetsPtr,
	}, nil
}

func makeUpdateProjectEndpoint(app *projectapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(projectapp.UpdateProject)

		if err := app.Commands.UpdateProject.Handle(ctx, cmd); err != nil {
			return nil, err
		}

		proj, err := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
			ProjectID:   cmd.ProjectID,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			return nil, err
		}

		return toProjectDTO(proj), nil
	}
}

func decodeUpdateProjectBudgetRequest(ctx context.Context, r *http.Request) (any, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req UpdateProjectBudgetDTO
	if err := decodeJSONBody(r, &req); err != nil && err != io.EOF {
		return nil, exceptions.NewValidationException("invalid json", err)
	}

	return projectapp.UpdateProjectBudget{
		ProjectID:   projectUUID,
		RequesterID: personID,
		TagBudgets:  toTagBudgetInputs(req.TagBudgets),
	}, nil
}

func makeUpdateProjectBudgetEndpoint(app *projectapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd := request.(projectapp.UpdateProjectBudget)

		if err := app.Commands.UpdateProjectBudget.Handle(ctx, cmd); err != nil {
			return nil, err
		}

		proj, err := app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
			ProjectID:   cmd.ProjectID,
			RequesterID: cmd.RequesterID,
		})
		if err != nil {
			return nil, err
		}

		return toProjectDTO(proj), nil
	}
}

func decodeRemoveProjectRequest(ctx context.Context, r *http.Request) (any, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	vars := mux.Vars(r)
	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	return projectapp.RemoveProject{
		ProjectID:   projectUUID,
		RequesterID: personID,
	}, nil
}

func makeRemoveProjectEndpoint(app *projectapp.Application) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd, ok := request.(projectapp.RemoveProject)
		if !ok {
			return nil, exceptions.NewValidationException("invalid request", nil)
		}
		return nil, app.Commands.RemoveProject.Handle(ctx, cmd)
	}
}

func MakeProjectEndpoints(projectApp *projectapp.Application) (ProjectEndpoints, error) {
	if projectApp == nil {
		return ProjectEndpoints{}, nil
	}
	return ProjectEndpoints{
		CreateProject:       makeCreateProjectEndpoint(projectApp),
		GetProject:          makeGetProjectEndpoint(projectApp),
		AcceptShare:         makeAcceptShareEndpoint(projectApp),
		ListProjects:        makeListProjectsEndpoint(projectApp),
		ShowProjectTotals:   makeShowProjectTotalsEndpoint(projectApp),
		UpdateProject:       makeUpdateProjectEndpoint(projectApp),
		UpdateProjectBudget: makeUpdateProjectBudgetEndpoint(projectApp),
		RemoveProject:       makeRemoveProjectEndpoint(projectApp),
	}, nil
}
