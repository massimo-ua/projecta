package asset

import (
	"context"
	"sort"
	"strings"
	"sync"

	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

type edgeKey struct {
	parentID core.AssetID
	childID  core.AssetID
}

// MemoryAssetRepository is a thread-safe, value-copy in-memory Repository
// suitable for Domain-First exploration and fast unit/application tests.
type MemoryAssetRepository struct {
	mu     sync.RWMutex
	assets map[core.AssetID]Asset
	edges  map[edgeKey]float64
}

func NewMemoryAssetRepository() *MemoryAssetRepository {
	return &MemoryAssetRepository{
		assets: make(map[core.AssetID]Asset),
		edges:  make(map[edgeKey]float64),
	}
}

func cloneAsset(stored Asset) *Asset {
	cp := ReconstituteAsset(
		stored.ID(),
		stored.Name(),
		stored.Description(),
		stored.ProjectID(),
		stored.Price(),
		stored.AcquiredAt(),
		stored.Owner(),
		stored.Status(),
		stored.StartDate(),
		stored.CompletedDate(),
		stored.TargetCompletionDate(),
		stored.TargetPrice(),
		stored.Tags(),
	)
	cp.AssignValuation(stored.DirectCost(), stored.TotalCost(), stored.Children(), stored.Parents())
	return cp
}

func (m *MemoryAssetRepository) hydrateLinksLocked(a *Asset) {
	var children []ChildAssetLink
	var parents []ParentAssetLink
	for k, share := range m.edges {
		if k.parentID == a.ID() {
			if child, ok := m.assets[k.childID]; ok {
				children = append(children, ChildAssetLink{
					ChildID:         k.childID,
					ChildName:       child.Name(),
					SharePercentage: share,
					TotalCost:       child.TotalCost(),
				})
			}
		}
		if k.childID == a.ID() {
			if parent, ok := m.assets[k.parentID]; ok {
				parents = append(parents, ParentAssetLink{
					ParentID:        k.parentID,
					ParentName:      parent.Name(),
					SharePercentage: share,
				})
			}
		}
	}
	a.AssignValuation(a.DirectCost(), a.TotalCost(), children, parents)
}

func (m *MemoryAssetRepository) Save(_ context.Context, a *Asset) error {
	if a == nil {
		return exceptions.NewValidationException("asset is required", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.assets[a.ID()] = *cloneAsset(*a)
	return nil
}

func (m *MemoryAssetRepository) Remove(_ context.Context, a *Asset) error {
	if a == nil {
		return exceptions.NewValidationException("asset is required", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.assets[a.ID()]; !ok {
		return exceptions.NewNotFoundException("asset not found", nil)
	}
	delete(m.assets, a.ID())
	for k := range m.edges {
		if k.parentID == a.ID() || k.childID == a.ID() {
			delete(m.edges, k)
		}
	}
	return nil
}

func (m *MemoryAssetRepository) FindOne(_ context.Context, filter Filter) (*Asset, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stored, ok := m.assets[filter.ID]
	if !ok {
		return nil, exceptions.NewNotFoundException("asset not found", nil)
	}
	cp := cloneAsset(stored)
	if !filter.ProjectID.IsNil() && cp.ProjectID() != filter.ProjectID {
		return nil, exceptions.NewNotFoundException("asset not found", nil)
	}
	if filter.Name != "" && !strings.HasPrefix(strings.ToLower(cp.Name()), strings.ToLower(filter.Name)) {
		return nil, exceptions.NewNotFoundException("asset not found", nil)
	}
	m.hydrateLinksLocked(cp)
	return cp, nil
}

func (m *MemoryAssetRepository) UpdateAsset(
	ctx context.Context,
	assetID core.AssetID,
	projectID core.ProjectID,
	_ core.PersonID,
	updateFn func(ctx context.Context, a *Asset) (*Asset, error),
) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	stored, ok := m.assets[assetID]
	if !ok {
		return exceptions.NewNotFoundException("asset not found", nil)
	}
	candidate := cloneAsset(stored)
	if !projectID.IsNil() && candidate.ProjectID() != projectID {
		return exceptions.NewNotFoundException("asset not found", nil)
	}
	m.hydrateLinksLocked(candidate)

	updated, err := updateFn(ctx, candidate)
	if err != nil {
		return err
	}
	if updated == nil {
		return exceptions.NewValidationException("updated asset cannot be nil", nil)
	}
	m.assets[assetID] = *cloneAsset(*updated)
	return nil
}

func (m *MemoryAssetRepository) Find(_ context.Context, filter CollectionFilter) (*Collection, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*Asset
	for _, stored := range m.assets {
		cp := cloneAsset(stored)
		if !filter.ProjectID.IsNil() && cp.ProjectID() != filter.ProjectID {
			continue
		}
		if filter.Name != "" && !strings.HasPrefix(strings.ToLower(cp.Name()), strings.ToLower(filter.Name)) {
			continue
		}
		m.hydrateLinksLocked(cp)
		matched = append(matched, cp)
	}

	sort.Slice(matched, func(i, j int) bool {
		if matched[i].AcquiredAt().Equal(matched[j].AcquiredAt()) {
			return matched[i].ID().String() > matched[j].ID().String()
		}
		return matched[i].AcquiredAt().After(matched[j].AcquiredAt())
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
	for _, a := range matched[start:end] {
		col.Add(a)
	}
	return col, nil
}

func (m *MemoryAssetRepository) AddChild(_ context.Context, parentID core.AssetID, childID core.AssetID, sharePercentage float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.edges[edgeKey{parentID: parentID, childID: childID}] = sharePercentage
	return nil
}

func (m *MemoryAssetRepository) RemoveChild(_ context.Context, parentID core.AssetID, childID core.AssetID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.edges, edgeKey{parentID: parentID, childID: childID})
	return nil
}

func (m *MemoryAssetRepository) FindChildren(_ context.Context, parentID core.AssetID) ([]ChildAssetLink, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []ChildAssetLink
	for k, share := range m.edges {
		if k.parentID == parentID {
			name := ""
			if ch, ok := m.assets[k.childID]; ok {
				name = ch.Name()
			}
			res = append(res, ChildAssetLink{ChildID: k.childID, ChildName: name, SharePercentage: share})
		}
	}
	return res, nil
}

func (m *MemoryAssetRepository) FindParents(_ context.Context, childID core.AssetID) ([]ParentAssetLink, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []ParentAssetLink
	for k, share := range m.edges {
		if k.childID == childID {
			name := ""
			if p, ok := m.assets[k.parentID]; ok {
				name = p.Name()
			}
			res = append(res, ParentAssetLink{ParentID: k.parentID, ParentName: name, SharePercentage: share})
		}
	}
	return res, nil
}

func (m *MemoryAssetRepository) FindAncestors(_ context.Context, assetID core.AssetID) ([]core.AssetID, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	visited := make(map[core.AssetID]bool)
	queue := []core.AssetID{assetID}
	var ancestors []core.AssetID

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]
		for k := range m.edges {
			if k.childID == curr && !visited[k.parentID] {
				visited[k.parentID] = true
				ancestors = append(ancestors, k.parentID)
				queue = append(queue, k.parentID)
			}
		}
	}
	return ancestors, nil
}
