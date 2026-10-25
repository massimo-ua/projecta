package dal

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

var ErrFailedToSaveAsset = errors.New("failed to save asset")
var ErrAssetNotFound = errors.Join(exceptions.NotFoundError, errors.New("asset not found"))

type PgAssetRepository struct {
	db *PgRepository
}

func NewPgAssetRepository(conn *PgDbConnection) *PgAssetRepository {
	return &PgAssetRepository{
		db: &PgRepository{db: conn},
	}
}

func (r *PgAssetRepository) Save(ctx context.Context, anAsset *asset.Asset) error {
	var targetPriceVal any
	var targetCurrVal any
	if anAsset.TargetPrice() != nil {
		targetPriceVal = anAsset.TargetPrice().Amount()
		targetCurrVal = anAsset.TargetPrice().Currency().Code
	} else if anAsset.Price() != nil {
		targetPriceVal = anAsset.Price().Amount()
		targetCurrVal = anAsset.Price().Currency().Code
	}

	tags := anAsset.Tags()
	if tags == nil {
		tags = []string{}
	}

	sql := `
		INSERT INTO projecta_assets (
			asset_id, name, description, project_id, owner_id,
			status, start_date, completed_date, target_price,
			target_currency, acquired_at, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (asset_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			start_date = EXCLUDED.start_date,
			completed_date = EXCLUDED.completed_date,
			target_price = EXCLUDED.target_price,
			target_currency = EXCLUDED.target_currency,
			acquired_at = EXCLUDED.acquired_at,
			tags = EXCLUDED.tags
	`

	if _, err := r.db.Exec(
		ctx,
		sql,
		anAsset.ID().String(),
		anAsset.Name(),
		anAsset.Description(),
		anAsset.ProjectID().String(),
		anAsset.Owner().PersonID.String(),
		anAsset.Status().String(),
		anAsset.StartDate(),
		anAsset.CompletedDate(),
		targetPriceVal,
		targetCurrVal,
		anAsset.AcquiredAt(),
		tags,
	); err != nil {
		return errors.Join(ErrFailedToSaveAsset, err)
	}

	return nil
}

func (r *PgAssetRepository) Remove(ctx context.Context, asset *asset.Asset) error {
	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_assets")
	qb.Where(qb.Equal("asset_id", asset.ID().String()))
	qb.Where(qb.Equal("project_id", asset.ProjectID().String()))

	sql, args := qb.Build()

	res, err := r.db.Exec(ctx, sql, args...)

	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return ErrAssetNotFound
	}

	return nil
}

func (r *PgAssetRepository) AddChild(ctx context.Context, parentID core.AssetID, childID core.AssetID, sharePercentage float64) error {
	sql := `
		INSERT INTO projecta_asset_compositions (parent_asset_id, child_asset_id, share_percentage)
		VALUES ($1, $2, $3)
		ON CONFLICT (parent_asset_id, child_asset_id) DO UPDATE SET share_percentage = EXCLUDED.share_percentage
	`
	_, err := r.db.Exec(ctx, sql, parentID.String(), childID.String(), sharePercentage)
	return err
}

func (r *PgAssetRepository) RemoveChild(ctx context.Context, parentID core.AssetID, childID core.AssetID) error {
	sql := `DELETE FROM projecta_asset_compositions WHERE parent_asset_id = $1 AND child_asset_id = $2`
	_, err := r.db.Exec(ctx, sql, parentID.String(), childID.String())
	return err
}

