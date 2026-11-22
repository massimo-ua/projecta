package asset

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

const (
	failedToCreateAsset = "failed to create asset"
	failedToFindAsset   = "failed to find asset"
	failedToUpdateAsset = "failed to update asset"
)

type ServiceImpl struct {
	db           core.DbConnection
	assets       Repository
	people       OwnerReader
	projects     ProjectReader
	rateProvider CurrencyConverter
	investments  InvestmentSource
}

func NewService(
	db core.DbConnection,
	assets Repository,
	people OwnerReader,
	projects ProjectReader,
	investmentSource InvestmentSource,
	rateProvider ...CurrencyConverter,
) *ServiceImpl {
	var rp CurrencyConverter
	if len(rateProvider) > 0 {
		rp = rateProvider[0]
	}
	return &ServiceImpl{
		db:           db,
		assets:       assets,
		people:       people,
		projects:     projects,
		investments:  investmentSource,
		rateProvider: rp,
	}
}

func (s *ServiceImpl) resolveProjectCurrency(ctx context.Context, projectID core.ProjectID) string {
	if s.projects != nil && !projectID.IsNil() {
		if proj, err := s.projects.FindProject(ctx, projectID); err == nil && proj.MainCurrency != "" {
			return proj.MainCurrency
		}
	}
	return "UAH"
}

func (s *ServiceImpl) Find(ctx context.Context, filter CollectionFilter) (*Collection, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToFindAsset, err)
	}

	filter.OwnerID = personID

	collection, err := s.assets.Find(ctx, filter)
	if err != nil {
		return nil, preserveOrWrapError(err, failedToFindAsset)
	}

	if collection != nil {
		var defaultHomeCurrency string
		if !filter.ProjectID.IsNil() {
			defaultHomeCurrency = s.resolveProjectCurrency(ctx, filter.ProjectID)
		}
		if err := s.reconcileAndEvaluateAssets(ctx, personID, filter.ProjectID, defaultHomeCurrency, collection.Elements()); err != nil {
			return nil, preserveOrWrapError(err, failedToFindAsset)
		}
	}

	return collection, nil
}

func (s *ServiceImpl) FindOne(ctx context.Context, filter Filter) (*Asset, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToFindAsset, err)
	}

	filter.OwnerID = personID

	anAsset, err := s.assets.FindOne(ctx, filter)
	if err != nil {
		return nil, preserveOrWrapError(err, failedToFindAsset)
	}
	if anAsset == nil {
		return nil, exceptions.NewNotFoundException("asset not found", nil)
	}

	projID := filter.ProjectID
	if projID.IsNil() {
		projID = anAsset.ProjectID()
	}
	homeCurrency := s.resolveProjectCurrency(ctx, projID)
	if err := s.reconcileAndEvaluateAssets(ctx, personID, projID, homeCurrency, []*Asset{anAsset}); err != nil {
		return nil, preserveOrWrapError(err, failedToFindAsset)
	}

	return anAsset, nil
}

func (s *ServiceImpl) Create(ctx context.Context, command CreateAssetCommand) (*Asset, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToCreateAsset, err)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	proj, err := s.projects.FindProject(ctx, command.ProjectID)
	if err != nil {
		return nil, preserveOrWrapError(err, failedToCreateAsset)
	}

	price := command.TargetPrice
	if price == nil {
		price = command.Price
	}

	startDate := command.StartDate
	if startDate.IsZero() {
		startDate = core.DateOrNow(command.AcquiredAt)
	}

	anAsset, err := CreateAsset(
		core.NewAssetID(),
		command.Name,
		command.Description,
		command.ProjectID,
		price,
		startDate,
		owner,
		command.Status,
		command.CompletedDate,
		command.Tags,
	)
	if err != nil {
		return nil, err
	}

	if command.TargetCompletionDate != nil {
		if err := anAsset.UpdateSchedule(nil, command.TargetCompletionDate, false); err != nil {
			return nil, err
		}
	}

	if command.WithPayment && (price == nil || price.Amount() <= 0) {
		return nil, exceptions.NewValidationException("price must be greater than 0 when creating an asset with payment", nil)
	}

	homeCurrency := "UAH"
	if proj.MainCurrency != "" {
		homeCurrency = proj.MainCurrency
	}
	anAsset.EvaluateValuation(homeCurrency, s.rateProvider)

	paymentDescription := command.Description
	if paymentDescription == "" {
		paymentDescription = command.Name
	}

	if command.WithPayment {
		_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
			if err = s.assets.Save(ctx, anAsset); err != nil {
				return nil, err
			}

			if s.investments != nil {
				if err = s.investments.CreateInitialInvestment(ctx, InitialInvestment{
					ID:            core.NewInvestmentID(),
					ProjectID:     command.ProjectID,
					AssetID:       anAsset.ID(),
					ContributorID: owner.PersonID,
					Amount:        price,
					Date:          startDate,
					Description:   paymentDescription,
					Tags:          command.Tags,
				}); err != nil {
					return nil, err
				}
			}

			return nil, nil
		})

		if err != nil {
			return nil, preserveOrWrapError(err, failedToCreateAsset)
		}

		return anAsset, nil
	}

	err = s.assets.Save(ctx, anAsset)
	if err != nil {
		return nil, preserveOrWrapError(err, failedToCreateAsset)
	}

	return anAsset, nil
}

