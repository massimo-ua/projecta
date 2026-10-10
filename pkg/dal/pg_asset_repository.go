package dal

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jackc/pgx/v5"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

var ErrFailedToSaveAsset = errors.New("failed to save asset")
var ErrAssetNotFound = errors.New("asset not found")

type PgAssetRepository struct {
	db *PgRepository
}

func NewPgAssetRepository(conn *PgDbConnection) *PgAssetRepository {
	return &PgAssetRepository{
		db: &PgRepository{db: conn},
	}
}

func (r *PgAssetRepository) Save(ctx context.Context, anAsset *asset.Asset) error {
	_, err := r.FindOne(ctx, asset.Filter{
		ID: anAsset.ID(),
	})

	if err != nil {
		if errors.Is(err, ErrAssetNotFound) {
			return r.create(ctx, anAsset)
		}

		return err
	}

	return r.update(ctx, anAsset)
}

func (r *PgAssetRepository) create(ctx context.Context, anAsset *asset.Asset) error {
	qb := sqlbuilder.PostgreSQL.NewInsertBuilder()

	qb.InsertInto("projecta_assets")
	qb.Cols(
		"asset_id",
		"name",
		"description",
		"project_id",
		"owner_id",
		"status",
		"start_date",
		"completed_date",
		"target_price",
		"target_currency",
		"acquired_at",
		"tags")

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

	qb.Values(
		anAsset.ID().String(),
		anAsset.Name(),
		anAsset.Description(),
		anAsset.Project().ProjectID.String(),
		anAsset.Owner().PersonID.String(),
		anAsset.Status().String(),
		anAsset.StartDate(),
		anAsset.CompletedDate(),
		targetPriceVal,
		targetCurrVal,
		anAsset.AcquiredAt(),
		tags)

	sql, args := qb.Build()

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return errors.Join(ErrFailedToSaveAsset, err)
	}

	return nil
}

func (r *PgAssetRepository) update(ctx context.Context, anAsset *asset.Asset) error {
	qb := sqlbuilder.PostgreSQL.NewUpdateBuilder()

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

	qb.Update("projecta_assets")
	qb.Set(
		qb.Assign("name", anAsset.Name()),
		qb.Assign("description", anAsset.Description()),
		qb.Assign("status", anAsset.Status().String()),
		qb.Assign("start_date", anAsset.StartDate()),
		qb.Assign("completed_date", anAsset.CompletedDate()),
		qb.Assign("target_price", targetPriceVal),
		qb.Assign("target_currency", targetCurrVal),
		qb.Assign("acquired_at", anAsset.AcquiredAt()),
		qb.Assign("tags", tags),
	)
	qb.Where(qb.Equal("asset_id", anAsset.ID().String()))
	qb.Where(qb.Equal("owner_id", anAsset.Owner().PersonID.String()))

	sql, args := qb.Build()

	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return errors.Join(ErrFailedToSaveAsset, err)
	}

	return nil
}

func (r *PgAssetRepository) Remove(ctx context.Context, asset *asset.Asset) error {
	qb := sqlbuilder.PostgreSQL.NewDeleteBuilder()
	qb.DeleteFrom("projecta_assets")
	qb.Where(qb.Equal("asset_id", asset.ID().String()))
	qb.Where(qb.Equal("owner_id", asset.Owner().PersonID.String()))

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

func (r *PgAssetRepository) AddChild(ctx context.Context, parentID uuid.UUID, childID uuid.UUID, sharePercentage float64) error {
	sql := `
		INSERT INTO projecta_asset_compositions (parent_asset_id, child_asset_id, share_percentage)
		VALUES ($1, $2, $3)
		ON CONFLICT (parent_asset_id, child_asset_id) DO UPDATE SET share_percentage = EXCLUDED.share_percentage
	`
	_, err := r.db.Exec(ctx, sql, parentID.String(), childID.String(), sharePercentage)
	return err
}

func (r *PgAssetRepository) RemoveChild(ctx context.Context, parentID uuid.UUID, childID uuid.UUID) error {
	sql := `DELETE FROM projecta_asset_compositions WHERE parent_asset_id = $1 AND child_asset_id = $2`
	_, err := r.db.Exec(ctx, sql, parentID.String(), childID.String())
	return err
}

func (r *PgAssetRepository) FindChildren(ctx context.Context, parentID uuid.UUID) ([]asset.ChildAssetLink, error) {
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
			ChildID:         uuid.MustParse(cid),
			ChildName:       cname,
			SharePercentage: share,
		})
	}
	return links, nil
}

func (r *PgAssetRepository) FindParents(ctx context.Context, childID uuid.UUID) ([]asset.ParentAssetLink, error) {
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
			ParentID:        uuid.MustParse(pid),
			ParentName:      pname,
			SharePercentage: share,
		})
	}
	return links, nil
}

func (r *PgAssetRepository) FindAncestors(ctx context.Context, assetID uuid.UUID) ([]uuid.UUID, error) {
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

	var ancestors []uuid.UUID
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err == nil {
			ancestors = append(ancestors, uuid.MustParse(pid))
		}
	}
	return ancestors, nil
}

