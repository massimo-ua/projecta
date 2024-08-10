package main

import (
	"errors"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"gitlab.com/massimo-ua/projecta/gui/api"
	"gitlab.com/massimo-ua/projecta/gui/core"
	"gitlab.com/massimo-ua/projecta/gui/ui"
)

const (
	baseURL = "http://127.0.0.1:8000"
)

func main() {
	auth, err := api.NewHttpAuthProvider(baseURL)

	if err != nil {
		panic(errors.New("failed to initiate auth provider"))
	}

	projectRepository, err := api.NewHttpProjectRepository(baseURL, auth)

	if err != nil {
		panic(errors.New("failed to initiate project repository"))
	}

	a := app.New()
	w := a.NewWindow("Project Manager")

	showProjectDetails := func(projectID string) {
		// Navigate to project details page
		//detailsPage := ui.ProjectDetailsPage(auth, w, projectID)
		//w.SetContent(detailsPage)
	}

	showProjects := func() {
		core.WithAuth(auth, func(isAuthorised bool) {
			projectsPage := ui.ProjectsPage(projectRepository, w, showProjectDetails)
			w.SetContent(projectsPage)
		})
	}

	authPage := ui.AuthPage(auth, w, showProjects)
	w.SetContent(authPage)

	w.Resize(fyne.NewSize(800, 600))
	w.ShowAndRun()
}
