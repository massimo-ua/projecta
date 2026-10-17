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
	"gitlab.com/massimo-ua/projecta/internal/investment"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/project"
	"gitlab.com/massimo-ua/projecta/internal/statement"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
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
	case exceptions.Forbidden:
		return http.StatusForbidden
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
	invitationService people.InvitationService,
	projectService project.ProjectService,
	statementService statement.Service,
	assetService asset.Service,
	rateProvider currency.CurrencyRateProvider,
	investmentService ...investment.Service,
) (http.Handler, error) {
	r := mux.NewRouter()
	createSwaggerHandler(r)
	peopleEndpoints, err := MakeUserEndpoints(peopleService, authService)
	invitationEndpoints := MakeInvitationEndpoints(invitationService)
	projectEndpoints, err := MakeProjectEndpoints(
		projectService,
		assetService,
		rateProvider,
		investmentService...,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create purchase endpoints: %s", err.Error())
	}

	assetEndpoints := MakeAssetEndpoints(assetService)
	statementEndpoints := MakeStatementEndpoints(statementService)

	var invSvc investment.Service
	if len(investmentService) > 0 {
		invSvc = investmentService[0]
	}
	investmentEndpoints := MakeInvestmentEndpoints(invSvc)

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

	r.Methods(http.MethodPut).Path("/profile").Handler(ht.NewServer(
		loggedInOnly(peopleEndpoints.UpdateProfile),
		decodeUpdateProfileRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/refresh").Handler(ht.NewServer(
		peopleEndpoints.RefreshToken,
		decodeRefreshUserToken,
		encodeJSON(http.StatusOK),
		options...,
	))

	r.Methods(http.MethodGet).Path("/users").Handler(ht.NewServer(
		requireRole(string(people.RoleAdministrator), peopleEndpoints.ListUsers),
		decodeListUsersRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPut).Path("/users/{user_id}/roles").Handler(ht.NewServer(
		requireRole(string(people.RoleAdministrator), peopleEndpoints.AssignRoles),
		decodeAssignRolesRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/invitations").Handler(ht.NewServer(
		requireRole(string(people.RoleAdministrator), invitationEndpoints.CreateInvitation),
		decodeCreateInvitationRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/invitations").Handler(ht.NewServer(
		requireRole(string(people.RoleAdministrator), invitationEndpoints.ListInvitations),
		decodeListInvitationsRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodDelete).Path("/invitations/{invitation_id}").Handler(ht.NewServer(
		requireRole(string(people.RoleAdministrator), invitationEndpoints.DeleteInvitation),
		decodeDeleteInvitationRequest,
		encodeJSON(http.StatusNoContent),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/invitations/code/{code}").Handler(ht.NewServer(
		invitationEndpoints.ValidateInvitation,
		decodeValidateInvitationRequest,
		encodeJSON(http.StatusOK),
		options...,
	))

	r.Methods(http.MethodPost).Path("/projects").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), projectEndpoints.CreateProject),
		DecodeCreateProjectRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), projectEndpoints.ListProjects),
		decodeListProjectsRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), projectEndpoints.GetProject),
		DecodeGetProjectRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPatch).Path("/projects/{project_id}").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), projectEndpoints.UpdateProject),
		decodeUpdateProjectRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/projects/share/{share_token}").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), projectEndpoints.AcceptShare),
		DecodeAcceptShareRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}/totals").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), projectEndpoints.ShowProjectTotals),
		decodeProjectTotalsRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	if statementEndpoints.ParseStatement != nil {
		r.Methods(http.MethodPost).Path("/projects/{project_id}/payments/parse-statement").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), statementEndpoints.ParseStatement),
			decodeParseStatementRequest,
			encodeJSON(http.StatusOK),
			withAuth...,
		))

		r.Methods(http.MethodPost).Path("/projects/{project_id}/statements/parse").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), statementEndpoints.ParseStatement),
			decodeParseStatementRequest,
			encodeJSON(http.StatusOK),
			withAuth...,
		))
	}

	r.Methods(http.MethodPost).Path("/projects/{project_id}/assets/from-payments").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), assetEndpoints.CreateAssetFromPayments),
		decodeCreateAssetFromPaymentsRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodPost).Path("/projects/{project_id}/assets").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), assetEndpoints.CreateAsset),
		decodeCreateAssetRequest,
		encodeJSON(http.StatusCreated),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}/assets").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), assetEndpoints.ListAssets),
		decodeListAssetsRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	r.Methods(http.MethodDelete).Path("/projects/{project_id}/assets/{asset_id}").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), assetEndpoints.RemoveAsset),
		decodeProjectResourceRemoveCommand("project_id", "asset_id"),
		encodeJSON(http.StatusNoContent),
		withAuth...,
	))

	r.Methods(http.MethodPut).Path("/projects/{project_id}/assets/{asset_id}").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), assetEndpoints.UpdateAsset),
		decodeUpdateAssetRequest,
		encodeJSON(http.StatusNoContent),
		withAuth...,
	))

	r.Methods(http.MethodGet).Path("/projects/{project_id}/assets/{asset_id}").Handler(ht.NewServer(
		requireRole(string(people.RoleUser), assetEndpoints.GetAsset),
		decodeGetAssetRequest,
		encodeJSON(http.StatusOK),
		withAuth...,
	))

	if assetEndpoints.LinkChildAsset != nil {
		r.Methods(http.MethodPost).Path("/projects/{project_id}/assets/{asset_id}/children").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), assetEndpoints.LinkChildAsset),
			decodeLinkChildAssetRequest,
			encodeJSON(http.StatusOK),
			withAuth...,
		))

		r.Methods(http.MethodDelete).Path("/projects/{project_id}/assets/{asset_id}/children/{child_asset_id}").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), assetEndpoints.UnlinkChildAsset),
			decodeUnlinkChildAssetRequest,
			encodeJSON(http.StatusNoContent),
			withAuth...,
		))
	}

	if assetEndpoints.AssignInvestments != nil {
		r.Methods(http.MethodPost).Path("/projects/{project_id}/assets/{asset_id}/investments").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), assetEndpoints.AssignInvestments),
			decodeAssignInvestmentsRequest,
			encodeJSON(http.StatusOK),
			withAuth...,
		))
	}

	if assetEndpoints.GroupAssets != nil {
		r.Methods(http.MethodPost).Path("/projects/{project_id}/assets/group").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), assetEndpoints.GroupAssets),
			decodeGroupAssetsRequest,
			encodeJSON(http.StatusCreated),
			withAuth...,
		))
	}

	if assetEndpoints.LinkChildren != nil {
		r.Methods(http.MethodPost).Path("/projects/{project_id}/assets/{asset_id}/children/batch").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), assetEndpoints.LinkChildren),
			decodeLinkChildrenRequest,
			encodeJSON(http.StatusOK),
			withAuth...,
		))
	}

	if investmentEndpoints.CreateInvestment != nil {
		r.Methods(http.MethodPost).Path("/projects/{project_id}/investments").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), investmentEndpoints.CreateInvestment),
			decodeCreateInvestmentRequest,
			encodeJSON(http.StatusCreated),
			withAuth...,
		))

		r.Methods(http.MethodGet).Path("/projects/{project_id}/investments").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), investmentEndpoints.ListInvestments),
			decodeListInvestmentsRequest,
			encodeJSON(http.StatusOK),
			withAuth...,
		))

		r.Methods(http.MethodGet).Path("/projects/{project_id}/investments/{investment_id}").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), investmentEndpoints.GetInvestment),
			decodeGetInvestmentRequest,
			encodeJSON(http.StatusOK),
			withAuth...,
		))

		r.Methods(http.MethodPut).Path("/projects/{project_id}/investments/{investment_id}").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), investmentEndpoints.UpdateInvestment),
			decodeUpdateInvestmentRequest,
			encodeJSON(http.StatusNoContent),
			withAuth...,
		))

		r.Methods(http.MethodDelete).Path("/projects/{project_id}/investments/{investment_id}").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), investmentEndpoints.RemoveInvestment),
			decodeProjectResourceRemoveCommand("project_id", "investment_id"),
			encodeJSON(http.StatusNoContent),
			withAuth...,
		))

		r.Methods(http.MethodPost).Path("/projects/{project_id}/investments/batch").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), investmentEndpoints.CreateBatchInvestments),
			decodeCreateBatchInvestmentsRequest,
			encodeJSON(http.StatusCreated),
			withAuth...,
		))

		r.Methods(http.MethodGet).Path("/projects/{project_id}/tags").Handler(ht.NewServer(
			requireRole(string(people.RoleUser), investmentEndpoints.ListTags),
			decodeListTagsRequest,
			encodeJSON(http.StatusOK),
			withAuth...,
		))
	}

	return r, nil
}
