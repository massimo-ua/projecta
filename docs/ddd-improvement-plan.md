# Domain-Driven Design (DDD) 7-Step Improvement Plan

This document provides a comprehensive, file-by-file engineering blueprint to resolve the **7 DDD implementation gaps** identified in the Projecta codebase. Each step is structured with:
1. **Root Cause & Current State Analysis** (exact file locations and line numbers)
2. **DDD Principle & Architectural Goal**
3. **Detailed File-by-File Implementation Guide (Before vs. After Code)**
4. **Edge Cases & Backward-Compatibility Considerations**
5. **Verification & Automated Test Specifications**

---

## Executive Summary & Verification Roadmap

```text
1. [Step 1: Enforce Layer & Context Boundaries]    → ✅ COMPLETED: zero cross-context imports across all bounded contexts (`internal/asset`, `internal/investment`, `internal/project`, `internal/people`, `internal/statement`)
2. [Step 2: Decouple Domain from pkg/currency]     → ✅ COMPLETED: `go list -f '{{.Imports}}' ./internal/...` contains zero `pkg/currency` imports; domain uses pure `*money.Money` CurrencyConverter
3. [Step 3: Consolidate Validation & Strip JSON]   → ✅ COMPLETED: domain price/share validation enforced; zero `encoding/json` imports or `json:` tags in `internal/` domain types
4. [Step 4: Harden Aggregate Encapsulation]        → ✅ COMPLETED: defensive copying across all slice getters/setters; `NewAsset` removed; `UpdateDetails` & `UpdateValuation` return domain errors
5. [Step 5: Unify Authorization Boundaries]        → ✅ COMPLETED: zero `core.AuthGuard` calls in `pkg/dal`; shared-project participants can read/update assets; investment endpoints enforce project membership
6. [Step 6: Enforce Transactional Boundaries]      → ✅ COMPLETED: all multi-statement writes execute inside `db.Tx`; `PgAssetRepository.Save` & `PgInvestmentRepository.Save` use atomic `ON CONFLICT` upserts
7. [Step 7: Complete Use Cases & Clean Adapters]   → ✅ COMPLETED: `ProjectService.Remove` implemented; `GetTotals` moved to domain port; N+1 eliminated via pagination and batch ID filtering
```

| Step | Gap Addressed | Affected Packages | Status |
| :--- | :--- | :--- | :---: |
| **Step 1** | Gap 1: Flat Package Layout & Cross-Context Coupling | `internal/asset`, `internal/investment`, `internal/project`, `internal/people`, `cmd/web`, `pkg/web` | ✅ **COMPLETED** |
| **Step 2** | Gap 2: Domain Coupled to `pkg/currency` & Dual Money Types | `internal/asset`, `internal/investment`, `pkg/currency`, `cmd/web`, `pkg/web` | ✅ **COMPLETED** |
| **Step 3** | Gap 3: Business Validation in HTTP Decoders & JSON in Domain | `pkg/web`, `internal/asset`, `internal/investment`, `internal/project`, `internal/people`, `internal/statement` | ✅ **COMPLETED** |
| **Step 4** | Gap 4: Encapsulation Leaks & Silent Validation Failures | `internal/asset`, `internal/investment`, `internal/project` | ✅ **COMPLETED** |
| **Step 5** | Gap 5: Inconsistent Authorization Boundaries Across Layers | `internal/project`, `internal/asset`, `internal/investment`, `pkg/dal` | ✅ **COMPLETED** |
| **Step 6** | Gap 6: Missing Transactional Boundaries (Unit of Work Gaps) | `internal/asset`, `internal/investment`, `pkg/dal` | ✅ **COMPLETED** |
| **Step 7** | Gap 7: Incomplete Domain Use Cases & Cross-Context Quirks | `internal/project`, `internal/investment`, `pkg/dal`, `pkg/web`, `cmd/web` | ✅ **COMPLETED** |

> [!TIP]
> **Why execute Steps 2–7 before Step 1?**
> Steps 2 through 7 clean up method signatures, imports, validation rules, authorization filters, and cross-context interfaces in place. Once those boundaries are clean, splitting `domain` and `app` subpackages in Step 1 becomes a mechanical file move with zero circular-dependency friction.

---

## Step 1: Enforce Layer & Context Boundaries (`domain` vs. `app` Isolation) — ✅ COMPLETED

> [!NOTE]
> **Completion Note (Step 1):**
> - **Design Decision**: Kept a single cohesive Go package per bounded context (`internal/asset`, `internal/investment`, `internal/project`, `internal/people`, `internal/statement`) and achieved **100% bounded-context isolation** via consumer-owned ports and composition-root adapters.
> - **Applied Changes**:
>   - `internal/asset/ports.go` & `internal/asset/service.go`: Added consumer-owned `ProjectDescriptor`, `ProjectReader`, and `OwnerReader`; removed unused `SetInvestmentSource`; removed `internal/project` import from `service.go` and `service_test.go`.
>   - `internal/investment/ports.go` & `internal/investment/service.go`: Added consumer-owned `AssetDescriptor`, `AssetReader`, `ProjectDescriptor`, `ProjectReader`, and `ContributorReader`; removed `internal/asset` and `internal/project` imports from `service.go` and `service_test.go`.
>   - `internal/investment/totals.go` -> `pkg/web/project.go`: Moved multi-context pagination orchestration `calculateProjectTotals` into `pkg/web/project.go` (delegating pure math to `asset.CalculateProjectTotals`).
>   - `internal/project/people_service.go` -> `cmd/web/adapters.go`: Moved `projectPeopleAdapter` (`people.Repository` -> `project.PeopleService`) to `cmd/web/adapters.go`, eliminating `internal/people` import from `internal/project`.
>   - `cmd/web/adapters.go`, `cmd/web/main.go`, & `cmd/web/main_test.go`: Wired and unit-tested `projectPeopleAdapter`, `assetProjectReaderAdapter`, `assetOwnerReaderAdapter`, `investmentAssetReaderAdapter`, `investmentProjectReaderAdapter`, and `investmentContributorReaderAdapter`.
> - **Verification**: `go list -f '{{.ImportPath}} -> {{.Imports}}' ./internal/...` confirmed zero cross-context imports across all bounded contexts and test suites; `go test -count=1 ./internal/... ./pkg/... ./cmd/...` passed 100%.

### 1.1. Root Cause & Current State Analysis
Currently, each bounded context (`internal/asset`, `internal/investment`, `internal/project`, `internal/people`) places domain aggregates, value objects, application services, commands, filters, and repository interfaces inside a **single flat Go package**:
- **Intra-Context Leak**: Because `asset.ServiceImpl` (`internal/asset/service.go`) is in `package asset`, it has direct visibility into all private fields of `Asset` (`a.id`, `a.price`, `a.status`, `a.children`) and `AssetHierarchy` (`h.assets`, `h.childrenMap`). The Go compiler cannot enforce aggregate encapsulation against application orchestration code in the same package.
- **Cross-Context Coupling**:
  - `internal/investment/service.go:10,13` imports `"gitlab.com/massimo-ua/projecta/internal/asset"` and `"gitlab.com/massimo-ua/projecta/internal/project"` solely because `AssetReader` and `ProjectReader` (`lines 23-30`) use foreign filter and aggregate types (`asset.Filter`, `*asset.Asset`, `project.ProjectFilter`, `*project.Project`) and `ServiceImpl` (`line 37`) uses `project.PeopleService`.
  - `internal/asset/service.go:11` imports `"gitlab.com/massimo-ua/projecta/internal/project"` because `ProjectReader` (`lines 22-24`) uses `project.ProjectFilter` and `*project.Project`, and `ServiceImpl` (`line 29`) uses `project.PeopleService`.
  - `internal/project/people_service.go:6` imports `"gitlab.com/massimo-ua/projecta/internal/people"` to map `*people.Person` into `*project.Owner`.

### 1.2. DDD Principle & Architectural Goal
- **Bounded Context Autonomy**: A bounded context's application layer should depend only on its own consumer-owned ports using local Value Objects / primitives, bridged at the composition root (`cmd/web/adapters.go`).
- **Strict Layer Dependency Direction**:
  ```mermaid
  flowchart TD
      Web["pkg/web (Inbound HTTP Ports)"]
      DAL["pkg/dal (Outbound Persistence Adapters)"]
      Cmd["cmd/web (Composition Root & Cross-Context Bridges)"]
      App["internal/<context>/app (Use-Case Services & Commands)"]
      Domain["internal/<context> (Pure Domain Aggregates & Value Objects)"]

      Cmd --> Web
      Cmd --> DAL
      Cmd --> App
      Web --> App
      Web --> Domain
      DAL --> App
      DAL --> Domain
      App --> Domain
  ```

### 1.3. Detailed File-by-File Implementation Guide

#### Sub-task 1.3.1: Eliminate Foreign Context Imports in `internal/investment`
In `internal/investment/ports.go` and `internal/investment/service.go`, replace `asset.Filter`, `*asset.Asset`, `project.ProjectFilter`, `*project.Project`, and `project.PeopleService` with consumer-owned types defined inside `internal/investment`:

```go
// BEFORE: internal/investment/service.go (Lines 10-39)
import (
	"gitlab.com/massimo-ua/projecta/internal/asset"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
	"gitlab.com/massimo-ua/projecta/internal/project"
)

type AssetReader interface {
	FindOne(ctx context.Context, filter asset.Filter) (*asset.Asset, error)
}

type ProjectReader interface {
	FindOne(ctx context.Context, filter project.ProjectFilter) (*project.Project, error)
}

type ServiceImpl struct {
	db           core.DbConnection
	investments  Repository
	assets       AssetReader
	projects     ProjectReader
	people       project.PeopleService
	rateProvider CurrencyConverter
}
```

```go
// AFTER: internal/investment/ports.go & internal/investment/service.go
// No imports of internal/asset or internal/project!

// AssetDescriptor represents the read-only asset projection needed by the investment context.
type AssetDescriptor struct {
	ID   uuid.UUID
	Name string
}

// AssetReader is a consumer-owned port for verifying and resolving assets in a project.
type AssetReader interface {
	FindAsset(ctx context.Context, projectID, assetID uuid.UUID) (AssetDescriptor, error)
}

// ProjectDescriptor represents the read-only project projection needed by the investment context.
type ProjectDescriptor struct {
	ID           uuid.UUID
	MainCurrency string
}

// ProjectReader is a consumer-owned port for verifying project access and retrieving base currency.
type ProjectReader interface {
	FindProject(ctx context.Context, projectID uuid.UUID) (ProjectDescriptor, error)
}

// ContributorReader is a consumer-owned port for resolving theauthenticated contributor.
type ContributorReader interface {
	FindContributor(ctx context.Context, personID uuid.UUID) (ContributorRef, error)
}

type ServiceImpl struct {
	db           core.DbConnection
	investments  Repository
	assets       AssetReader
	projects     ProjectReader
	contributors ContributorReader
	rateProvider CurrencyConverter
}
```

Update `resolveAllocations` in `internal/investment/service.go:348-397` to call `s.assets.FindAsset(ctx, projectID, aInput.AssetID)` and use `ast.ID` and `ast.Name`.

#### Sub-task 1.3.2: Eliminate Foreign Context Imports in `internal/asset`
In `internal/asset/service.go`, replace `project.ProjectFilter`, `*project.Project`, and `project.PeopleService` with consumer-owned ports:

```go
// AFTER: internal/asset/ports.go & internal/asset/service.go
// No import of internal/project!

type ProjectDescriptor struct {
	ID           uuid.UUID
	MainCurrency string
}

type ProjectReader interface {
	FindProject(ctx context.Context, projectID uuid.UUID) (ProjectDescriptor, error)
}

type OwnerReader interface {
	FindOwner(ctx context.Context, personID uuid.UUID) (OwnerRef, error)
}

type ServiceImpl struct {
	db           core.DbConnection
	assets       Repository
	owners       OwnerReader
	projects     ProjectReader
	rateProvider CurrencyConverter
	investments  InvestmentSource
}
```

#### Sub-task 1.3.3: Wire Cross-Context Adapters in `cmd/web/adapters.go`
Add lightweight bridge adapters in `cmd/web/adapters.go` that adapt `project.ProjectService`, `asset.Service`, and `people.UserService` to the narrow consumer-owned interfaces of `asset` and `investment`:

```go
// cmd/web/adapters.go
type assetProjectReaderAdapter struct {
	projects project.ProjectService
}

func (a *assetProjectReaderAdapter) FindProject(ctx context.Context, projectID uuid.UUID) (asset.ProjectDescriptor, error) {
	proj, err := a.projects.FindOne(ctx, project.ProjectFilter{ProjectID: projectID})
	if err != nil {
		return asset.ProjectDescriptor{}, err
	}
	return asset.ProjectDescriptor{
		ID:           proj.ID(),
		MainCurrency: proj.MainCurrency(),
	}, nil
}

type investmentAssetReaderAdapter struct {
	assets asset.Repository
}

func (a *investmentAssetReaderAdapter) FindAsset(ctx context.Context, projectID, assetID uuid.UUID) (investment.AssetDescriptor, error) {
	ast, err := a.assets.FindOne(ctx, asset.Filter{ID: assetID, ProjectID: projectID})
	if err != nil {
		return investment.AssetDescriptor{}, err
	}
	return investment.AssetDescriptor{
		ID:   ast.ID(),
		Name: ast.Name(),
	}, nil
}
```

#### Sub-task 1.3.4: Separate Domain vs. Application Package Scope (Optional Physical Split)
For contexts with complex orchestration (`internal/asset` and `internal/investment`):
- Keep pure domain files (`asset.go`, `hierarchy.go`, `valuation.go`, `filters.go`) in `internal/asset` (package `asset`).
- If stricter compiler-enforced encapsulation between `ServiceImpl` and `Asset` private fields is desired, move `service.go` and `service_test.go` into `internal/asset/app` (or keep `service.go` in `internal/asset` while enforcing via linter/code review that `service.go` never accesses `a.<privateField>` directly).

### 1.4. Edge Cases & Backward-Compatibility Considerations
- **Delete Unused Setter**: Remove `func (s *ServiceImpl) SetInvestmentSource(is InvestmentSource)` (`internal/asset/service.go:35-37`), which is a leftover from an earlier circular dependency before `investmentSourceAdapter` was introduced.
- **Test Mocks**: Update `mockProjectRepo` and `mockPeopleService` in `internal/asset/service_test.go` and `internal/investment/service_test.go` to implement the new consumer-owned `ProjectReader`, `OwnerReader`, and `ContributorReader` interfaces.

### 1.5. Verification & Test Specifications
1. Verify zero cross-context imports between `asset`, `investment`, `project`, and `people`:
   ```bash
   go list -f '{{.ImportPath}} -> {{.Imports}}' ./internal/...
   ```
2. Run all unit tests:
   ```bash
   go test -v -count=1 ./internal/... ./cmd/...
   ```

---

## Step 2: Decouple the Domain Layer from `pkg/currency` & Unify Monetary Types — ✅ COMPLETED

> [!NOTE]
> **Completion Note (Step 2):**
> - **Design Decision**: Defined narrow, consumer-owned `CurrencyConverter` domain interfaces (`ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error)`) inside `internal/asset` and `internal/investment`. Kept `pkg/currency` agnostic of `go-money` by creating a `currencyConverterAdapter` in `cmd/web/adapters.go` and `moneyRateAdapter` in `pkg/web/project.go` that bridge `*money.Money` to `pkg/currency.CurrencyRateProvider`.
> - **Applied Changes**:
>   - `internal/asset/valuation.go`, `internal/asset/asset.go`, `internal/asset/service.go`: Switched all conversion signatures from `pkg/currency` types to `ConvertMoney(*money.Money, string)`; eliminated all imports of `pkg/currency` from `internal/asset`.
>   - `internal/investment/investment.go`: Switched `CurrencyConverter` and `EvaluateHomeAmount` to `ConvertMoney`; eliminated all imports of `pkg/currency` from `internal/investment`.
>   - `cmd/web/adapters.go` & `cmd/web/main.go`: Created `currencyConverterAdapter` and wired it into `asset.NewService` and `investment.NewService`.
>   - `pkg/web/project.go`: Created `moneyRateAdapter` and bridged `rateProvider` into `asset.CalculateProjectTotals`.
>   - Tests: Updated `mockConverter` and `mockRateProvider` in `internal/asset`, `cmd/web`, and `pkg/web` to implement `ConvertMoney`.
> - **Verification**: `go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/... | grep "pkg/currency"` confirmed zero imports of `pkg/currency` across all domain and application packages in `internal/`; `go test -count=1 ./internal/... ./pkg/... ./cmd/...` passed 100%.

