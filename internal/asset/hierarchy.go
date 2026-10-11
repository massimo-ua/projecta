package asset

import (
	"math"

	"github.com/Rhymond/go-money"
	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

var (
	ErrSelfReference      = exceptions.NewValidationException("an asset cannot be a child of itself", nil)
	ErrInvalidShare       = exceptions.NewValidationException("share percentage must be between 0 and 100", nil)
	ErrCircularDependency = exceptions.NewValidationException("cannot link child asset: circular dependency detected", nil)
)

type ChildEdge struct {
	ChildID         uuid.UUID
	SharePercentage float64
}

type ParentEdge struct {
	ParentID        uuid.UUID
	SharePercentage float64
}

// AssetHierarchy encapsulates a Directed Acyclic Graph (DAG) of assets and their composition links.
type AssetHierarchy struct {
	assets      map[uuid.UUID]*Asset
	childrenMap map[uuid.UUID][]ChildEdge
	parentsMap  map[uuid.UUID][]ParentEdge
}

func NewAssetHierarchy() *AssetHierarchy {
	return &AssetHierarchy{
		assets:      make(map[uuid.UUID]*Asset),
		childrenMap: make(map[uuid.UUID][]ChildEdge),
		parentsMap:  make(map[uuid.UUID][]ParentEdge),
	}
}

func (h *AssetHierarchy) AddAsset(a *Asset) {
	if a != nil {
		h.assets[a.ID()] = a
	}
}

func (h *AssetHierarchy) GetAsset(id uuid.UUID) (*Asset, bool) {
	a, ok := h.assets[id]
	return a, ok
}

func (h *AssetHierarchy) AddEdge(parentID, childID uuid.UUID, sharePercentage float64) error {
	if parentID == childID {
		return ErrSelfReference
	}
	if sharePercentage <= 0 || sharePercentage > 100 {
		return ErrInvalidShare
	}

	ancestors := h.Ancestors(parentID)
	for _, aID := range ancestors {
		if aID == childID {
			return ErrCircularDependency
		}
	}

	h.childrenMap[parentID] = append(h.childrenMap[parentID], ChildEdge{
		ChildID:         childID,
		SharePercentage: sharePercentage,
	})
	h.parentsMap[childID] = append(h.parentsMap[childID], ParentEdge{
		ParentID:        parentID,
		SharePercentage: sharePercentage,
	})

	return nil
}

// Ancestors returns all ancestor asset IDs of the given node ID.
func (h *AssetHierarchy) Ancestors(nodeID uuid.UUID) []uuid.UUID {
	visited := make(map[uuid.UUID]bool)
	var result []uuid.UUID

	var dfs func(curr uuid.UUID)
	dfs = func(curr uuid.UUID) {
		for _, p := range h.parentsMap[curr] {
			if !visited[p.ParentID] {
				visited[p.ParentID] = true
				result = append(result, p.ParentID)
				dfs(p.ParentID)
			}
		}
	}

	dfs(nodeID)
	return result
}

// CalculateRollups computes the rolled-up total cost for all assets in the hierarchy using memoized recursion.
func (h *AssetHierarchy) CalculateRollups(directCosts map[uuid.UUID]*money.Money, defaultCurrency string) map[uuid.UUID]*money.Money {
	if defaultCurrency == "" {
		defaultCurrency = "UAH"
	}
	totalCosts := make(map[uuid.UUID]*money.Money)
	memo := make(map[uuid.UUID]*money.Money)
	visiting := make(map[uuid.UUID]bool)

	var compute func(id uuid.UUID) *money.Money
	compute = func(id uuid.UUID) *money.Money {
		if val, exists := memo[id]; exists {
			return val
		}
		if visiting[id] {
			return money.New(0, defaultCurrency)
		}
		visiting[id] = true

		curr := defaultCurrency
		var baseAmount int64
		if dc, ok := directCosts[id]; ok && dc != nil {
			baseAmount = dc.Amount()
			if dc.Currency().Code != "" {
				curr = dc.Currency().Code
			}
		}

		amount := baseAmount
		for _, ch := range h.childrenMap[id] {
			chCost := compute(ch.ChildID)
			if chCost != nil {
				contrib := int64(math.Round(float64(chCost.Amount()) * ch.SharePercentage / 100.0))
				amount += contrib
			}
		}
		visiting[id] = false

		res := money.New(amount, curr)
		memo[id] = res
		return res
	}

	for id := range h.assets {
		totalCosts[id] = compute(id)
	}

	return totalCosts
}

// ApplyToAssets updates all assets in the hierarchy with calculated valuations and links.
func (h *AssetHierarchy) ApplyToAssets(directCosts map[uuid.UUID]*money.Money, defaultCurrency string) {
	totalCosts := h.CalculateRollups(directCosts, defaultCurrency)

	for id, a := range h.assets {
		dc := directCosts[id]
		if dc == nil {
			dc = money.New(0, defaultCurrency)
		}
		tc := totalCosts[id]
		if tc == nil {
			tc = dc
		}

		children := h.childrenMap[id]
		resolvedChildren := make([]ChildAssetLink, 0, len(children))
		for _, ch := range children {
			childName := ""
			if childAsset, ok := h.assets[ch.ChildID]; ok && childAsset != nil {
				childName = childAsset.Name()
			}
			resolvedChildren = append(resolvedChildren, ChildAssetLink{
				ChildID:         ch.ChildID,
				ChildName:       childName,
				SharePercentage: ch.SharePercentage,
				TotalCost:       totalCosts[ch.ChildID],
			})
		}

		parents := h.parentsMap[id]
		resolvedParents := make([]ParentAssetLink, 0, len(parents))
		for _, p := range parents {
			parentName := ""
			if parentAsset, ok := h.assets[p.ParentID]; ok && parentAsset != nil {
				parentName = parentAsset.Name()
			}
			resolvedParents = append(resolvedParents, ParentAssetLink{
				ParentID:        p.ParentID,
				ParentName:      parentName,
				SharePercentage: p.SharePercentage,
			})
		}

		a.AssignValuation(dc, tc, resolvedChildren, resolvedParents)
	}
}

// RootAssets returns assets that have no parent assets.
func (h *AssetHierarchy) RootAssets() []*Asset {
	var roots []*Asset
	for id, a := range h.assets {
		if len(h.parentsMap[id]) == 0 {
			roots = append(roots, a)
		}
	}
	return roots
}