func (r *PgAssetRepository) FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_assets")

	setupSelectQueryBuilder(qb)

	qb.Where(qb.Equal("projecta_assets.asset_id", filter.ID.String()))

	if filter.OwnerID != uuid.Nil {
		qb.Where(qb.Equal("projecta_assets.owner_id", filter.OwnerID.String()))
	}

	if filter.Name != "" {
		qb.Where(qb.ILike("projecta_assets.name", fmt.Sprintf("%s%%", filter.Name)))
	}

	sql, args := qb.Build()

	var (
		assetID             string
		name                string
		description         string
		projectID           string
		projectName         string
		projectDescription  string
		projectMainCurrency string
		acquiredAt          time.Time
		ownerID             string
		ownerFirstName      string
		ownerDisplayName    string
		status              string
		startDate           time.Time
		completedDate       *time.Time
		targetPrice         *int64
		targetCurrency      *string
		tags                []string
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
		&projectName,
		&projectDescription,
		&projectMainCurrency,
		&acquiredAt,
		&ownerID,
		&ownerFirstName,
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
		projectName,
		projectDescription,
		projectMainCurrency,
		acquiredAt,
		ownerID,
		ownerFirstName,
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

	r.loadCompositionsAndRollup(ctx, []*asset.Asset{a}, a.Project().ProjectID)
	return a, nil
}