func (s *ServiceImpl) CreateFromInvestments(ctx context.Context, command CreateAssetFromInvestmentsCommand) (*Asset, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException(failedToCreateAsset, err)
	}

	if command.ProjectID.IsNil() {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	if len(command.InvestmentIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one investment must be selected", nil)
	}

	if strings.TrimSpace(command.Name) == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException(failedToCreateAsset, err)
	}

	proj, err := s.projects.FindProject(ctx, command.ProjectID)
	if err != nil {
		return nil, preserveOrWrapError(err, failedToCreateAsset)
	}

	if s.investments == nil {
		return nil, exceptions.NewInternalException("investment source not configured", nil)
	}

	invItems, err := s.investments.FindInvestments(ctx, command.ProjectID, command.InvestmentIDs)
	if err != nil {
		return nil, preserveOrWrapError(err, failedToCreateAsset)
	}
	if len(invItems) != len(command.InvestmentIDs) {
		return nil, exceptions.NewValidationException("one or more investments not found in project", nil)
	}

	mainCurrency := "UAH"
	if proj.MainCurrency != "" {
		mainCurrency = proj.MainCurrency
	}

	anAsset, err := CreateAssetFromInvestments(
		core.NewAssetID(),
		command.Name,
		command.Description,
		command.ProjectID,
		mainCurrency,
		command.TargetCurrency,
		command.AcquiredAt,
		owner,
		invItems,
		s.rateProvider,
	)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err = s.assets.Save(txCtx, anAsset); err != nil {
			return nil, err
		}

		if err = s.investments.AssignToAsset(txCtx, anAsset.ID(), command.InvestmentIDs); err != nil {
			return nil, err
		}

		return nil, nil
	})
	if err != nil {
		return nil, preserveOrWrapError(err, failedToCreateAsset)
	}

	return anAsset, nil
}

// Deprecated: Use CreateFromInvestments instead.
func (s *ServiceImpl) CreateFromPayments(ctx context.Context, command CreateAssetFromPaymentsCommand) (*Asset, error) {
	return s.CreateFromInvestments(ctx, CreateAssetFromInvestmentsCommand{
		ProjectID:      command.ProjectID,
		InvestmentIDs:  command.PaymentIDs,
		Name:           command.Name,
		Description:    command.Description,
		TypeID:         command.TypeID,
		AcquiredAt:     command.AcquiredAt,
		TargetCurrency: command.TargetCurrency,
	})
}

func (s *ServiceImpl) Remove(ctx context.Context, command RemoveAssetCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException(failedToFindAsset, err)
	}

	anAsset, err := s.assets.FindOne(ctx, Filter{ID: command.AssetID, ProjectID: command.ProjectID, OwnerID: personID})
	if err != nil {
		return preserveOrWrapError(err, failedToFindAsset)
	}

	_, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		if err := s.assets.Remove(txCtx, anAsset); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return preserveOrWrapError(err, "failed to remove asset")
}

