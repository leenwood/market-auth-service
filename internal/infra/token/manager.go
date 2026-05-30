package token

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/leenwood/market-auth-service/internal/core/domain"
	"github.com/leenwood/market-auth-service/internal/core/port"
)

type claims struct {
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	accessTTL  time.Duration
	guestTTL   time.Duration
	jwksJSON   []byte
}

func NewManager(privatePEM, publicPEM string, accessTTL, guestTTL time.Duration) (*Manager, error) {
	priv, err := parsePrivateKey(privatePEM)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	pub, err := parsePublicKey(publicPEM)
	if err != nil {
		return nil, fmt.Errorf("parse public key: %w", err)
	}

	m := &Manager{
		privateKey: priv,
		publicKey:  pub,
		accessTTL:  accessTTL,
		guestTTL:   guestTTL,
	}

	m.jwksJSON, err = buildJWKS(pub)
	if err != nil {
		return nil, fmt.Errorf("build jwks: %w", err)
	}

	return m, nil
}

// issueToken is the shared JWT signing helper.
func (m *Manager) issueToken(userID uuid.UUID, email, role string, ttl time.Duration) (string, error) {
	now := time.Now()
	c := claims{
		Email: email,
		Role:  role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodRS256, c).SignedString(m.privateKey)
}

// IssueAccessToken implements port.TokenManager.
func (m *Manager) IssueAccessToken(userID uuid.UUID, email, role string) (string, error) {
	return m.issueToken(userID, email, role, m.accessTTL)
}

// IssueGuestToken implements port.TokenManager.
func (m *Manager) IssueGuestToken(guestID uuid.UUID) (string, error) {
	return m.issueToken(guestID, "", domain.RoleGuest, m.guestTTL)
}

// GuestTTLSeconds implements port.TokenManager.
func (m *Manager) GuestTTLSeconds() int64 {
	return int64(m.guestTTL.Seconds())
}

// IssueRefreshToken implements port.TokenManager.
func (m *Manager) IssueRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// AccessTTLSeconds implements port.TokenManager.
func (m *Manager) AccessTTLSeconds() int64 {
	return int64(m.accessTTL.Seconds())
}

// ParseAccessToken implements port.TokenParser.
func (m *Manager) ParseAccessToken(tokenStr string) (*port.ParsedToken, error) {
	t, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return m.publicKey, nil
	})
	if err != nil {
		return nil, err
	}

	c, ok := t.Claims.(*claims)
	if !ok || !t.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	return &port.ParsedToken{
		Subject: c.Subject,
		Email:   c.Email,
		Role:    c.Role,
	}, nil
}

// JWKS implements port.JWKSProvider.
func (m *Manager) JWKS() []byte {
	return m.jwksJSON
}

func parsePrivateKey(pemStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return x509.ParsePKCS1PrivateKey(block.Bytes)
	}
	rsaKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA private key")
	}
	return rsaKey, nil
}

func parsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}
	return rsaKey, nil
}

func buildJWKS(pub *rsa.PublicKey) ([]byte, error) {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())

	type jwk struct {
		Kty string `json:"kty"`
		Use string `json:"use"`
		Alg string `json:"alg"`
		Kid string `json:"kid"`
		N   string `json:"n"`
		E   string `json:"e"`
	}
	type jwks struct {
		Keys []jwk `json:"keys"`
	}

	return json.Marshal(jwks{Keys: []jwk{{
		Kty: "RSA", Use: "sig", Alg: "RS256", Kid: "1", N: n, E: e,
	}}})
}
