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
3. [Step 3: Consolidate Validation & Strip JSON]   → verify: negative prices/shares rejected by domain; zero `encoding/json` imports or `json:` tags in `internal/` domain types
4. [Step 4: Harden Aggregate Encapsulation]        → verify: slice/pointer mutation tests prove immutability; `NewAsset` removed; `UpdateValuation`/`UpdateDetails` return errors
5. [Step 5: Unify Authorization Boundaries]        → verify: zero `core.AuthGuard` calls in `pkg/dal`; shared-project participants can read/update assets; investment endpoints enforce project membership
6. [Step 6: Enforce Transactional Boundaries]      → verify: all multi-statement writes execute inside `db.Tx`; `PgAssetRepository.Save` & `PgInvestmentRepository.Save` use atomic `ON CONFLICT` upserts
7. [Step 7: Complete Use Cases & Clean Adapters]   → verify: `ProjectService.Remove` implemented; `CalculateProjectTotals` moved to `project`; N+1 & silent error swallowing in `cmd/web/adapters.go` eliminated
```

| Step | Gap Addressed | Affected Packages | Status / Recommended Order |
| :--- | :--- | :--- | :---: |
| **Step 1** | Gap 1: Flat Package Layout & Cross-Context Coupling | `internal/asset`, `internal/investment`, `internal/project`, `internal/people`, `cmd/web`, `pkg/web` | ✅ **COMPLETED** |
| **Step 2** | Gap 2: Domain Coupled to `pkg/currency` & Dual Money Types | `internal/asset`, `internal/investment`, `pkg/currency`, `cmd/web`, `pkg/web` | ✅ **COMPLETED** |
| **Step 3** | Gap 3: Business Validation in HTTP Decoders & JSON in Domain | `pkg/web`, `internal/asset`, `internal/investment`, `internal/project`, `internal/people`, `internal/statement` | **Next** |
| **Step 4** | Gap 4: Encapsulation Leaks & Silent Validation Failures | `internal/asset`, `internal/investment`, `internal/project` | **2nd** |
| **Step 5** | Gap 5: Inconsistent Authorization Boundaries Across Layers | `internal/project`, `internal/asset`, `internal/investment`, `pkg/dal` | **4th** |
| **Step 6** | Gap 6: Missing Transactional Boundaries (Unit of Work Gaps) | `internal/asset`, `internal/investment`, `pkg/dal` | **5th** |
| **Step 7** | Gap 7: Incomplete Domain Use Cases & Cross-Context Quirks | `internal/project`, `internal/investment`, `pkg/dal`, `pkg/web`, `cmd/web` | **6th** |

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
