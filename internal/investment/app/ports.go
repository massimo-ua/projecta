package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
)

const (
	failedToCreateInvestment = "failed to create investment"
	failedToFindInvestment   = "failed to find investment"
	failedToUpdateInvestment = "failed to update investment"
)

// AssetDescriptor represents the narrow read view of an asset needed by the investment context.
type AssetDescriptor struct {
	ID        core.AssetID
	ProjectID core.ProjectID
	Name      string
}

// AssetReader is a consumer-owned port providing read access to assets.
type AssetReader interface {
	FindAsset(ctx context.Context, projectID core.ProjectID, assetID core.AssetID) (AssetDescriptor, error)
}

// ProjectDescriptor represents the narrow read view of a project needed by the investment context.
type ProjectDescriptor struct {
	ID           core.ProjectID
	MainCurrency string
}

// ProjectReader is a consumer-owned port verifying project access and reading project currency.
type ProjectReader interface {
	FindProject(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) (ProjectDescriptor, error)
}

// ContributorReader is a consumer-owned port providing read access to contributors.
type ContributorReader interface {
	FindContributor(ctx context.Context, personID core.PersonID) (investment.ContributorRef, error)
}

// TagsReadModel is a narrow read-model port returning distinct project tags across resources.
type TagsReadModel interface {
	FindTags(ctx context.Context, projectID core.ProjectID) ([]string, error)
}

type AssetAllocationInput struct {
	AssetID         core.AssetID
	SharePercentage float64
}

func verifyProjectAccess(ctx context.Context, projects ProjectReader, projectID core.ProjectID, requesterID core.PersonID) (ProjectDescriptor, error) {
	if requesterID.IsNil() {
		return ProjectDescriptor{}, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	if projectID.IsNil() {
		return ProjectDescriptor{}, exceptions.NewValidationException("project id is required", nil)
	}
	proj, err := projects.FindProject(ctx, projectID, requesterID)
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return ProjectDescriptor{}, exceptions.NewNotFoundException("project not found", err)
		}
		return ProjectDescriptor{}, preserveOrWrapError(err, "failed to verify project access")
	}
	if proj.MainCurrency == "" {
		proj.MainCurrency = "UAH"
	}
	return proj, nil
}

func resolveAllocations(
	ctx context.Context,
	assets AssetReader,
	projectID core.ProjectID,
	primaryID core.AssetID,
	inputs []AssetAllocationInput,
) ([]investment.InvestmentAssetLink, core.AssetID, string, error) {
	if len(inputs) > 1 {
		allocs := make([]investment.InvestmentAssetLink, 0, len(inputs))
		var primaryName string
		for _, aInput := range inputs {
			if aInput.AssetID.IsNil() {
				return nil, core.NilAssetID, "", exceptions.NewValidationException("asset id is required for each allocation", nil)
			}
			ast, err := assets.FindAsset(ctx, projectID, aInput.AssetID)
			if err != nil {
				return nil, core.NilAssetID, "", exceptions.NewValidationException(fmt.Sprintf("asset %s not found in project", aInput.AssetID), err)
			}
			if primaryName == "" {
				primaryName = ast.Name
			}
			allocs = append(allocs, investment.InvestmentAssetLink{
				AssetID:         ast.ID,
				AssetName:       ast.Name,
				SharePercentage: aInput.SharePercentage,
			})
		}
		return allocs, inputs[0].AssetID, primaryName, nil
	}

	targetID := primaryID
	if len(inputs) == 1 {
		targetID = inputs[0].AssetID
	}
	if targetID.IsNil() {
		return nil, core.NilAssetID, "", exceptions.NewValidationException("asset id or asset allocations are required", nil)
	}

	ast, err := assets.FindAsset(ctx, projectID, targetID)
	if err != nil {
		return nil, core.NilAssetID, "", exceptions.NewValidationException("asset not found in project", err)
	}
	allocs := []investment.InvestmentAssetLink{
		{
			AssetID:         ast.ID,
			AssetName:       ast.Name,
			SharePercentage: 100.0,
		},
	}
	return allocs, ast.ID, ast.Name, nil
}

func preserveOrWrapError(err error, fallbackMsg string) error {
	if err == nil {
		return nil
	}
	var exc exceptions.Exception
	if errors.As(err, &exc) {
		return err
	}
	if errors.Is(err, exceptions.NotFoundError) {
		return exceptions.NewNotFoundException(fallbackMsg, err)
	}
	return exceptions.NewInternalException(fallbackMsg, err)
}

func logExecution(logger core.Logger, kind, name string, start time.Time, err error) {
	if logger == nil {
		return
	}
	fields := map[string]any{
		"kind":     kind,
		"handler":  name,
		"duration": time.Since(start).String(),
	}
	if err != nil {
		logger.Error(kind+" execution failed", err, fields)
		return
	}
	logger.Info(kind+" executed", fields)
}
