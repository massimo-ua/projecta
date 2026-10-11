package dal

import (
	"context"
	types "database/sql"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/people"
	"gitlab.com/massimo-ua/projecta/internal/projecta"
)

type mockRow struct {
	row []any
	err error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	for i, v := range dest {
		if i >= len(m.row) {
			break
		}
		val := m.row[i]
		if val == nil {
			continue
		}
		switch d := v.(type) {
		case *string:
			if str, ok := val.(string); ok {
				*d = str
			} else {
				*d = fmt.Sprintf("%v", val)
			}
		case *int64:
			switch n := val.(type) {
			case int64:
				*d = n
			case int:
				*d = int64(n)
			}
		case *int:
			switch n := val.(type) {
			case int:
				*d = n
			case int64:
				*d = int(n)
			}
		case *bool:
			*d = val.(bool)
		case *time.Time:
			if tVal, ok := val.(time.Time); ok {
				*d = tVal
			} else if str, ok := val.(string); ok {
				if parsed, err := time.Parse(time.RFC3339, str); err == nil {
					*d = parsed
				}
			}
		case *types.NullString:
			d.String = val.(string)
			d.Valid = true
		}
	}
	return nil
}

type mockRows struct {
	data [][]any
	idx  int
	err  error
}

func (m *mockRows) Close()                                      {}
func (m *mockRows) Err() error                                  { return m.err }
func (m *mockRows) CommandTag() pgconn.CommandTag               { return pgconn.CommandTag{} }
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription { return nil }
func (m *mockRows) Next() bool {
	if m.idx < len(m.data) {
		m.idx++
		return true
	}
	return false
}
func (m *mockRows) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	row := m.data[m.idx-1]
	for i, d := range dest {
		if i >= len(row) {
			break
		}
		val := row[i]
		if val == nil {
			continue
		}
		switch target := d.(type) {
		case *string:
			if str, ok := val.(string); ok {
				*target = str
			} else {
				*target = fmt.Sprintf("%v", val)
			}
		case *int64:
			switch n := val.(type) {
			case int64:
				*target = n
			case int:
				*target = int64(n)
			}
		case *int:
			switch n := val.(type) {
			case int:
				*target = n
			case int64:
				*target = int(n)
			}
		case *bool:
			*target = val.(bool)
		case *time.Time:
			if tVal, ok := val.(time.Time); ok {
				*target = tVal
			} else if str, ok := val.(string); ok {
				if parsed, err := time.Parse(time.RFC3339, str); err == nil {
					*target = parsed
				}
			}
		case *types.NullString:
			target.String = val.(string)
			target.Valid = true
		}
	}
	return nil
}
func (m *mockRows) Values() ([]any, error) { return nil, nil }
func (m *mockRows) RawValues() [][]byte    { return nil }
func (m *mockRows) Conn() *pgx.Conn        { return nil }
func (m *mockRows) TypeMap() *pgtype.Map   { return nil }

type mockPgDb struct {
	execErr    error
	execTag    pgconn.CommandTag
	queryErr   error
	rowsData   [][]any
	rowVal     []any
	rowErr     error
	countErr   error
	isNotFound bool
	zeroTotal  bool
}

func (m *mockPgDb) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	if m.execErr != nil {
		return pgconn.CommandTag{}, m.execErr
	}
	if m.execTag.String() != "" {
		return m.execTag, nil
	}
	return pgconn.NewCommandTag("UPDATE 1"), nil
}

func (m *mockPgDb) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return &mockRows{data: m.rowsData}, nil
}

func (m *mockPgDb) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if m.rowErr != nil {
		return &mockRow{err: m.rowErr}
	}
	if m.isNotFound {
		return &mockRow{err: pgx.ErrNoRows}
	}
	if strings.Contains(sql, "EXISTS") || strings.Contains(sql, "exists") {
		return &mockRow{row: []any{true}}
	}
	if strings.Contains(sql, "COUNT") || strings.Contains(sql, "count") {
		if m.countErr != nil {
			return &mockRow{err: m.countErr}
		}
		if m.zeroTotal {
			return &mockRow{row: []any{int(0), int64(0)}}
		}
		return &mockRow{row: []any{int(1), int64(1)}}
	}
	return &mockRow{row: m.rowVal}
}

