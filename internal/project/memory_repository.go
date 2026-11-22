package project

import (
	"context"
	"strings"
	"sync"

	"github.com/google/uuid"
	"gitlab.com/massimo-ua/projecta/internal/core"
	"gitlab.com/massimo-ua/projecta/internal/exceptions"
)

// MemoryProjectRepository is a thread-safe, value-copy in-memory ProjectRepository
// suitable for Domain-First exploration and fast unit/application tests.
// Storing Project values (not pointers) guarantees failed UpdateProject closures
// can never corrupt stored aggregate state.
type MemoryProjectRepository struct {
	mu       sync.RWMutex
	projects map[core.ProjectID]Project
}

func NewMemoryProjectRepository() *MemoryProjectRepository {
	return &MemoryProjectRepository{
		projects: make(map[core.ProjectID]Project),
	}
}

func cloneProjectForRequester(stored Project, requesterID core.PersonID) *Project {
	isShared := !requesterID.IsNil() && !stored.IsOwnedByPerson(requesterID) && stored.HasSharedAccess(requesterID)
	cp := ReconstituteProjectWithShares(
		stored.ID(),
		stored.Name(),
		stored.Description(),
		stored.Owner(),
		stored.StartDate(),
		stored.EndDate(),
		stored.ShareToken(),
		isShared,
		stored.MainCurrency(),
		stored.Participants(),
		stored.SharedWith(),
		stored.TagBudgets(),
	)
	return cp
}

func (m *MemoryProjectRepository) Create(_ context.Context, p *Project) error {
	if p == nil {
		return exceptions.NewValidationException("project is required", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.projects[p.ID()] = *cloneProjectForRequester(*p, core.NilPersonID)
	return nil
}

func (m *MemoryProjectRepository) Update(_ context.Context, p *Project) error {
	if p == nil {
		return exceptions.NewValidationException("project is required", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.projects[p.ID()]; !ok {
		return exceptions.NewNotFoundException("project not found", nil)
	}
	m.projects[p.ID()] = *cloneProjectForRequester(*p, core.NilPersonID)
	return nil
}

func (m *MemoryProjectRepository) GetProject(_ context.Context, projectID core.ProjectID, requesterID core.PersonID) (*Project, error) {
	if requesterID.IsNil() {
		return nil, exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	stored, ok := m.projects[projectID]
	if !ok {
		return nil, exceptions.NewNotFoundException("project not found", nil)
	}
	if err := CanPersonViewProject(requesterID, stored); err != nil {
		return nil, err
	}
	return cloneProjectForRequester(stored, requesterID), nil
}

func (m *MemoryProjectRepository) UpdateProject(
	ctx context.Context,
	projectID core.ProjectID,
	requesterID core.PersonID,
	updateFn func(ctx context.Context, p *Project) (*Project, error),
) error {
	if requesterID.IsNil() {
		return exceptions.NewUnauthorizedException("failed to identify requester", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	stored, ok := m.projects[projectID]
	if !ok {
		return exceptions.NewNotFoundException("project not found", nil)
	}
	if err := CanPersonViewProject(requesterID, stored); err != nil {
		return err
	}

	candidate := cloneProjectForRequester(stored, requesterID)
	updated, err := updateFn(ctx, candidate)
	if err != nil {
		return err
	}
	if updated == nil {
		return exceptions.NewValidationException("updated project cannot be nil", nil)
	}
	m.projects[projectID] = *cloneProjectForRequester(*updated, core.NilPersonID)
	return nil
}

func (m *MemoryProjectRepository) FindOne(_ context.Context, filter ProjectFilter) (*Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, stored := range m.projects {
		if !filter.ProjectID.IsNil() && stored.ID() != filter.ProjectID {
			continue
		}
		if filter.Name != "" && stored.Name() != filter.Name {
			continue
		}
		if !filter.RequesterID.IsNil() {
			if err := CanPersonViewProject(filter.RequesterID, stored); err != nil {
				continue
			}
		}
		return cloneProjectForRequester(stored, filter.RequesterID), nil
	}
	return nil, exceptions.NewNotFoundException("project not found", nil)
}

func (m *MemoryProjectRepository) Find(_ context.Context, filter ProjectCollectionFilter) ([]*Project, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []*Project
	for _, stored := range m.projects {
		if filter.Name != "" && !strings.Contains(stored.Name(), filter.Name) {
			continue
		}
		if !filter.RequesterID.IsNil() {
			if err := CanPersonViewProject(filter.RequesterID, stored); err != nil {
				continue
			}
		}
		matched = append(matched, cloneProjectForRequester(stored, filter.RequesterID))
	}
	if filter.Offset >= len(matched) {
		return []*Project{}, nil
	}
	start := filter.Offset
	end := len(matched)
	if filter.Limit > 0 && start+filter.Limit < end {
		end = start + filter.Limit
	}
	return matched[start:end], nil
}

func (m *MemoryProjectRepository) FindByShareToken(_ context.Context, token uuid.UUID) (*Project, error) {
	if token == uuid.Nil {
		return nil, exceptions.NewValidationException("invalid share token", nil)
	}
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, stored := range m.projects {
		if stored.ShareToken() == token {
			return cloneProjectForRequester(stored, core.NilPersonID), nil
		}
	}
	return nil, exceptions.NewNotFoundException("project not found", nil)
}

func (m *MemoryProjectRepository) Remove(_ context.Context, p *Project) error {
	if p == nil {
		return exceptions.NewValidationException("project is required", nil)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.projects, p.ID())
	return nil
}

func (m *MemoryProjectRepository) CreateShareRecord(_ context.Context, projectID core.ProjectID, personID core.PersonID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	stored, ok := m.projects[projectID]
	if !ok {
		return exceptions.NewNotFoundException("project not found", nil)
	}
	candidate := cloneProjectForRequester(stored, core.NilPersonID)
	candidate.GrantShare(personID)
	m.projects[projectID] = *cloneProjectForRequester(*candidate, core.NilPersonID)
	return nil
}
