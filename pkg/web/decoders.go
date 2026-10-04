package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
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

func decodeListUsersRequest(_ context.Context, r *http.Request) (any, error) {
	var err error
	var limit, offset int
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

	return core.Pagination{
		Limit:  limit,
		Offset: offset,
	}, nil
}

type AssignRolesRequest struct {
	UserID uuid.UUID
	Roles  []string
}

func decodeAssignRolesRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	userIDStr, ok := vars["user_id"]
	if !ok {
		return nil, exceptions.NewValidationException("missing user_id", nil)
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid user_id", err)
	}

	var req AssignRolesDTO
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid request payload", err)
	}

	return AssignRolesRequest{
		UserID: userID,
		Roles:  req.Roles,
	}, nil
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

	fromDateStr := r.URL.Query().Get("from_date")
	toDateStr := r.URL.Query().Get("to_date")

	fromDate, err := parseFilterDate(fromDateStr, false)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid from_date", err)
	}

	toDate, err := parseFilterDate(toDateStr, true)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid to_date", err)
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
		FromDate:   fromDate,
		ToDate:     toDate,
	}

	return filter, nil
}

func parseFilterDate(s string, endOfDay bool) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err == nil {
		return t, nil
	}
	t, err = time.Parse("2006-01-02", s)
	if err == nil {
		if endOfDay {
			return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location()), nil
		}
		return t, nil
	}
	return time.Time{}, exceptions.NewValidationException("invalid date format", err)
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