### 2.1. Root Cause & Current State Analysis
- **Domain Importing Outer Package**: Four files in `internal/` import `"gitlab.com/massimo-ua/projecta/pkg/currency"`:
  1. `internal/asset/asset.go:12` (used in `CreateAssetFromInvestments` lines 223–224 and `EvaluateValuation` lines 428–429)
  2. `internal/asset/valuation.go:5` (used in `CurrencyConverter` line 10, `CalculateBudgetProgress` lines 37–38 & 55–56, and `CalculateProjectTotals` lines 106–107 & 129–130)
  3. `internal/asset/service.go:12` (used in `ServiceImpl.rateProvider currency.CurrencyRateProvider` line 31 and `NewService` line 45)
  4. `internal/investment/investment.go:13` (used in `CurrencyConverter` line 17 and `EvaluateHomeAmount` lines 343–344)
- **Dual Money Representations**: The domain uses `*money.Money` (`github.com/Rhymond/go-money`) for all entity attributes (`price`, `targetPrice`, `directCost`, `totalCost`, `amount`, `timeHourlyRate`), yet every time it converts currencies, it unboxes `*money.Money` into `currency.NewCurrency(amt, code)`, calls `converter.Convert(...)`, and re-boxes `converted.Amount` back into `*money.Money` or `int64`.

### 2.2. DDD Principle & Architectural Goal
- **Dependency Inversion Principle (DIP)**: The `domain` layer must own the `CurrencyConverter` port signature using domain-native types (`*money.Money`), and the outer `pkg/currency` infrastructure adapter must conform to the domain—never the reverse.

### 2.3. Detailed File-by-File Implementation Guide

#### Sub-task 2.3.1: Define a Pure Domain `CurrencyConverter` Using `*money.Money`
In `internal/core` (e.g., `internal/core/currency_converter.go`) or directly in `internal/asset/valuation.go` and `internal/investment/investment.go`, define `CurrencyConverter` without `pkg/currency`:

```go
// BEFORE: internal/asset/valuation.go (Lines 1-11)
package asset

import (
	"github.com/Rhymond/go-money"
	"gitlab.com/massimo-ua/projecta/pkg/currency"
)

type CurrencyConverter interface {
	Convert(source, target currency.Currency) (currency.Currency, error)
}
```

```go
// AFTER: internal/asset/valuation.go (Lines 1-10)
package asset

import (
	"github.com/Rhymond/go-money"
)

// CurrencyConverter is a narrow domain interface for converting monetary amounts to a target currency code.
type CurrencyConverter interface {
	ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error)
}
```

#### Sub-task 2.3.2: Simplify Currency Conversions in `internal/asset/valuation.go`
Replace all `currency.NewCurrency(...)` calls in `CalculateBudgetProgress` (`lines 35-62`) and `CalculateProjectTotals` (`lines 104-137`):

```go
// BEFORE: internal/asset/valuation.go (Lines 35-43)
if converter != nil && targetCurr != homeCurrency && targetAmt > 0 {
	converted, err := converter.Convert(
		currency.NewCurrency(targetAmt, targetCurr),
		currency.NewCurrency(0, homeCurrency),
	)
	if err == nil {
		targetHomeVal = converted.Amount
	}
}
```

```go
// AFTER: internal/asset/valuation.go (Lines 35-41)
if converter != nil && targetCurr != homeCurrency && targetAmt > 0 {
	if converted, err := converter.ConvertMoney(targetPrice, homeCurrency); err == nil && converted != nil {
		targetHomeVal = converted.Amount()
	}
}
```

And in `CalculateProjectTotals` (`internal/asset/valuation.go:104-114`):
```go
// AFTER: internal/asset/valuation.go
if converter != nil && inv.Currency().Code != homeCurrency {
	converted, err := converter.ConvertMoney(inv, homeCurrency)
	if err != nil {
		return ProjectTotals{}, err
	}
	amt = converted.Amount()
}
```

#### Sub-task 2.3.3: Remove `pkg/currency` from `internal/asset/asset.go` and `internal/asset/service.go`
1. In `internal/asset/asset.go`:
   - Remove `"gitlab.com/massimo-ua/projecta/pkg/currency"` from imports (`line 12`).
   - In `CreateAssetFromInvestments` (`lines 222-229`), replace `converter.Convert(...)` with:
     ```go
     converted, err := converter.ConvertMoney(inv.Amount, targetCurrency)
     if err != nil {
         return nil, exceptions.NewValidationException(fmt.Sprintf("failed to convert currency %s to %s: %s", currCode, targetCurrency, err.Error()), err)
     }
     totalAmount += converted.Amount()
     ```
   - In `EvaluateValuation` (`lines 426-434`), replace `converter.Convert(...)` with:
     ```go
     if converter != nil && priceCurrency != homeCurrency && priceAmount > 0 {
         if converted, err := converter.ConvertMoney(a.Price(), homeCurrency); err == nil && converted != nil {
             homeAmount = converted.Amount()
         }
     }
     ```
2. In `internal/asset/service.go`:
   - Remove `"gitlab.com/massimo-ua/projecta/pkg/currency"` from imports (`line 12`).
   - Change `rateProvider currency.CurrencyRateProvider` (`line 31`) and `NewService(..., rateProvider ...currency.CurrencyRateProvider)` (`line 45`) to use the domain interface `CurrencyConverter`:
     ```go
     type ServiceImpl struct {
         db           core.DbConnection
         assets       Repository
         people       project.PeopleService
         projects     ProjectReader
         rateProvider CurrencyConverter
         investments  InvestmentSource
     }

     func NewService(
         db core.DbConnection,
         assets Repository,
         people project.PeopleService,
         projects ProjectReader,
         investmentSource InvestmentSource,
         rateProvider ...CurrencyConverter,
     ) *ServiceImpl
     ```

#### Sub-task 2.3.4: Remove `pkg/currency` from `internal/investment/investment.go`
1. Remove `"gitlab.com/massimo-ua/projecta/pkg/currency"` (`line 13`).
2. Update `CurrencyConverter` (`lines 16-18`) and `EvaluateHomeAmount` (`lines 334-350`):
   ```go
   type CurrencyConverter interface {
       ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error)
   }

   func (i *Investment) EvaluateHomeAmount(converter CurrencyConverter) {
       homeCurr := i.HomeCurrency()
       if i.amount == nil {
           i.homeAmount = 0
           return
       }
       i.homeAmount = i.amount.Amount()
       if converter != nil && i.amount.Currency() != nil && i.amount.Currency().Code != homeCurr {
           if converted, err := converter.ConvertMoney(i.amount, homeCurr); err == nil && converted != nil {
               i.homeAmount = converted.Amount()
           }
       }
   }
   ```

#### Sub-task 2.3.5: Implement `ConvertMoney` on `NBUCurrencyRateProvider` (or Adapter)
In `pkg/currency/nbu_provider.go` (or in `cmd/web/adapters.go`), add `ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error)`:

```go
// pkg/currency/nbu_provider.go (or cmd/web/adapters.go)
func (p *NBUCurrencyRateProvider) ConvertMoney(amount *money.Money, targetCurrency string) (*money.Money, error) {
	if amount == nil {
		return money.New(0, targetCurrency), nil
	}
	res, err := p.Convert(
		NewCurrency(amount.Amount(), amount.Currency().Code),
		NewCurrency(0, targetCurrency),
	)
	if err != nil {
		return nil, err
	}
	return money.New(res.Amount, res.Code), nil
}
```

### 2.4. Edge Cases & Backward-Compatibility Considerations
- Keep `Convert(currencyA Currency, currencyB Currency) (Currency, error)` on `NBUCurrencyRateProvider` in `pkg/currency` so existing `pkg/currency/nbu_provider_test.go` tests continue to pass without modification.
- Update `mockRateProvider` in `internal/asset/service_test.go:151-158`, `internal/asset/valuation_test.go`, and `internal/investment/service_test.go` to implement `ConvertMoney`.

### 2.5. Verification & Test Specifications
1. Check that no package under `internal/` imports `pkg/currency`:
   ```bash
   go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/... | grep "pkg/currency"
   # Expected output: empty
   ```
2. Run domain and service tests:
   ```bash
   go test -v -count=1 ./internal/asset/... ./internal/investment/... ./pkg/currency/...
   ```

---

## Step 3: Consolidate Business Validation in the Domain & Strip Transport Concerns from `internal/` — ✅ COMPLETED

> [!NOTE]
> **Completion Note (Step 3):**
> - **Design Decision**: Enforced asset invariants (`price > 0` and currency required) directly in `CreateAsset` for standalone assets, while allowing `CreateAssetGroup` to initialize group containers with zero price (derived from child rollups). Replaced `float64` in `LinkChildAssetDTO.SharePercentage` with `*float64` so explicit `0` or negative values pass through to domain `ValidateEdge` instead of being masked to `100.0`. Stripped JSON serialization concerns (`encoding/json` and `json:"..."` struct tags) completely out of domain aggregates and value objects (`Participant`, `DisplayName`, `StatementParseItem`, `StatementParseResult`), introducing transport DTOs in `pkg/web`.
> - **Applied Changes**:
>   - `internal/asset/asset.go`: Added domain price and currency validation in `CreateAsset`; factored internal constructor `createAsset` to support group container initialization.
>   - `pkg/web/assets.go`: Updated `LinkChildAssetDTO.SharePercentage` to `*float64`; only default to `100.0` when omitted (`nil`).
>   - `internal/project/participant.go` & `internal/people/display_name.go`: Removed `encoding/json` imports and `MarshalJSON`/`UnmarshalJSON` methods.
>   - `pkg/web/project.go` & `pkg/web/people.go`: Changed `ProjectDTO.Participants` to `[]string` and `UpdateProfileDTO.DisplayName` to `string`; validate display name in decoder via `people.NewDisplayName`.
>   - `internal/statement/service.go` & `pkg/web/statements.go`: Stripped `json:"..."` tags from `StatementParseItem` and `StatementParseResult`; created `StatementParseItemDTO` and `StatementParseResultDTO` in `pkg/web/statements.go`.
>   - Unit tests: Updated tests in `internal/project`, `internal/people`, `internal/asset`, `pkg/web/statement_endpoints_test.go`, and added comprehensive tests in `pkg/web/decoders_and_endpoints_test.go`.
> - **Verification**: Verified zero imports of `encoding/json` in domain packages under `internal/`; verified zero `json:"..."` struct tags in domain aggregates; all unit tests passed (`go test -v -count=1 ./internal/... ./pkg/... ./cmd/...`).

### 3.1. Root Cause & Current State Analysis
1. **Asset Price Invariant Located in HTTP Decoder Instead of Domain**:
   - `pkg/web/assets.go:276-282` (`decodeCreateAssetRequest`) and `pkg/web/assets.go:482-488` (`decodeUpdateAssetRequest`) validate:
     ```go
     if req.Price <= 0 && req.TargetPrice <= 0 {
         return nil, exceptions.NewValidationException("price must be greater than 0", nil)
     }
     if req.Currency == "" && req.TargetCurrency == "" {
         return nil, exceptions.NewValidationException("currency is required", nil)
     }
     ```
   - Meanwhile, `asset.CreateAsset` (`internal/asset/asset.go:107-156`) accepts `price == nil` or `price.Amount() <= 0` without error. Any caller outside HTTP (e.g., gRPC, CLI, message consumer, or service test) can create an invalid standalone asset with a negative or nil price.
2. **HTTP Decoder Masking Domain Share Validation**:
   - In `pkg/web/assets.go:635-638` (`decodeLinkChildAssetRequest`):
     ```go
     share := req.SharePercentage
     if share <= 0 {
         share = 100.0
     }
     ```
     Because `LinkChildAssetDTO.SharePercentage` is a non-pointer `float64`, the decoder cannot distinguish between an omitted field (`0.0`) and an explicit `0` or `-25.0`. It overwrites any `<= 0` value with `100.0`, rendering `ValidateEdge`'s `sharePercentage <= 0` check (`internal/asset/hierarchy.go:58`) unreachable from HTTP!
3. **Duplicated Pre-Checks in Application Services**:
   - `internal/asset/service.go:217-227` (`CreateFromInvestments`) validates `command.ProjectID == uuid.Nil`, `len(command.InvestmentIDs) == 0`, and `strings.TrimSpace(command.Name) == ""`—all three are immediately re-checked in `CreateAssetFromInvestments` (`internal/asset/asset.go:171-179`).
   - `internal/investment/service.go:69-79` (`Create`) validates `command.ProjectID == uuid.Nil`, `command.AssetID == uuid.Nil && len(command.AssetAllocations) == 0`, and `strings.TrimSpace(command.Description) == ""`—all three are checked in `NewInvestment` (`internal/investment/investment.go:106-111,162`) and `resolveAllocations`.
4. **JSON Serialization Coupled to Domain & Application Structs**:
   - `internal/project/participant.go:4,38-53` imports `"encoding/json"` and implements `MarshalJSON` / `UnmarshalJSON` because `pkg/web/project.go:38` uses `[]project.Participant` directly in `ProjectDTO`.
   - `internal/people/display_name.go:4,34-49` imports `"encoding/json"` and implements `MarshalJSON` / `UnmarshalJSON` because `pkg/web/people.go:44` uses `people.DisplayName` directly in `UpdateProfileDTO`.
   - `internal/statement/service.go:29-43` defines `json:"..."` tags on `StatementParseItem` and `StatementParseResult`.

### 3.2. DDD Principle & Architectural Goal
- **Single Source of Truth for Invariants**: Domain constructors and behavior methods must own all business invariants; HTTP decoders should only parse wire formats (JSON/URL params) into Command structs.
- **Transport-Agnostic Domain**: Domain value objects (`Participant`, `DisplayName`) must not know how they are serialized over HTTP/JSON.

### 3.3. Detailed File-by-File Implementation Guide

#### Sub-task 3.3.1: Enforce Price & Currency Invariants in `asset.CreateAsset`
In `internal/asset/asset.go:107-156`, validate that `price` is present, positive, and has a valid currency code (while `CreateAssetGroup` constructs its group container directly or via an internal helper since group assets derive their cost from child rollups):

```go
// AFTER: internal/asset/asset.go (Lines 118-156)
func CreateAsset(
	id uuid.UUID,
	name string,
	description string,
	projectID uuid.UUID,
	price *money.Money,
	startDate time.Time,
	owner OwnerRef,
	status AssetStatus,
	completedDate *time.Time,
	tags []string,
) (*Asset, error) {
	if id == uuid.Nil {
		id = uuid.New()
	}
	if projectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}
	trimmedName := strings.TrimSpace(name)
	if trimmedName == "" {
		return nil, exceptions.NewValidationException("name is required", nil)
	}
	if price == nil || price.Amount() <= 0 {
		return nil, exceptions.NewValidationException("price must be greater than 0", nil)
	}
	if price.Currency() == nil || strings.TrimSpace(price.Currency().Code) == "" {
		return nil, exceptions.NewValidationException("currency is required", nil)
	}
	// ... construct and return *Asset
}
```

And in `CreateAssetGroup` (`internal/asset/asset.go:270-292`), validate `projectID != uuid.Nil` and `strings.TrimSpace(name) != ""` and construct the group `&Asset{...}` with `zeroPrice := money.New(0, "UAH")` so group assets can intentionally have a `0` direct target price.

#### Sub-task 3.3.2: Fix `LinkChildAssetDTO` Pointer Semantics in `pkg/web/assets.go`
Change `LinkChildAssetDTO` (`pkg/web/assets.go:230-233`) and `decodeLinkChildAssetRequest` (`pkg/web/assets.go:625-645`):

