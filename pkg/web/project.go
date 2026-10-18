package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type CreateProjectDTO struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	MainCurrency string `json:"mainCurrency,omitempty"`
}

type OwnerDTO struct {
	PersonID    string `json:"person_id"`
	DisplayName string `json:"display_name"`
}

type ProjectDTO struct {
	ProjectID    string   `json:"project_id"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Owner        OwnerDTO `json:"owner"`
	ShareToken   string   `json:"share_token,omitempty"`
	IsShared     bool     `json:"is_shared,omitempty"`
	MainCurrency string   `json:"mainCurrency,omitempty"`
	Participants []string `json:"participants"`
}

type UpdateProjectDTO struct {
	Name         string `json:"name,omitempty"`
	Description  string `json:"description,omitempty"`
	MainCurrency string `json:"mainCurrency,omitempty"`
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
	dto := ProjectDTO{
		ProjectID:    p.ID().String(),
		Name:         p.Name(),
		Description:  p.Description(),
		ShareToken:   p.ShareToken().String(),
		IsShared:     p.IsShared(),
		MainCurrency: mainCurrency,
		Participants: participants,
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
	CreateProject     endpoint.Endpoint
	GetProject        endpoint.Endpoint
	AcceptShare       endpoint.Endpoint
	ListProjects      endpoint.Endpoint
	ShowProjectTotals endpoint.Endpoint
	UpdateProject     endpoint.Endpoint
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

	return project.CreateProjectCommand{
		PersonID:     personID,
		Name:         req.Name,
		Description:  req.Description,
		MainCurrency: req.MainCurrency,
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
		projectID := request.(uuid.UUID)

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

	return project.UpdateProjectCommand{
		ProjectID:    projectUUID,
		PersonID:     personID,
		Name:         req.Name,
		Description:  req.Description,
		MainCurrency: req.MainCurrency,
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

func MakeProjectEndpoints(
	projectService project.ProjectService,
) (ProjectEndpoints, error) {
	return ProjectEndpoints{
		CreateProject:     makeCreateProjectEndpoint(projectService),
		GetProject:        makeGetProjectEndpoint(projectService),
		AcceptShare:       makeAcceptShareEndpoint(projectService),
		ListProjects:      makeListProjectsEndpoint(projectService),
		ShowProjectTotals: makeShowProjectTotalsEndpoint(projectService),
		UpdateProject:     makeUpdateProjectEndpoint(projectService),
	}, nil
}
