package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

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
	mainCurrency := project.MainCurrency()
	if mainCurrency == "" {
		mainCurrency = "UAH"
	}
	participants := project.Participants()
	if participants == nil {
		participants = make([]projecta.Participant, 0)
	}
	dto := ProjectDTO{
		ProjectID:    project.ID().String(),
		Name:         project.Name(),
		Description:  project.Description(),
		ShareToken:   project.ShareToken().String(),
		IsShared:     project.IsShared(),
		MainCurrency: mainCurrency,
		Participants: participants,
	}
	if project.Owner() != nil {
		dto.Owner = OwnerDTO{
			PersonID:    project.Owner().PersonID.String(),
			DisplayName: project.Owner().DisplayName,
		}
	}
	return dto
}

type ProjectEndpoints struct {
	CreateProject           endpoint.Endpoint
	GetProject              endpoint.Endpoint
	AcceptShare             endpoint.Endpoint
	ListProjects            endpoint.Endpoint
	ShowProjectTotals       endpoint.Endpoint
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
	CreateInvestment        endpoint.Endpoint
	ListInvestments         endpoint.Endpoint
	GetInvestment           endpoint.Endpoint
	UpdateInvestment        endpoint.Endpoint
	RemoveInvestment        endpoint.Endpoint
	CreateBatchInvestments  endpoint.Endpoint
	ListTags                endpoint.Endpoint
	UpdateProject           endpoint.Endpoint
	ParseStatement          endpoint.Endpoint
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

func makeShowProjectTotalsEndpoint(
	projectSvc projecta.ProjectService,
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

		homeCurrency := proj.MainCurrency()
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
		var invAmounts []*money.Money
		var rootAssets []*asset.Asset
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
					invAmounts = append(invAmounts, e.Amount())
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

				hasAssets = true
				rootAssets = append(rootAssets, e)
			}

			if len(page.Elements()) < limit {
				next = false
			}

			offset += limit
		}

		calculatedTotals, err := asset.CalculateProjectTotals(invAmounts, rootAssets, homeCurrency, rateProvider)
		if err != nil {
			return nil, err
		}

		totals := make([]TotalDTO, 0)

		totals = append(totals, TotalDTO{
			Title:    "Total Invested",
			Amount:   calculatedTotals.TotalInvested,
			Currency: homeCurrency,
		})

		if hasAssets || hasInvestments {
			totals = append(totals, TotalDTO{
				Title:    "Project Balance",
				Amount:   calculatedTotals.TotalInvested - calculatedTotals.TotalValuation,
				Currency: homeCurrency,
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
	statementService projecta.StatementService,
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

	var epParseStatement endpoint.Endpoint
	if statementService != nil {
		epParseStatement = makeParseStatementEndpoint(statementService)
	}

	return ProjectEndpoints{
		CreateProject:           makeCreateProjectEndpoint(projectService),
		GetProject:              makeGetProjectEndpoint(projectService),
		AcceptShare:             makeAcceptShareEndpoint(projectService),
		ListProjects:            makeListProjectsEndpoint(projectService),
		ShowProjectTotals:       makeShowProjectTotalsEndpoint(projectService, assetService, rateProvider, invSvc),
		CreateAsset:             makeCreateAssetEndpoint(assetService, rateProvider),
		CreateAssetFromPayments: makeCreateAssetFromPaymentsEndpoint(assetService, rateProvider),
		RemoveAsset:             makeRemoveAssetEndpoint(assetService),
		ListAssets:              makeListAssetsEndpoint(assetService, rateProvider),
		UpdateAsset:             makeUpdateAssetEndpoint(assetService),
		GetAsset:                makeGetAssetEndpoint(assetService, rateProvider),
		LinkChildAsset:          makeLinkChildAssetEndpoint(assetService),
		UnlinkChildAsset:        makeUnlinkChildAssetEndpoint(assetService),
		AssignInvestments:       makeAssignInvestmentsEndpoint(assetService),
		GroupAssets:             makeGroupAssetsEndpoint(assetService, rateProvider),
		LinkChildren:            makeLinkChildrenEndpoint(assetService),
		CreateInvestment:        epCreateInv,
		ListInvestments:         epListInv,
		GetInvestment:           epGetInv,
		UpdateInvestment:        epUpdInv,
		RemoveInvestment:        epRemInv,
		CreateBatchInvestments:  epBatchInv,
		ListTags:                epListTags,
		UpdateProject:           makeUpdateProjectEndpoint(projectService),
		ParseStatement:          epParseStatement,
	}, nil
}
