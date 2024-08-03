package people

import "errors"

type Credentials struct {
	provider       IdentityProvider
	identifier     string
	registrationID string
}

type IdentityProvider = string

const (
	GOOGLE   IdentityProvider = "GOOGLE"
	FACEBOOK IdentityProvider = "FACEBOOK"
	LOCAL    IdentityProvider = "LOCAL"
)

func ToIdentityProvider(s string) (IdentityProvider, error) {
	switch s {
	case "GOOGLE":
		return GOOGLE, nil
	case "FACEBOOK":
		return FACEBOOK, nil
	case "LOCAL":
		return LOCAL, nil
	default:
		return "", errors.New("invalid identity provider")
	}
}

func (c Credentials) Equals(other Credentials) bool {
	return c.provider == other.Provider() && c.identifier == other.Identifier()
}

func (c Credentials) Provider() IdentityProvider {
	return c.provider
}

func (c Credentials) Identifier() string {
	return c.identifier
}

func (c Credentials) RegistrationID() string {
	return c.registrationID
}

func (c Credentials) SetIdentifier(identifier string) Credentials {
	return Credentials{
		provider:       c.provider,
		identifier:     identifier,
		registrationID: c.registrationID,
	}
}

func NewCredentials(provider string, id string, identity string) (Credentials, error) {
	p, err := ToIdentityProvider(provider)
	if err != nil {
		return Credentials{}, err
	}

	if id == "" && p == LOCAL {
		return Credentials{}, errors.New("invalid credentials")
	}

	if identity == "" {
		return Credentials{}, errors.New("invalid credentials")
	}

	return Credentials{
		provider:       p,
		identifier:     identity,
		registrationID: id,
	}, nil
}