func (r *PgAssetRepository) FindChildren(ctx context.Context, parentID core.AssetID) ([]asset.ChildAssetLink, error) {
	sql := `
		SELECT c.child_asset_id, a.name, c.share_percentage
		FROM projecta_asset_compositions c
		JOIN projecta_assets a ON a.asset_id = c.child_asset_id
		WHERE c.parent_asset_id = $1
	`
	rows, err := r.db.Query(ctx, sql, parentID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []asset.ChildAssetLink
	for rows.Next() {
		var (
			cid   string
			cname string
			share float64
		)
		if err := rows.Scan(&cid, &cname, &share); err != nil {
			return nil, err
		}
		links = append(links, asset.ChildAssetLink{
			ChildID:         core.MustParseAssetID(cid),
			ChildName:       cname,
			SharePercentage: share,
		})
	}
	return links, nil
}

func (r *PgAssetRepository) FindParents(ctx context.Context, childID core.AssetID) ([]asset.ParentAssetLink, error) {
	sql := `
		SELECT c.parent_asset_id, a.name, c.share_percentage
		FROM projecta_asset_compositions c
		JOIN projecta_assets a ON a.asset_id = c.parent_asset_id
		WHERE c.child_asset_id = $1
	`
	rows, err := r.db.Query(ctx, sql, childID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var links []asset.ParentAssetLink
	for rows.Next() {
		var (
			pid   string
			pname string
			share float64
		)
		if err := rows.Scan(&pid, &pname, &share); err != nil {
			return nil, err
		}
		links = append(links, asset.ParentAssetLink{
			ParentID:        core.MustParseAssetID(pid),
			ParentName:      pname,
			SharePercentage: share,
		})
	}
	return links, nil
}

func (r *PgAssetRepository) FindAncestors(ctx context.Context, assetID core.AssetID) ([]core.AssetID, error) {
	sql := `
		WITH RECURSIVE ancestors AS (
			SELECT parent_asset_id FROM projecta_asset_compositions WHERE child_asset_id = $1
			UNION
			SELECT c.parent_asset_id FROM projecta_asset_compositions c
			INNER JOIN ancestors a ON c.child_asset_id = a.parent_asset_id
		)
		SELECT parent_asset_id FROM ancestors
	`
	rows, err := r.db.Query(ctx, sql, assetID.String())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ancestors []core.AssetID
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err == nil {
			ancestors = append(ancestors, core.MustParseAssetID(pid))
		}
	}
	return ancestors, nil
}

func (r *PgAssetRepository) FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_assets")

	setupSelectQueryBuilder(qb)

	qb.Where(qb.Equal("projecta_assets.asset_id", filter.ID.String()))

	if !filter.ProjectID.IsNil() {
		qb.Where(qb.Equal("projecta_assets.project_id", filter.ProjectID.String()))
	}

	if !filter.OwnerID.IsNil() {
		qb.Where(fmt.Sprintf("(projecta_assets.owner_id = %s OR projecta_assets.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))", qb.Var(filter.OwnerID.String()), qb.Var(filter.OwnerID.String())))
	}

	if filter.Name != "" {
		qb.Where(qb.ILike("projecta_assets.name", fmt.Sprintf("%s%%", filter.Name)))
	}

	sql, args := qb.Build()

	var (
		assetID          string
		name             string
		description      string
		projectID        string
		acquiredAt       time.Time
		ownerID          string
		ownerDisplayName string
		status           string
		startDate        time.Time
		completedDate    *time.Time
		targetPrice      *int64
		targetCurrency   *string
		tags             []string
	)

	if err := r.db.QueryRow(
		ctx,
		sql,
		args...,
	).Scan(
		&assetID,
		&name,
		&description,
		&projectID,
		&acquiredAt,
		&ownerID,
		&ownerDisplayName,
		&status,
		&startDate,
		&completedDate,
		&targetPrice,
		&targetCurrency,
		&tags,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}

		return nil, err
	}

	a, err := toAssetFromPg(
		assetID,
		name,
		description,
		projectID,
		acquiredAt,
		ownerID,
		ownerDisplayName,
		status,
		startDate,
		completedDate,
		targetPrice,
		targetCurrency,
		tags,
	)

	if err != nil {
		return nil, errors.Join(ErrAssetNotFound, err)
	}

	r.loadCompositionsAndRollup(ctx, []*asset.Asset{a}, a.ProjectID())
	return a, nil
}

