package main

import (
	"context"

	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	assetapp "gitlab.com/massimo-ua/projecta/internal/asset/app"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/investment"
	investmentapp "gitlab.com/massimo-ua/projecta/internal/investment/app"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/project"
	projectapp "gitlab.com/massimo-ua/projecta/internal/project/app"
	"gitlab.com/massimo-ua/projecta/internal/statement"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

type currencyConverterAdapter struct {
	provider currency.CurrencyRateProvider
}

func (a *currencyConverterAdapter) ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error) {
	if amount == nil {
		return money.New(0, targetCurrency), nil
	}
	currCode := amount.Currency().Code
	if currCode == targetCurrency {
		return amount, nil
	}
	if a.provider == nil {
		return nil, exceptions.NewValidationException("currency rate provider not configured", nil)
	}
	converted, err := a.provider.Convert(
		currency.NewCurrency(amount.Amount(), currCode),
		currency.NewCurrency(0, targetCurrency),
	)
	if err != nil {
		return nil, err
	}
	return money.New(converted.Amount, converted.Code), nil
}

type personReader interface {
	FindByID(ctx context.Context, personID core.PersonID) (*people.Person, error)
}

type projectPeopleAdapter struct {
	repo personReader
}

func (a *projectPeopleAdapter) FindOwner(ctx context.Context, personID core.PersonID) (*project.Owner, error) {
	person, err := a.repo.FindByID(ctx, personID)
	if err != nil {
		return nil, err
	}
	return &project.Owner{
		PersonID:        person.ID(),
		DisplayName:     person.DisplayName(),
		FirstName:       person.FirstName(),
		CanHaveProjects: person.CanHaveProjects(),
	}, nil
}

type assetProjectReaderAdapter struct {
	projects     project.ProjectRepository
	rateProvider currency.CurrencyRateProvider
}

