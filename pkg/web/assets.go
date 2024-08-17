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
	"strconv"
	"time"
)

type AssetDTO struct {
	AssetID     string      `json:"asset_id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Price       int64       `json:"price"`
	Currency    string      `json:"currency"`
	AcquiredAt  string      `json:"acquired_at"`
	Owner       OwnerDTO    `json:"owner"`
	Project     ProjectDTO  `json:"project"`
	Type        TypeDTO     `json:"type"`
	Category    CategoryDTO `json:"category"`
}

type CreateAssetDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	TypeID      string `json:"type_id"`
	Price       int64  `json:"price"`
	Currency    string `json:"currency"`
	AcquiredAt  string `json:"acquired_at,omitempty"`
	WithPayment bool   `json:"with_payment"`
}

type ListAssetsResponse struct {
	Assets []AssetDTO `json:"assets"`
	PaginationDTO
}

func decodeCreateAssetRequest(_ context.Context, r *http.Request) (interface{}, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("invalid project id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project id", err)
	}

	var req CreateAssetDTO
	err = json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}

	if req.Name == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	typeUUID, err := uuid.Parse(req.TypeID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid type id", err)
	}

	if req.Price <= 0 {
		return nil, exceptions.NewValidationException("price must be greater than 0", nil)
	}

	if req.Currency == "" {
		return nil, exceptions.NewValidationException("currency is required", nil)
	}

	price := money.New(req.Price, req.Currency)

	date, err := time.Parse(time.RFC3339, req.AcquiredAt)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid acquired at date", err)
	}

	return asset.CreateAssetCommand{
		Name:        req.Name,
		Description: req.Description,
		ProjectID:   projectUUID,
		TypeID:      typeUUID,
		Price:       price,
		AcquiredAt:  date,
		WithPayment: req.WithPayment,
	}, nil
}

func decodeListAssetsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var err error
	var limit, offset int
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("missing project_id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project_id", err)
	}

	typeID := r.URL.Query().Get("type_id")
	var typeUUID uuid.UUID

	if typeID != "" {
		typeUUID, err = uuid.Parse(typeID)

		if err != nil {
			return nil, exceptions.NewValidationException("invalid type_id", err)
		}
	}

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")

	if limitStr != "" {
		limit, err = strconv.Atoi(limitStr)

		if err != nil {
			return nil, exceptions.NewValidationException("invalid limit", err)
		}
	} else {
		limit = core.DefaultLimit
	}

	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)

		if err != nil {
			return nil, exceptions.NewValidationException("invalid offset", err)
		}
	}

	orderBy := r.URL.Query().Get("order_by")

	order := core.ToOrder(r.URL.Query().Get("order"))

	filter := asset.CollectionFilter{
		ProjectID: projectUUID,
		Name:      r.URL.Query().Get("name"),
		TypeID:    typeUUID,
		Pagination: core.Pagination{
			Limit:  limit,
			Offset: offset,
		},
		Sorting: core.Sorting{
			OrderBy: orderBy,
			Order:   order,
		},
	}

	return filter, nil
}

func makeCreateAssetEndpoint(s asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		cmd := request.(asset.CreateAssetCommand)
		a, err := s.Create(ctx, cmd)

		if err != nil {
			return nil, err
		}

		owner := OwnerDTO{
			PersonID:    a.Owner.PersonID.String(),
			DisplayName: a.Owner.DisplayName,
		}

		category := CategoryDTO{
			CategoryID:  a.Type.Category.ID.String(),
			Name:        a.Type.Category.Name,
			Description: a.Type.Category.Description,
		}

		return AssetDTO{
			AssetID:     a.ID.String(),
			Name:        a.Name,
			Description: a.Description,
			Price:       a.Price.Amount(),
			Currency:    a.Price.Currency().Code,
			AcquiredAt:  a.AcquiredAt.Format(time.RFC3339),
			Owner:       owner,
			Project: ProjectDTO{
				ProjectID:   a.Project.ProjectID.String(),
				Name:        a.Project.Name,
				Description: a.Project.Description,
				Owner:       owner,
			},
			Type: TypeDTO{
				TypeID:      a.Type.ID.String(),
				Name:        a.Type.Name,
				Description: a.Type.Description,
				Category: TypeCategoryDTO{
					CategoryID: category.CategoryID,
					Name:       category.Name,
				},
			},
			Category: category,
		}, nil
	}
}

func makeRemoveAssetEndpoint(svc asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		command, ok := request.(projecta.RemoveProjectResourceCommand)

		if !ok {
			return nil, exceptions.NewValidationException("invalid request", nil)
		}

		err := svc.Remove(ctx, asset.RemoveAssetCommand{
			AssetID:   command.ResourceID,
			ProjectID: command.ProjectID,
		})

		return nil, err
	}
}

func makeListAssetsEndpoint(svc asset.Service) endpoint.Endpoint {
	return func(ctx context.Context, request any) (any, error) {
		filter := request.(asset.CollectionFilter)

		collection, err := svc.Find(ctx, filter)

		if err != nil {
			return nil, err
		}

		var list []AssetDTO = make([]AssetDTO, 0)

		for _, e := range collection.Elements() {
			owner := OwnerDTO{
				PersonID:    e.Owner.PersonID.String(),
				DisplayName: e.Owner.DisplayName,
			}

			category := CategoryDTO{
				CategoryID:  e.Type.Category.ID.String(),
				Name:        e.Type.Category.Name,
				Description: e.Type.Category.Description,
			}

			list = append(list, AssetDTO{
				AssetID: e.ID.String(),
				Project: ProjectDTO{
					ProjectID:   e.Project.ProjectID.String(),
					Name:        e.Project.Name,
					Description: e.Project.Description,
					Owner:       owner,
				},
				Owner: owner,
				Type: TypeDTO{
					TypeID:      e.Type.ID.String(),
					Name:        e.Type.Name,
					Description: e.Type.Description,
					Category: TypeCategoryDTO{
						CategoryID: category.CategoryID,
						Name:       category.Name,
					},
				},
				Category:    category,
				Name:        e.Name,
				Description: e.Description,
				Price:       e.Price.Amount(),
				Currency:    e.Price.Currency().Code,
				AcquiredAt:  e.AcquiredAt.Format(time.RFC3339),
			})
		}

		return ListAssetsResponse{
			Assets: list,
			PaginationDTO: PaginationDTO{
				Limit:  filter.Limit,
				Offset: filter.Offset,
				Total:  collection.Total(),
			},
		}, err
	}
}
