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
	}

	tags := anAsset.Tags()
	if tags == nil {
		tags = []string{}
	}

	sql := `
		INSERT INTO projecta_assets (
			asset_id, name, description, project_id, owner_id,
			status, start_date, completed_date, target_completion_date,
			target_price, target_currency, acquired_at, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (asset_id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			status = EXCLUDED.status,
			start_date = EXCLUDED.start_date,
			completed_date = EXCLUDED.completed_date,
			target_completion_date = EXCLUDED.target_completion_date,
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
		anAsset.TargetCompletionDate(),
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
	return r.findOneInternal(ctx, filter, false)
}

func (r *PgAssetRepository) UpdateAsset(
	ctx context.Context,
	assetID core.AssetID,
	projectID core.ProjectID,
	ownerID core.PersonID,
	updateFn func(ctx context.Context, a *asset.Asset) (*asset.Asset, error),
) error {
	execFn := func(txCtx context.Context) (any, error) {
		existing, err := r.findOneInternal(txCtx, asset.Filter{
			ID:        assetID,
			ProjectID: projectID,
			OwnerID:   ownerID,
		}, true)
		if err != nil {
			return nil, err
		}
		updated, err := updateFn(txCtx, existing)
		if err != nil {
			return nil, err
		}
		if updated == nil {
			return nil, exceptions.NewValidationException("updated asset cannot be nil", nil)
		}
		if err := r.Save(txCtx, updated); err != nil {
			return nil, err
		}
		return nil, nil
	}

	if r.db != nil && r.db.db != nil {
		_, err := r.db.db.Tx(ctx, execFn)
		return err
	}
	_, err := execFn(ctx)
	return err
}

type assetRecord struct {
	assetID              string
	name                 string
	description          string
	projectID            string
	acquiredAt           time.Time
	ownerID              string
	ownerDisplayName     string
	status               string
	startDate            time.Time
	completedDate        *time.Time
	targetPrice          *int64
	targetCurrency       *string
	tags                 []string
	targetCompletionDate *time.Time
}

func (r *assetRecord) scanFields() []any {
	return []any{
		&r.assetID,
		&r.name,
		&r.description,
		&r.projectID,
		&r.acquiredAt,
		&r.ownerID,
		&r.ownerDisplayName,
		&r.status,
		&r.startDate,
		&r.completedDate,
		&r.targetPrice,
		&r.targetCurrency,
		&r.tags,
		&r.targetCompletionDate,
	}
}

func (r *assetRecord) toAsset() (*asset.Asset, error) {
	aID, err := core.ParseAssetID(r.assetID)
	if err != nil {
		return nil, fmt.Errorf("invalid asset id: %w", err)
	}
	pID, err := core.ParseProjectID(r.projectID)
	if err != nil {
		return nil, fmt.Errorf("invalid project id: %w", err)
	}
	oID, err := core.ParsePersonID(r.ownerID)
	if err != nil {
		return nil, fmt.Errorf("invalid owner id: %w", err)
	}

	owner := asset.OwnerRef{
		PersonID:    oID,
		DisplayName: r.ownerDisplayName,
	}

	var targetPriceMoney *money.Money
	if r.targetPrice != nil && r.targetCurrency != nil && *r.targetCurrency != "" {
		targetPriceMoney = money.New(*r.targetPrice, *r.targetCurrency)
	}

	return asset.ReconstituteAsset(
		aID,
		r.name,
		r.description,
		pID,
		targetPriceMoney,
		r.acquiredAt,
		owner,
		asset.ToAssetStatus(r.status),
		r.startDate,
		r.completedDate,
		r.targetCompletionDate,
		targetPriceMoney,
		r.tags,
	), nil
}

func (r *PgAssetRepository) findOneInternal(ctx context.Context, filter asset.Filter, forUpdate bool) (*asset.Asset, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_assets")

	setupSelectQueryBuilder(qb)

	qb.Where(qb.Equal("projecta_assets.asset_id", filter.ID.String()))

	if !filter.ProjectID.IsNil() {
		qb.Where(qb.Equal("projecta_assets.project_id", filter.ProjectID.String()))
	}

	applyAssetOwnerFilter(qb, filter.OwnerID)

	if filter.Name != "" {
		qb.Where(qb.ILike("projecta_assets.name", fmt.Sprintf("%s%%", filter.Name)))
	}

	sql, args := qb.Build()
	if forUpdate {
		sql += " FOR UPDATE OF projecta_assets"
	}

	var rec assetRecord
	if err := r.db.QueryRow(ctx, sql, args...).Scan(rec.scanFields()...); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}

		return nil, err
	}

	a, err := rec.toAsset()
	if err != nil {
		return nil, errors.Join(ErrAssetNotFound, err)
	}

	r.loadCompositionsAndRollup(ctx, []*asset.Asset{a}, a.ProjectID())
	return a, nil
}

func (r *PgAssetRepository) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_assets")

	applyAssetOwnerFilter(qb, filter.OwnerID)

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
	if filter.Limit > 0 {
		qb.Limit(filter.Limit)
	}

	validAssetOrderBy := map[string]bool{
		"name":           true,
		"acquired_at":    true,
		"start_date":     true,
		"completed_date": true,
		"status":         true,
		"target_price":   true,
	}
	if filter.OrderBy != "" && filter.Order != "" && validAssetOrderBy[filter.OrderBy] {
		orderDir := core.ToOrder(filter.Order.String()).String()
		qb.OrderBy(fmt.Sprintf("projecta_assets.%s %s, projecta_assets.asset_id %s", filter.OrderBy, orderDir, orderDir))
	} else {
		qb.OrderBy("projecta_assets.acquired_at DESC, projecta_assets.asset_id DESC")
	}

	sql, args = qb.Build()

	rows, err := r.db.Query(ctx, sql, args...)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collection := asset.NewCollection(total)

	for rows.Next() {
		var rec assetRecord
		if err = rows.Scan(rec.scanFields()...); err != nil {
			return nil, err
		}

		a, err := rec.toAsset()
		if err != nil {
			return nil, err
		}

		collection.Add(a)
	}
	rows.Close()

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

func applyAssetOwnerFilter(qb *sqlbuilder.SelectBuilder, ownerID core.PersonID) {
	if ownerID.IsNil() {
		return
	}
	qb.Where(fmt.Sprintf(
		"projecta_assets.project_id IN (SELECT project_id FROM projecta_projects WHERE owner_id = %s UNION SELECT project_id FROM projecta_project_shares WHERE person_id = %s)",
		qb.Var(ownerID.String()),
		qb.Var(ownerID.String()),
	))
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
		"projecta_assets.target_completion_date",
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
	targetCompletionDate *time.Time,
	targetPrice *int64,
	targetCurrency *string,
	tags []string,
) (*asset.Asset, error) {
	rec := assetRecord{
		assetID:              assetID,
		name:                 name,
		description:          description,
		projectID:            projectID,
		acquiredAt:           acquiredAt,
		ownerID:              ownerID,
		ownerDisplayName:     ownerDisplayName,
		status:               status,
		startDate:            startDate,
		completedDate:        completedDate,
		targetCompletionDate: targetCompletionDate,
		targetPrice:          targetPrice,
		targetCurrency:       targetCurrency,
		tags:                 tags,
	}
	return rec.toAsset()
}
