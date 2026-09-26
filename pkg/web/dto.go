package web

import "time"

type PaginationDTO struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

type LoginDTO struct {
	ID               string `json:"id"`
	IdentityProvider string `json:"identity_provider"`
	Token            string `json:"token"`
	InvitationCode   string `json:"invitation_code,omitempty"`
}

type CreateInvitationDTO struct {
	Email string `json:"email"`
}

type InvitationDTO struct {
	ID          string     `json:"id"`
	Email       string     `json:"email"`
	Status      string     `json:"status"`
	ExpiresAt   time.Time  `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type InvitationCreatedResponseDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

type ValidateInvitationResponseDTO struct {
	Email     string    `json:"email"`
	ExpiresAt time.Time `json:"expires_at"`
}

type ListInvitationsResponse struct {
	Invitations []InvitationDTO `json:"invitations"`
}

type RefreshTokenDTO struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

type AssignRolesDTO struct {
	Roles []string `json:"roles"`
}

type ListUsersResponse struct {
	Users []UserDTO `json:"users"`
	PaginationDTO
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

type ListPaymentsResponse struct {
	Payments []PaymentDTO `json:"payments"`
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
