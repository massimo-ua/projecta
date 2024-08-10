package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"gitlab.com/massimo-ua/projecta/gui/core"
	"net/http"
)

type HttpProjectRepository struct {
	url  string
	auth core.AuthProvider
}

func (r *HttpProjectRepository) FindByID(ID string) (core.Project, error) {
	//TODO implement me
	panic("implement me")
}

type ProjectDTO struct {
	ID          string `json:"project_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type FindProjectsResponse struct {
	Projects []core.Project `json:"projects"`
	Total    int            `json:"total"`
}

func NewHttpProjectRepository(url string, auth core.AuthProvider) (*HttpProjectRepository, error) {
	if url == "" {
		return nil, errors.New("url cannot be empty")
	}

	if auth == nil {
		return nil, errors.New("auth provider cannot be nil")
	}

	return &HttpProjectRepository{
		url:  url,
		auth: auth,
	}, nil
}

func (r *HttpProjectRepository) Find(offset int, limit int) (projects []core.Project, total int, err error) {
	accessToken, err := r.auth.GetAccessToken()
	if err != nil {
		return nil, 0, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects?offset=%d&limit=%d", r.url, offset, limit), nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, errors.New("authentication failed")
	}

	var response FindProjectsResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return nil, 0, err
	}

	result := make([]core.Project, len(response.Projects))

	for i, project := range response.Projects {
		result[i] = core.Project{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
		}
	}

	return result, response.Total, nil
}
