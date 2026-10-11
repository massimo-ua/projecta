# Domain-Driven Design (DDD) Refactoring Notes & Playbook

This document records the exact steps, patterns, and anti-pattern resolutions applied during the DDD refactoring of Projecta's backend (specifically `internal/asset` and `internal/investment`). Follow this playbook for any subsequent bounded contexts (e.g., `internal/projecta`, `internal/people`) to avoid repeating duplicate exploratory steps.

---

## 1. Core Architectural Principles Applied

1. **Aggregate Encapsulation & Private State**:
   Aggregates protect their invariants. All struct fields are private (`unexported`). Entities cannot be mutated arbitrarily from outside.
2. **Reference Other Aggregates by Identity Only**:
   An aggregate must NEVER hold a direct pointer to another aggregate (e.g., `Investment` holding `*asset.Asset`). Reference by ID (`uuid.UUID`) and read-only snapshot attributes (`string`).
3. **Dedicated Reconstitution Factory**:
   Persistence hydration must NOT run business invariants or trigger timestamp generation. Provide `Reconstitute<Entity>(...) *Entity` alongside `New<Entity>(...) (*Entity, error)`.
4. **Consumer-Owned Ports**:
   Application services must not depend on wide repositories from foreign bounded contexts. Define a narrow consumer interface in the calling package (e.g., `AssetReader`).
5. **Decoupled DAL Records**:
   Repositories must not construct dummy aggregate stubs during queries. Use intermediate scan structs (`<entity>Record`), fetch auxiliary tables, and invoke `Reconstitute<Entity>`.

---

## 2. Standard 6-Step Transformation Recipe

### Step 1: Encapsulate the Aggregate Root (`internal/<context>/<entity>.go`)
1. **Unexport fields**: Make state private (`id`, `project`, `amount`, `allocations`, etc.).
2. **Replace cross-aggregate pointers**:
   - *Before*: `Asset *asset.Asset`
   - *After*: `primaryAssetID uuid.UUID`, `primaryAssetName string`
3. **Consolidate invariants into domain methods**:
   - Move validation rules out of application services into entity methods:
     - `Allocate(allocations []InvestmentAssetLink) error`: Enforces 100% total sum, positive shares, non-zero values, and duplicate prevention.
     - `UpdateDetails(...)`
     - `UpdateValuation(...)`
4. **Create persistence factory**:
   - `Reconstitute<Entity>(...) *Entity`: Hydrates private fields directly.
5. **Expose read-only getters**:
   - `ID() uuid.UUID`, `Amount() *money.Money`, `PrimaryAssetID() uuid.UUID`, etc.

### Step 2: Implement Pure Domain Tests (`internal/<context>/<entity>_test.go`)
1. Write fast, pure unit tests without database or mock dependencies.
2. Verify:
   - Constructor validation and error cases.
   - Domain invariants (e.g., share percentages != 100%, duplicate allocations, invalid dates).
   - Invariant mutation methods (`Allocate`, `UpdateValuation`).
   - Reconstitution factory retains field fidelity.

### Step 3: Narrow Consumer-Owned Ports & Clean Service (`internal/<context>/service.go`)
1. **Define narrow consumer-owned interfaces**:
   ```go
   // AssetReader provides only what the investment service needs from assets.
   type AssetReader interface {
       FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error)
   }
   ```
2. **Eliminate duplicate invariant logic in service**:
   - Replace manual validation loops with aggregate calls (`inv.Allocate(...)`, `inv.UpdateDetails(...)`).
3. **Update unit tests (`service_test.go`)**:
   - Update mocks to only satisfy the narrow consumer interface.
   - Replace field accesses with getters (`inv.Amount().Amount()`, `inv.PrimaryAssetID()`).

### Step 4: Refactor Persistence Layer (`pkg/dal/pg_<entity>_repository.go`)
1. **Use getters in mutation operations**:
   - `Save`, `create`, `update`, `Remove` call getters (`inv.ID().String()`, `inv.PrimaryAssetID().String()`).
2. **Introduce intermediate DB records**:
   ```go
   type investmentRecord struct {
       id             uuid.UUID
       project        *projecta.Project
       primaryAssetID uuid.UUID
       ...
   }
   ```
3. **Scan into record and reconstitute**:
   - Eliminate fake stub creations (e.g., delete `asset.NewAsset(astID, astName, ...)` in investment DAL).
   - In `FindOne` / `Find`:
     1. Scan rows into `<entity>Record`.
     2. Load auxiliary relations (e.g. allocations).
     3. Call `investment.ReconstituteInvestment(rec..., allocs)`.
   - Remove post-hoc mutation setters (e.g., delete `inv.SetAssetAllocations`).

### Step 5: Align Transport & Serialization (`pkg/web/<entity>.go`)
1. **Update DTO serializer**:
   - In `to<Entity>DTO`, replace direct struct field access with getters (`inv.ID()`, `inv.PrimaryAssetID()`, `inv.Amount()`).
2. **Keep decoders focused on request parsing**:
   - Decoders map HTTP payloads into typed application Commands.

### Step 6: Verification & Knowledge Graph
1. **Run full Go test suite**:
   ```bash
   go test -count=1 ./internal/... ./pkg/... ./cmd/...
   ```
2. **Build Web UI**:
   ```bash
   npm --prefix web-ui run build
   ```
3. **Update knowledge graph**:
   ```bash
   graphify update .
   ```
4. **Rebuild & test live Docker containers**:
   ```bash
   docker compose build web && docker compose up -d web
   docker compose logs --tail=30 web
   ```

---

## 3. Anti-Patterns Eliminated

| Anti-Pattern | Where It Existed | Resolved Solution |
| :--- | :--- | :--- |
| **Cross-Aggregate Pointer Coupling** | `Investment` held `Asset *asset.Asset` | Replaced with `primaryAssetID uuid.UUID` and `primaryAssetName string`. Aggregates reference by ID only. |
| **Fake Stub Hydration in DAL** | `pg_investment_repository.go` constructed fake `asset.NewAsset(...)` | Used intermediate `investmentRecord` and hydrated `Investment` directly via `ReconstituteInvestment`. |
| **Duplicated Invariant Checks** | Invariant allocation loops copy-pasted in `Create` and `Update` in `service.go` | Consolidated in domain method `Investment.Allocate`. |
| **Wide Foreign Dependencies** | `investment.Service` imported full `asset.Repository` | Narrowed to consumer-owned `AssetReader` interface. |
| **Post-Hoc Aggregate Mutation in DAL** | `inv.SetAssetAllocations(allocs)` after DB scan | Pass `allocations` directly into `ReconstituteInvestment`. |

---

## 4. Checklist for Future Refactorings

When refactoring the next context (e.g., `internal/projecta` or `internal/people`):
- [ ] Are all aggregate fields unexported?
- [ ] Are other aggregates referenced solely by UUID / snapshot values?
- [ ] Is there an invariant-free `Reconstitute<Entity>` factory for DAL?
- [ ] Are all business calculations encapsulated inside entity methods rather than in service handlers?
- [ ] Does the service define narrow consumer interfaces instead of importing wide foreign repositories?
- [ ] Does the DAL scan into an intermediate struct without instantiating dummy foreign entities?
- [ ] Does `graphify update .` succeed?
- [ ] Do all tests pass across `./internal/...`, `./pkg/...`, and `./cmd/...`?
