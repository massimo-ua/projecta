package web

type PaginationDTO struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type LoginDTO struct {
	ID               string `json:"id"`
	IdentityProvider string `json:"identity_provider"`
	Token            string `json:"token"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type ListProjectsResponse struct {
	Projects []ProjectDTO `json:"projects"`
	PaginationDTO
}

type ListTypesResponse struct {
	Types []TypeDTO `json:"types"`
	PaginationDTO
}

type ListCategoriesResponse struct {
	Categories []CategoryDTO `json:"categories"`
	PaginationDTO
}

type ListExpensesResponse struct {
	Expenses []ExpenseDTO `json:"expenses"`
	PaginationDTO
}

type TotalDTO struct {
	Title    string `json:"title"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
}

type ProjectTotalsDTO struct {
	Totals []TotalDTO `json:"totals"`
}

type CreateTypeDTO struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  string `json:"category_id"`
}
