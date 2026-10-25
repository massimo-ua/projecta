package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
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

	filter := project.ProjectCollectionFilter{
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

func decodeProjectTotalsRequest(_ context.Context, r *http.Request) (any, error) {
	vars := mux.Vars(r)

	projectID, ok := vars["project_id"]

	if !ok {
		return nil, exceptions.NewValidationException("missing project_id", nil)
	}

	projectUUID, err := core.ParseProjectID(projectID)

	if err != nil {
		return nil, exceptions.NewValidationException("invalid project_id", err)
	}

	return projectUUID, nil
}
