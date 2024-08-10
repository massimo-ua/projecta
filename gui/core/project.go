package core

type Project struct {
	ID          string
	Name        string
	Description string
}

type Category struct {
	ID          string
	Name        string
	Description string
}

type Type struct {
	ID          string
	Name        string
	Description string
	Category    Category
}

type Expense struct {
	ID string
}

type ProjectRepository interface {
	Find(offset int, limit int) (projects []Project, total int, err error)
	FindByID(ID string) (Project, error)
}
