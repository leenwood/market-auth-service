package port

import "github.com/google/uuid"

// TokenManager is used by the auth usecase to issue tokens.
type TokenManager interface {
	IssueAccessToken(userID uuid.UUID, email, role string) (string, error)
	IssueRefreshToken() (string, error)
	AccessTTLSeconds() int64
}

// ParsedToken holds the claims extracted from an access token.
type ParsedToken struct {
	Subject string
	Email   string
	Role    string
}

// TokenParser is used by the HTTP auth middleware to validate Bearer tokens.
type TokenParser interface {
	ParseAccessToken(tokenStr string) (*ParsedToken, error)
}

// JWKSProvider publishes the RSA public key for gateway verification.
type JWKSProvider interface {
	JWKS() []byte
}
