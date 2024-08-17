package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"net/http"
	"strconv"
)

func decodeRegisterUser(_ context.Context, r *http.Request) (any, error) {
	var req RegisterUserDTO
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeLoginUser(_ context.Context, r *http.Request) (any, error) {
	var req LoginDTO
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeRefreshUserToken(_ context.Context, r *http.Request) (any, error) {
	var req RefreshTokenDTO
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

func decodeListProjectsRequest(_ context.Context, r *http.Request) (any, error) {
	var err error
	var limit, offset int
	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	name := r.URL.Query().Get("name")

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

	filter := projecta.ProjectCollectionFilter{
		Pagination: core.Pagination{
			Limit:  limit,
			Offset: offset,
		},
	}

	if name != "" {
		filter.Name = name
	}

	return filter, nil
}

func decodeListTypesRequest(_ context.Context, r *http.Request) (any, error) {
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

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	name := r.URL.Query().Get("name")

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

	filter := projecta.TypeCollectionFilter{
		Pagination: core.Pagination{
			Limit:  limit,
			Offset: offset,
		},
		ProjectID: projectUUID,
	}

	if name != "" {
		filter.Name = name
	}

	return filter, nil
}

func decodeListCategoriesRequest(_ context.Context, r *http.Request) (any, error) {
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

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	name := r.URL.Query().Get("name")

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

	filter := projecta.CategoryCollectionFilter{
		Pagination: core.Pagination{
			Limit:  limit,
			Offset: offset,
		},
		ProjectID: projectUUID,
	}

	if name != "" {
		filter.Name = name
	}

	return filter, nil
}

func decodeListPaymentsRequest(_ context.Context, r *http.Request) (any, error) {
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

	offsetStr := r.URL.Query().Get("offset")
	limitStr := r.URL.Query().Get("limit")
	categoryIDStr := r.URL.Query().Get("category_id")
	typeIDStr := r.URL.Query().Get("type_id")
	sortBy := r.URL.Query().Get("order_by")
	order := core.ToOrder(r.URL.Query().Get("order"))

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

	var categoryID uuid.UUID
	var typeID uuid.UUID

	if categoryIDStr != "" {
		categoryID, err = uuid.Parse(categoryIDStr)

		if err != nil {
			return nil, exceptions.NewValidationException("invalid category_id", err)
		}
	}

	if typeIDStr != "" {
		typeID, err = uuid.Parse(typeIDStr)

		if err != nil {
			return nil, exceptions.NewValidationException("invalid type_id", err)
		}
	}

	filter := projecta.PaymentCollectionFilter{
		Pagination: core.Pagination{
			Limit:  limit,
			Offset: offset,
		},
		Sorting: core.Sorting{
			OrderBy: sortBy,
			Order:   order,
		},
		ProjectID:  projectUUID,
		CategoryID: categoryID,
		TypeID:     typeID,
	}

	return filter, nil
}

func decodeProjectTotalsRequest(_ context.Context, r *http.Request) (any, error) {
	var err error
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("missing project_id", nil)
	}

	projectUUID, err := uuid.Parse(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project_id", err)
	}

	return projectUUID, nil
}

func decodeProjectResourceRemoveCommand(projectIDKey string, resourceIDKey string) func(context.Context, *http.Request) (any, error) {
	return func(ctx context.Context, r *http.Request) (any, error) {
		var err error
		vars := mux.Vars(r)

		projectID, ok := vars[projectIDKey]

		if !ok {
			return nil, exceptions.NewValidationException(fmt.Sprintf("missing %s", projectIDKey), nil)
		}

		projectUUID, err := uuid.Parse(projectID)

		if err != nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("invalid %s", projectIDKey), err)
		}

		resourceID, ok := vars[resourceIDKey]

		if !ok {
			return nil, exceptions.NewValidationException(fmt.Sprintf("missing %s", resourceIDKey), nil)
		}

		resourceUUID, err := uuid.Parse(resourceID)

		if err != nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("invalid %s", resourceIDKey), err)
		}

		return projecta.RemoveProjectResourceCommand{
			ProjectID:  projectUUID,
			ResourceID: resourceUUID,
		}, nil
	}
}