func withMockDb(ctx context.Context, db PgDb) context.Context {
	return context.WithValue(ctx, txKey{}, db)
}

func TestPgDbConnectionAndRepo(t *testing.T) {
	t.Run("NewPgDbConnection success and error", func(t *testing.T) {
		_, err := NewPgDbConnection("invalid connection string")
		if err == nil {
			t.Errorf("expected error for invalid config")
		}

		conn, err := NewPgDbConnection("postgres://user:pass@localhost:5432/dbname")
		if err != nil || conn == nil {
			t.Fatalf("unexpected error creating PgDbConnection: %v", err)
		}
		conn.Close()
	})

	t.Run("errorRow Scan", func(t *testing.T) {
		e := &errorRow{err: errors.New("scan error")}
		if err := e.Scan(); err == nil {
			t.Errorf("expected scan error")
		}
	})

	t.Run("PgDbConnection Tx, Close, Ping", func(t *testing.T) {
		emptyConn := &PgDbConnection{}

		// Tx when GetConnection returns error (nil pool)
		_, err := emptyConn.Tx(context.Background(), func(ctx context.Context) (any, error) { return nil, nil })
		if err == nil {
			t.Errorf("expected Tx error when pool is nil")
		}

		// Tx with mockDb in context
		mockDb := &mockPgDb{}
		ctxTx := context.WithValue(context.Background(), txKey{}, mockDb)
		executed := false
		res, err := emptyConn.Tx(ctxTx, func(ctx context.Context) (any, error) {
			executed = true
			return "success", nil
		})
		if err != nil || !executed || res != "success" {
			t.Errorf("expected Tx execution for mock db in context")
		}

		// Tx returning function error
		txErr := errors.New("tx error")
		_, err = emptyConn.Tx(ctxTx, func(ctx context.Context) (any, error) {
			return nil, txErr
		})
		if !errors.Is(err, txErr) {
			t.Errorf("expected txErr to be returned")
		}

		emptyConn.Close()
		_ = emptyConn.Ping(context.Background())

		// PgRepository with nil db
		repo := &PgRepository{db: nil}
		_, err = repo.Exec(context.Background(), "SELECT 1")
		if err == nil {
			t.Errorf("expected error on nil db Exec")
		}
		_, err = repo.Query(context.Background(), "SELECT 1")
		if err == nil {
			t.Errorf("expected error on nil db Query")
		}
		row := repo.QueryRow(context.Background(), "SELECT 1")
		if err = row.Scan(); err == nil {
			t.Errorf("expected error on nil db QueryRow Scan")
		}
	})
}

