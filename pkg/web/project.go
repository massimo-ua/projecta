package web

import (
	"context"
	"encoding/json"
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
)

type TagBudgetDTO struct {
	Tag            string   `json:"tag"`
	Frequency      string   `json:"frequency"`
	TargetMonth    int      `json:"target_month,omitempty"`
	ResourceType   string   `json:"resource_type"`
	Amount         int64    `json:"amount"`
	TimeHours      *float64 `json:"time_hours,omitempty"`
	TimeHourlyRate *int64   `json:"time_hourly_rate,omitempty"`
	GoodsQuantity  *float64 `json:"goods_quantity,omitempty"`
	GoodsUnit      string   `json:"goods_unit,omitempty"`
	GoodsItemName  string   `json:"goods_item_name,omitempty"`
	Description    string   `json:"description,omitempty"`
	AnnualTotal    int64    `json:"annual_total"`
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
			Tag:            b.Tag(),
			Frequency:      b.Frequency().String(),
			TargetMonth:    b.TargetMonth(),
			ResourceType:   b.ResourceType().String(),
			Amount:         b.Amount(),
			TimeHours:      b.TimeHours(),
			TimeHourlyRate: b.TimeHourlyRate(),
			GoodsQuantity:  b.GoodsQuantity(),
			GoodsUnit:      b.GoodsUnit(),
			GoodsItemName:  b.GoodsItemName(),
			Description:    b.Description(),
			AnnualTotal:    b.AnnualTotal(p.StartDate(), p.EndDate()),
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

	dto := ProjectDTO{
		ProjectID:           p.ID().String(),
		Name:                p.Name(),
		Description:         p.Description(),
		StartDate:           startDateStr,
		EndDate:             endDateStr,
		ShareToken:          p.ShareToken().String(),
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
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

	return project.CreateProjectCommand{
		PersonID:     personID,
		Name:         req.Name,
		Description:  req.Description,
		MainCurrency: req.MainCurrency,
		StartDate:    startDate,
		EndDate:      endPtr,
		TagBudgets:   toTagBudgetInputs(req.TagBudgets),
	}, nil
}

func makeCreateProjectEndpoint(svc project.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(project.CreateProjectCommand)

		project, err := svc.Create(ctx, command)
		if err != nil {
			return nil, err
		}

		return toProjectDTO(project), nil
	}
}

func makeListProjectsEndpoint(svc project.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(project.ProjectCollectionFilter)

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

func makeShowProjectTotalsEndpoint(
	projectSvc project.ProjectService,
) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		projectID := request.(core.ProjectID)

		calculatedTotals, err := projectSvc.GetTotals(ctx, projectID)
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

type AcceptShareCommand struct {
	ShareToken uuid.UUID
	PersonID   core.PersonID
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

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	return project.ProjectFilter{
		ProjectID: projectUUID,
	}, nil
}

func makeGetProjectEndpoint(svc project.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(project.ProjectFilter)
		project, err := svc.FindOne(ctx, filter)
		if err != nil {
			return nil, err
		}

		return toProjectDTO(project), nil
	}
}

func makeAcceptShareEndpoint(svc project.ProjectService) endpoint.Endpoint {
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
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

	return project.UpdateProjectCommand{
		ProjectID:    projectUUID,
		PersonID:     personID,
		Name:         req.Name,
		Description:  req.Description,
		MainCurrency: req.MainCurrency,
		StartDate:    startPtr,
		EndDate:      endPtr,
		TagBudgets:   tagBudgetsPtr,
	}, nil
}

func makeUpdateProjectEndpoint(svc project.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(project.UpdateProjectCommand)

		project, err := svc.Update(ctx, command)
		if err != nil {
			return nil, err
		}

		return toProjectDTO(project), nil
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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && err != io.EOF {
		return nil, exceptions.NewValidationException("invalid json", err)
	}

	return project.UpdateProjectBudgetCommand{
		ProjectID:  projectUUID,
		PersonID:   personID,
		TagBudgets: toTagBudgetInputs(req.TagBudgets),
	}, nil
}

func makeUpdateProjectBudgetEndpoint(svc project.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command := request.(project.UpdateProjectBudgetCommand)

		project, err := svc.UpdateBudget(ctx, command)
		if err != nil {
			return nil, err
		}

		return toProjectDTO(project), nil
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

	return project.RemoveProjectCommand{
		ProjectID: projectUUID,
		PersonID:  personID,
	}, nil
}

func makeRemoveProjectEndpoint(svc project.ProjectService) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		cmd, ok := request.(project.RemoveProjectCommand)
		if !ok {
			return nil, exceptions.NewValidationException("invalid request", nil)
		}
		err := svc.Remove(ctx, cmd)
		return nil, err
	}
}

func MakeProjectEndpoints(
	projectService project.ProjectService,
) (ProjectEndpoints, error) {
	return ProjectEndpoints{
		CreateProject:       makeCreateProjectEndpoint(projectService),
		GetProject:          makeGetProjectEndpoint(projectService),
		AcceptShare:         makeAcceptShareEndpoint(projectService),
		ListProjects:        makeListProjectsEndpoint(projectService),
		ShowProjectTotals:   makeShowProjectTotalsEndpoint(projectService),
		UpdateProject:       makeUpdateProjectEndpoint(projectService),
		UpdateProjectBudget: makeUpdateProjectBudgetEndpoint(projectService),
		RemoveProject:       makeRemoveProjectEndpoint(projectService),
	}, nil
}
