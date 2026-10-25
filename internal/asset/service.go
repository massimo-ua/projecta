package asset

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		for _, a := range collection.Elements() {
			if a != nil {
				homeCurrency := defaultHomeCurrency
				if homeCurrency == "" {
					homeCurrency = s.resolveProjectCurrency(ctx, a.ProjectID())
				}
				a.EvaluateValuation(homeCurrency, s.rateProvider)
			}
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

	if anAsset != nil {
		projID := filter.ProjectID
		if projID.IsNil() {
			projID = anAsset.ProjectID()
		}
		anAsset.EvaluateValuation(s.resolveProjectCurrency(ctx, projID), s.rateProvider)
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

	if !command.StartDate.IsZero() {
		anAsset.Reschedule(command.StartDate)
	} else if !command.AcquiredAt.IsZero() {
		anAsset.Reschedule(command.AcquiredAt)
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