func TestPgPeopleRepository(t *testing.T) {
	dbConn := &PgDbConnection{}
	repo := NewPgPeopleRepository(dbConn)
	pID := uuid.New()
	cred, _ := people.NewCredentials("LOCAL", "john@example.com", "secret")
	person, _ := people.NewPerson(pID, "John", "Doe", "J.D.", []people.Credentials{cred})

	t.Run("Register success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{}
		ctx := withMockDb(context.Background(), mockDb)

		err := repo.Register(ctx, person)
		if err != nil {
			t.Errorf("unexpected Register error: %v", err)
		}

		personNoCreds, _ := people.NewPerson(uuid.New(), "Jane", "Doe", "J.D.", nil, people.RoleUser)
		if err := repo.Register(ctx, personNoCreds); err != nil {
			t.Errorf("unexpected Register without credentials error: %v", err)
		}

		mockDbErr := &mockPgDb{execErr: errors.New("exec error")}
		ctxErr := withMockDb(context.Background(), mockDbErr)
		err = repo.Register(ctxErr, person)
		if err == nil {
			t.Errorf("expected exec error")
		}
	})

	t.Run("FindCredentials success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{rowVal: []any{pID.String(), "john@example.com"}}
		ctx := withMockDb(context.Background(), mockDb)

		gotID, identity, err := repo.FindCredentials(ctx, "LOCAL", "john@example.com")
		if err != nil || gotID != pID || identity != "john@example.com" {
			t.Errorf("FindCredentials error: %v", err)
		}

		mockDbErr := &mockPgDb{rowErr: pgx.ErrNoRows}
		ctxErr := withMockDb(context.Background(), mockDbErr)
		_, _, err = repo.FindCredentials(ctxErr, "LOCAL", "bad")
		if err == nil {
			t.Errorf("expected not found error")
		}

		mockDbBadUUID := &mockPgDb{rowVal: []any{"not-uuid", "john@example.com"}}
		ctxBadUUID := withMockDb(context.Background(), mockDbBadUUID)
		_, _, err = repo.FindCredentials(ctxBadUUID, "LOCAL", "john@example.com")
		if err == nil {
			t.Errorf("expected bad UUID error")
		}
	})

	t.Run("FindByID success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{rowVal: []any{"John", "Doe", "J.D."}}
		ctx := withMockDb(context.Background(), mockDb)

		p, err := repo.FindByID(ctx, pID)
		if err != nil || p == nil {
			t.Errorf("FindByID error: %v", err)
		}

		mockDbErr := &mockPgDb{rowErr: pgx.ErrNoRows}
		ctxErr := withMockDb(context.Background(), mockDbErr)
		_, err = repo.FindByID(ctxErr, pID)
		if err == nil {
			t.Errorf("expected not found error")
		}

		mockDbOtherErr := &mockPgDb{rowErr: errors.New("db failure")}
		ctxOther := withMockDb(context.Background(), mockDbOtherErr)
		_, err = repo.FindByID(ctxOther, pID)
		if err == nil {
			t.Errorf("expected internal error on db failure")
		}
	})

	t.Run("toPersonFromPg test", func(t *testing.T) {
		p, err := toPersonFromPg(pID.String(), "John", "Doe", "J.D.", people.RoleUser, people.RoleAdministrator)
		if err != nil || p.FirstName() != "John" || !p.IsAdministrator() || !p.IsUser() {
			t.Errorf("toPersonFromPg error: %v", err)
		}
	})

	t.Run("SaveRoles success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{}
		ctx := withMockDb(context.Background(), mockDb)

		err := repo.SaveRoles(ctx, pID, []people.Role{people.RoleUser, people.RoleAdministrator})
		if err != nil {
			t.Errorf("unexpected SaveRoles error: %v", err)
		}

		err = repo.SaveRoles(ctx, pID, []people.Role{})
		if err != nil {
			t.Errorf("unexpected SaveRoles with empty roles error: %v", err)
		}

		mockDbErr := &mockPgDb{execErr: errors.New("delete failure")}
		ctxErr := withMockDb(context.Background(), mockDbErr)
		err = repo.SaveRoles(ctxErr, pID, []people.Role{people.RoleUser})
		if err == nil {
			t.Errorf("expected error on delete failure")
		}
	})

	t.Run("FindAll success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{
			rowVal:   []any{1}, // count(*)
			rowsData: [][]any{{pID.String(), "John", "Doe", "J.D."}},
		}
		ctx := withMockDb(context.Background(), mockDb)

		list, total, err := repo.FindAll(ctx, core.Pagination{Limit: 10, Offset: 0})
		if err != nil || total != 1 || len(list) != 1 {
			t.Fatalf("FindAll failed: %v", err)
		}

		mockDbCountErr := &mockPgDb{rowErr: errors.New("count failed")}
		ctxCountErr := withMockDb(context.Background(), mockDbCountErr)
		_, _, err = repo.FindAll(ctxCountErr, core.Pagination{Limit: 10, Offset: 0})
		if err == nil {
			t.Errorf("expected count error")
		}

		mockDbQueryErr := &mockPgDb{rowVal: []any{1}, queryErr: errors.New("query failed")}
		ctxQueryErr := withMockDb(context.Background(), mockDbQueryErr)
		_, _, err = repo.FindAll(ctxQueryErr, core.Pagination{Limit: 10, Offset: 0})
		if err == nil {
			t.Errorf("expected query error")
		}
	})
}

