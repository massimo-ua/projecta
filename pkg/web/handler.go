package web

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	ht "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
	"net/http"
)

func errorCodeToHttpStatus(e exceptions.Exception) int {
	switch e.Code {
	case exceptions.NotFound:
		return http.StatusNotFound
	case exceptions.ValidationFailed:
		return http.StatusBadRequest
	case exceptions.Internal:
		return http.StatusInternalServerError
	case exceptions.Unauthorized:
		return http.StatusUnauthorized
	default:
		return http.StatusInternalServerError
	}
}

func encodeErrorResponse(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	var exception exceptions.Exception
	if ok := errors.As(err, &exception); ok {
		w.WriteHeader(errorCodeToHttpStatus(exception))
		_ = json.NewEncoder(w).Encode(exception)
		return
	}

	w.WriteHeader(http.StatusInternalServerError)
	_ = json.NewEncoder(w).Encode(exceptions.NewApplicationError("Internal Error", exceptions.Internal, err))

}

func MakeHTTPHandler(
	peopleService people.UserService,
	authTokenProvider core.AuthTokenProvider,
	authService people.AuthService,
	projectService projecta.ProjectService,
	categoryService projecta.CategoryService,
	typeService projecta.TypeService,
	expenseService projecta.PaymentService,
	assetService asset.Service,
) (http.Handler, error) {
	r := mux.NewRouter()
	createSwaggerHandler(r)
	peopleEndpoints, err := MakeCustomerEndpoints(peopleService, authService)
	projectEndpoints, err := MakeProjectEndpoints(
		projectService,
		categoryService,
		typeService,
		expenseService,
		assetService,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create purchase endpoints: %s", err.Error())
	}

	options := []ht.ServerOption{
		// TODO: add logging
		// ht.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		ht.ServerErrorEncoder(encodeErrorResponse),
	}

	withAuth := append(options, ht.ServerBefore(jwtMiddleware(authTokenProvider)))

	r.Methods(http.MethodPost).Path("/register").Handler(ht.NewServer(
		peopleEndpoints.Register,
		decodeRegisterUser,
		encodeJSON(http.StatusCreated),
		options...,
	))

	r.Methods(http.MethodPost).Path("/login").Handler(ht.NewServer(
		peopleEndpoints.Login,
		decodeLoginUser,
		encodeJSON(http.StatusOK),
		options...,
	))

	r.Methods(http.MethodGet).Path("/profile").Handler(ht.NewServer(
		loggedInOnly(peopleEndpoints.Profile),
		decodeProfileRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/refresh").Handler(ht.NewServer(
		peopleEndpoints.RefreshToken,
		decodeRefreshUserToken,
		encodeJSON(http.StatusOK),
		options...,
	))

	r.Methods(http.MethodPost).Path("/projects").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.CreateProject),
		DecodeCreateProjectRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.ListProjects),
		decodeListProjectsRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/projects/{project_id}/categories").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.CreateCategory),
		DecodeCreateCategoryRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}/categories").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.ListCategories),
		decodeListCategoriesRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/projects/{project_id}/types").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.CreateType),
		DecodeCreateTypeRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}/types").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.ListTypes),
		decodeListTypesRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodDelete).Path("/projects/{project_id}/types/{type_id}").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.RemoveType),
		decodeProjectResourceRemoveCommand("project_id", "type_id"),
		encodeJSON(http.StatusNoContent),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}/totals").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.ShowProjectTotals),
		decodeProjectTotalsRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/projects/{project_id}/payments").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.CreatePayment),
		DecodeCreatePaymentRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}/payments").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.ListPayments),
		decodeListPaymentsRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodDelete).Path("/projects/{project_id}/payments/{payment_id}").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.RemovePayment),
		decodeProjectResourceRemoveCommand("project_id", "payment_id"),
		encodeJSON(http.StatusNoContent),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/projects/{project_id}/assets").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.CreateAsset),
		decodeCreateAssetRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}/assets").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.ListAssets),
		decodeListAssetsRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodDelete).Path("/projects/{project_id}/assets/{asset_id}").Handler(ht.NewServer(
		loggedInOnly(projectEndpoints.RemoveAsset),
		decodeProjectResourceRemoveCommand("project_id", "asset_id"),
		encodeJSON(http.StatusNoContent),
		withAuth...,
	))

	return r, nil
}
