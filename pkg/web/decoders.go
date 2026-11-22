package web

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	projectapp "gitlab.com/massimo-ua/projecta/internal/project/app"
)

const (
	maxJSONBodyBytes   int64 = 1 << 20 // 1 MiB
	maxPaginationLimit int   = 1000
)

func decodeJSONBody(r *http.Request, dst any) error {
	if r == nil || r.Body == nil {
		return io.EOF
	}
	return json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxJSONBodyBytes)).Decode(dst)
}

func parsePagination(r *http.Request, defaultLimit int) (core.Pagination, error) {
	limit := defaultLimit
	offset := 0

	q := r.URL.Query()
	if limitStr := q.Get("limit"); limitStr != "" {
		val, err := strconv.Atoi(limitStr)
		if err != nil || val <= 0 {
			return core.Pagination{}, exceptions.NewValidationException("invalid limit", err)
		}
		if val > maxPaginationLimit {
			val = maxPaginationLimit
		}
		limit = val
	}

	if offsetStr := q.Get("offset"); offsetStr != "" {
		val, err := strconv.Atoi(offsetStr)
		if err != nil || val < 0 {
			return core.Pagination{}, exceptions.NewValidationException("invalid offset", err)
		}
		offset = val
	}

	return core.Pagination{
		Limit:  limit,
		Offset: offset,
	}, nil
}

func decodeRegisterUser(_ context.Context, r *http.Request) (any, error) {
	var req RegisterUserDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}
	return req, nil
}

func decodeLoginUser(_ context.Context, r *http.Request) (any, error) {
	var req LoginDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}
	return req, nil
}

func decodeRefreshUserToken(_ context.Context, r *http.Request) (any, error) {
	var req RefreshTokenDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request", err)
	}
	return req, nil
}

func decodeListUsersRequest(_ context.Context, r *http.Request) (any, error) {
	return parsePagination(r, core.DefaultLimit)
}

type AssignRolesRequest struct {
	UserID core.PersonID
	Roles  []string
}

func decodeAssignRolesRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)
	userIDStr, ok := vars["user_id"]
	if !ok {
		return nil, exceptions.NewValidationException("missing user_id", nil)
	}

	userID, err := core.ParsePersonID(userIDStr)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid user_id", err)
	}

	var req AssignRolesDTO
	if err := decodeJSONBody(r, &req); err != nil {
		return nil, exceptions.NewValidationException("invalid request payload", err)
	}

	return AssignRolesRequest{
		UserID: userID,
		Roles:  req.Roles,
	}, nil
}

func decodeListProjectsRequest(ctx context.Context, r *http.Request) (any, error) {
	pagination, err := parsePagination(r, core.DefaultLimit)
	if err != nil {
		return nil, err
	}

	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return projectapp.ListProjects{
		Pagination:  pagination,
		Name:        r.URL.Query().Get("name"),
		RequesterID: personID,
	}, nil
}

func decodeProjectTotalsRequest(ctx context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]
	if !ok {
		return nil, exceptions.NewValidationException("missing project_id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)
	if err != nil {
		return nil, exceptions.NewValidationException("invalid project_id", err)
	}

	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
	}

	return projectapp.GetProjectTotals{
		ProjectID:   projectUUID,
		RequesterID: personID,
	}, nil
}
