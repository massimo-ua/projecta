package components

import tea "github.com/charmbracelet/bubbletea"

type categoriesPage struct{}

func NewCategoriesPageModel() tea.Model {
	return categoriesPage{}
}

func (m categoriesPage) Init() tea.Cmd {
	return nil
}

func (m categoriesPage) Update(_ tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m categoriesPage) View() string {
	return "Categories Page"
}
