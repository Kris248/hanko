package session

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	hankoJwk "github.com/teamhanko/hanko/crypto/jwk"
	hankoJwt "github.com/teamhanko/hanko/crypto/jwt"
	"time"
)

type Manager interface {
	Generate(uuid.UUID) (string, error)
	Verify(string) (jwt.Token, error)
}

// Manager is used to create and verify session JWTs
type manager struct {
	jwtGenerator  hankoJwt.Generator
	sessionLength time.Duration
}

// NewManager returns a new Manager which will be used to create and verify sessions JWTs
func NewManager(jwkManager hankoJwk.Manager) (Manager, error) {
	signatureKey, err := jwkManager.GetSigningKey()
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	verificationKeys, err := jwkManager.GetPublicKeys()
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	g, err := hankoJwt.NewGenerator(signatureKey, verificationKeys)
	if err != nil {
		return nil, fmt.Errorf("failed to create session generator: %w", err)
	}
	return &manager{
		jwtGenerator:  g,
		sessionLength: time.Minute * 60, // TODO: should come from config
	}, nil
}

// Generate creates a new session JWT for the given user
func (g *manager) Generate(userId uuid.UUID) (string, error) {
	issuedAt := time.Now()
	expiration := issuedAt.Add(g.sessionLength)

	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, userId.String())
	_ = token.Set(jwt.IssuedAtKey, issuedAt)
	_ = token.Set(jwt.ExpirationKey, expiration)
	//_ = token.Set(jwt.AudienceKey, []string{"http://localhost"})

	signed, err := g.jwtGenerator.Sign(token)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}

// Verify verifies the given JWT and returns a parsed one if verification was successful
func (g *manager) Verify(token string) (jwt.Token, error) {
	parsedToken, err := g.jwtGenerator.Verify([]byte(token))
	if err != nil {
		return nil, fmt.Errorf("failed to verify session token: %w", err)
	}

	return parsedToken, nil
}