func (s *ServiceImpl) Update(ctx context.Context, command UpdateAssetCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException(failedToUpdateAsset, err)
	}

	proj, err := s.projects.FindProject(ctx, command.ProjectID)
	if err != nil {
		return preserveOrWrapError(err, failedToUpdateAsset)
	}

	anAsset, err := s.assets.FindOne(ctx, Filter{ID: command.AssetID, ProjectID: command.ProjectID, OwnerID: personID})
	if err != nil {
		return preserveOrWrapError(err, failedToUpdateAsset)
	}

	price := command.TargetPrice
	if price == nil {
		price = command.Price
	}

	if err := anAsset.UpdateDetails(command.Name, command.Description, price, command.Tags); err != nil {
		return err
	}
	anAsset.TransitionStatus(command.Status, command.CompletedDate)

	var schedStart *time.Time
	if !command.StartDate.IsZero() {
		sDate := command.StartDate
		schedStart = &sDate
	} else if !command.AcquiredAt.IsZero() {
		aDate := command.AcquiredAt
		schedStart = &aDate
	}
	if err := anAsset.UpdateSchedule(schedStart, command.TargetCompletionDate, command.ClearTargetDate); err != nil {
		return err
	}

	homeCurrency := "UAH"
	if proj.MainCurrency != "" {
		homeCurrency = proj.MainCurrency
	}
	anAsset.EvaluateValuation(homeCurrency, s.rateProvider)

	err = s.assets.Save(ctx, anAsset)
	return preserveOrWrapError(err, failedToUpdateAsset)
}

func (s *ServiceImpl) LinkChild(ctx context.Context, command LinkChildCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException("failed to link child asset", err)
	}

	if err := ValidateEdge(command.ParentID, command.ChildID, command.SharePercentage, nil); err != nil {
		return err
	}

	_, err = s.assets.FindOne(ctx, Filter{ID: command.ParentID, ProjectID: command.ProjectID, OwnerID: personID})
	if err != nil {
		return exceptions.NewValidationException("parent asset not found", err)
	}

	_, err = s.assets.FindOne(ctx, Filter{ID: command.ChildID, ProjectID: command.ProjectID, OwnerID: personID})
	if err != nil {
		return exceptions.NewValidationException("child asset not found", err)
	}

	ancestors, err := s.assets.FindAncestors(ctx, command.ParentID)
	if err != nil {
		return exceptions.NewInternalException("failed to check hierarchy cycles", err)
	}
	if err := ValidateEdge(command.ParentID, command.ChildID, command.SharePercentage, ancestors); err != nil {
		return err
	}

	err = s.assets.AddChild(ctx, command.ParentID, command.ChildID, command.SharePercentage)
	return preserveOrWrapError(err, "failed to link child asset")
}

func (s *ServiceImpl) UnlinkChild(ctx context.Context, command UnlinkChildCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException("failed to unlink child asset", err)
	}

	if _, err := s.assets.FindOne(ctx, Filter{
		ID:        command.ParentID,
		ProjectID: command.ProjectID,
		OwnerID:   personID,
	}); err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException("parent asset not found", err)
		}
		return preserveOrWrapError(err, "failed to unlink child asset")
	}

	err = s.assets.RemoveChild(ctx, command.ParentID, command.ChildID)
	return preserveOrWrapError(err, "failed to unlink child asset")
}

func (s *ServiceImpl) AssignInvestments(ctx context.Context, command AssignInvestmentsCommand) error {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return exceptions.NewUnauthorizedException("failed to assign investments to asset", err)
	}

	if command.AssetID.IsNil() {
		return exceptions.NewValidationException("asset id is required", nil)
	}
	if command.ProjectID.IsNil() {
		return exceptions.NewValidationException("project id is required", nil)
	}
	if len(command.InvestmentIDs) == 0 {
		return exceptions.NewValidationException("at least one investment must be selected", nil)
	}

	_, err = s.assets.FindOne(ctx, Filter{
		ID:        command.AssetID,
		ProjectID: command.ProjectID,
		OwnerID:   personID,
	})
	if err != nil {
		return exceptions.NewValidationException("asset not found in project", err)
	}

	if s.investments != nil {
		items, err := s.investments.FindInvestments(ctx, command.ProjectID, command.InvestmentIDs)
		if err != nil || len(items) != len(command.InvestmentIDs) {
			return exceptions.NewValidationException("one or more investments not found in project", err)
		}

		if err = s.investments.AssignToAsset(ctx, command.AssetID, command.InvestmentIDs); err != nil {
			return preserveOrWrapError(err, "failed to assign investments to asset")
		}
	}

	return nil
}

