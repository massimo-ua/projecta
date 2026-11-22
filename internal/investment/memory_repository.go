package investment

import (
	"context"
	"sort"
	"sync"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

// MemoryInvestmentRepository is a thread-safe, value-copy in-memory Repository
// suitable for Domain-First exploration and fast unit/application tests.
type MemoryInvestmentRepository struct {
	mu          sync.RWMutex
	investments map[core.InvestmentID]Investment
}

func NewMemoryInvestmentRepository() *MemoryInvestmentRepository {
	return &MemoryInvestmentRepository{
		investments: make(map[core.InvestmentID]Investment),
	}
}

func cloneInvestment(stored Investment) *Investment {
	return ReconstituteInvestment(
		stored.ID(),
		stored.ProjectID(),
		stored.HomeCurrency(),
		stored.PrimaryAssetID(),
		stored.PrimaryAssetName(),
		stored.AssetAllocations(),
		stored.Contributor(),
		stored.ResourceType(),
		stored.Amount(),
		stored.TimeHours(),
		stored.TimeHourlyRate(),
		stored.GoodsQuantity(),
		stored.GoodsUnit(),
		stored.GoodsItemName(),
		stored.Description(),
		stored.Date(),
		stored.Tags(),
	)
}

func (m *MemoryInvestmentRepository) Save(_ context.Context, inv *Investment) error {
	if inv == nil {
		return exceptions.NewValidationException("investment is required", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.investments[inv.ID()] = *cloneInvestment(*inv)
	return nil
}

func (m *MemoryInvestmentRepository) Remove(_ context.Context, inv *Investment) error {
	if inv == nil {
		return exceptions.NewValidationException("investment is required", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.investments[inv.ID()]; !ok {
		return exceptions.NewNotFoundException("investment not found", nil)
	}
	delete(m.investments, inv.ID())
	return nil
}

func matchesInvestmentAsset(inv *Investment, assetID core.AssetID) bool {
	if assetID.IsNil() {
		return true
	}
	if inv.PrimaryAssetID() == assetID {
		return true
	}
	for _, a := range inv.AssetAllocations() {
		if a.AssetID == assetID {
			return true
		}
	}
	return false
}

func (m *MemoryInvestmentRepository) FindOne(_ context.Context, filter Filter) (*Investment, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stored, ok := m.investments[filter.ID]
	if !ok {
		return nil, exceptions.NewNotFoundException("investment not found", nil)
	}
	cp := cloneInvestment(stored)
	if !filter.ProjectID.IsNil() && cp.ProjectID() != filter.ProjectID {
		return nil, exceptions.NewNotFoundException("investment not found", nil)
	}
	if !matchesInvestmentAsset(cp, filter.AssetID) {
		return nil, exceptions.NewNotFoundException("investment not found", nil)
	}
	return cp, nil
}

func (m *MemoryInvestmentRepository) UpdateInvestment(
	ctx context.Context,
	id core.InvestmentID,
	projectID core.ProjectID,
	updateFn func(ctx context.Context, inv *Investment) (*Investment, error),
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stored, ok := m.investments[id]
	if !ok {
		return exceptions.NewNotFoundException("investment not found", nil)
	}
	candidate := cloneInvestment(stored)
	if !projectID.IsNil() && candidate.ProjectID() != projectID {
		return exceptions.NewNotFoundException("investment not found", nil)
	}

	updated, err := updateFn(ctx, candidate)
	if err != nil {
		return err
	}
	if updated == nil {
		return exceptions.NewValidationException("updated investment cannot be nil", nil)
	}
	m.investments[id] = *cloneInvestment(*updated)
	return nil
}

func (m *MemoryInvestmentRepository) Find(_ context.Context, filter CollectionFilter) (*Collection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	idSet := make(map[core.InvestmentID]struct{}, len(filter.IDs))
	for _, id := range filter.IDs {
		idSet[id] = struct{}{}
	}

	var matched []*Investment
	for _, stored := range m.investments {
		cp := cloneInvestment(stored)
		if len(idSet) > 0 {
			if _, ok := idSet[cp.ID()]; !ok {
				continue
			}
		}
		if !filter.ProjectID.IsNil() && cp.ProjectID() != filter.ProjectID {
			continue
		}
		if !matchesInvestmentAsset(cp, filter.AssetID) {
			continue
		}
		if !filter.OwnerID.IsNil() && cp.Contributor().PersonID != filter.OwnerID {
			continue
		}
		if filter.ResourceType != "" && cp.ResourceType() != filter.ResourceType {
			continue
		}
		if filter.Tag != "" {
			hasTag := false
			for _, t := range cp.Tags() {
				if t == filter.Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}
		if filter.StartDate != nil && cp.Date().Before(*filter.StartDate) {
			continue
		}
		if filter.EndDate != nil && cp.Date().After(*filter.EndDate) {
			continue
		}
		matched = append(matched, cp)
	}

	sort.Slice(matched, func(i, j int) bool {
		if matched[i].Date().Equal(matched[j].Date()) {
			return matched[i].ID().String() > matched[j].ID().String()
		}
		return matched[i].Date().After(matched[j].Date())
	})

	total := len(matched)
	col := NewCollection(total)
	if filter.Offset >= total {
		return col, nil
	}
	start := filter.Offset
	end := total
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}
	for _, inv := range matched[start:end] {
		col.Add(inv)
	}
	return col, nil
}

func (m *MemoryInvestmentRepository) FindTags(_ context.Context, projectID core.ProjectID) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	seen := make(map[string]struct{})
	for _, stored := range m.investments {
		if stored.ProjectID() != projectID {
			continue
		}
		for _, t := range stored.Tags() {
			if t != "" {
				seen[t] = struct{}{}
			}
		}
	}
	tags := make([]string, 0, len(seen))
	for t := range seen {
		tags = append(tags, t)
	}
	sort.Strings(tags)
	return tags, nil
}
