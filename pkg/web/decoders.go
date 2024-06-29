package web

import (
	"context"
	"encoding/json"
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
		limit = core.DEFAULT_LIMIT
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
		limit = core.DEFAULT_LIMIT
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
