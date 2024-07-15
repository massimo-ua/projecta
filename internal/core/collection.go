package core

type PaginatedCollection[T any] struct {
	elements []T
	total    int
}

func NewPaginatedCollection[T any](total int) *PaginatedCollection[T] {
	return &PaginatedCollection[T]{
		elements: make([]T, 0),
		total:    total,
	}
}

func (l *PaginatedCollection[T]) Add(elements ...T) {
	l.elements = append(l.elements, elements...)
}

func (l *PaginatedCollection[T]) Elements() []T {
	return l.elements
}

func (l *PaginatedCollection[T]) Total() int {
	return l.total
}
