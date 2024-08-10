package core

type AuthProvider interface {
	Login(username, password string) error
	GetAccessToken() (string, error)
	IsAuthorised() bool
}

func WithAuth(auth AuthProvider, callback func(isAuthorised bool)) {
	callback(auth.IsAuthorised())
}