func (s *ServiceImpl) Group(ctx context.Context, command GroupAssetsCommand) (*Asset, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to group assets", err)
	}

	if command.ProjectID.IsNil() {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}

	if strings.TrimSpace(command.Name) == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}

	if len(command.ChildAssetIDs) == 0 {
		return nil, exceptions.NewValidationException("at least one asset must be selected to group", nil)
	}

	owner, err := s.people.FindOwner(ctx, personID)
	if err != nil {
		return nil, exceptions.NewInternalException("failed to group assets", err)
	}

	proj, err := s.projects.FindProject(ctx, command.ProjectID)
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to group assets")
	}

	// Verify all child assets exist and belong to the project
	for _, childID := range command.ChildAssetIDs {
		child, err := s.assets.FindOne(ctx, Filter{ID: childID, ProjectID: command.ProjectID, OwnerID: personID})
		if err != nil || child == nil {
			return nil, exceptions.NewValidationException(fmt.Sprintf("child asset %s not found in project", childID), err)
		}
	}

	parentID := core.NewAssetID()
	parentAsset, err := CreateAssetGroup(
		parentID,
		command.Name,
		command.Description,
		command.ProjectID,
		owner,
		command.Tags,
	)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Tx(ctx, func(ctx context.Context) (any, error) {
		if err := s.assets.Save(ctx, parentAsset); err != nil {
			return nil, err
		}

		for _, childID := range command.ChildAssetIDs {
			if err := s.assets.AddChild(ctx, parentID, childID, 100.0); err != nil {
				return nil, err
			}
		}

		return nil, nil
	})

	if err != nil {
		return nil, preserveOrWrapError(err, "failed to group assets")
	}

	res, err := s.assets.FindOne(ctx, Filter{ID: parentID, ProjectID: command.ProjectID, OwnerID: personID})
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to find grouped asset")
	}
	if res != nil {
		homeCurrency := "UAH"
		if proj.MainCurrency != "" {
			homeCurrency = proj.MainCurrency
		}
		res.EvaluateValuation(homeCurrency, s.rateProvider)
	}
	return res, nil
}

func (s *ServiceImpl) LinkChildren(ctx context.Context, command LinkChildrenCommand) error {
	_, err := s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
		for _, childID := range command.ChildAssetIDs {
			if err := s.LinkChild(txCtx, LinkChildCommand{
				ParentID:        command.ParentID,
				ChildID:         childID,
				ProjectID:       command.ProjectID,
				SharePercentage: 100.0,
			}); err != nil {
				return nil, err
			}
		}
		return nil, nil
	})
	return err
}

// UpdateSchedule updates start_date and/or target_completion_date for an asset, returning non-blocking DAG schedule warnings.
func (s *ServiceImpl) UpdateSchedule(ctx context.Context, command RescheduleAssetCommand) (*Asset, []string, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, nil, exceptions.NewUnauthorizedException(failedToUpdateAsset, err)
	}

	proj, err := s.projects.FindProject(ctx, command.ProjectID)
	if err != nil {
		return nil, nil, preserveOrWrapError(err, failedToUpdateAsset)
	}

	anAsset, err := s.assets.FindOne(ctx, Filter{ID: command.AssetID, ProjectID: command.ProjectID, OwnerID: personID})
	if err != nil {
		return nil, nil, preserveOrWrapError(err, failedToUpdateAsset)
	}
	if anAsset == nil {
		return nil, nil, exceptions.NewNotFoundException("asset not found", nil)
	}

	if err := anAsset.UpdateSchedule(command.StartDate, command.TargetCompletionDate, command.ClearTargetDate); err != nil {
		return nil, nil, err
	}

	homeCurrency := "UAH"
	if proj.MainCurrency != "" {
		homeCurrency = proj.MainCurrency
	}
	if err := s.reconcileAndEvaluateAssets(ctx, personID, command.ProjectID, homeCurrency, []*Asset{anAsset}); err != nil {
		return nil, nil, preserveOrWrapError(err, failedToUpdateAsset)
	}

	if err := s.assets.Save(ctx, anAsset); err != nil {
		return nil, nil, preserveOrWrapError(err, failedToUpdateAsset)
	}

	warnings := s.computeScheduleWarnings(ctx, personID, anAsset)
	return anAsset, warnings, nil
}