func TestPgProjectRepository(t *testing.T) {
	dbConn := &PgDbConnection{}
	repo := NewPgProjectRepository(dbConn)

	pID := uuid.New()
	ownerID := uuid.New()
	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "John Doe"}
	now := time.Now()
	proj, _ := projecta.NewProject(pID, "Project A", "Desc", owner, now, now)

	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, ownerID)

	t.Run("FindOne unauthenticated", func(t *testing.T) {
		_, err := repo.FindOne(context.Background(), projecta.ProjectFilter{})
		if err == nil {
			t.Errorf("expected unauthenticated error")
		}
	})

	t.Run("FindOne success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{rowVal: []any{pID.String(), "Project A", "Desc", ownerID.String(), now, now, "John", "Doe", "J.D.", uuid.New().String(), "UAH"}}
		ctx := withMockDb(authedCtx, mockDb)

		p, err := repo.FindOne(ctx, projecta.ProjectFilter{ProjectID: pID, Name: "Project A"})
		if err != nil || p == nil {
			t.Errorf("FindOne error: %v", err)
		}

		// Shared project (different requester)
		sharedPersonID := uuid.New()
		sharedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, sharedPersonID)
		ctxShared := withMockDb(sharedCtx, mockDb)
		pShared, err := repo.FindOne(ctxShared, projecta.ProjectFilter{ProjectID: pID})
		if err != nil || !pShared.IsShared() {
			t.Errorf("expected shared project")
		}

		mockDbErr := &mockPgDb{rowErr: pgx.ErrNoRows}
		ctxErr := withMockDb(authedCtx, mockDbErr)
		_, err = repo.FindOne(ctxErr, projecta.ProjectFilter{})
		if err == nil {
			t.Errorf("expected not found error")
		}

		mockDbQueryErr := &mockPgDb{rowErr: errors.New("db error")}
		ctxQueryErr := withMockDb(authedCtx, mockDbQueryErr)
		_, err = repo.FindOne(ctxQueryErr, projecta.ProjectFilter{})
		if err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("FindOne with participants from shares", func(t *testing.T) {
		mockDb := &mockPgDb{
			rowVal:   []any{pID.String(), "Project A", "Desc", ownerID.String(), now, now, "John", "Doe", "Owner Display", uuid.New().String(), "UAH"},
			rowsData: [][]any{{"Jane", "Smith", "Jane Member"}},
		}
		ctx := withMockDb(authedCtx, mockDb)

		p, err := repo.FindOne(ctx, projecta.ProjectFilter{ProjectID: pID})
		if err != nil || p == nil {
			t.Fatalf("FindOne error: %v", err)
		}
		if len(p.Participants()) != 2 {
			t.Fatalf("expected 2 participants, got %d", len(p.Participants()))
		}
		if p.Participants()[0].DisplayName() != "Owner Display" {
			t.Errorf("expected owner 'Owner Display', got '%s'", p.Participants()[0].DisplayName())
		}
		if p.Participants()[1].DisplayName() != "Jane Member" {
			t.Errorf("expected shared participant 'Jane Member', got '%s'", p.Participants()[1].DisplayName())
		}
	})

	t.Run("Create, Update, Remove", func(t *testing.T) {
		mockDb := &mockPgDb{}
		ctx := withMockDb(authedCtx, mockDb)

		err := repo.Create(ctx, proj)
		if err != nil {
			t.Errorf("Create error: %v", err)
		}

		token := uuid.New()
		projWithToken := projecta.ReconstituteProject(pID, "Project A", "Desc", owner, now, now, token, false, "UAH", nil)
		err = repo.Create(ctx, projWithToken)
		if err != nil {
			t.Errorf("Create error with existing share token: %v", err)
		}

		projEmptyCurr := projecta.ReconstituteProject(pID, "Project A", "Desc", owner, now, now, token, false, "", nil)
		err = repo.Create(ctx, projEmptyCurr)
		if err != nil {
			t.Errorf("Create error with empty currency: %v", err)
		}

		err = repo.Update(ctx, projEmptyCurr)
		if err != nil {
			t.Errorf("Update error: %v", err)
		}

		err = repo.Remove(ctx, proj)
		if err != nil {
			t.Errorf("Remove error: %v", err)
		}
	})

	t.Run("Find collection success and errors", func(t *testing.T) {
		mockDb := &mockPgDb{rowsData: [][]any{{pID.String(), "Project A", "Desc", ownerID.String(), now, now, "John", "Doe", "J.D.", uuid.New().String(), "UAH"}}}
		ctx := withMockDb(authedCtx, mockDb)

		projects, err := repo.Find(ctx, projecta.ProjectCollectionFilter{Name: "Project A", Pagination: core.Pagination{Limit: 10, Offset: 0}})
		if err != nil || len(projects) != 1 {
			t.Errorf("Find error: %v", err)
		}

		// Unauthenticated
		_, err = repo.Find(context.Background(), projecta.ProjectCollectionFilter{})
		if err == nil {
			t.Errorf("expected unauthenticated error")
		}

		// Query error
		mockDbErr := &mockPgDb{queryErr: errors.New("query error")}
		ctxErr := withMockDb(authedCtx, mockDbErr)
		_, err = repo.Find(ctxErr, projecta.ProjectCollectionFilter{})
		if err == nil {
			t.Errorf("expected query error")
		}
	})

	t.Run("FindByShareToken", func(t *testing.T) {
		token := uuid.New()
		mockDb := &mockPgDb{rowVal: []any{pID.String(), "Project A", "Desc", ownerID.String(), now, now, "John", "Doe", "J.D.", token.String(), "UAH"}}
		ctx := withMockDb(authedCtx, mockDb)

		p, err := repo.FindByShareToken(ctx, token)
		if err != nil || p == nil {
			t.Errorf("FindByShareToken error: %v", err)
		}

		// Invalid share token
		_, err = repo.FindByShareToken(ctx, uuid.Nil)
		if err == nil {
			t.Errorf("expected validation error for nil token")
		}

		// Not found
		mockDbNotFound := &mockPgDb{rowErr: pgx.ErrNoRows}
		ctxNotFound := withMockDb(authedCtx, mockDbNotFound)
		_, err = repo.FindByShareToken(ctxNotFound, token)
		if err == nil {
			t.Errorf("expected not found error")
		}

		// Other DB error
		mockDbErr := &mockPgDb{rowErr: errors.New("db error")}
		ctxErr := withMockDb(authedCtx, mockDbErr)
		_, err = repo.FindByShareToken(ctxErr, token)
		if err == nil {
			t.Errorf("expected db error")
		}
	})

	t.Run("CreateShareRecord", func(t *testing.T) {
		mockDb := &mockPgDb{}
		ctx := withMockDb(authedCtx, mockDb)

		err := repo.CreateShareRecord(ctx, pID, ownerID)
		if err != nil {
			t.Errorf("CreateShareRecord error: %v", err)
		}

		mockDbErr := &mockPgDb{execErr: errors.New("exec error")}
		ctxErr := withMockDb(authedCtx, mockDbErr)
		err = repo.CreateShareRecord(ctxErr, pID, ownerID)
		if err == nil {
			t.Errorf("expected exec error")
		}
	})

	t.Run("toProject test", func(t *testing.T) {
		token := uuid.New()
		rec := projectRecord{
			projectID:     pID.String(),
			name:          "Name",
			description:   "Desc",
			ownerID:       ownerID.String(),
			firstName:     "John",
			lastName:      "Doe",
			displayName:   types.NullString{String: "J.D.", Valid: true},
			startedAt:     now,
			endedAt:       now,
			shareTokenStr: token.String(),
			mainCurrency:  types.NullString{String: "UAH", Valid: true},
		}
		p, err := rec.toProject(false, nil)
		if err != nil || p == nil || p.ShareToken() != token {
			t.Errorf("toProject error: %v", err)
		}

		// Invalid project id
		recInvalidPID := rec
		recInvalidPID.projectID = "invalid"
		_, err = recInvalidPID.toProject(false, nil)
		if err == nil {
			t.Errorf("expected error for invalid project id")
		}

		// Invalid owner id
		recInvalidOID := rec
		recInvalidOID.ownerID = "invalid"
		_, err = recInvalidOID.toProject(false, nil)
		if err == nil {
			t.Errorf("expected error for invalid owner id")
		}
	})
}