```go
// BEFORE: pkg/web/assets.go (Lines 230-233 & 635-645)
type LinkChildAssetDTO struct {
	ChildAssetID    string  `json:"child_asset_id"`
	SharePercentage float64 `json:"share_percentage"`
}
// ...
	share := req.SharePercentage
	if share <= 0 {
		share = 100.0
	}
```

```go
// AFTER: pkg/web/assets.go (Lines 230-233 & 635-645)
type LinkChildAssetDTO struct {
	ChildAssetID    string   `json:"child_asset_id"`
	SharePercentage *float64 `json:"share_percentage,omitempty"`
}
// ...
	share := 100.0
	if req.SharePercentage != nil {
		share = *req.SharePercentage
	}

	return asset.LinkChildCommand{
		ParentID:        parentID,
		ChildID:         childID,
		ProjectID:       projectID,
		SharePercentage: share,
	}, nil
```

#### Sub-task 3.3.3: Remove Redundant Validation in `asset/service.go` and `investment/service.go`
1. In `internal/asset/service.go:211-228` (`CreateFromInvestments`):
   - Remove the duplicate `command.Name == ""` and `len(command.InvestmentIDs) == 0` checks, or call `CreateAssetFromInvestments` / validate before hitting the DB.
2. In `internal/investment/service.go:63-80` (`Create`):
   - Remove the duplicate `if strings.TrimSpace(command.Description) == ""` and `if command.AssetID == uuid.Nil && len(command.AssetAllocations) == 0` checks; `resolveAllocations` and `NewInvestment` already enforce both.

#### Sub-task 3.3.4: Remove JSON Serialization from `Participant`, `DisplayName`, and `statement`
1. **`internal/project/participant.go` & `pkg/web/project.go`**:
   - Delete `MarshalJSON` and `UnmarshalJSON` (`lines 38-53`) and `"encoding/json"` import from `internal/project/participant.go`.
   - In `pkg/web/project.go:30-75`, change `ProjectDTO.Participants` from `[]project.Participant` to `[]string`:
     ```go
     // AFTER: pkg/web/project.go
     type ProjectDTO struct {
         ProjectID    string   `json:"project_id"`
         Name         string   `json:"name"`
         Description  string   `json:"description"`
         Owner        OwnerDTO `json:"owner"`
         ShareToken   string   `json:"share_token,omitempty"`
         IsShared     bool     `json:"is_shared,omitempty"`
         MainCurrency string   `json:"mainCurrency,omitempty"`
         Participants []string `json:"participants"`
     }

     func toProjectDTO(p *project.Project) ProjectDTO {
         // ...
         rawParticipants := p.Participants()
         participants := make([]string, 0, len(rawParticipants))
         for _, part := range rawParticipants {
             participants = append(participants, part.DisplayName())
         }
         // ...
     }
     ```
2. **`internal/people/display_name.go` & `pkg/web/people.go`**:
   - Delete `MarshalJSON` and `UnmarshalJSON` (`lines 34-49`) and `"encoding/json"` import from `internal/people/display_name.go`.
   - In `pkg/web/people.go:43-45,72-88`, change `UpdateProfileDTO` to use `string` and construct `people.NewDisplayName` in `decodeUpdateProfileRequest`:
     ```go
     // AFTER: pkg/web/people.go
     type UpdateProfileDTO struct {
         DisplayName string `json:"display_name"`
     }

     func decodeUpdateProfileRequest(ctx context.Context, r *http.Request) (any, error) {
         requesterID, ok := ctx.Value(core.RequesterIDContextKey).(uuid.UUID)
         if !ok {
             return nil, exceptions.NewUnauthorizedException("failed to authorize profile request", nil)
         }

         var dto UpdateProfileDTO
         if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
             return nil, exceptions.NewValidationException("invalid request body", err)
         }

         dn, err := people.NewDisplayName(dto.DisplayName)
         if err != nil {
             return nil, exceptions.NewValidationException("invalid display name", err)
         }

         return UpdateProfileRequest{
             PersonID:    requesterID,
             DisplayName: dn,
         }, nil
     }
     ```
3. **`internal/statement/service.go` & `pkg/web/statements.go`**:
   - Remove `json:"..."` struct tags from `StatementParseItem` and `StatementParseResult` in `internal/statement/service.go:29-43`.
   - Add `StatementParseItemDTO`, `StatementParseResultDTO`, and `toStatementParseResultDTO` in `pkg/web/statements.go`.

### 3.4. Edge Cases & Backward-Compatibility Considerations
- **JSON Wire Compatibility**: Mapping `[]project.Participant` to `[]string` in `ProjectDTO` produces the exact same JSON array of strings (`["Alice", "Bob"]`) that `Participant.MarshalJSON` previously produced, so the frontend (`web-ui`) requires zero changes.
- **Validation Error Mapping**: In `decodeUpdateProfileRequest`, when `people.NewDisplayName` fails (e.g., `> 255` characters), wrapping it in `exceptions.NewValidationException` ensures HTTP `400 Bad Request` is returned consistently.

### 3.5. Verification & Test Specifications
1. Update `internal/project/participant_test.go` and `internal/people/display_name_test.go` to test pure domain construction/equality rather than `json.Marshal`.
2. Add test cases in `pkg/web/decoders_and_endpoints_test.go` verifying:
   - Omitted `share_percentage` in `POST /projects/{id}/assets/{id}/children` defaults to `100.0`.
   - Explicit `"share_percentage": -5` or `"share_percentage": 0` returns `400 Bad Request` (`ErrInvalidShare`).

---

## Step 4: Harden Aggregate Encapsulation, Defensive Copying & Mutation Invariants — ✅ COMPLETED

> [!NOTE]
> **Completion Note (Step 4):**
> - **Design Decision**: Completely removed unvalidated legacy constructor `asset.NewAsset(...)`, migrating all test callers to `CreateAsset` or `ReconstituteAsset`. Implemented defensive copying across all slice-returning getters (`Asset.Children()`, `Asset.Parents()`, `Asset.Tags()`, `Investment.AssetAllocations()`, `Investment.Tags()`) and slice-accepting methods/constructors (`AssignValuation`, `createAsset`, `ReconstituteAsset`, `Allocate`, `ReconstituteInvestment`). Returned a pointer to a struct copy in `Project.Owner()`. Enforced strict error returns on `Asset.UpdateDetails` (rejecting empty name or non-positive price) and `Investment.UpdateValuation` (rejecting negative hours, negative rates, negative quantities, negative/zero amounts, or invalid resource types).
> - **Applied Changes**:
>   - `internal/asset/asset.go`: Deleted `NewAsset`; added defensive copying in `createAsset`, `Children()`, `Parents()`, `Tags()`, `AssignValuation()`, `UpdateDetails()`, `ReconstituteAsset()`; updated `UpdateDetails` signature to return `error` with business validation.
>   - `internal/asset/service.go`: Handled error returned from `anAsset.UpdateDetails(...)` in `ServiceImpl.Update`.
>   - `internal/investment/investment.go`: Added defensive copying in `Allocate()`, `ReconstituteInvestment()`, `AssetAllocations()`, `Tags()`; updated `UpdateValuation` to strictly validate inputs and return domain validation exceptions.
>   - `internal/project/project.go`: Updated `Project.Owner()` to return a defensive copy of `*p.owner`.
>   - Test suites: Migrated test callers from `NewAsset` to `CreateAsset`/`CreateAssetGroup` in `cmd/web/main_test.go`, `internal/asset/asset_test.go`, `internal/asset/hierarchy_test.go`, `internal/asset/service_test.go`, `internal/asset/valuation_test.go`, `pkg/dal/dal_test.go`, `pkg/web/decoders_and_endpoints_test.go`, `pkg/web/web_test.go`; added unit tests verifying getter slice immutability and mutation validation errors in `asset_test.go`, `investment_test.go`, and `project_test.go`.
> - **Verification**: Zero occurrences of `NewAsset(` in codebase; all unit tests pass (`go test -v -count=1 ./internal/... ./pkg/... ./cmd/...`); zero cross-context import leaks.

### 4.1. Root Cause & Current State Analysis
1. **Leaky Slice Getters**:
   - In `internal/asset/asset.go:392-405`:
     ```go
     func (a *Asset) Children() []ChildAssetLink { return a.children }
     func (a *Asset) Parents() []ParentAssetLink { return a.parents }
     func (a *Asset) Tags() []string             { return a.tags }
     ```
   - In `internal/investment/investment.go:372,383`:
     ```go
     func (i *Investment) AssetAllocations() []InvestmentAssetLink { return i.assetAllocations }
     func (i *Investment) Tags() []string                          { return i.tags }
     ```
   - Any external caller holding the returned slice can mutate `inv.AssetAllocations()[0].SharePercentage = 999` or `a.Tags()[0] = ""` in memory, violating aggregate invariants without calling `Allocate` or `UpdateDetails`.
2. **Mutable Pointer Getter on `Project`**:
   - `Project.Owner()` (`internal/project/project.go:204`) returns `*Owner`, where `Owner` (`internal/project/owner.go:5-10`) has public mutable fields (`PersonID`, `DisplayName`, `FirstName`, `CanHaveProjects`). Calling `proj.Owner().CanHaveProjects = false` mutates the aggregate's internal owner state.
3. **Unvalidated Legacy Constructor**:
   - `asset.NewAsset` (`internal/asset/asset.go:74-104`) constructs an `*Asset` without checking `projectID != uuid.Nil`, `name != ""`, or `price > 0`.
4. **Silent Failure in Mutation Methods**:
   - `Investment.UpdateValuation` (`internal/investment/investment.go:232-272`) has signature `error`, yet if a caller passes an invalid negative amount (`money.New(-500, "UAH")`), lines 268–270 simply skip updating `i.amount` and return `nil`:
     ```go
     if finalAmount != nil && finalAmount.Amount() > 0 {
         i.amount = finalAmount
     }
     return nil
     ```
   - `Asset.UpdateDetails` (`internal/asset/asset.go:475-485`) silently ignores an empty/whitespace `name` (`if name != "" { a.name = name }`) and does not return an `error` if `targetPrice` is negative.

### 4.2. DDD Principle & Architectural Goal
- **Keep a Valid State in Memory at All Times**: Aggregates must never expose internal mutable references (slices, maps, pointers to mutable structs) and must fail fast with domain validation errors when a state transition is invalid.

### 4.3. Detailed File-by-File Implementation Guide

#### Sub-task 4.3.1: Return Defensive Copies from All Slice Getters & Constructors
In `internal/asset/asset.go`:
```go
// AFTER: internal/asset/asset.go (Lines 392-415)
func (a *Asset) Children() []ChildAssetLink {
	if a.children == nil {
		return make([]ChildAssetLink, 0)
	}
	cp := make([]ChildAssetLink, len(a.children))
	copy(cp, a.children)
	return cp
}

func (a *Asset) Parents() []ParentAssetLink {
	if a.parents == nil {
		return make([]ParentAssetLink, 0)
	}
	cp := make([]ParentAssetLink, len(a.parents))
	copy(cp, a.parents)
	return cp
}

func (a *Asset) Tags() []string {
	if a.tags == nil {
		return make([]string, 0)
	}
	cp := make([]string, len(a.tags))
	copy(cp, a.tags)
	return cp
}
```
Also copy slices on input in `AssignValuation` (`internal/asset/asset.go:458-472`).

In `internal/investment/investment.go`:
```go
// AFTER: internal/investment/investment.go (Lines 372-395)
func (i *Investment) AssetAllocations() []InvestmentAssetLink {
	if i.assetAllocations == nil {
		return make([]InvestmentAssetLink, 0)
	}
	cp := make([]InvestmentAssetLink, len(i.assetAllocations))
	copy(cp, i.assetAllocations)
	return cp
}

func (i *Investment) Tags() []string {
	if i.tags == nil {
		return make([]string, 0)
	}
	cp := make([]string, len(i.tags))
	copy(cp, i.tags)
	return cp
}
```
Also copy `allocations` inside `Investment.Allocate` (`line 214`) and `ReconstituteInvestment` (`line 299`).

#### Sub-task 4.3.2: Return a Defensive Copy in `Project.Owner()`
In `internal/project/project.go:204`:
```go
// BEFORE: internal/project/project.go (Line 204)
func (p *Project) Owner() *Owner { return p.owner }

// AFTER: internal/project/project.go (Lines 204-210)
func (p *Project) Owner() *Owner {
	if p.owner == nil {
		return nil
	}
	cp := *p.owner
	return &cp
}
```

#### Sub-task 4.3.3: Remove Legacy `asset.NewAsset`
1. Delete `func NewAsset(...) *Asset` from `internal/asset/asset.go:74-104`.
2. Update any remaining test callers in `internal/asset/asset_test.go`, `internal/asset/hierarchy_test.go`, `internal/asset/valuation_test.go`, `internal/asset/service_test.go`, and `pkg/web/web_test.go` to use `CreateAsset` or `ReconstituteAsset`.

#### Sub-task 4.3.4: Enforce Error Returns in `Investment.UpdateValuation` and `Asset.UpdateDetails`
1. In `internal/investment/investment.go:232-272`:
   ```go
   // AFTER: internal/investment/investment.go
   func (i *Investment) UpdateValuation(
       resType ResourceType,
       amount *money.Money,
       timeHours *float64,
       timeHourlyRate *money.Money,
       goodsQuantity *float64,
       goodsUnit string,
       goodsItemName string,
   ) error {
       targetType := i.resourceType
       if resType != "" {
           if _, err := ToResourceType(string(resType)); err != nil {
               return err
           }
           targetType = resType
       }

       if timeHours != nil && *timeHours <= 0 {
           return exceptions.NewValidationException("time hours must be greater than 0", nil)
       }
       if timeHourlyRate != nil && timeHourlyRate.Amount() <= 0 {
           return exceptions.NewValidationException("time hourly rate must be greater than 0", nil)
       }
       if goodsQuantity != nil && *goodsQuantity <= 0 {
           return exceptions.NewValidationException("goods quantity must be greater than 0", nil)
       }

       finalAmount := amount
       effHours := timeHours
       if effHours == nil {
           effHours = i.timeHours
       }
       effRate := timeHourlyRate
       if effRate == nil {
           effRate = i.timeHourlyRate
       }
       if (finalAmount == nil || finalAmount.Amount() <= 0) && targetType == ResourceTypeTime && effHours != nil && effRate != nil {
           if *effHours > 0 && effRate.Amount() > 0 {
               calc := int64(*effHours * float64(effRate.Amount()))
               finalAmount = money.New(calc, effRate.Currency().Code)
           }
       }

       if amount != nil && (finalAmount == nil || finalAmount.Amount() <= 0) {
           return exceptions.NewValidationException("investment amount must be greater than 0", nil)
       }

       i.resourceType = targetType
       if timeHours != nil {
           i.timeHours = timeHours
       }
       if timeHourlyRate != nil {
           i.timeHourlyRate = timeHourlyRate
       }
       if goodsQuantity != nil {
           i.goodsQuantity = goodsQuantity
       }
       if goodsUnit != "" {
           i.goodsUnit = goodsUnit
       }
       if goodsItemName != "" {
           i.goodsItemName = goodsItemName
       }
       if finalAmount != nil && finalAmount.Amount() > 0 {
           i.amount = finalAmount
       }
       return nil
   }
   ```
2. In `internal/asset/asset.go:475-485`:
   ```go
   // AFTER: internal/asset/asset.go
   func (a *Asset) UpdateDetails(name, description string, targetPrice *money.Money, tags []string) error {
       trimmed := strings.TrimSpace(name)
       if trimmed == "" {
           return exceptions.NewValidationException("name is required", nil)
       }
       if targetPrice != nil && targetPrice.Amount() <= 0 {
           return exceptions.NewValidationException("price must be greater than 0", nil)
       }
       a.name = trimmed
       a.description = description
       a.targetPrice = targetPrice
       if tags != nil {
           cp := make([]string, len(tags))
           copy(cp, tags)
           a.tags = cp
       }
       a.EvaluateValuation(a.homeCurrency, nil)
       return nil
   }
   ```
   And check the returned error in `asset.ServiceImpl.Update` (`internal/asset/service.go:331`):
   ```go
   if err := anAsset.UpdateDetails(command.Name, command.Description, price, command.Tags); err != nil {
       return err
   }
   ```

