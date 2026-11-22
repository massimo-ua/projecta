package app

import (
	"context"
	"errors"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
	failedToCreateAsset = "failed to create asset"
	failedToFindAsset   = "failed to find asset"
	failedToUpdateAsset = "failed to update asset"
)

// OwnerReader is a consumer-owned port providing read access to asset owners.
type OwnerReader interface {
	FindOwner(ctx context.Context, personID core.PersonID) (asset.OwnerRef, error)
}

// ProjectReader is a consumer-owned port providing authorized read access to project metadata and ceilings.
type ProjectReader interface {
	FindProject(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) (asset.ProjectDescriptor, error)
	FindMonthlyCashCeilings(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID, months []string, tag string) (map[string]int64, error)
}

// InvestmentSource is a consumer-owned port bridging asset workflows with the investment context.
type InvestmentSource interface {
	FindInvestments(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID, ids []core.InvestmentID) ([]asset.InvestmentItem, error)
	ListProjectInvestments(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) ([]asset.InvestmentRecordSummary, error)
	AssignToAsset(ctx context.Context, assetID core.AssetID, requesterID core.PersonID, investmentIDs []core.InvestmentID) error
	CreateInitialInvestment(ctx context.Context, item asset.InitialInvestment) error
}

func resolveProjectCurrency(ctx context.Context, projects ProjectReader, projectID core.ProjectID, requesterID core.PersonID) string {
	if projects != nil && !projectID.IsNil() && !requesterID.IsNil() {
		if proj, err := projects.FindProject(ctx, projectID, requesterID); err == nil && proj.MainCurrency != "" {
			return proj.MainCurrency
		}
	}
	return "UAH"
}

func reconcileAndEvaluateAssets(
	ctx context.Context,
	repo asset.Repository,
	projects ProjectReader,
	investments InvestmentSource,
	converter asset.CurrencyConverter,
	requesterID core.PersonID,
	defaultProjID core.ProjectID,
	defaultHomeCurrency string,
	targets []*asset.Asset,
) error {
	if len(targets) == 0 {
		return nil
	}

	if investments == nil {
		for _, a := range targets {
			if a == nil {
				continue
			}
			hc := defaultHomeCurrency
			if hc == "" {
				hc = resolveProjectCurrency(ctx, projects, a.ProjectID(), requesterID)
			}
			a.EvaluateValuation(hc, converter)
		}
		return nil
	}

	projAssetsMap := make(map[core.ProjectID][]*asset.Asset)
	for _, a := range targets {
		if a == nil {
			continue
		}
		pid := a.ProjectID()
		if pid.IsNil() {
			pid = defaultProjID
		}
		projAssetsMap[pid] = append(projAssetsMap[pid], a)
	}

	for pid, projTargets := range projAssetsMap {
		hc := defaultHomeCurrency
		if hc == "" || pid != defaultProjID {
			hc = resolveProjectCurrency(ctx, projects, pid, requesterID)
		}

		invs, err := investments.ListProjectInvestments(ctx, pid, requesterID)
		if err != nil {
			return err
		}
		if len(invs) > 0 {
			allProjectCol, err := repo.Find(ctx, asset.CollectionFilter{OwnerID: requesterID, ProjectID: pid})
			if err != nil {
				return err
			}
			hierarchy := asset.NewAssetHierarchy()
			assetByID := make(map[core.AssetID]*asset.Asset)
			if allProjectCol != nil {
				for _, pa := range allProjectCol.Elements() {
					if pa != nil {
						hierarchy.AddAsset(pa)
						assetByID[pa.ID()] = pa
					}
				}
			}
			for _, ta := range projTargets {
				if ta != nil && assetByID[ta.ID()] == nil {
					hierarchy.AddAsset(ta)
					assetByID[ta.ID()] = ta
				}
			}
			for _, pa := range assetByID {
				for _, ch := range pa.Children() {
					_ = hierarchy.AddEdge(pa.ID(), ch.ChildID, ch.SharePercentage)
				}
			}

			directAll, _ := asset.NormalizeProjectInvestments(invs, hc, converter)
			hierarchy.ApplyToAssets(directAll, hc)

			for _, ta := range projTargets {
				if ta == nil {
					continue
				}
				if updated, ok := assetByID[ta.ID()]; ok && updated != nil && updated != ta {
					ta.AssignValuation(updated.DirectCost(), updated.TotalCost(), updated.Children(), updated.Parents())
				}
				ta.EvaluateValuation(hc, converter)
			}
			continue
		}

		for _, ta := range projTargets {
			if ta != nil {
				ta.EvaluateValuation(hc, converter)
			}
		}
	}

	return nil
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