func TestPgAssetRepository(t *testing.T) {
	dbConn := &PgDbConnection{}
	projRepo := NewPgProjectRepository(dbConn)
	astRepo := NewPgAssetRepository(dbConn)

	pID := uuid.New()
	astID := uuid.New()
	ownerID := uuid.New()
	now := time.Now()

	owner := &projecta.Owner{PersonID: ownerID, DisplayName: "John"}
	proj, _ := projecta.NewProject(pID, "Project", "Desc", owner, now, now)
	ast := asset.NewAsset(astID, "Laptop", "Desc", proj, money.New(1000, money.USD), now, owner)

	authedCtx := context.WithValue(context.Background(), core.RequesterIDContextKey, ownerID)

	t.Run("PgAssetRepository methods and zero total branch", func(t *testing.T) {
		targetPrice := int64(1000)
		targetCurr := "USD"
		mockDb := &mockPgDb{
			rowVal: []any{
				astID.String(), "Laptop", "Desc", pID.String(), "Project", "Proj Desc", "USD",
				now, ownerID.String(), "John", "J.D.", "ACTIVE", now, nil,
				&targetPrice, &targetCurr, []string{"tag1"},
			},
			rowsData: [][]any{{
				astID.String(), "Laptop", "Desc", pID.String(), "Project", "Proj Desc", "USD",
				now, ownerID.String(), "John", "J.D.", "ACTIVE", now, nil,
				&targetPrice, &targetCurr, []string{"tag1"},
			}},
		}
		ctx := withMockDb(authedCtx, mockDb)

		err := astRepo.Save(ctx, ast)
		if err != nil {
			t.Errorf("Save asset error: %v", err)
		}

		// Save calling create when ErrAssetNotFound
		mockDbCreate := &mockPgDb{rowErr: ErrAssetNotFound}
		ctxCreate := withMockDb(authedCtx, mockDbCreate)
		err = astRepo.Save(ctxCreate, ast)
		if err != nil {
			t.Errorf("Save asset create branch error: %v", err)
		}

		// Save update failure (exec error)
		mockDbUpdErr := &mockPgDb{
			rowVal: []any{
				astID.String(), "Laptop", "Desc", pID.String(), "Project", "Proj Desc", "USD",
				now, ownerID.String(), "John", "J.D.", "ACTIVE", now, nil,
				&targetPrice, &targetCurr, []string{"tag1"},
			},
			execErr: errors.New("exec update error"),
		}
		ctxUpdErr := withMockDb(authedCtx, mockDbUpdErr)
		err = astRepo.Save(ctxUpdErr, ast)
		if err == nil {
			t.Errorf("expected error on asset update failure")
		}

		a, err := astRepo.FindOne(ctx, asset.Filter{ID: astID, ProjectID: pID})
		if err != nil || a == nil {
			t.Errorf("FindOne asset error: %v", err)
		}

		// FindOne non-ErrAssetNotFound error
		mockDbOtherErr := &mockPgDb{rowErr: errors.New("db failure")}
		ctxOtherErr := withMockDb(authedCtx, mockDbOtherErr)
		_, err = astRepo.FindOne(ctxOtherErr, asset.Filter{ID: astID, ProjectID: pID})
		if err == nil {
			t.Errorf("expected error on asset FindOne db failure")
		}

		cols, err := astRepo.Find(ctx, asset.CollectionFilter{
			ProjectID:  pID,
			Name:       "Laptop",
			Sorting:    core.Sorting{OrderBy: "price", Order: core.ASC},
			Pagination: core.Pagination{Limit: 10, Offset: 0},
		})
		if err != nil || cols == nil {
			t.Errorf("Find asset collection error: %v", err)
		}

		// Find zero total
		mockDbZero := &mockPgDb{zeroTotal: true}
		ctxZero := withMockDb(authedCtx, mockDbZero)
		zeroCols, err := astRepo.Find(ctxZero, asset.CollectionFilter{ProjectID: pID})
		if err != nil || zeroCols.Total() != 0 {
			t.Errorf("expected 0 total assets")
		}

		err = astRepo.Remove(ctx, ast)
		if err != nil {
			t.Errorf("Remove asset error: %v", err)
		}

		// Remove zero rows affected error
		mockDbRemZero := &mockPgDb{execTag: pgconn.NewCommandTag("DELETE 0")}
		ctxRemZero := withMockDb(authedCtx, mockDbRemZero)
		err = astRepo.Remove(ctxRemZero, ast)
		if err == nil {
			t.Errorf("expected error on 0 rows deleted for asset")
		}
	})

	t.Run("toAssetFromPg test", func(t *testing.T) {
		tp := int64(1000)
		tc := "USD"
		a, err := toAssetFromPg(astID.String(), "Laptop", "Desc", pID.String(), "Proj", "Desc", "USD", now, ownerID.String(), "John", "J.D.", "ACTIVE", now, nil, &tp, &tc, []string{"tag1"})
		if err != nil || a == nil {
			t.Errorf("toAssetFromPg error: %v", err)
		}
	})

	t.Run("Repository additional FindOne and error branches", func(t *testing.T) {
		// toPersonFromPg with empty fields
		_, err := toPersonFromPg(pID.String(), "", "", "")
		if err == nil {
			t.Errorf("expected error for empty person fields")
		}

		// toProject with valid fields
		rec := projectRecord{
			projectID:    pID.String(),
			name:         "Project 1",
			description:  "Description",
			ownerID:      ownerID.String(),
			firstName:    "John",
			lastName:     "Doe",
			displayName:  types.NullString{String: "J.D.", Valid: true},
			startedAt:    now,
			endedAt:      now,
			mainCurrency: types.NullString{String: "UAH", Valid: true},
		}
		p1, err := rec.toProject(false, nil)
		if err != nil || p1 == nil {
			t.Errorf("toProject error: %v", err)
		}

		mockDbProj := &mockPgDb{rowVal: []any{pID.String(), "Project A", "Desc", ownerID.String(), now, now, "John", "Doe", "J.D."}}
		ctxProj := withMockDb(authedCtx, mockDbProj)
		p, err := projRepo.FindOne(ctxProj, projecta.ProjectFilter{ProjectID: pID, Name: "Project A"})
		if err != nil || p == nil {
			t.Errorf("FindOne project error: %v", err)
		}

		targetPriceVal := int64(1000)
		targetCurrVal := "USD"
		mockDbAst := &mockPgDb{rowVal: []any{
			astID.String(), "Laptop", "Desc", pID.String(), "Project", "Proj Desc", "USD",
			now, ownerID.String(), "John", "J.D.", "ACTIVE", now, nil,
			&targetPriceVal, &targetCurrVal, []string{"tag1"},
		}}
		ctxAst := withMockDb(authedCtx, mockDbAst)
		astOne, err := astRepo.FindOne(ctxAst, asset.Filter{ID: astID, ProjectID: pID, Name: "Laptop"})
		if err != nil || astOne == nil {
			t.Errorf("FindOne asset by all filters error: %v", err)
		}
	})

	t.Run("Repository Save row error branches", func(t *testing.T) {
		mockDbRowErr := &mockPgDb{rowErr: errors.New("query row error")}
		ctxRowErr := withMockDb(authedCtx, mockDbRowErr)

		err := astRepo.Save(ctxRowErr, ast)
		if err == nil {
			t.Errorf("expected Save asset row error")
		}
	})

	t.Run("Repository Find count error branches", func(t *testing.T) {
		mockDbCountErr := &mockPgDb{countErr: errors.New("count query error")}
		ctxCountErr := withMockDb(authedCtx, mockDbCountErr)

		_, err := astRepo.Find(ctxCountErr, asset.CollectionFilter{ProjectID: pID})
		if err == nil {
			t.Errorf("expected Find asset count error")
		}
	})
}