### 4.4. Verification & Test Specifications
1. Add unit tests in `internal/asset/asset_test.go` and `internal/investment/investment_test.go` verifying:
   - Mutating slices returned by `Children()`, `Parents()`, `Tags()`, and `AssetAllocations()` does not affect subsequent getter calls.
   - `Asset.UpdateDetails("", ...)` and `Asset.UpdateDetails("Valid", "", money.New(-100, "UAH"), nil)` return `ValidationFailed`.
   - `Investment.UpdateValuation(ResourceTypeMoney, money.New(-50, "UAH"), ...)` returns `ValidationFailed`.

---

## Step 5: Standardize Authorization Boundaries & Fix Shared-Project Access Control — ✅ COMPLETED

> [!NOTE]
> **Completion Note (Step 5):**
> - **Design Decision**: Completely decoupled the persistence layer (`pkg/dal`) from HTTP request context authentication by removing all calls to `core.AuthGuard(ctx)` from `PgProjectRepository`. Moved identity extraction into `ProjectServiceImpl` and passed it explicitly via `filter.RequesterID`. Fixed shared-project authorization across `PgAssetRepository` (`FindOne`, `update`, `Remove`) so participants can inspect and modify assets. Added project membership verification in `investment.ServiceImpl` (`verifyProjectAccess`) across all commands and queries, preventing unauthorized cross-project access.
> - **Applied Changes**:
>   - `internal/project/filters.go`: Added `RequesterID uuid.UUID` to `ProjectFilter` and `ProjectCollectionFilter`.
>   - `internal/project/project_service.go`: Authenticated caller via `core.AuthGuard(ctx)` in `FindOne` and `Find` when `RequesterID` is not pre-populated, and passed `RequesterID` in `Update`.
>   - `pkg/dal/pg_project_repository.go`: Removed `core.AuthGuard(ctx)` and `internal/core` import; filtered by requester ownership or share membership only when `filter.RequesterID != uuid.Nil`. Calculated `isShared` dynamically using `filter.RequesterID`.
>   - `pkg/dal/pg_asset_repository.go`: Scoped `update` and `Remove` by `asset_id` and `project_id`; added project share membership subquery in `FindOne` when `filter.OwnerID != uuid.Nil`.
>   - `internal/investment/service.go`: Added `verifyProjectAccess` helper; verified project access in `Create`, `CreateBatch`, `Update`, `Remove`, `FindOne`, `Find`, and `FindTags`.
>   - `cmd/web/adapters.go`: Passed `RequesterID` via `core.AuthGuard(ctx)` in `assetProjectReaderAdapter` and `investmentProjectReaderAdapter`.
>   - `internal/project/service_test.go` & `pkg/dal/dal_test.go`: Verified unauthorized rejections at application service level and explicit `RequesterID` handling at DAL level.

### 5.1. Root Cause & Current State Analysis
There are three distinct authorization inconsistencies across `internal/` and `pkg/dal/`:

1. **DAL Calling `core.AuthGuard(ctx)` in `PgProjectRepository`**:
   - `ProjectServiceImpl.FindOne` (`internal/project/project_service.go:18-20`) and `Find` (`lines 22-24`) simply pass `ctx` and `filter` straight to `s.repository`.
   - `PgProjectRepository.FindOne` (`pkg/dal/pg_project_repository.go:119`) and `PgProjectRepository.Find` (`line 360`) extract the authenticated user from `ctx` via `personID, err := core.AuthGuard(ctx)`.
   - **Why this is a bug/smell**: The persistence adapter (`pkg/dal`) is coupled to HTTP request context authentication (`core.AuthGuard`), making it impossible for background workers, CLI tools, or system tasks to query `PgProjectRepository` without spoofing a `RequesterIDContextKey` in `context.Context`.

2. **Shared Project Participants Blocked on Asset Operations (`PgAssetRepository`)**:
   - In `PgAssetRepository.Find` (`pkg/dal/pg_asset_repository.go:358-360`), the query checks both project ownership and shared project participation:
     ```go
     if filter.OwnerID != uuid.Nil {
         qb.Where(fmt.Sprintf("(projecta_assets.owner_id = %s OR projecta_assets.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))", qb.Var(filter.OwnerID.String()), qb.Var(filter.OwnerID.String())))
     }
     ```
   - However, in `PgAssetRepository.FindOne` (`pkg/dal/pg_asset_repository.go:278-280`), it checks **only** direct asset ownership:
     ```go
     if filter.OwnerID != uuid.Nil {
         qb.Where(qb.Equal("projecta_assets.owner_id", filter.OwnerID.String()))
     }
     ```
   - And `PgAssetRepository.update` (`line 132`) and `Remove` (`line 147`) also add `WHERE owner_id = anAsset.Owner().PersonID`.
   - Because `asset.ServiceImpl` (`internal/asset/service.go:96-103, 302, 321, 353, 358, 402, 452`) always sets `filter.OwnerID = personID` (the logged-in requester), a participant who joined a shared project via `AcceptShare` can see the list of assets (`Find`), but gets `ErrAssetNotFound` / `500 Internal Error` when clicking on an asset (`FindOne`), editing it (`Update`), grouping assets (`Group`), linking children (`LinkChild`), or assigning investments (`AssignInvestments`)!

3. **Missing Project Membership Check in `investment.ServiceImpl`**:
   - In `internal/investment/service.go`:
     - `Create` (`line 86`) and `CreateBatch` (`line 152`) call `s.projects.FindOne(ctx, project.ProjectFilter{ProjectID: command.ProjectID})`, which (transitively via `PgProjectRepository.FindOne`) verifies the user has access to the project.
     - However, `Update` (`line 216`), `Remove` (`line 283`), `FindOne` (`line 303`), `Find` (`line 319`), and `FindTags` (`line 339`) call `core.AuthGuard(ctx)` (confirming *some* user is logged in) and then query `s.investments` directly **without checking if the logged-in user owns or participates in `ProjectID`**! Any authenticated user who knows or guesses a `project_id` UUID can list, view, edit, or delete investments in another user's private project.

### 5.2. DDD Principle & Architectural Goal
- **Application Layer Owns Use-Case Authorization**: Application services must authenticate the caller (`core.AuthGuard`) and verify project access/ownership before delegating to repositories.
- **Repositories Filter by Explicit Criteria**: Repositories should only apply SQL predicates based on fields explicitly passed in `Filter` / `CollectionFilter` structs, never by inspecting `core.RequesterIDContextKey` directly.

### 5.3. Detailed File-by-File Implementation Guide

#### Sub-task 5.3.1: Move `AuthGuard` from `PgProjectRepository` to `ProjectServiceImpl`
1. Add `RequesterID uuid.UUID` to `ProjectFilter` and `ProjectCollectionFilter` in `internal/project/filters.go`:
   ```go
   // AFTER: internal/project/filters.go
   type ProjectFilter struct {
       Name        string
       ProjectID   uuid.UUID
       RequesterID uuid.UUID
   }

   type ProjectCollectionFilter struct {
       core.Pagination
       Name        string
       RequesterID uuid.UUID
   }
   ```
2. Update `ProjectServiceImpl.FindOne`, `Find`, `Update`, and `Create` in `internal/project/project_service.go`:
   ```go
   // AFTER: internal/project/project_service.go
   func (s *ProjectServiceImpl) FindOne(ctx context.Context, filter ProjectFilter) (*Project, error) {
       personID, err := core.AuthGuard(ctx)
       if err != nil {
           return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
       }
       filter.RequesterID = personID
       return s.repository.FindOne(ctx, filter)
   }

   func (s *ProjectServiceImpl) Find(ctx context.Context, filter ProjectCollectionFilter) ([]*Project, error) {
       personID, err := core.AuthGuard(ctx)
       if err != nil {
           return nil, exceptions.NewUnauthorizedException("failed to identify requester", err)
       }
       filter.RequesterID = personID
       return s.repository.Find(ctx, filter)
   }
   ```
3. Update `PgProjectRepository.FindOne` (`pkg/dal/pg_project_repository.go:118-175`) and `Find` (`lines 359-421`) to use `filter.RequesterID` instead of calling `core.AuthGuard(ctx)`:
   ```go
   // AFTER: pkg/dal/pg_project_repository.go (Lines 148-154 & 173)
   if filter.RequesterID != uuid.Nil {
       qb.Where(fmt.Sprintf(
           "(projecta_projects.owner_id = %s OR projecta_projects.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))",
           qb.Var(filter.RequesterID.String()),
           qb.Var(filter.RequesterID.String()),
       ))
   }
   // ...
   isShared := (filter.RequesterID != uuid.Nil && rec.ownerID != filter.RequesterID.String())
   ```

#### Sub-task 5.3.2: Fix Shared-Project Access in `PgAssetRepository`
In `pkg/dal/pg_asset_repository.go:278-285` (`FindOne`), replace the strict `owner_id` check with the shared-project access predicate (and add `filter.ProjectID` filtering if provided):

```go
// BEFORE: pkg/dal/pg_asset_repository.go (Lines 276-280)
qb.Where(qb.Equal("projecta_assets.asset_id", filter.ID.String()))

if filter.OwnerID != uuid.Nil {
	qb.Where(qb.Equal("projecta_assets.owner_id", filter.OwnerID.String()))
}
```

```go
// AFTER: pkg/dal/pg_asset_repository.go (Lines 276-288)
qb.Where(qb.Equal("projecta_assets.asset_id", filter.ID.String()))

if filter.ProjectID != uuid.Nil {
	qb.Where(qb.Equal("projecta_assets.project_id", filter.ProjectID.String()))
}

if filter.OwnerID != uuid.Nil {
	qb.Where(fmt.Sprintf(
		"(projecta_assets.owner_id = %s OR projecta_assets.project_id IN (SELECT project_id FROM projecta_project_shares WHERE person_id = %s))",
		qb.Var(filter.OwnerID.String()),
		qb.Var(filter.OwnerID.String()),
	))
}
```

And in `PgAssetRepository.update` (`pkg/dal/pg_asset_repository.go:131-133`), scope the `UPDATE` by `asset_id` and `project_id` rather than `owner_id` so shared project participants can update an asset created by another participant in the same project:
```go
// AFTER: pkg/dal/pg_asset_repository.go (Lines 131-133)
qb.Where(qb.Equal("asset_id", anAsset.ID().String()))
qb.Where(qb.Equal("project_id", anAsset.ProjectID().String()))
```

#### Sub-task 5.3.3: Enforce Project Access in `investment.ServiceImpl`
In `internal/investment/service.go`, add a helper `verifyProjectAccess` and call it in `Update`, `Remove`, `FindOne`, `Find`, and `FindTags`:

```go
// AFTER: internal/investment/service.go
func (s *ServiceImpl) verifyProjectAccess(ctx context.Context, projectID uuid.UUID) (*project.Project, error) {
	if projectID == uuid.Nil {
		return nil, exceptions.NewValidationException("project id is required", nil)
	}
	proj, err := s.projects.FindOne(ctx, project.ProjectFilter{ProjectID: projectID})
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return nil, exceptions.NewNotFoundException("project not found", err)
		}
		return nil, exceptions.NewInternalException("failed to verify project access", err)
	}
	return proj, nil
}
```
Invoke `if _, err := s.verifyProjectAccess(ctx, ...); err != nil { return ... }` at the start of `Update`, `Remove`, `FindOne` (when `filter.ProjectID != uuid.Nil`), `Find`, and `FindTags`.

### 5.4. Edge Cases & Backward-Compatibility Considerations
- **Internal Repository Calls to `FindOne` (`Filter{ID: ...}`)**: When `PgAssetRepository.Save` or `investment.ServiceImpl.AssignToAsset` calls `FindOne` with `OwnerID: uuid.Nil`, the `if filter.OwnerID != uuid.Nil` guard is skipped, preserving internal lookup behavior.
- **Error Code Translation in `asset.ServiceImpl`**: In `internal/asset/service.go:103-106`, `FindOne` currently wraps all repository errors in `exceptions.NewInternalException(failedToFindAsset, err)`. Check `if errors.Is(err, exceptions.NotFoundError)` (or wrap `ErrAssetNotFound` with `exceptions.NotFoundError`) so missing/unauthorized assets return `404 Not Found` instead of `500 Internal Server Error`.

### 5.5. Verification & Test Specifications
1. Verify `pkg/dal` contains zero references to `core.AuthGuard`:
   ```bash
   go list -f '{{.Imports}}' ./pkg/dal/...
   ```
2. Add tests in `pkg/dal/dal_test.go` and `internal/investment/service_test.go`:
   - Shared user (`personB`) calls `PgAssetRepository.FindOne` and `Update` on an asset owned by `personA` in a shared project → succeeds.
   - Unshared user (`personC`) calls `investment.ServiceImpl.Find` or `Remove` on `personA`'s project → returns `NotFound` / `Unauthorized`.

---

## Step 6: Enforce Transactional Boundaries (Unit of Work) & Atomic Upserts — ✅ COMPLETED

> [!NOTE]
> **Completion Note (Step 6):**
> - **Design Decision**: Enforced atomic transactional boundaries across all multi-aggregate mutations and switched persistence to atomic upserts. Multi-step operations (`asset.CreateFromInvestments`, `asset.LinkChildren`, `asset.Remove`, `investment.Create`, `investment.Update`, `investment.AssignToAsset`, `investment.Remove`) are wrapped in `s.db.Tx(ctx, ...)`, ensuring all changes rollback on failure. In `pkg/dal`, eliminated the non-atomic, query-heavy `FindOne`-before-`Save` anti-pattern in `PgAssetRepository.Save` and `PgInvestmentRepository.Save`, replacing them with atomic PostgreSQL `INSERT ... ON CONFLICT (id) DO UPDATE SET ...` statements.
> - **Applied Changes**:
>   - `internal/asset/service.go`: Wrapped `CreateFromInvestments`, `LinkChildren`, and `Remove` in `s.db.Tx`.
>   - `internal/investment/service.go`: Wrapped `Create`, `Update`, `AssignToAsset`, and `Remove` in `s.db.Tx`.
>   - `pkg/dal/pg_asset_repository.go`: Replaced separate `FindOne` + `create`/`update` with a single atomic PostgreSQL `INSERT ... ON CONFLICT (asset_id) DO UPDATE SET ...` statement; removed unexported `create` and `update` methods.
>   - `pkg/dal/pg_investment_repository.go`: Replaced separate `FindOne` + `create`/`update` with an atomic `INSERT ... ON CONFLICT (investment_id) DO UPDATE SET ...` followed by `saveAllocations`; removed unexported `create` and `update` methods.
>   - `pkg/dal/dal_test.go`: Added `TestPgInvestmentRepository` testing atomic upsert, allocations, find, tags, and remove; updated `TestPgAssetRepository` to assert atomic upsert execution.
>   - `internal/asset/service_test.go` & `internal/investment/service_test.go`: Verified `db.txCalled == true` across transactional mutations.

### 6.1. Root Cause & Current State Analysis
1. **Multi-Aggregate / Multi-Table Writes Without `s.db.Tx`**:
   - `asset.ServiceImpl.CreateFromInvestments` (`internal/asset/service.go:272-278`):
     ```go
     if err = s.assets.Save(ctx, anAsset); err != nil {
         return nil, exceptions.NewInternalException(failedToCreateAsset, err)
     }
     if err = s.investments.AssignToAsset(ctx, anAsset.ID(), command.InvestmentIDs); err != nil {
         return nil, exceptions.NewInternalException(failedToCreateAsset, err)
     }
     ```
     If `AssignToAsset` fails on the 2nd investment, `anAsset` and the 1st reassigned investment remain committed in the database.
   - `asset.ServiceImpl.LinkChildren` (`internal/asset/service.go:503-515`): Loops over `command.ChildAssetIDs` calling `s.LinkChild(ctx, ...)` without a transaction. If linking the 3rd child fails due to `ErrCircularDependency`, the first 2 children remain linked.
   - `investment.ServiceImpl.AssignToAsset` (`internal/investment/service.go:261-278`): Loops over `investmentIDs`, loading and saving each `Investment` outside `s.db.Tx`.
   - `investment.ServiceImpl.Create` (`line 126`) and `Update` (`line 248`): Call `s.investments.Save(ctx, inv)` outside `s.db.Tx`. Inside `PgInvestmentRepository.create` (`pkg/dal/pg_investment_repository.go:98-106`) and `update` (`lines 145-153`), `Save` executes:
     1. `INSERT` or `UPDATE` on `projecta_investments`
     2. `DELETE FROM projecta_investment_assets WHERE investment_id = $1`
     3. `INSERT INTO projecta_investment_assets ...` in a loop (`saveAllocations`, `lines 157-185`)
     Without `s.db.Tx(ctx, ...)`, those statements run on separate pool connections in auto-commit mode! If an `INSERT INTO projecta_investment_assets` fails, all previous allocations for that investment are permanently deleted (`DELETE` already auto-committed).