func (r *PgAssetRepository) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_assets")

	if !filter.OwnerID.IsNil() {
		qb.Where(fmt.Sprintf("(projecta_assets.owner_id = %s OR projecta_assets.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))", qb.Var(filter.OwnerID.String()), qb.Var(filter.OwnerID.String())))
	}

	if !filter.ProjectID.IsNil() {
		qb.Where(qb.Equal("projecta_assets.project_id", filter.ProjectID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.ILike("projecta_assets.name", fmt.Sprintf("%s%%", filter.Name)))
	}

	qb.Select(qb.As("COUNT(*)", "total"))

	sql, args := qb.Build()

	var total int

	if err := r.db.QueryRow(ctx, sql, args...).Scan(&total); err != nil {
		return nil, err
	}

	qb.Select() // reset select

	setupSelectQueryBuilder(qb)

	qb.Offset(filter.Offset)
	qb.Limit(filter.Limit)

	validAssetOrderBy := map[string]bool{
		"name":           true,
		"acquired_at":    true,
		"start_date":     true,
		"completed_date": true,
		"status":         true,
		"target_price":   true,
	}
	if filter.OrderBy != "" && filter.Order != "" && validAssetOrderBy[filter.OrderBy] {
		qb.OrderBy(fmt.Sprintf("projecta_assets.%s %s", filter.OrderBy, filter.Order.String()))
	} else {
		qb.OrderBy("projecta_assets.acquired_at DESC")
	}

	sql, args = qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collection := asset.NewCollection(total)

	for rows.Next() {
		var (
			assetID          string
			name             string
			description      string
			projectID        string
			acquiredAt       time.Time
			ownerID          string
			ownerDisplayName string
			status           string
			startDate        time.Time
			completedDate    *time.Time
			targetPrice      *int64
			targetCurrency   *string
			tags             []string
		)

		if err = rows.Scan(
			&assetID,
			&name,
			&description,
			&projectID,
			&acquiredAt,
			&ownerID,
			&ownerDisplayName,
			&status,
			&startDate,
			&completedDate,
			&targetPrice,
			&targetCurrency,
			&tags,
		); err != nil {
			return nil, err
		}

		a, err := toAssetFromPg(
			assetID,
			name,
			description,
			projectID,
			acquiredAt,
			ownerID,
			ownerDisplayName,
			status,
			startDate,
			completedDate,
			targetPrice,
			targetCurrency,
			tags,
		)

		if err != nil {
			return nil, err
		}

		collection.Add(a)
	}

	r.loadCompositionsAndRollup(ctx, collection.Elements(), filter.ProjectID)

	return collection, nil
}