func TestPgInvitationRepositoryAndPeopleUpdates(t *testing.T) {
	dbConn := &PgDbConnection{}
	peopleRepo := NewPgPeopleRepository(dbConn)
	invRepo := NewPgInvitationRepository(dbConn)

	pID := uuid.New()
	adminID := uuid.New()
	invID := uuid.New()
	now := time.Now()
	codeHash := "sha256_hash_value_123"

	t.Run("DeletePerson, UpdateProfile, SaveCredentials", func(t *testing.T) {
		mockDb := &mockPgDb{}
		ctx := withMockDb(context.Background(), mockDb)

		err := peopleRepo.DeletePerson(ctx, pID)
		if err != nil {
			t.Errorf("DeletePerson error: %v", err)
		}

		err = peopleRepo.UpdateProfile(ctx, pID, "Alice", "Smith", "Alice Smith")
		if err != nil {
			t.Errorf("UpdateProfile error: %v", err)
		}

		cred, _ := people.NewCredentials(people.GOOGLE, "sub_123", "sub_123")
		err = peopleRepo.SaveCredentials(ctx, pID, cred)
		if err != nil {
			t.Errorf("SaveCredentials error: %v", err)
		}
	})

	t.Run("PgInvitationRepository Create, FindByID, FindByCodeHash, FindByCreator, Complete, Delete, EmailExists", func(t *testing.T) {
		inv, _ := people.NewInvitation(
			invID,
			"user@example.com",
			codeHash,
			now.Add(7*24*time.Hour),
			adminID,
			&pID,
			people.InvitationStatusPending,
			now,
			nil,
		)

		mockDb := &mockPgDb{
			rowVal: []any{
				invID.String(),
				"user@example.com",
				codeHash,
				now.Add(7 * 24 * time.Hour),
				adminID.String(),
				pID.String(),
				"PENDING",
				now,
				now,
			},
			rowsData: [][]any{{
				invID.String(),
				"user@example.com",
				codeHash,
				now.Add(7 * 24 * time.Hour),
				adminID.String(),
				pID.String(),
				"PENDING",
				now,
				now,
			}},
		}
		ctx := withMockDb(context.Background(), mockDb)

		// Create
		err := invRepo.Create(ctx, inv)
		if err != nil {
			t.Errorf("Create invitation error: %v", err)
		}

		// FindByID
		found, err := invRepo.FindByID(ctx, invID)
		if err != nil || found == nil || found.Email() != "user@example.com" {
			t.Errorf("FindByID invitation error: %v", err)
		}

		// FindByCodeHash
		foundHash, err := invRepo.FindByCodeHash(ctx, codeHash)
		if err != nil || foundHash == nil {
			t.Errorf("FindByCodeHash error: %v", err)
		}

		// FindByCreator
		list, err := invRepo.FindByCreator(ctx, adminID)
		if err != nil || len(list) != 1 {
			t.Errorf("FindByCreator error: %v", err)
		}

		// Complete
		err = invRepo.Complete(ctx, invID)
		if err != nil {
			t.Errorf("Complete error: %v", err)
		}

		// Delete
		err = invRepo.Delete(ctx, invID)
		if err != nil {
			t.Errorf("Delete error: %v", err)
		}

		// EmailExists
		exists, err := invRepo.EmailExists(ctx, "user@example.com")
		if err != nil || !exists {
			t.Errorf("EmailExists error: %v", err)
		}
	})
}