2. **Inefficient & Non-Atomic `FindOne`-Before-`Save` Pattern**:
   - `PgAssetRepository.Save` (`pkg/dal/pg_asset_repository.go:29-43`) and `PgInvestmentRepository.Save` (`pkg/dal/pg_investment_repository.go:29-43`) execute `r.FindOne(ctx, ...)` before every write:
     ```go
     func (r *PgAssetRepository) Save(ctx context.Context, anAsset *asset.Asset) error {
         _, err := r.FindOne(ctx, asset.Filter{ID: anAsset.ID()})
         if err != nil {
             if errors.Is(err, ErrAssetNotFound) {
                 return r.create(ctx, anAsset)
             }
             return err
         }
         return r.update(ctx, anAsset)
     }
     ```
     In `PgAssetRepository.FindOne`, this runs 3 SQL queries (`SELECT` asset + `SELECT` direct costs + `SELECT` compositions) and builds an `AssetHierarchy` DAG in memory just to check if the row exists before running `INSERT` or `UPDATE`!

### 6.2. DDD Principle & Architectural Goal
- **Aggregate Atomicity & Unit of Work**: Saving a single aggregate (`Investment` + its `InvestmentAssetLink` value objects, or `Person` + `Credentials` + `Role`) must always be atomic. When an application use case coordinates multiple aggregates in one request (`CreateFromInvestments`, `Group`, `LinkChildren`), the application service must wrap the operation in `s.db.Tx`.
- **Idempotent Upsert Persistence**: `Save(ctx, aggregate)` should use PostgreSQL `INSERT ... ON CONFLICT DO UPDATE` (already used in `PgPeopleRepository.Save` at `pkg/dal/pg_people_repository.go:51-57` and `PgInvitationRepository.Save` at `pkg/dal/pg_invitation_repository.go:82-90`).

### 6.3. Detailed File-by-File Implementation Guide

#### Sub-task 6.3.1: Wrap Multi-Step Application Use Cases in `s.db.Tx`
1. **`internal/asset/service.go` (`CreateFromInvestments`, Lines 272-280)**:
   ```go
   // AFTER: internal/asset/service.go
   _, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
       if err := s.assets.Save(txCtx, anAsset); err != nil {
           return nil, exceptions.NewInternalException(failedToCreateAsset, err)
       }
       if err := s.investments.AssignToAsset(txCtx, anAsset.ID(), command.InvestmentIDs); err != nil {
           return nil, exceptions.NewInternalException(failedToCreateAsset, err)
       }
       return nil, nil
   })
   if err != nil {
       return nil, err
   }
   ```
2. **`internal/asset/service.go` (`LinkChildren`, Lines 503-515)**:
   ```go
   // AFTER: internal/asset/service.go
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
   ```
3. **`internal/investment/service.go` (`Create`, `Update`, `AssignToAsset`)**:
   - Wrap `s.investments.Save` in `Create` (`line 126`) and `Update` (`line 248`) with `s.db.Tx`.
   - Wrap the loop in `AssignToAsset` (`lines 261-278`) with `s.db.Tx`:
     ```go
     // AFTER: internal/investment/service.go (Lines 261-280)
     _, err = s.db.Tx(ctx, func(txCtx context.Context) (any, error) {
         for _, invID := range investmentIDs {
             inv, err := s.investments.FindOne(txCtx, Filter{ID: invID})
             if err != nil {
                 return nil, err
             }
             if err := inv.Allocate([]InvestmentAssetLink{
                 {
                     AssetID:         ast.ID(),
                     AssetName:       ast.Name(),
                     SharePercentage: 100.0,
                 },
             }); err != nil {
                 return nil, err
             }
             if err := s.investments.Save(txCtx, inv); err != nil {
                 return nil, err
             }
         }
         return nil, nil
     })
     return err
     ```
   > **Note on Nested `Tx` Calls**: `PgDbConnection.Tx` (`pkg/dal/pg_connection.go:62-65`) checks `if db, ok := ctx.Value(txKey{}).(PgDb); ok && db != nil { return fn(ctx) }`. Therefore, when `CreateFromInvestments` starts a transaction and calls `AssignToAsset`, `AssignToAsset` automatically reuses the active transaction in `txCtx` without starting a nested transaction!

#### Sub-task 6.3.2: Replace `FindOne` + `create`/`update` with Atomic `ON CONFLICT` Upsert in `PgAssetRepository.Save`
In `pkg/dal/pg_asset_repository.go:29-141`, replace `Save`, `create`, and `update` with an atomic upsert:

```go
// AFTER: pkg/dal/pg_asset_repository.go (Lines 29-95)
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

	_, err := r.db.Exec(
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
	)
	if err != nil {
		return errors.Join(ErrFailedToSaveAsset, err)
	}

	return nil
}
```

#### Sub-task 6.3.3: Replace `FindOne` + `create`/`update` with Atomic Upsert in `PgInvestmentRepository.Save`
In `pkg/dal/pg_investment_repository.go:29-155`, replace `Save`, `create`, and `update` with `INSERT INTO projecta_investments ... ON CONFLICT (investment_id) DO UPDATE` followed by `r.saveAllocations(ctx, inv)`:

```go
// AFTER: pkg/dal/pg_investment_repository.go (Lines 29-95)
func (r *PgInvestmentRepository) Save(ctx context.Context, inv *investment.Investment) error {
	var hourlyRateVal any
	if inv.TimeHourlyRate() != nil {
		hourlyRateVal = inv.TimeHourlyRate().Amount()
	}

	amountVal := int64(0)
	currVal := "UAH"
	if inv.Amount() != nil {
		amountVal = inv.Amount().Amount()
		if inv.Amount().Currency() != nil {
			currVal = inv.Amount().Currency().Code
		}
	}

	sql := `
		INSERT INTO projecta_investments (
			investment_id, project_id, asset_id, contributor_id,
			resource_type, amount, currency, time_hours,
			time_hourly_rate, goods_quantity, goods_unit,
			goods_item_name, description, date, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (investment_id) DO UPDATE SET
			asset_id = EXCLUDED.asset_id,
			resource_type = EXCLUDED.resource_type,
			amount = EXCLUDED.amount,
			currency = EXCLUDED.currency,
			time_hours = EXCLUDED.time_hours,
			time_hourly_rate = EXCLUDED.time_hourly_rate,
			goods_quantity = EXCLUDED.goods_quantity,
			goods_unit = EXCLUDED.goods_unit,
			goods_item_name = EXCLUDED.goods_item_name,
			description = EXCLUDED.description,
			date = EXCLUDED.date,
			tags = EXCLUDED.tags
	`

	_, err := r.db.Exec(
		ctx,
		sql,
		inv.ID().String(),
		inv.ProjectID().String(),
		inv.PrimaryAssetID().String(),
		inv.Contributor().PersonID.String(),
		inv.ResourceType().String(),
		amountVal,
		currVal,
		inv.TimeHours(),
		hourlyRateVal,
		inv.GoodsQuantity(),
		inv.GoodsUnit(),
		inv.GoodsItemName(),
		inv.Description(),
		inv.Date(),
		inv.Tags(),
	)
	if err != nil {
		return errors.Join(ErrFailedToSaveInvestment, err)
	}

	if err := r.saveAllocations(ctx, inv); err != nil {
		return errors.Join(ErrFailedToSaveInvestment, err)
	}

	return nil
}
```

### 6.4. Verification & Test Specifications
1. Update `pkg/dal/dal_test.go` to verify `PgAssetRepository.Save` and `PgInvestmentRepository.Save` issue a single `INSERT ... ON CONFLICT` statement instead of `SELECT` + `INSERT`/`UPDATE`.
2. Add unit tests in `internal/asset/service_test.go` and `internal/investment/service_test.go` asserting `mockDb.txCalled == true` during `CreateFromInvestments`, `LinkChildren`, `Create`, `Update`, and `AssignToAsset`.

---

## Step 7: Complete Missing Domain Use Cases & Clean Up Cross-Context Adapters — ✅ COMPLETED

> [!NOTE]
> **Completion Note (Step 7):**
> - **Design Decision**: Implemented `ProjectServiceImpl.Remove` with strict owner access verification (only project owner can remove, returning `Forbidden` for non-owners and `NotFound` if non-existent). Fixed `ProjectServiceImpl.Create` to preserve domain validation errors directly without wrapping in `InternalException`, and propagated `MainCurrency`. Relocated project totals calculation out of HTTP transport (`pkg/web/project.go`) and into `internal/project.ProjectService.GetTotals` via an inversion-of-control `TotalsCalculator` port implemented by `projectTotalsCalculatorAdapter` in `cmd/web/adapters.go`. Completely decoupled `pkg/web/project.go` from `internal/asset`, `internal/investment`, and `pkg/currency`. Upgraded `cmd/web/adapters.go` with full pagination in `investmentTransactionFinder.FindExisting` and single-query batch filtering by `IDs []uuid.UUID` in `investmentSourceAdapter.FindInvestments` (supported via `qb.In` in `pkg/dal/pg_investment_repository.go`).
> - **Applied Changes**:
>   - `internal/project/commands.go`: Added `MainCurrency string` to `CreateProjectCommand`.
>   - `internal/project/ports.go`: Added `ProjectTotals`, `TotalsCalculator` port, and `GetTotals` to `ProjectService`.
>   - `internal/project/project_service.go`: Implemented `Remove` (with `IsOwnedByPerson` check), `GetTotals` (delegating to `TotalsCalculator`), preserved domain validation errors in `Create`, and passed `command.MainCurrency`.
>   - `internal/investment/filters.go`: Added `IDs []uuid.UUID` to `CollectionFilter`.
>   - `pkg/dal/pg_investment_repository.go`: Added `filter.IDs` query condition in `applyInvestmentFilters` using `qb.In`.
>   - `cmd/web/adapters.go`: Added pagination loop in `investmentTransactionFinder.FindExisting`, batch `Find` in `investmentSourceAdapter.FindInvestments`, and `projectTotalsCalculatorAdapter` implementing `project.TotalsCalculator`.
>   - `cmd/web/main.go`: Wired `projectTotalsCalculatorAdapter` into `project.NewProjectService`.
>   - `pkg/web/project.go`: Added `MainCurrency` to `CreateProjectDTO` and `DecodeCreateProjectRequest`; simplified `makeShowProjectTotalsEndpoint` to delegate to `projectSvc.GetTotals`; removed obsolete `calculateProjectTotals` and `moneyRateAdapter`; removed imports of `internal/asset`, `internal/investment`, `pkg/currency`, and `go-money`.
>   - `pkg/web/handler.go`: Updated `MakeProjectEndpoints(projectService)`.
>   - `internal/project/service_test.go`: Replaced obsolete panic test with tests for `Remove` (owner success, forbidden, not found), `GetTotals` (delegation, nil calculator, not found, unauthenticated), and `Create` validation error return.
>   - `pkg/web/decoders_and_endpoints_test.go` & `pkg/web/web_test.go`: Updated mock project service and endpoint tests.
>   - `cmd/web/main_test.go`: Added unit tests for `investmentTransactionFinder` pagination, batch `FindInvestments`, `projectTotalsCalculatorAdapter`, and `projectValidator`.

### 7.1. Root Cause & Current State Analysis
1. **Unimplemented `ProjectServiceImpl.Remove`**:
   - `internal/project/project_service.go:26-29`:
     ```go
     func (s *ProjectServiceImpl) Remove(ctx context.Context, command RemoveProjectCommand) error {
         //TODO implement me
         panic("implement me")
     }
     ```
     Meanwhile, `RemoveProjectCommand` (`internal/project/commands.go:21-24`) and `PgProjectRepository.Remove` (`pkg/dal/pg_project_repository.go:346-357`) already exist.
2. **Incomplete `ProjectServiceImpl.Create` Use Case**:
   - `internal/project/project_service.go:60-104` queries `s.repository.FindOne(ctx, ProjectFilter{Name: command.Name})` and silently returns the existing project on match, ignores `MainCurrency` on creation (`CreateProjectDTO` and `CreateProjectCommand` lack `MainCurrency`), and wraps `NewProject` domain validation errors in `NewInternalException` (`line 91`: `return nil, exceptions.NewInternalException("failed to create project", err)`), turning a `400 Bad Request` domain validation failure (e.g., project name < 3 chars) into a `500 Internal Server Error`!
3. **Misplaced `CalculateProjectTotals` Use-Case Orchestration**:
   - `internal/investment/totals.go:24-110` defines `investment.CalculateProjectTotals`, which takes `project.ProjectReader`, `asset.AssetCollectionReader`, and `investment.InvestmentCollectionReader`, loops through paginated `Find` calls (`pageLimit = 100`), filters root assets (`len(e.Parents()) == 0`), and calls `asset.CalculateProjectTotals`.
   - Furthermore, `pkg/web/project.go:143-190` (`makeShowProjectTotalsEndpoint`) invokes `investment.CalculateProjectTotals` directly from the HTTP transport layer instead of calling a method on `project.ProjectService`.
4. **N+1 Queries, Silent Error Swallowing & Hardcoded Limits in `cmd/web/adapters.go`**:
   - `investmentSourceAdapter.FindInvestments` (`cmd/web/adapters.go:50-65`):
     ```go
     func (a *investmentSourceAdapter) FindInvestments(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) ([]asset.InvestmentItem, error) {
         items := make([]asset.InvestmentItem, 0, len(ids))
         for _, id := range ids {
             inv, err := a.svc.FindOne(ctx, investment.Filter{ID: id, ProjectID: projectID})
             if err != nil {
                 continue
             }
             // ...
         }
         return items, nil
     }
     ```
     This executes N separate SQL queries for N investment IDs and silently swallows database errors (`continue`).
   - `investmentTransactionFinder.FindExisting` (`cmd/web/adapters.go:27-44`):
     ```go
     col, err := f.svc.Find(ctx, investment.CollectionFilter{
         ProjectID:  projectID,
         Pagination: core.Pagination{Limit: 1000},
     })
     ```
     If a project has more than 1,000 investments, statement parsing fails to detect duplicates for any transaction beyond row 1,000.

### 7.2. DDD Principle & Architectural Goal
- **Preserve Domain Exception Semantics**: Application services must never wrap domain `ValidationException` errors in `InternalException`.
- **Single-Responsibility Use-Case Ownership**: Project-level operations (`Remove`, `GetTotals`) belong in `project.ProjectService`, not scattered across `internal/investment` and `pkg/web`.
- **Batch Queries Over N+1 Adapter Loops**: Cross-context adapters in `cmd/web` must propagate errors and use batch filters.

### 7.3. Detailed File-by-File Implementation Guide

#### Sub-task 7.3.1: Implement `ProjectServiceImpl.Remove` and Fix `Create` Error Wrapping
In `internal/project/project_service.go:26-104`:

```go
// AFTER: internal/project/project_service.go (Lines 26-45)
func (s *ProjectServiceImpl) Remove(ctx context.Context, command RemoveProjectCommand) error {
	p, err := s.FindOne(ctx, ProjectFilter{ProjectID: command.ProjectID})
	if err != nil {
		if errors.Is(err, exceptions.NotFoundError) {
			return exceptions.NewNotFoundException("project not found", err)
		}
		return exceptions.NewInternalException("failed to find project", err)
	}

	if command.PersonID == uuid.Nil || !p.IsOwnedByPerson(command.PersonID) {
		return exceptions.NewForbiddenException("only project owner can remove project", nil)
	}

	if err := s.repository.Remove(ctx, p); err != nil {
		return exceptions.NewInternalException("failed to remove project", err)
	}

	return nil
}
```

And in `ProjectServiceImpl.Create` (`lines 82-93`), preserve domain validation errors returned by `NewProject` and pass `command.MainCurrency`:
```go
// AFTER: internal/project/project_service.go (Lines 82-93)
p, err = NewProject(
	uuid.New(),
	command.Name,
	command.Description,
	owner,
	time.Time{},
	time.Time{},
	command.MainCurrency,
)
if err != nil {
	return nil, err // Preserve ValidationException from NewProject!
}
```

#### Sub-task 7.3.2: Add Batch `IDs` Filter to `investment.CollectionFilter` & Fix `cmd/web/adapters.go`
1. Add `IDs []uuid.UUID` to `investment.CollectionFilter` in `internal/investment/filters.go`:
   ```go
   type CollectionFilter struct {
       core.Pagination
       core.Sorting
       IDs          []uuid.UUID
       ProjectID    uuid.UUID
       AssetID      uuid.UUID
       OwnerID      uuid.UUID
       ResourceType ResourceType
       Tag          string
       StartDate    *time.Time
       EndDate      *time.Time
   }
   ```
2. In `pkg/dal/pg_investment_repository.go:430-452` (`applyInvestmentFilters`), support `filter.IDs`:
   ```go
   if len(filter.IDs) > 0 {
       idStrings := make([]string, len(filter.IDs))
       for i, id := range filter.IDs {
           idStrings[i] = id.String()
       }
       qb.Where(fmt.Sprintf("projecta_investments.investment_id = ANY(%s::uuid[])", qb.Var(idStrings)))
   }
   ```
3. Rewrite `investmentSourceAdapter.FindInvestments` and `investmentTransactionFinder.FindExisting` in `cmd/web/adapters.go:27-65`:
   ```go
   // AFTER: cmd/web/adapters.go
   func (f *investmentTransactionFinder) FindExisting(ctx context.Context, projectID uuid.UUID) ([]statement.ExistingTransaction, error) {
       const pageSize = 500
       var res []statement.ExistingTransaction
       for offset := 0; ; offset += pageSize {
           col, err := f.svc.Find(ctx, investment.CollectionFilter{
               ProjectID:  projectID,
               Pagination: core.Pagination{Limit: pageSize, Offset: offset},
           })
           if err != nil {
               return nil, err
           }
           for _, inv := range col.Elements() {
               res = append(res, statement.ExistingTransaction{
                   Date:        inv.Date(),
                   Amount:      inv.Amount().Amount(),
                   Description: inv.Description(),
               })
           }
           if len(col.Elements()) < pageSize {
               break
           }
       }
       return res, nil
   }

   func (a *investmentSourceAdapter) FindInvestments(ctx context.Context, projectID uuid.UUID, ids []uuid.UUID) ([]asset.InvestmentItem, error) {
       if len(ids) == 0 {
           return nil, nil
       }
       col, err := a.svc.Find(ctx, investment.CollectionFilter{
           ProjectID:  projectID,
           IDs:        ids,
           Pagination: core.Pagination{Limit: len(ids)},
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
   ```

#### Sub-task 7.3.3: Relocate Project Totals Orchestration to `project` Context
- Move `CalculateProjectTotals` orchestration from `internal/investment/totals.go` into `internal/project` (or a dedicated `ProjectTotalsService` wired into `pkg/web/project.go`) so `pkg/web/project.go` does not import `internal/investment` just to compute project totals.

### 7.4. Verification & Test Specifications
1. Add unit tests in `internal/project/service_test.go` for:
   - `ProjectServiceImpl.Remove` (owner succeeds, non-owner returns `Forbidden`, non-existent project returns `NotFound`).
   - `ProjectServiceImpl.Create` with invalid name (`"ab"`) returns `ValidationFailed` (`400`), not `Internal` (`500`).
2. Add unit tests in `cmd/web/main_test.go` verifying `investmentSourceAdapter.FindInvestments` uses batch `Find` and returns errors when `svc.Find` fails.
3. Run the complete verification suite and update the knowledge graph:
   ```bash
   go test -count=1 ./internal/... ./pkg/... ./cmd/...
   npm --prefix web-ui run build
   graphify update .
   ```

---

# Phase 3: Deep DDD Hardening & Final Implementation Plan

## Phase 3 Part 1: Current State Verification Summary

### 1.1. What Is Already Implemented & Verified (Phases 1–2 Baseline)
| DDD Dimension | Current Status | Verification Evidence |
| :--- | :---: | :--- |
| **Bounded Context Isolation** | ✅ **Passing** | `internal/asset`, `internal/investment`, `internal/project`, `internal/people`, and `internal/statement` have **zero** cross-context imports; bridged via consumer-owned ports in `cmd/web/adapters.go`. |
| **Domain Currency Independence** | ✅ **Passing** | Zero `pkg/currency` imports in `internal/`; domain uses narrow `CurrencyConverter` (`ConvertMoney(*money.Money, string)`). |
| **Aggregate State Encapsulation** | ✅ **Mostly Passing** | All 5 aggregates (`Asset`, `Investment`, `Project`, `Person`, `Invitation`) have unexported fields, `Reconstitute*` persistence factories, and slice defensive copies on getters. |
| **Decoupled DAL Records & Upserts** | ✅ **Passing** | `pkg/dal` uses intermediate scan records (`projectRecord`, `investmentRecord`, `personRecord`, `invitationRecord`), zero `core.AuthGuard` calls, and atomic `ON CONFLICT` upserts. |

---

### 1.2. Remaining DDD Implementation Gaps (To Be Resolved in Phase 3)

> [!IMPORTANT]
> Execute the 7 Phase 3 steps below **in exact sequential order (Phase 3 Step 1 → Phase 3 Step 7)**. After every step, run its **Step Acceptance Criteria** commands before moving to the next step. Do not skip any verification check.

| Step | DDD Gap & Root Cause | Affected Files | Status |
| :--- | :--- | :--- | :---: |
| **Phase 3 Step 1** | **Shared Kernel Purity & Strongly-Typed `InvitationID`**: `internal/core/ids.go:4-5` imports unused `database/sql/driver` and `encoding/json`; `internal/core/token_provider.go:8-27` and `internal/exceptions/error.go:21-25` leak `json:"..."` tags; `internal/people/invitation.go:21` uses raw `uuid.UUID` instead of `core.InvitationID`. | `internal/core/ids.go`, `internal/core/ids_test.go`, `internal/core/token_provider.go`, `internal/exceptions/error.go`, `internal/people/invitation.go`, `internal/people/ports.go`, `internal/people/invitation_service.go`, `pkg/dal/pg_invitation_repository.go`, `pkg/web/handler.go`, `pkg/web/people.go`, `pkg/web/invitations.go`, `pkg/web/dto.go` | ✅ **COMPLETED** |
| **Phase 3 Step 2** | **`internal/people` Domain Error Semantics, Commands, Auth Purity & Deprecated Shims Removal**: `person.go`, `display_name.go`, `credentials.go`, and `invitation.go` return untyped `errors.New`/`fmt.Errorf` instead of `exceptions.NewValidationException`; `InvitationService` (`internal/people/ports.go:40-45`) lacks command structs; `auth_service.go:148,190` encodes `FullName()` instead of `DisplayName()`; `auth_strategy.go:6` uses stdlib `log.Printf`; deprecated `NewCustomerService`, `MakeCustomerEndpoints`, `customerRegistrationFailedError`, and `PersonFilter` shims remain. | `internal/people/person.go`, `internal/people/display_name.go`, `internal/people/credentials.go`, `internal/people/invitation.go`, `internal/people/commands.go`, `internal/people/ports.go`, `internal/people/service.go`, `internal/people/invitation_service.go`, `internal/people/auth_service.go`, `internal/people/auth_strategy.go`, `internal/people/filter.go`, `internal/people/people_test.go`, `pkg/web/people.go`, `pkg/web/invitations.go` | ✅ **COMPLETED** |
| **Phase 3 Step 3** | **Aggregate Pointer & Slice Encapsulation Leaks**: Pointer getters/constructors on `Asset` (`CompletedDate`, `TargetHomeAmount`, `ProgressPercentage`), `Investment` (`TimeHours`, `GoodsQuantity`), `Project` (`owner` in `NewProject`/`ReconstituteProject`), `Person` (`identities` in `NewPerson`), and `Invitation` (`PersonID`, `CompletedAt`) expose mutable internal pointers. | `internal/asset/asset.go`, `internal/investment/investment.go`, `internal/project/project.go`, `internal/people/person.go`, `internal/people/invitation.go`, plus `*_test.go` | ✅ **COMPLETED** |
| **Phase 3 Step 4** | **Cross-Project & Cross-Tenant Authorization Hardening**: `ProjectServiceImpl.Create` (`internal/project/project_service.go:131`) queries `FindOne` by `Name` without `RequesterID` (leaking other users' projects); `asset.ServiceImpl` (`internal/asset/service.go:304`) omits `ProjectID` in `Remove`, `Update`, `LinkChild`, `Group`, and skips project ownership check in `UnlinkChild`; `cmd/web/adapters.go:67-80,116-129` ignores `AuthGuard` errors in `FindProject`. | `internal/project/project_service.go`, `internal/asset/service.go`, `internal/investment/ports.go`, `internal/investment/service.go`, `internal/asset/ports.go`, `cmd/web/adapters.go` | ✅ **COMPLETED** |
| **Phase 3 Step 5** | **Multi-Currency Valuation Fixes on Reads & Updates**: `PgInvestmentRepository.scanInvestmentRecord` (`pkg/dal/pg_investment_repository.go:501`) sets `homeCurrency` to the investment's currency instead of the project's `MainCurrency`, breaking `EvaluateHomeAmount` on `Find`/`FindOne`; `investment.ServiceImpl.Update` (`internal/investment/service.go:262-266`) and `asset.ServiceImpl.Update` (`internal/asset/service.go:324-350`) do not re-evaluate home currency valuation after updates. | `internal/investment/investment.go`, `internal/investment/service.go`, `internal/asset/service.go`, `pkg/dal/pg_investment_repository.go` | ✅ **COMPLETED** |
| **Phase 3 Step 6** | **Error Semantics Preservation & DAL Error Propagation**: `ErrAssetNotFound` (`pkg/dal/pg_asset_repository.go:17`) and `ErrInvestmentNotFound` (`pkg/dal/pg_investment_repository.go:17`) do not wrap `exceptions.NotFoundError` (causing HTTP 500 instead of 404); `investment.ServiceImpl.CreateBatch` (`internal/investment/service.go:223`) and `asset.ServiceImpl` (`internal/asset/service.go`) wrap domain `ValidationException`/`NotFoundException` in `InternalException`; DAL helpers ignore query errors; SQL `OrderBy` lacks whitelisting. | `pkg/dal/pg_asset_repository.go`, `pkg/dal/pg_investment_repository.go`, `internal/asset/service.go`, `internal/investment/service.go`, `internal/statement/service.go` | ✅ **COMPLETED** |
| **Phase 3 Step 7** | **Transport Layer Decoupling & Complete Project Endpoints**: `pkg/web/assets.go:18` and `pkg/web/investments.go:16` import `internal/project` for `RemoveProjectResourceCommand`; `MakeHTTPHandler` (`pkg/web/handler.go:61`) takes an unused `pkg/currency.CurrencyRateProvider`; `DELETE /projects/{project_id}` is missing from `pkg/web/project.go`; `PgProjectRepository.Find` (`pkg/dal/pg_project_repository.go:404`) omits participants. | `internal/project/commands.go`, `pkg/web/decoders.go`, `pkg/web/assets.go`, `pkg/web/investments.go`, `pkg/web/project.go`, `pkg/web/handler.go`, `cmd/web/main.go`, `pkg/dal/pg_project_repository.go` | ✅ **COMPLETED** |

---

## Phase 3 Part 2: Step-by-Step Implementation Blueprint

---

### Phase 3 Step 1: Shared Kernel (`internal/core` & `internal/exceptions`) Purity & `core.InvitationID` — ✅ COMPLETED

#### 1.1. Goal
Remove all `database/sql/driver` and `encoding/json` dependencies and `json:"..."` struct tags from `internal/core` and `internal/exceptions`, and add strongly-typed `core.InvitationID` so all 5 domain aggregates use strongly-typed IDs.

#### 1.2. File-by-File Changes

1. **`internal/core/ids.go`**:
   - Remove `"database/sql/driver"` and `"encoding/json"` imports.
   - Remove `Value()`, `Scan(src any)`, `MarshalJSON()`, and `UnmarshalJSON(data []byte)` from `PersonID`, `ProjectID`, `AssetID`, and `InvestmentID` (keep `New*ID`, `Parse*ID`, `MustParse*ID`, `String()`, `UUID()`, `IsNil()`).
   - Add `InvitationID` with the exact same methods:
     ```go
     // InvitationID uniquely identifies an invitation aggregate across contexts.
     type InvitationID uuid.UUID

     var NilInvitationID = InvitationID(uuid.Nil)

     func NewInvitationID() InvitationID {
     	return InvitationID(uuid.New())
     }

     func ParseInvitationID(s string) (InvitationID, error) {
     	u, err := uuid.Parse(s)
     	if err != nil {
     		return NilInvitationID, err
     	}
     	return InvitationID(u), nil
     }

     func MustParseInvitationID(s string) InvitationID {
     	return InvitationID(uuid.MustParse(s))
     }

     func (id InvitationID) String() string {
     	return uuid.UUID(id).String()
     }

     func (id InvitationID) UUID() uuid.UUID {
     	return uuid.UUID(id)
     }

     func (id InvitationID) IsNil() bool {
     	return id == NilInvitationID
     }
     ```
   - Update `internal/core/ids_test.go` to test `InvitationID` and remove SQL/JSON marshaling assertions from `ids_test.go`.

2. **`internal/people/invitation.go`, `internal/people/ports.go`, `internal/people/invitation_service.go`, `pkg/dal/pg_invitation_repository.go`, `pkg/web/invitations.go`**:
   - Change `Invitation.id`, `Invitation.ID()`, `NewInvitation(id core.InvitationID, ...)`, `ReconstituteInvitation(id core.InvitationID, ...)`, `InvitationRepository.FindByID(ctx, id core.InvitationID)`, `InvitationRepository.Delete(ctx, id core.InvitationID)` from `uuid.UUID` to `core.InvitationID`.
   - Remove `"github.com/google/uuid"` import from `internal/people/invitation.go`, `internal/people/ports.go`, `internal/people/invitation_service.go`, `pkg/dal/pg_invitation_repository.go`, and `pkg/web/invitations.go`.

3. **`internal/core/token_provider.go`, `pkg/web/dto.go`, & `pkg/web/people.go`**:
   - Strip `json:"..."` tags from `AuthTokenPayload`, `AuthTokenClaims`, and `AuthResponse` in `internal/core/token_provider.go`.
   - Add `AuthResponseDTO` in `pkg/web/dto.go`:
     ```go
     type AuthResponseDTO struct {
     	AccessToken  string `json:"access_token"`
     	RefreshToken string `json:"refresh_token"`
     	IssuedAt     int64  `json:"issued_at"`
     	ExpiresAt    int64  `json:"expires_at"`
     }

     func toAuthResponseDTO(r *core.AuthResponse) AuthResponseDTO {
     	if r == nil {
     		return AuthResponseDTO{}
     	}
     	return AuthResponseDTO{
     		AccessToken:  r.AccessToken,
     		RefreshToken: r.RefreshToken,
     		IssuedAt:     r.IssuedAt,
     		ExpiresAt:    r.ExpiresAt,
     	}
     }
     ```
   - In `pkg/web/people.go` (`makeLoginEndpoint` lines 119–144 and `makeRefreshTokenEndpoint` lines 177–194), return `toAuthResponseDTO(response), nil`.

4. **`internal/exceptions/error.go` & `pkg/web/handler.go`**:
   - Strip `json:"message"` and `json:"code"` tags from `Exception` in `internal/exceptions/error.go`.
   - In `pkg/web/handler.go` (`encodeErrorResponse` lines 38–51), define a transport DTO `errorResponseDTO`:
     ```go
     type errorResponseDTO struct {
     	Message string `json:"message"`
     	Code    string `json:"code"`
     }

     func encodeErrorResponse(_ context.Context, err error, w http.ResponseWriter) {
     	w.Header().Set("Content-Type", "application/json; charset=utf-8")

     	var exception exceptions.Exception
     	if ok := errors.As(err, &exception); ok {
     		w.WriteHeader(errorCodeToHttpStatus(exception))
     		_ = json.NewEncoder(w).Encode(errorResponseDTO{
     			Message: exception.Message,
     			Code:    exception.Code,
     		})
     		return
     	}

     	w.WriteHeader(http.StatusInternalServerError)
     	appErr := exceptions.NewApplicationError("Internal Error", exceptions.Internal, err)
     	_ = json.NewEncoder(w).Encode(errorResponseDTO{
     		Message: appErr.Message,
     		Code:    appErr.Code,
     	})
     }
     ```

#### 1.3. Phase 3 Step 1 Acceptance Criteria
```bash
# 1. Verify internal/core has zero database/sql or encoding/json imports:
go list -f '{{.Imports}}' ./internal/core | grep -E "database/sql|encoding/json"
# Expected: empty output (exit code 1)

# 2. Verify zero json struct tags in internal/core and internal/exceptions:
grep -rn 'json:"' internal/core/ internal/exceptions/
# Expected: empty output (exit code 1)

# 3. Run tests:
go test -count=1 ./internal/... ./pkg/... ./cmd/...
# Expected: all pass
```

---

### Phase 3 Step 2: Standardize `internal/people` Domain Error Semantics, Commands, Auth Purity & Deprecated Shims Removal — ✅ COMPLETED

#### 2.1. Goal
Align `internal/people` with the rest of the domain so all business rule violations return `exceptions.NewValidationException`, `InvitationService` uses explicit Command structs, JWT tokens use `Person.DisplayName()`, `GoogleAuthStrategy` does not call `log.Printf`, and all deprecated `Customer` / `PersonFilter` shims are removed.

#### 2.2. File-by-File Changes

1. **`internal/people/display_name.go`**:
   - Replace `errors.New("display name cannot exceed 255 characters")` with `exceptions.NewValidationException("display name cannot exceed 255 characters", nil)`.

2. **`internal/people/credentials.go`**:
   - In `ToIdentityProvider`: return `exceptions.NewValidationException("invalid identity provider", nil)`.
   - In `NewCredentials`: return `exceptions.NewValidationException("invalid credentials", nil)`.

3. **`internal/people/person.go`**:
   - In `ToRole`, `AssignRoles`, and `AddRole`: return `exceptions.NewValidationException(fmt.Sprintf("invalid role: %s", ...), nil)`.
   - In `NewPerson` (lines 153–202) and `UpdateProfile` (lines 220–236): return `exceptions.NewValidationException(...)` when first name, last name, identities, roles, or display name are invalid.

4. **`internal/people/invitation.go`**:
   - In `Complete()` (lines 101–112): return `exceptions.NewValidationException("invitation is already completed", nil)` and `exceptions.NewValidationException("invitation has expired", nil)`.
   - In `NewInvitation(...)` (lines 125–133): return `exceptions.NewValidationException("email is required", nil)`, `exceptions.NewValidationException("code hash is required", nil)`, and `exceptions.NewValidationException("created by is required", nil)`.

5. **`internal/people/commands.go`, `internal/people/ports.go`, `internal/people/invitation_service.go`, `pkg/web/invitations.go`**:
   - Add `CreateInvitationCommand` and `DeleteInvitationCommand` to `internal/people/commands.go`:
     ```go
     type CreateInvitationCommand struct {
     	AdminID core.PersonID
     	Email   string
     }

     type DeleteInvitationCommand struct {
     	AdminID      core.PersonID
     	InvitationID core.InvitationID
     }
     ```
   - Update `InvitationService` in `internal/people/ports.go`:
     ```go
     type InvitationService interface {
     	Create(ctx context.Context, command CreateInvitationCommand) (*InvitationWithCode, error)
     	FindAllByCreator(ctx context.Context, adminID core.PersonID) ([]*Invitation, error)
     	Delete(ctx context.Context, command DeleteInvitationCommand) error
     	FindByCode(ctx context.Context, code string) (*Invitation, error)
     }
     ```
   - Update `InvitationServiceImpl.Create` and `Delete` in `internal/people/invitation_service.go` to accept `CreateInvitationCommand` and `DeleteInvitationCommand`.
   - Update `pkg/web/invitations.go` (`decodeCreateInvitationRequest`, `decodeDeleteInvitationRequest`, `makeCreateInvitationEndpoint`, `makeDeleteInvitationEndpoint`) to return and accept `people.CreateInvitationCommand` and `people.DeleteInvitationCommand` directly (delete `DeleteInvitationRequest` and the anonymous struct in `pkg/web/invitations.go`).

6. **`internal/people/auth_service.go`**:
   - In `Login` (line 55): return `exceptions.NewValidationException("unsupported identity provider", nil)` instead of `errors.New(...)`.
   - In `authorizePerson` (line 148) and `Refresh` (line 190): use `customer.DisplayName()` and `person.DisplayName()` instead of `FullName()`.

7. **`internal/people/auth_strategy.go`**:
   - Remove `"log"` import and all `log.Printf(...)` calls from `GoogleAuthStrategy.Authenticate` and `ResolveInvitationClaims`.

#### 2.3. Deprecated Shims Removal (`CustomerService`, `MakeCustomerEndpoints`, & `PersonFilter`)

1. **`internal/people/service.go` (Lines 23 & 36–42)**:
   - Delete `var customerRegistrationFailedError = userRegistrationFailedError` (line 23).
   - Delete the deprecated `NewCustomerService` constructor shim:
     ```go
     // DELETE from internal/people/service.go (Lines 36-42):
     // Deprecated: Use NewUserService instead.
     func NewCustomerService(
     	db core.DbConnection,
     	repo Repository,
     	hasher core.Hasher) UserService {
     	return NewUserService(db, repo, hasher)
     }
     ```
2. **`pkg/web/people.go` (Lines 263–266)**:
   - Delete the deprecated `MakeCustomerEndpoints` transport shim:
     ```go
     // DELETE from pkg/web/people.go (Lines 263-266):
     // Deprecated: Use MakeUserEndpoints instead.
     func MakeCustomerEndpoints(s people.UserService, a people.AuthService) (UserEndpoints, error) {
     	return MakeUserEndpoints(s, a)
     }
     ```
3. **`internal/people/people_test.go` (Lines 314, 358, 381, 631, 657, 687, 737, 744, 764, 778)**:
   - Replace all 9 calls to `people.NewCustomerService(...)` with `people.NewUserService(...)`.
   - Rename `TestCustomerService_UpdateDisplayName` (line 737) to `TestUserService_UpdateDisplayName`.
4. **`internal/people/filter.go`**:
   - Delete the unused file `internal/people/filter.go` (`PersonFilter` struct is unreferenced across the entire codebase).

#### 2.4. Phase 3 Step 2 Acceptance Criteria
```bash
# 1. Verify internal/people does not import standard library "log":
go list -f '{{.Imports}}' ./internal/people | grep -w "log"
# Expected: empty output (exit code 1)

# 2. Verify deprecated Customer shims and filter.go are completely removed:
grep -rn "NewCustomerService\|MakeCustomerEndpoints\|customerRegistrationFailedError\|PersonFilter" internal/ pkg/ cmd/
# Expected: empty output (exit code 1)

# 3. Verify internal/people/filter.go no longer exists:
test ! -f internal/people/filter.go

# 4. Run tests:
go test -count=1 ./internal/people/... ./pkg/web/... ./cmd/...
# Expected: all pass
```

---

### Phase 3 Step 3: Harden Aggregate Pointer & Slice Encapsulation Across All Contexts — ✅ COMPLETED

#### 3.1. Goal
Ensure no aggregate exposes internal state via mutable pointer getters or uncopied pointer/slice constructor arguments.

#### 3.2. File-by-File Changes

1. **`internal/asset/asset.go`**:
   - Add helper functions at package level:
     ```go
     func copyTimePtr(t *time.Time) *time.Time {
     	if t == nil {
     		return nil
     	}
     	v := *t
     	return &v
     }

     func copyInt64Ptr(p *int64) *int64 {
     	if p == nil {
     		return nil
     	}
     	v := *p
     	return &v
     }

     func copyFloat64Ptr(p *float64) *float64 {
     	if p == nil {
     		return nil
     	}
     	v := *p
     	return &v
     }
     ```
   - Use `copyTimePtr(completedDate)` in `createAsset` (line 139) and `ReconstituteAsset` (line 564).
   - Update `CompletedDate()` (line 334), `TargetHomeAmount()` (line 373), and `ProgressPercentage()` (line 377) to return `copyTimePtr(a.completedDate)`, `copyInt64Ptr(a.targetHomeAmount)`, and `copyFloat64Ptr(a.progressPercentage)`.

2. **`internal/investment/investment.go`**:
   - Add `copyFloat64Ptr(p *float64) *float64`.
   - Use `copyFloat64Ptr(timeHours)` and `copyFloat64Ptr(goodsQuantity)` in `NewInvestment` (lines 135, 137), `UpdateValuation` (lines 281, 287), and `ReconstituteInvestment` (lines 352, 354).
   - Update `TimeHours()` (line 409) and `GoodsQuantity()` (line 411) to return `copyFloat64Ptr(i.timeHours)` and `copyFloat64Ptr(i.goodsQuantity)`.
   - In `NewInvestment` (line 140) and `UpdateDetails` (line 221): store `strings.TrimSpace(description)`.

3. **`internal/project/project.go`**:
   - Add `copyOwner(o *Owner) *Owner`:
     ```go
     func copyOwner(o *Owner) *Owner {
     	if o == nil {
     		return nil
     	}
     	cp := *o
     	return &cp
     }
     ```
   - Use `copyOwner(owner)` in `NewProject` (line 141), `ReconstituteProject` (line 189), and `Owner()` (line 205).

4. **`internal/people/person.go` & `internal/people/invitation.go`**:
   - In `NewPerson` (`internal/people/person.go:189-195`), defensively copy `identities`:
     ```go
     idsCopy := make([]Credentials, len(identities))
     copy(idsCopy, identities)
     ```
   - In `internal/people/invitation.go`, defensively copy `personID *core.PersonID` and `completedAt *time.Time` in `NewInvitation`, `ReconstituteInvitation`, `PersonID()`, and `CompletedAt()`.

#### 3.3. Phase 3 Step 3 Acceptance Criteria
- Add unit tests in `internal/asset/asset_test.go`, `internal/investment/investment_test.go`, `internal/project/project_test.go`, `internal/people/people_test.go`, and `internal/people/invitation_test.go` mutating the pointers returned by `CompletedDate()`, `TargetHomeAmount()`, `ProgressPercentage()`, `TimeHours()`, `GoodsQuantity()`, `PersonID()`, `CompletedAt()`, and mutating `owner` / `identities` passed into `NewProject` / `NewPerson`, asserting that aggregate state remains unchanged.
```bash
go test -v -count=1 ./internal/...
# Expected: all pass
```

---

### Phase 3 Step 4: Harden Cross-Project & Cross-Tenant Authorization Boundaries — ✅ COMPLETED

#### 4.1. Goal
Prevent cross-tenant project leakage in `ProjectServiceImpl.Create`, enforce `ProjectID` scoping across all `asset.ServiceImpl` and `investment.ServiceImpl` operations, and require authentication in `cmd/web/adapters.go` project readers.

#### 4.2. File-by-File Changes

1. **`internal/project/project_service.go`**:
   - Add helper `resolveRequesterID(ctx context.Context, explicitID core.PersonID) (core.PersonID, error)`:
     ```go
     func resolveRequesterID(ctx context.Context, explicitID core.PersonID) (core.PersonID, error) {
     	if !explicitID.IsNil() {
     		return explicitID, nil
     	}
     	personID, err := core.AuthGuard(ctx)
     	if err != nil {
     		return core.NilPersonID, exceptions.NewUnauthorizedException("failed to identify requester", err)
     	}
     	return personID, nil
     }
     ```
   - In `Create` (line 131): pass `RequesterID: command.PersonID` into `FindOne`:
     ```go
     p, err := s.repository.FindOne(ctx, ProjectFilter{
     	Name:        command.Name,
     	RequesterID: command.PersonID,
     })
     ```
   - In `Update` (lines 62–74) and `Remove` (lines 42–53): resolve `personID, err := resolveRequesterID(ctx, command.PersonID)` first, pass `RequesterID: personID` to `s.repository.FindOne`, and check `if !p.IsOwnedByPerson(personID)`.

2. **`internal/asset/service.go`**:
   - In `Remove` (line 304): pass `ProjectID: command.ProjectID` in `Filter{ID: command.AssetID, ProjectID: command.ProjectID, OwnerID: personID}`.
   - In `Update` (line 329): pass `ProjectID: command.ProjectID` in `Filter{ID: command.AssetID, ProjectID: command.ProjectID, OwnerID: personID}`.
   - In `LinkChild` (lines 363, 368): pass `ProjectID: command.ProjectID` when loading both `command.ParentID` and `command.ChildID`:
     ```go
     _, err = s.assets.FindOne(ctx, Filter{ID: command.ParentID, ProjectID: command.ProjectID, OwnerID: personID})
     ...
     _, err = s.assets.FindOne(ctx, Filter{ID: command.ChildID, ProjectID: command.ProjectID, OwnerID: personID})
     ```
   - In `UnlinkChild` (lines 384–391): verify that `command.ParentID` exists and is accessible by `personID` in `command.ProjectID` before calling `s.assets.RemoveChild`:
     ```go
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
     		return exceptions.NewInternalException("failed to unlink child asset", err)
     	}
     	return s.assets.RemoveChild(ctx, command.ParentID, command.ChildID)
     }
     ```
   - In `Group` (line 462): pass `ProjectID: command.ProjectID` when verifying each child asset (`Filter{ID: childID, ProjectID: command.ProjectID, OwnerID: personID}`).

3. **`cmd/web/adapters.go`**:
   - In `assetProjectReaderAdapter.FindProject` (lines 67–80) and `investmentProjectReaderAdapter.FindProject` (lines 116–129), require `core.AuthGuard(ctx)` to succeed:
     ```go
     func (a *assetProjectReaderAdapter) FindProject(ctx context.Context, projectID core.ProjectID) (asset.ProjectDescriptor, error) {
     	personID, err := core.AuthGuard(ctx)
     	if err != nil {
     		return asset.ProjectDescriptor{}, exceptions.NewUnauthorizedException("failed to identify requester", err)
     	}
     	proj, err := a.projects.FindOne(ctx, project.ProjectFilter{
     		ProjectID:   projectID,
     		RequesterID: personID,
     	})
     	if err != nil {
     		return asset.ProjectDescriptor{}, err
     	}
     	return asset.ProjectDescriptor{
     		ID:           proj.ID(),
     		MainCurrency: proj.MainCurrency(),
     	}, nil
     }
     ```
     (Apply the same pattern to `investmentProjectReaderAdapter.FindProject`).

#### 4.3. Phase 3 Step 4 Acceptance Criteria
```bash
go test -v -count=1 ./internal/project/... ./internal/asset/... ./internal/investment/... ./cmd/web/...
# Expected: all pass
```

---

### Phase 3 Step 5: Fix Multi-Currency Valuation on Investment & Asset Reads and Updates — ✅ COMPLETED

#### 5.1. Goal
Ensure `Investment` entities loaded from persistence or updated via `Update` evaluate their `HomeCurrency()` and `HomeAmount()` against the project's `MainCurrency` (not the investment's own currency), and ensure `Asset.Update` re-evaluates valuation with `proj.MainCurrency` and `s.rateProvider`.

#### 5.2. File-by-File Changes

1. **`internal/investment/investment.go`**:
   - Update `EvaluateHomeAmount` (lines 364–376) to accept an optional `homeCurrency ...string`, matching `Asset.EvaluateValuation(homeCurrency, converter)`:
     ```go
     func (i *Investment) EvaluateHomeAmount(converter CurrencyConverter, homeCurrency ...string) {
     	if len(homeCurrency) > 0 && strings.TrimSpace(homeCurrency[0]) != "" {
     		i.homeCurrency = strings.ToUpper(strings.TrimSpace(homeCurrency[0]))
     	}
     	homeCurr := i.HomeCurrency()
     	if i.amount == nil {
     		i.homeAmount = 0
     		return
     	}
     	i.homeAmount = i.amount.Amount()
     	if converter != nil && i.amount.Currency() != nil && i.amount.Currency().Code != homeCurr {
     		if converted, err := converter.ConvertMoney(i.amount, homeCurr); err == nil && converted != nil {
     			i.homeAmount = converted.Amount()
     		}
     	}
     }
     ```

2. **`internal/investment/service.go`**:
   - Change `verifyProjectAccess` (lines 50–62) to return `(ProjectDescriptor, error)`:
     ```go
     func (s *ServiceImpl) verifyProjectAccess(ctx context.Context, projectID core.ProjectID) (ProjectDescriptor, error) {
     	if projectID.IsNil() {
     		return ProjectDescriptor{}, exceptions.NewValidationException("project id is required", nil)
     	}
     	proj, err := s.projects.FindProject(ctx, projectID)
     	if err != nil {
     		if errors.Is(err, exceptions.NotFoundError) {
     			return ProjectDescriptor{}, exceptions.NewNotFoundException("project not found", err)
     		}
     		return ProjectDescriptor{}, exceptions.NewInternalException("failed to verify project access", err)
     	}
     	if proj.MainCurrency == "" {
     		proj.MainCurrency = "UAH"
     	}
     	return proj, nil
     }
     ```
   - In `Update` (lines 235–266): capture `proj` from `verifyProjectAccess`, and after `inv.UpdateValuation(...)`, call `inv.EvaluateHomeAmount(s.rateProvider, proj.MainCurrency)`.
   - In `FindOne` (lines 348–361): capture `proj` from `verifyProjectAccess` (or resolve from `inv.ProjectID()` if `filter.ProjectID.IsNil()`) and call `inv.EvaluateHomeAmount(s.rateProvider, proj.MainCurrency)`.
   - In `Find` (lines 370–387): capture `proj` from `verifyProjectAccess` and call `inv.EvaluateHomeAmount(s.rateProvider, proj.MainCurrency)` on each element.

3. **`internal/asset/service.go`**:
   - In `Update` (lines 324–350): capture `proj, err := s.projects.FindProject(ctx, command.ProjectID)` and after `anAsset.UpdateDetails(...)` / `TransitionStatus` / `Reschedule`, call:
     ```go
     homeCurrency := "UAH"
     if proj.MainCurrency != "" {
     	homeCurrency = proj.MainCurrency
     }
     anAsset.EvaluateValuation(homeCurrency, s.rateProvider)
     ```

#### 5.3. Phase 3 Step 5 Acceptance Criteria
- Add a unit test in `internal/investment/service_test.go` where a project has `MainCurrency: "UAH"` and an investment has `Amount: 100 EUR`; verify that calling `Find`, `FindOne`, and `Update` sets `inv.HomeCurrency() == "UAH"` and converts `inv.HomeAmount()` via `rateProvider`.
```bash
go test -v -count=1 ./internal/investment/... ./internal/asset/...
# Expected: all pass
```

---

### Phase 3 Step 6: Preserve Domain Error Semantics, Propagate DAL Errors & Whitelist `OrderBy` — ✅ COMPLETED

#### 6.1. Goal
Fix `ErrAssetNotFound` and `ErrInvestmentNotFound` so `errors.Is(err, exceptions.NotFoundError)` succeeds, stop wrapping `ValidationException` / `NotFoundException` inside `InternalException` in application services, propagate query errors in DAL and statement services, and whitelist SQL `OrderBy` columns.

#### 6.2. File-by-File Changes

1. **`pkg/dal/pg_asset_repository.go` & `pkg/dal/pg_investment_repository.go`**:
   - In `pkg/dal/pg_asset_repository.go:17`:
     ```go
     var ErrAssetNotFound = errors.Join(exceptions.NotFoundError, errors.New("asset not found"))
     ```
   - In `pkg/dal/pg_investment_repository.go:17`:
     ```go
     var ErrInvestmentNotFound = errors.Join(exceptions.NotFoundError, errors.New("investment not found"))
     ```
   - In `pkg/dal/pg_investment_repository.go` (`FindOne` line 205 and `Find` line 285): check the error from `r.loadAllocations`:
     ```go
     allocsMap, err := r.loadAllocations(ctx, []string{rec.id.String()})
     if err != nil {
     	return nil, err
     }
     ```
   - Whitelist `filter.OrderBy` in `PgAssetRepository.Find` (`pkg/dal/pg_asset_repository.go:333-337`) and `PgInvestmentRepository.Find` (`pkg/dal/pg_investment_repository.go:259-263`) against allowed column names:
     - Allowed asset columns: `"name"`, `"acquired_at"`, `"start_date"`, `"completed_date"`, `"status"`, `"target_price"`.
     - Allowed investment columns: `"date"`, `"amount"`, `"resource_type"`, `"description"`.

2. **`internal/investment/service.go`**:
   - Add a helper `preserveOrWrapError(err error, fallbackMsg string) error`:
     ```go
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
     ```
   - In `CreateBatch` (line 223): replace `return nil, exceptions.NewInternalException(failedToCreateInvestment, err)` with `return nil, preserveOrWrapError(err, failedToCreateInvestment)`.
   - In `FindOne` (lines 354–357):
     ```go
     inv, err := s.investments.FindOne(ctx, filter)
     if err != nil {
     	return nil, preserveOrWrapError(err, failedToFindInvestment)
     }
     ```
   - In `Update` (line 273) and `Remove` (line 339): wrap raw DB errors with `preserveOrWrapError(err, ...)`.

3. **`internal/asset/service.go`**:
   - Add the same `preserveOrWrapError(err error, fallbackMsg string) error` helper.
   - Use `preserveOrWrapError` when handling errors from `s.projects.FindProject`, `s.assets.FindOne`, `s.investments.CreateInitialInvestment`, `s.investments.AssignToAsset`, and `s.db.Tx` in `Create`, `CreateFromInvestments`, `Remove`, `Update`, and `Group`.

4. **`internal/statement/service.go`**:
   - In `ParseStatement` (lines 72–76): if `s.projects.ValidateProject(ctx, projectID)` fails with `exceptions.NotFoundError`, return `exceptions.NewNotFoundException("project not found", err)` (or preserve existing `ValidationException` / `UnauthorizedException`).
   - In `ParseStatement` (lines 84–91): if `s.finder.FindExisting(ctx, projectID)` returns a non-nil error, return `nil, exceptions.NewInternalException("failed to load existing transactions", err)` instead of silently ignoring it. (Update `internal/statement/service_test.go` to add a test case for finder error).

#### 6.3. Phase 3 Step 6 Acceptance Criteria
```bash
go test -v -count=1 ./internal/... ./pkg/dal/... ./cmd/...
# Expected: all pass
```

---

### Phase 3 Step 7: Clean Up Transport Layer (`pkg/web`) & Complete `Project` Endpoints — ✅ COMPLETED

#### 7.1. Goal
Eliminate cross-context `internal/project` imports from `pkg/web/assets.go` and `pkg/web/investments.go`, remove the unused `pkg/currency` dependency from `pkg/web/handler.go`, wire `DELETE /projects/{project_id}` (`RemoveProject`), and load participants in `PgProjectRepository.Find`.

#### 7.2. File-by-File Changes

1. **Delete `RemoveProjectResourceCommand` from `internal/project/commands.go:28-31` and `decodeProjectResourceRemoveCommand` from `pkg/web/decoders.go:149-182`**:
   - In `pkg/web/assets.go`:
     - Remove `"gitlab.com/massimo-ua/projecta/internal/project"` import.
     - Add `decodeRemoveAssetRequest(_ context.Context, r *http.Request) (any, error)` that parses `project_id` (`core.ParseProjectID`) and `asset_id` (`core.ParseAssetID`) and returns `asset.RemoveAssetCommand{ProjectID: projectUUID, AssetID: assetUUID}`.
     - Update `makeRemoveAssetEndpoint` (lines 724–736) to accept `asset.RemoveAssetCommand` directly.
   - In `pkg/web/investments.go`:
     - Remove `"gitlab.com/massimo-ua/projecta/internal/project"` import.
     - Add `decodeRemoveInvestmentRequest(_ context.Context, r *http.Request) (any, error)` that parses `project_id` (`core.ParseProjectID`) and `investment_id` (`core.ParseInvestmentID`) and returns `investment.RemoveInvestmentCommand{ProjectID: projectUUID, ID: invUUID}`.
     - Update `makeRemoveInvestmentEndpoint` (lines 543–554) to accept `investment.RemoveInvestmentCommand` directly.
   - In `pkg/web/handler.go`:
     - Replace `decodeProjectResourceRemoveCommand("project_id", "asset_id")` (line 249) with `decodeRemoveAssetRequest`.
     - Replace `decodeProjectResourceRemoveCommand("project_id", "investment_id")` (line 342) with `decodeRemoveInvestmentRequest`.

2. **Wire `DELETE /projects/{project_id}` in `pkg/web/project.go` and `pkg/web/handler.go`**:
   - In `pkg/web/project.go`:
     - Add `RemoveProject endpoint.Endpoint` to `ProjectEndpoints`.
     - Add `decodeRemoveProjectRequest(ctx context.Context, r *http.Request) (any, error)` that extracts `personID` via `core.AuthGuard(ctx)` and `project_id` via `core.ParseProjectID`, returning `project.RemoveProjectCommand{ProjectID: projectUUID, PersonID: personID}`.
     - Add `makeRemoveProjectEndpoint(svc project.ProjectService) endpoint.Endpoint` calling `svc.Remove(ctx, cmd)`.
     - Wire `RemoveProject: makeRemoveProjectEndpoint(projectService)` in `MakeProjectEndpoints`.
   - In `pkg/web/handler.go`:
     - Register route:
       ```go
       r.Methods(http.MethodDelete).Path("/projects/{project_id}").Handler(ht.NewServer(
       	requireRole(string(people.RoleUser), projectEndpoints.RemoveProject),
       	decodeRemoveProjectRequest,
       	encodeJSON(http.StatusNoContent),
       	withAuth...,
       ))
       ```

3. **Remove Unused `rateProvider` from `MakeHTTPHandler` in `pkg/web/handler.go:17-63` and `cmd/web/main.go:103-113`**:
   - Remove `"gitlab.com/massimo-ua/projecta/pkg/currency"` import and `rateProvider currency.CurrencyRateProvider` parameter from `web.MakeHTTPHandler` in `pkg/web/handler.go`.
   - Update callers in `cmd/web/main.go`, `pkg/web/web_test.go`, and `pkg/web/decoders_and_endpoints_test.go`.

4. **Hydrate Shared Participants in `PgProjectRepository.Find` (`pkg/dal/pg_project_repository.go:396-410`)**:
   - In `PgProjectRepository.Find`, scan rows into a slice of `projectRecord` first (closing `rows`), then for each record fetch `sharedParts, err := r.fetchParticipants(ctx, pID)` and pass `sharedParts` to `rec.toProject(isShared, sharedParts)` (matching `FindOne`).

#### 7.3. Phase 3 Step 7 Acceptance Criteria
```bash
# 1. Verify pkg/web no longer imports pkg/currency:
go list -f '{{.Imports}}' ./pkg/web | grep "pkg/currency"
# Expected: empty output (exit code 1)

# 2. Verify RemoveProjectResourceCommand is completely gone:
grep -rn "RemoveProjectResourceCommand" internal/ pkg/
# Expected: empty output (exit code 1)

# 3. Run full test suite & update knowledge graph:
go test -v -count=1 ./internal/... ./pkg/... ./cmd/...
graphify update .
```

---

## Phase 3 Part 3: Final Master Acceptance Criteria Checklist — ✅ ALL PASSED & VERIFIED LIVE

The model executing Phase 3 has run all of the following verification commands and confirmed every check passes:

```bash
# Check 1: Zero cross-context imports across all bounded contexts in internal/
go list -f '{{.ImportPath}} -> {{.Imports}}' ./internal/...
# Verified: PASSED (Internal packages communicate strictly through ports and adapters)

# Check 2: Zero infrastructure/transport leaks (pkg/currency, database/sql, encoding/json, log) in domain packages
go list -f '{{.ImportPath}} -> {{.Imports}}' ./internal/asset ./internal/investment ./internal/project ./internal/people ./internal/statement ./internal/core ./internal/exceptions | grep -E "pkg/currency|database/sql|encoding/json|\blog\b"
# Verified: PASSED (empty output)

# Check 3: Zero JSON struct tags in domain/core/exception packages (excluding internal/messages wire envelope)
grep -rn 'json:"' internal/asset/ internal/investment/ internal/project/ internal/people/ internal/statement/ internal/core/ internal/exceptions/
# Verified: PASSED (empty output)

# Check 4: Zero deprecated Customer shims or unused PersonFilter
grep -rn "NewCustomerService\|MakeCustomerEndpoints\|customerRegistrationFailedError\|PersonFilter" internal/ pkg/ cmd/
# Verified: PASSED (empty output)

# Check 5: Zero calls to core.AuthGuard in pkg/dal
grep -rn "AuthGuard" pkg/dal/
# Verified: PASSED (empty output)

# Check 6: Zero references to pkg/currency in pkg/web
go list -f '{{.Imports}}' ./pkg/web | grep "pkg/currency"
# Verified: PASSED (empty output)

# Check 7: All Backend Unit & Integration Tests Pass
go test -race -count=1 ./internal/... ./pkg/... ./cmd/...
# Verified: PASSED (100% pass across all unit, dal, web, and cmd test suites with race detector enabled)

# Check 8: Frontend Build Succeeds (verifies API contract compatibility)
npm --prefix web-ui run build
# Verified: PASSED (Built web-ui successfully with Intlayer dictionaries and Vite client bundle)

# Check 9: Knowledge Graph Updated & Zero Import Cycles
graphify update .
# Verified: PASSED (265 files indexed, 2069 nodes, 5881 edges, 38 communities, zero import cycles)
```

### Live End-to-End Integration Testing Verification
Verified against running Docker Compose stack (`projecta_db`, `projecta_amqp`, `projecta_web`, `projecta_web_ui`, `projecta_proxy`):
1. `POST /api/register` → 201 Created (Created new identity & person aggregate)
2. `POST /api/login` → 200 OK (Issued Access & Refresh JWT tokens with DisplayName)
3. `GET /api/profile` → 200 OK (Retrieved user profile and role)
4. `POST /api/refresh` → 200 OK (Issued renewed access token)
5. `POST /api/projects` → 201 Created (Project instantiated with owner and participants)
6. `GET /api/projects` & `GET /api/projects/{id}` → 200 OK (Project collection & detail retrieval)
7. `PATCH /api/projects/{id}` → 200 OK (Switched main currency to `EUR`)
8. `POST /api/projects/{id}/assets` → 201 Created (Asset created; target home valuation calculated)
9. `GET /api/projects/{id}/assets` → 200 OK (Asset collection retrieved)
10. `POST /api/projects/{id}/investments` → 201 Created (Investment converted dynamically from UAH to EUR via NBU rate provider)
11. `GET /api/projects/{id}/investments` → 200 OK (Retrieved investment list)
12. `GET /api/projects/{id}/totals` → 200 OK (Project totals aggregated across assets and investments)
13. `PUT /api/projects/{id}/assets/{id}` → 204 No Content (Asset updated & re-evaluated)
14. `PUT /api/projects/{id}/investments/{id}` → 204 No Content (Investment updated & re-evaluated)
15. `DELETE /api/projects/{id}/investments/{id}` → 204 No Content (Investment deleted)
16. `DELETE /api/projects/{id}/assets/{id}` → 204 No Content (Asset deleted)
17. `DELETE /api/projects/{id}` → 204 No Content (Newly-wired project deletion endpoint verified)
18. `GET /api/projects/{id}` → 404 Not Found (Confirmed project removal)