func (r *PgAssetRepository) loadCompositionsAndRollup(ctx context.Context, assets []*asset.Asset, projectID core.ProjectID) {
	if len(assets) == 0 {
		return
	}

	hierarchy := asset.NewAssetHierarchy()
	for _, a := range assets {
		hierarchy.AddAsset(a)
	}

	assetIDs := make([]string, len(assets))
	for i, a := range assets {
		assetIDs[i] = a.ID().String()
	}

	directCosts := make(map[core.AssetID]*money.Money)
	var defaultCurr string = "UAH"

	// 1. Direct costs
	var dcSql string
	var dcArgs []any
	if !projectID.IsNil() {
		dcSql = `
			SELECT ia.asset_id,
			       COALESCE(SUM(ROUND(inv.amount * ia.share_percentage / 100.0)), 0)::bigint,
			       COALESCE(MAX(inv.currency), 'UAH')
			FROM projecta_investment_assets ia
			JOIN projecta_investments inv ON inv.investment_id = ia.investment_id
			WHERE inv.project_id = $1
			GROUP BY ia.asset_id
		`
		dcArgs = []any{projectID.String()}
	} else {
		dcSql = `
			SELECT ia.asset_id,
			       COALESCE(SUM(ROUND(inv.amount * ia.share_percentage / 100.0)), 0)::bigint,
			       COALESCE(MAX(inv.currency), 'UAH')
			FROM projecta_investment_assets ia
			JOIN projecta_investments inv ON inv.investment_id = ia.investment_id
			WHERE ia.asset_id = ANY($1::uuid[])
			GROUP BY ia.asset_id
		`
		dcArgs = []any{assetIDs}
	}

	if rows, err := r.db.Query(ctx, dcSql, dcArgs...); err == nil {
		for rows.Next() {
			var (
				aid    string
				amount int64
				curr   string
			)
			if err := rows.Scan(&aid, &amount, &curr); err == nil {
				if parsedID, err := core.ParseAssetID(aid); err == nil {
					directCosts[parsedID] = money.New(amount, curr)
				}
			}
		}
		rows.Close()
	}

	// 2. Compositions
	var compSql string
	var compArgs []any
	if !projectID.IsNil() {
		compSql = `
			SELECT c.parent_asset_id, c.child_asset_id, c.share_percentage
			FROM projecta_asset_compositions c
			JOIN projecta_assets pa ON pa.asset_id = c.parent_asset_id
			WHERE pa.project_id = $1
		`
		compArgs = []any{projectID.String()}
	} else {
		compSql = `
			SELECT c.parent_asset_id, c.child_asset_id, c.share_percentage
			FROM projecta_asset_compositions c
			WHERE c.parent_asset_id = ANY($1::uuid[]) OR c.child_asset_id = ANY($1::uuid[])
		`
		compArgs = []any{assetIDs}
	}

	if rows, err := r.db.Query(ctx, compSql, compArgs...); err == nil {
		for rows.Next() {
			var (
				parentID string
				childID  string
				share    float64
			)
			if err := rows.Scan(&parentID, &childID, &share); err == nil {
				cID, err1 := core.ParseAssetID(childID)
				pID, err2 := core.ParseAssetID(parentID)
				if err1 == nil && err2 == nil {
					_ = hierarchy.AddEdge(pID, cID, share)
				}
			}
		}
		rows.Close()
	}

	// 3. Delegate valuation rollups & link resolution to domain aggregate
	hierarchy.ApplyToAssets(directCosts, defaultCurr)
}

func setupSelectQueryBuilder(qb *sqlbuilder.SelectBuilder) {
	qb.Select(
		"projecta_assets.asset_id",
		"projecta_assets.name",
		qb.As("COALESCE(projecta_assets.description, '')", "description"),
		"projecta_assets.project_id",
		"projecta_assets.acquired_at",
		"projecta_assets.owner_id",
		qb.As("COALESCE(people.display_name, people.first_name, '')", "display_name"),
		"projecta_assets.status",
		"projecta_assets.start_date",
		"projecta_assets.completed_date",
		"projecta_assets.target_price",
		"projecta_assets.target_currency",
		"projecta_assets.tags",
	)

	qb.Join("people", "people.person_id = projecta_assets.owner_id")
}

func toAssetFromPg(
	assetID string,
	name string,
	description string,
	projectID string,
	acquiredAt time.Time,
	ownerID string,
	ownerDisplayName string,
	status string,
	startDate time.Time,
	completedDate *time.Time,
	targetPrice *int64,
	targetCurrency *string,
	tags []string,
) (*asset.Asset, error) {
	owner := asset.OwnerRef{
		PersonID:    core.MustParsePersonID(ownerID),
		DisplayName: ownerDisplayName,
	}

	var targetPriceMoney *money.Money
	if targetPrice != nil && targetCurrency != nil && *targetCurrency != "" {
		targetPriceMoney = money.New(*targetPrice, *targetCurrency)
	}

	return asset.ReconstituteAsset(
		core.MustParseAssetID(assetID),
		name,
		description,
		core.MustParseProjectID(projectID),
		targetPriceMoney,
		acquiredAt,
		owner,
		asset.ToAssetStatus(status),
		startDate,
		completedDate,
		targetPriceMoney,
		tags,
	), nil
}
