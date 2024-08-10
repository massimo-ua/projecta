package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"projecta/tui/components"
)

type model struct {
	currentPageIdx int
	pages          []tea.Model
}

func NewModel() tea.Model {
	loginPage := components.NewLoginPageModel()
	categoriesPage := components.NewCategoriesPageModel()

	loginPage.Init()
	categoriesPage.Init()

	return model{
		currentPageIdx: 0,
		pages: []tea.Model{
			loginPage,
			categoriesPage,
		},
	}
}

func (m model) Init() tea.Cmd {
	return tea.ClearScreen
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+p":
			return m.nextPage(), nil
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	}

	updatedPage, cmd := m.pages[m.currentPageIdx].Update(msg)
	m.pages[m.currentPageIdx] = updatedPage

	return m, cmd
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\nPress ctrl+p to open next page.\nPress ctrl+c or q to quit.",
		m.pages[m.currentPageIdx].View(),
	)
}

func (m model) nextPage() tea.Model {
	nextPage := m.currentPageIdx + 1
	if nextPage >= len(m.pages) {
		nextPage = 0
	}

	m.currentPageIdx = nextPage
	return m
}
