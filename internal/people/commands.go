package people

type RegisterCommand struct {
    Login            string
    FirstName        string
    LastName         string
    IdentityProvider IdentityProvider
    Token            string
}