func (a *assetProjectReaderAdapter) FindProject(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) (asset.ProjectDescriptor, error) {
	if requesterID.IsNil() {
		return asset.ProjectDescriptor{}, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	proj, err := a.projects.GetProject(ctx, projectID, requesterID)
	if err != nil {
		return asset.ProjectDescriptor{}, err
	}
	return asset.ProjectDescriptor{
		ID:           proj.ID(),
		MainCurrency: proj.MainCurrency(),
		StartDate:    proj.StartDate(),
		EndDate:      proj.EndDate(),
	}, nil
}

func (a *assetProjectReaderAdapter) FindMonthlyCashCeilings(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID, months []string, tag string) (map[string]int64, error) {
	if requesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	proj, err := a.projects.GetProject(ctx, projectID, requesterID)
	if err != nil {
		return nil, err
	}

	if a.rateProvider != nil {
		proj.EvaluateTagBudgets(&currencyConverterAdapter{provider: a.rateProvider})
	} else {
		proj.EvaluateTagBudgets(nil)
	}

	return proj.MonthlyCashCeilings(months, tag), nil
}

type assetOwnerReaderAdapter struct {
	people projectapp.OwnerReader
}

func (a *assetOwnerReaderAdapter) FindOwner(ctx context.Context, personID core.PersonID) (asset.OwnerRef, error) {
	owner, err := a.people.FindOwner(ctx, personID)
	if err != nil {
		return asset.OwnerRef{}, err
	}
	return asset.OwnerRef{
		PersonID:    owner.PersonID,
		DisplayName: owner.DisplayName,
	}, nil
}

type investmentAssetReaderAdapter struct {
	assets asset.Repository
}

func (a *investmentAssetReaderAdapter) FindAsset(ctx context.Context, projectID core.ProjectID, assetID core.AssetID) (investmentapp.AssetDescriptor, error) {
	ast, err := a.assets.FindOne(ctx, asset.Filter{ID: assetID, ProjectID: projectID})
	if err != nil {
		return investmentapp.AssetDescriptor{}, err
	}
	return investmentapp.AssetDescriptor{
		ID:        ast.ID(),
		ProjectID: ast.ProjectID(),
		Name:      ast.Name(),
	}, nil
}

type investmentProjectReaderAdapter struct {
	projects project.ProjectRepository
}

func (a *investmentProjectReaderAdapter) FindProject(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) (investmentapp.ProjectDescriptor, error) {
	if requesterID.IsNil() {
		return investmentapp.ProjectDescriptor{}, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	proj, err := a.projects.GetProject(ctx, projectID, requesterID)
	if err != nil {
		return investmentapp.ProjectDescriptor{}, err
	}
	return investmentapp.ProjectDescriptor{
		ID:           proj.ID(),
		MainCurrency: proj.MainCurrency(),
	}, nil
}

type investmentContributorReaderAdapter struct {
	people projectapp.OwnerReader
}

func (a *investmentContributorReaderAdapter) FindContributor(ctx context.Context, personID core.PersonID) (investment.ContributorRef, error) {
	owner, err := a.people.FindOwner(ctx, personID)
	if err != nil {
		return investment.ContributorRef{}, err
	}
	return investment.ContributorRef{
		PersonID:    owner.PersonID,
		DisplayName: owner.DisplayName,
	}, nil
}

type projectValidator struct {
	app *projectapp.Application
}

func (v *projectValidator) ValidateProject(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) error {
	if requesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	_, err := v.app.Queries.GetProject.Handle(ctx, projectapp.GetProject{
		ProjectID:   projectID,
		RequesterID: requesterID,
	})
	return err
}

type investmentTransactionFinder struct {
	app *investmentapp.Application
}

func (f *investmentTransactionFinder) FindExisting(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) ([]statement.ExistingTransaction, error) {
	if requesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	pageSize := 500
	offset := 0
	var res []statement.ExistingTransaction

	for {
		col, err := f.app.Queries.ListInvestments.Handle(ctx, investmentapp.ListInvestments{
			RequesterID: requesterID,
			ProjectID:   projectID,
			Pagination: core.Pagination{
				Limit:  pageSize,
				Offset: offset,
			},
		})
		if err != nil {
			return nil, err
		}
		for _, inv := range col.Elements() {
			if inv == nil || inv.Amount() == nil {
				continue
			}
			res = append(res, statement.ExistingTransaction{
				Date:        inv.Date(),
				Amount:      inv.Amount().Amount(),
				Description: inv.Description(),
			})
		}
		if len(col.Elements()) < pageSize || len(res) >= col.Total() {
			break
		}
		offset += pageSize
	}
	return res, nil
}

type investmentSourceAdapter struct {
	app *investmentapp.Application
}

func (a *investmentSourceAdapter) FindInvestments(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID, ids []core.InvestmentID) ([]asset.InvestmentItem, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	col, err := a.app.Queries.ListInvestments.Handle(ctx, investmentapp.ListInvestments{
		RequesterID: requesterID,
		ProjectID:   projectID,
		IDs:         ids,
		Pagination:  core.Pagination{Limit: len(ids)},
	})
	if err != nil {
		return nil, err
	}
	items := make([]asset.InvestmentItem, 0, len(col.Elements()))
	for _, inv := range col.Elements() {
		items = append(items, asset.InvestmentItem{
			ID:     inv.ID(),
			Amount: inv.Amount(),
			Date:   inv.Date(),
			Tags:   inv.Tags(),
		})
	}
	return items, nil
}

func (a *investmentSourceAdapter) ListProjectInvestments(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID) ([]asset.InvestmentRecordSummary, error) {
	const pageSize = 500
	offset := 0
	var out []asset.InvestmentRecordSummary

	for {
		col, err := a.app.Queries.ListInvestments.Handle(ctx, investmentapp.ListInvestments{
			RequesterID: requesterID,
			ProjectID:   projectID,
			Pagination: core.Pagination{
				Limit:  pageSize,
				Offset: offset,
			},
		})
		if err != nil {
			return nil, err
		}
		if col == nil {
			break
		}
		for _, inv := range col.Elements() {
			if inv == nil {
				continue
			}
			var hours float64
			if th := inv.TimeHours(); th != nil {
				hours = *th
			}
			srcAllocs := inv.AssetAllocations()
			allocs := make([]asset.InvestmentAllocation, 0, len(srcAllocs))
			for _, al := range srcAllocs {
				allocs = append(allocs, asset.InvestmentAllocation{
					AssetID:         al.AssetID,
					SharePercentage: al.SharePercentage,
				})
			}
			if len(allocs) == 0 && !inv.PrimaryAssetID().IsNil() {
				allocs = append(allocs, asset.InvestmentAllocation{
					AssetID:         inv.PrimaryAssetID(),
					SharePercentage: 100.0,
				})
			}
			out = append(out, asset.InvestmentRecordSummary{
				ID:               inv.ID(),
				ResourceType:     string(inv.ResourceType()),
				Amount:           inv.Amount(),
				TimeHours:        hours,
				Date:             inv.Date(),
				AssetAllocations: allocs,
				Tags:             inv.Tags(),
			})
		}
		if len(col.Elements()) < pageSize || len(out) >= col.Total() {
			break
		}
		offset += pageSize
	}
	return out, nil
}

func (a *investmentSourceAdapter) AssignToAsset(ctx context.Context, assetID core.AssetID, requesterID core.PersonID, investmentIDs []core.InvestmentID) error {
	return a.app.Commands.AssignInvestmentsToAsset.Handle(ctx, investmentapp.AssignInvestmentsToAsset{
		RequesterID:   requesterID,
		AssetID:       assetID,
		InvestmentIDs: investmentIDs,
	})
}

func (a *investmentSourceAdapter) CreateInitialInvestment(ctx context.Context, item asset.InitialInvestment) error {
	return a.app.Commands.CreateInvestment.Handle(ctx, investmentapp.CreateInvestment{
		ID:           item.ID,
		RequesterID:  item.ContributorID,
		ProjectID:    item.ProjectID,
		AssetID:      item.AssetID,
		ResourceType: investment.ResourceTypeMoney,
		Amount:       item.Amount,
		Description:  item.Description,
		Date:         item.Date,
		Tags:         item.Tags,
	})
}

type projectTotalsCalculatorAdapter struct {
	assets       *assetapp.Application
	investments  *investmentapp.Application
	rateProvider currency.CurrencyRateProvider
}

func (a *projectTotalsCalculatorAdapter) CalculateTotals(ctx context.Context, projectID core.ProjectID, requesterID core.PersonID, homeCurrency string) (project.ProjectTotals, error) {
	var invAmounts []*money.Money
	var rootAssets []*asset.Asset
	hasInvestments := false
	hasAssets := false

	const pageLimit = 100

	if a.investments != nil {
		for offset := 0; ; offset += pageLimit {
			page, err := a.investments.Queries.ListInvestments.Handle(ctx, investmentapp.ListInvestments{
				RequesterID: requesterID,
				ProjectID:   projectID,
				Pagination: core.Pagination{
					Limit:  pageLimit,
					Offset: offset,
				},
			})
			if err != nil {
				return project.ProjectTotals{}, err
			}
			if page.Total() == 0 {
				break
			}
			for _, e := range page.Elements() {
				hasInvestments = true
				invAmounts = append(invAmounts, e.Amount())
			}
			if len(page.Elements()) < pageLimit {
				break
			}
		}
	}

	if a.assets != nil {
		for offset := 0; ; offset += pageLimit {
			page, err := a.assets.Queries.ListAssets.Handle(ctx, assetapp.ListAssets{
				RequesterID: requesterID,
				ProjectID:   projectID,
				Pagination: core.Pagination{
					Limit:  pageLimit,
					Offset: offset,
				},
			})
			if err != nil {
				return project.ProjectTotals{}, err
			}
			if page.Total() == 0 {
				break
			}
			for _, e := range page.Elements() {
				if len(e.Parents()) > 0 {
					continue
				}
				hasAssets = true
				rootAssets = append(rootAssets, e)
			}
			if len(page.Elements()) < pageLimit {
				break
			}
		}
	}

	var converter asset.CurrencyConverter
	if a.rateProvider != nil {
		converter = &currencyConverterAdapter{provider: a.rateProvider}
	}

	totals, err := asset.CalculateProjectTotals(invAmounts, rootAssets, homeCurrency, converter)
	if err != nil {
		return project.ProjectTotals{}, err
	}

	return project.ProjectTotals{
		TotalInvested: totals.TotalInvested,
		Balance:       totals.Balance(),
		HasActivity:   hasAssets || hasInvestments,
		Currency:      totals.Currency,
	}, nil
}