// GetCashFlowForecast calculates historical resource consumption, future cash demand, and ceiling alerts for a project.
func (s *ServiceImpl) GetCashFlowForecast(ctx context.Context, query GetCashFlowForecastQuery) (*CashFlowForecast, error) {
	personID, err := core.AuthGuard(ctx)
	if err != nil {
		return nil, exceptions.NewUnauthorizedException("failed to calculate cash flow forecast", err)
	}
	if query.ProjectID.IsNil() {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}
	if err := validateForecastTag(query.Tag); err != nil {
		return nil, err
	}

	proj, err := s.projects.FindProject(ctx, query.ProjectID)
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to calculate cash flow forecast")
	}

	col, err := s.assets.Find(ctx, CollectionFilter{
		OwnerID:   personID,
		ProjectID: query.ProjectID,
	})
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to calculate cash flow forecast")
	}
	var assets []*Asset
	if col != nil {
		assets = col.Elements()
	}

	var investments []InvestmentRecordSummary
	if s.investments != nil {
		investments, err = s.investments.ListProjectInvestments(ctx, query.ProjectID)
		if err != nil {
			return nil, preserveOrWrapError(err, "failed to calculate cash flow forecast")
		}
	}

	months, _, _, err := ResolveForecastMonths(proj, assets, investments, query)
	if err != nil {
		return nil, err
	}

	ceilings, err := s.projects.FindMonthlyCashCeilings(ctx, query.ProjectID, months, query.Tag)
	if err != nil {
		return nil, preserveOrWrapError(err, "failed to calculate cash flow forecast")
	}

	return CalculateProjectCashFlowForecast(proj, assets, investments, ceilings, months, query, s.rateProvider), nil
}

func (s *ServiceImpl) reconcileAndEvaluateAssets(
	ctx context.Context,
	personID core.PersonID,
	defaultProjID core.ProjectID,
	defaultHomeCurrency string,
	targets []*Asset,
) error {
	if len(targets) == 0 {
		return nil
	}

	if s.investments == nil {
		for _, a := range targets {
			if a == nil {
				continue
			}
			hc := defaultHomeCurrency
			if hc == "" {
				hc = s.resolveProjectCurrency(ctx, a.ProjectID())
			}
			a.EvaluateValuation(hc, s.rateProvider)
		}
		return nil
	}

	// Group targets by projectID so multi-project Find loads each project's investments once.
	projAssetsMap := make(map[core.ProjectID][]*Asset)
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
			hc = s.resolveProjectCurrency(ctx, pid)
		}

		invs, err := s.investments.ListProjectInvestments(ctx, pid)
		if err != nil {
			return err
		}
		if len(invs) > 0 {
			allProjectCol, err := s.assets.Find(ctx, CollectionFilter{OwnerID: personID, ProjectID: pid})
			if err != nil {
				return err
			}
			hierarchy := NewAssetHierarchy()
			assetByID := make(map[core.AssetID]*Asset)
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
					hierarchy.AddEdge(pa.ID(), ch.ChildID, ch.SharePercentage)
				}
			}

			directAll, _ := NormalizeProjectInvestments(invs, hc, s.rateProvider)
			hierarchy.ApplyToAssets(directAll, hc)

			for _, ta := range projTargets {
				if ta == nil {
					continue
				}
				if updated, ok := assetByID[ta.ID()]; ok && updated != nil && updated != ta {
					ta.AssignValuation(updated.DirectCost(), updated.TotalCost(), updated.Children(), updated.Parents())
				}
				ta.EvaluateValuation(hc, s.rateProvider)
			}
			continue
		}

		for _, ta := range projTargets {
			if ta != nil {
				ta.EvaluateValuation(hc, s.rateProvider)
			}
		}
	}

	return nil
}

func (s *ServiceImpl) computeScheduleWarnings(ctx context.Context, personID core.PersonID, target *Asset) []string {
	warnings := make([]string, 0)
	if target == nil || target.TargetCompletionDate() == nil {
		return warnings
	}
	parentTarget := utcCalendarDay(*target.TargetCompletionDate())
	for _, ch := range target.Children() {
		childAsset, err := s.assets.FindOne(ctx, Filter{ID: ch.ChildID, ProjectID: target.ProjectID(), OwnerID: personID})
		if err != nil || childAsset == nil || childAsset.TargetCompletionDate() == nil {
			continue
		}
		childTarget := utcCalendarDay(*childAsset.TargetCompletionDate())
		if childTarget.After(parentTarget) {
			warnings = append(warnings, "child_target_exceeds_parent:"+childAsset.ID().String())
		}
	}
	for _, p := range target.Parents() {
		parentAsset, err := s.assets.FindOne(ctx, Filter{ID: p.ParentID, ProjectID: target.ProjectID(), OwnerID: personID})
		if err != nil || parentAsset == nil || parentAsset.TargetCompletionDate() == nil {
			continue
		}
		if parentTarget.After(utcCalendarDay(*parentAsset.TargetCompletionDate())) {
			warnings = append(warnings, "child_target_exceeds_parent:"+parentAsset.ID().String())
		}
	}
	return warnings
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