func (r *PgAssetRepository) Find(ctx context.Context, filter asset.CollectionFilter) (*asset.Collection, error) {
	qb := sqlbuilder.PostgreSQL.NewSelectBuilder()
	qb.From("projecta_assets")

	if filter.OwnerID != uuid.Nil {
		qb.Where(fmt.Sprintf("(projecta_assets.owner_id = %s OR projecta_assets.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))", qb.Var(filter.OwnerID.String()), qb.Var(filter.OwnerID.String())))
	}

	if filter.ProjectID != uuid.Nil {
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

	if filter.OrderBy != "" && filter.Order != "" {
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
			assetID             string
			name                string
			description         string
			projectID           string
			projectName         string
			projectDescription  string
			projectMainCurrency string
			acquiredAt          time.Time
			ownerID             string
			ownerFirstName      string
			ownerDisplayName    string
			status              string
			startDate           time.Time
			completedDate       *time.Time
			targetPrice         *int64
			targetCurrency      *string
			tags                []string
		)

		if err = rows.Scan(
			&assetID,
			&name,
			&description,
			&projectID,
			&projectName,
			&projectDescription,
			&projectMainCurrency,
			&acquiredAt,
			&ownerID,
			&ownerFirstName,
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
			projectName,
			projectDescription,
			projectMainCurrency,
			acquiredAt,
			ownerID,
			ownerFirstName,
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

type rollupContext struct {
	directCosts map[string]*money.Money
	childrenMap map[string][]asset.ChildAssetLink
	parentsMap  map[string][]asset.ParentAssetLink
	memo        map[string]*money.Money
	visiting    map[string]bool
}

func (rc *rollupContext) compute(id string) *money.Money {
	if cost, ok := rc.memo[id]; ok {
		return cost
	}
	if rc.visiting[id] {
		if dc, ok := rc.directCosts[id]; ok {
			return dc
		}
		return money.New(0, "UAH")
	}
	rc.visiting[id] = true

	dc, ok := rc.directCosts[id]
	if !ok || dc == nil {
		dc = money.New(0, "UAH")
	}
	amount := dc.Amount()
	curr := dc.Currency().Code

	for _, ch := range rc.childrenMap[id] {
		chCost := rc.compute(ch.ChildID.String())
		if chCost != nil {
			contrib := int64(math.Round(float64(chCost.Amount()) * ch.SharePercentage / 100.0))
			amount += contrib
		}
	}
	rc.visiting[id] = false

	res := money.New(amount, curr)
	rc.memo[id] = res
	return res
}

func (r *PgAssetRepository) loadCompositionsAndRollup(ctx context.Context, assets []*asset.Asset, projectID uuid.UUID) {
	if len(assets) == 0 {
		return
	}

	rc := &rollupContext{
		directCosts: make(map[string]*money.Money),
		childrenMap: make(map[string][]asset.ChildAssetLink),
		parentsMap:  make(map[string][]asset.ParentAssetLink),
		memo:        make(map[string]*money.Money),
		visiting:    make(map[string]bool),
	}

	assetIDs := make([]string, len(assets))
	for i, a := range assets {
		assetIDs[i] = a.ID().String()
	}

	// 1. Direct costs
	var dcSql string
	var dcArgs []any
	if projectID != uuid.Nil {
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
				rc.directCosts[aid] = money.New(amount, curr)
			}
		}
		rows.Close()
	}

	// 2. Compositions
	var compSql string
	var compArgs []any
	if projectID != uuid.Nil {
		compSql = `
			SELECT c.parent_asset_id, c.child_asset_id, ca.name, pa.name, c.share_percentage
			FROM projecta_asset_compositions c
			JOIN projecta_assets ca ON ca.asset_id = c.child_asset_id
			JOIN projecta_assets pa ON pa.asset_id = c.parent_asset_id
			WHERE pa.project_id = $1
		`
		compArgs = []any{projectID.String()}
	} else {
		compSql = `
			SELECT c.parent_asset_id, c.child_asset_id, ca.name, pa.name, c.share_percentage
			FROM projecta_asset_compositions c
			JOIN projecta_assets ca ON ca.asset_id = c.child_asset_id
			JOIN projecta_assets pa ON pa.asset_id = c.parent_asset_id
			WHERE c.parent_asset_id = ANY($1::uuid[]) OR c.child_asset_id = ANY($1::uuid[])
		`
		compArgs = []any{assetIDs}
	}

	if rows, err := r.db.Query(ctx, compSql, compArgs...); err == nil {
		for rows.Next() {
			var (
				parentID   string
				childID    string
				childName  string
				parentName string
				share      float64
			)
			if err := rows.Scan(&parentID, &childID, &childName, &parentName, &share); err == nil {
				cUUID, err1 := uuid.Parse(childID)
				pUUID, err2 := uuid.Parse(parentID)
				if err1 == nil && err2 == nil {
					rc.childrenMap[parentID] = append(rc.childrenMap[parentID], asset.ChildAssetLink{
						ChildID:         cUUID,
						ChildName:       childName,
						SharePercentage: share,
					})
					rc.parentsMap[childID] = append(rc.parentsMap[childID], asset.ParentAssetLink{
						ParentID:        pUUID,
						ParentName:      parentName,
						SharePercentage: share,
					})
				}
			}
		}
		rows.Close()
	}

	// 3. Set rolled up total costs and components on each asset
	for _, a := range assets {
		aid := a.ID().String()
		dc, ok := rc.directCosts[aid]
		if !ok || dc == nil {
			dc = money.New(0, "UAH")
		}
		a.SetDirectCost(dc)

		tc := rc.compute(aid)
		a.SetTotalCost(tc)

		children := rc.childrenMap[aid]
		resolvedChildren := make([]asset.ChildAssetLink, len(children))
		for i, ch := range children {
			resolvedChildren[i] = ch
			resolvedChildren[i].TotalCost = rc.compute(ch.ChildID.String())
		}
		a.SetChildren(resolvedChildren)
		a.SetParents(rc.parentsMap[aid])
	}
}

func setupSelectQueryBuilder(qb *sqlbuilder.SelectBuilder) {
	qb.Select(
		"projecta_assets.asset_id",
		"projecta_assets.name",
		qb.As("COALESCE(projecta_assets.description, '')", "description"),
		"projecta_projects.project_id",
		qb.As("projecta_projects.name", "project_name"),
		qb.As("COALESCE(projecta_projects.description, '')", "project_description"),
		qb.As("COALESCE(projecta_projects.main_currency, 'UAH')", "project_main_currency"),
		"projecta_assets.acquired_at",
		"projecta_assets.owner_id",
		"people.first_name",
		qb.As("COALESCE(people.display_name, '')", "display_name"),
		"projecta_assets.status",
		"projecta_assets.start_date",
		"projecta_assets.completed_date",
		"projecta_assets.target_price",
		"projecta_assets.target_currency",
		"projecta_assets.tags",
	)

	qb.Join("people", "people.person_id = projecta_assets.owner_id")
	qb.Join("projecta_projects", "projecta_projects.project_id = projecta_assets.project_id")
}

func toAssetFromPg(
	assetID string,
	name string,
	description string,
	projectID string,
	projectName string,
	projectDescription string,
	projectMainCurrency string,
	acquiredAt time.Time,
	ownerID string,
	ownerFirstName string,
	ownerDisplayName string,
	status string,
	startDate time.Time,
	completedDate *time.Time,
	targetPrice *int64,
	targetCurrency *string,
	tags []string,
) (*asset.Asset, error) {
	owner := &projecta.Owner{
		PersonID:    uuid.MustParse(ownerID),
		FirstName:   ownerFirstName,
		DisplayName: ownerDisplayName,
	}

	project := &projecta.Project{
		ProjectID:    uuid.MustParse(projectID),
		Name:         projectName,
		Description:  projectDescription,
		MainCurrency: projectMainCurrency,
	}

	var targetPriceMoney *money.Money
	if targetPrice != nil && targetCurrency != nil && *targetCurrency != "" {
		targetPriceMoney = money.New(*targetPrice, *targetCurrency)
	}

	anAsset := asset.NewAsset(
		uuid.MustParse(assetID),
		name,
		description,
		project,
		nil,
		targetPriceMoney,
		acquiredAt,
		owner,
	)

	anAsset.SetStatus(asset.ToAssetStatus(status))
	anAsset.SetStartDate(startDate)
	anAsset.SetCompletedDate(completedDate)
	anAsset.SetTargetPrice(targetPriceMoney)
	anAsset.SetTags(tags)

	return anAsset, nil
}
