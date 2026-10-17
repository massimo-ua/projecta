package investment

import (
	"context"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

// AssetCollectionReader is a consumer-owned interface for paginating project assets.
type AssetCollectionReader interface {
	Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error)
}

// InvestmentCollectionReader is a consumer-owned interface for paginating project investments.
type InvestmentCollectionReader interface {
	Find(ctx context.Context, filter CollectionFilter) (*Collection, error)
}

// CalculateProjectTotals orchestrates pagination and currency-converted aggregation of project investments and root assets.
func CalculateProjectTotals(
	ctx context.Context,
	projectID uuid.UUID,
	projects ProjectReader,
	assets AssetCollectionReader,
	converter asset.CurrencyConverter,
	investments ...InvestmentCollectionReader,
) (asset.ProjectTotals, error) {
	proj, err := projects.FindOne(ctx, project.ProjectFilter{ProjectID: projectID})
	if err != nil {
		return asset.ProjectTotals{}, err
	}

	homeCurrency := proj.MainCurrency()
	if homeCurrency == "" {
		homeCurrency = "UAH"
	}

	var invSvc InvestmentCollectionReader
	if len(investments) > 0 {
		invSvc = investments[0]
	}

	const pageLimit = 100
	var invAmounts []*money.Money
	var rootAssets []*asset.Asset
	hasInvestments := false
	hasAssets := false

	if invSvc != nil {
		for offset := 0; ; offset += pageLimit {
			page, err := invSvc.Find(ctx, CollectionFilter{
				ProjectID: projectID,
				Pagination: core.Pagination{
					Limit:  pageLimit,
					Offset: offset,
				},
			})
			if err != nil {
				return asset.ProjectTotals{}, err
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

	for offset := 0; ; offset += pageLimit {
		page, err := assets.Find(ctx, asset.CollectionFilter{
			ProjectID: projectID,
			Pagination: core.Pagination{
				Limit:  pageLimit,
				Offset: offset,
			},
		})
		if err != nil {
			return asset.ProjectTotals{}, err
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

	totals, err := asset.CalculateProjectTotals(invAmounts, rootAssets, homeCurrency, converter)
	if err != nil {
		return asset.ProjectTotals{}, err
	}
	totals.HasActivity = hasAssets || hasInvestments
	return totals, nil
}
