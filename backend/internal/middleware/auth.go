package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/harshit110927/arnocodes/backend/internal/assessment"
)

type contextKey string

const userIDContextKey contextKey = "auth_user_id"

type JWKSAuthMiddleware struct {
	jwksURL  string
	issuer   string
	audience string
	client   *http.Client

	// Added repo to allow auto-provisioning of profiles
	repo *assessment.Repository

	mu         sync.RWMutex
	cachedKeys map[string]interface{}
	cachedAt   time.Time
}

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"` // RSA
	E   string `json:"e"` // RSA
	X   string `json:"x"` // EC
	Y   string `json:"y"` // EC
	Crv string `json:"crv"`
	Alg string `json:"alg"`
	Use string `json:"use"`
}

func NewJWKSAuthMiddleware(supabaseURL, audience string, repo *assessment.Repository) (*JWKSAuthMiddleware, error) {
	supabaseURL = strings.TrimSuffix(strings.TrimSpace(supabaseURL), "/")
	if supabaseURL == "" {
		return nil, fmt.Errorf("SUPABASE_URL is required")
	}
	if audience == "" {
		audience = "authenticated"
	}
	return &JWKSAuthMiddleware{
		jwksURL:    supabaseURL + "/auth/v1/.well-known/jwks.json",
		issuer:     supabaseURL + "/auth/v1",
		audience:   audience,
		client:     &http.Client{Timeout: 5 * time.Second},
		cachedKeys: map[string]interface{}{},
		repo:       repo, // Correctly assigning the repo here
	}, nil
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(userIDContextKey)
	uid, ok := v.(string)
	if !ok || strings.TrimSpace(uid) == "" {
		return "", false
	}
	return uid, true
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_, _ = w.Write([]byte(`{"error":"UNAUTHORIZED"}`))
}

func (m *JWKSAuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authz := strings.TrimSpace(r.Header.Get("Authorization"))
		if !strings.HasPrefix(authz, "Bearer ") {
			writeUnauthorized(w)
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(authz, "Bearer "))
		if tokenString == "" {
			writeUnauthorized(w)
			return
		}

		claims := jwt.RegisteredClaims{}
		token, err := jwt.ParseWithClaims(tokenString, &claims, m.keyFunc,
			jwt.WithIssuer(m.issuer),
			jwt.WithAudience(m.audience),
			jwt.WithValidMethods([]string{
				jwt.SigningMethodRS256.Alg(),
				jwt.SigningMethodES256.Alg(),
			}),
		)

		if err != nil || token == nil || !token.Valid {
			writeUnauthorized(w)
			return
		}
		if claims.Subject == "" {
			writeUnauthorized(w)
			return
		}

		// --- AUTO-PROVISION PROFILE ---
		// This ensures the userID exists in the 'profiles' table before the request continues.
		// This prevents foreign key violations in downstream diagnostic/test logic.
		err = m.repo.EnsureProfileExists(r.Context(), claims.Subject)
		if err != nil {
			// We log the error but allow the request to proceed; 
			// if the DB insert truly fails, downstream logic will catch it.
			fmt.Printf("Auth Hook: could not ensure profile for %s: %v\n", claims.Subject, err)
		}

		next.ServeHTTP(w, r.WithContext(WithUserID(r.Context(), claims.Subject)))
	})
}

func (m *JWKSAuthMiddleware) keyFunc(token *jwt.Token) (interface{}, error) {
	if token.Method == nil {
		return nil, errors.New("missing signing method")
	}

	alg := token.Method.Alg()
	if alg != jwt.SigningMethodRS256.Alg() && alg != jwt.SigningMethodES256.Alg() {
		return nil, errors.New("unexpected signing method")
	}

	kid, _ := token.Header["kid"].(string)
	if kid == "" {
		return nil, errors.New("missing kid")
	}

	key, err := m.getKey(kid)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func (m *JWKSAuthMiddleware) getKey(kid string) (interface{}, error) {
	m.mu.RLock()
	cachedKey, hadCachedKey := m.cachedKeys[kid]
	cacheFresh := time.Since(m.cachedAt) < 15*time.Minute
	m.mu.RUnlock()

	if cacheFresh && hadCachedKey {
		return cachedKey, nil
	}

	if err := m.refreshKeys(); err != nil {
		if hadCachedKey {
			return cachedKey, nil
		}
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()
	key, ok := m.cachedKeys[kid]
	if !ok {
		return nil, fmt.Errorf("kid not found")
	}
	return key, nil
}

func (m *JWKSAuthMiddleware) refreshKeys() error {
	resp, err := m.client.Get(m.jwksURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("jwks http status %d", resp.StatusCode)
	}

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return err
	}

	keys := make(map[string]interface{})

	for _, k := range jwks.Keys {
		switch k.Kty {
		case "RSA":
			if k.N == "" || k.E == "" {
				continue
			}
			pub, err := parseRSAPublicKey(k.N, k.E)
			if err == nil {
				keys[k.Kid] = pub
			}
		case "EC":
			if k.X == "" || k.Y == "" || k.Crv != "P-256" {
				continue
			}
			pub, err := parseECPublicKey(k.X, k.Y)
			if err == nil {
				keys[k.Kid] = pub
			}
		}
	}

	if len(keys) == 0 {
		return fmt.Errorf("no valid jwks keys")
	}

	m.mu.Lock()
	m.cachedKeys = keys
	m.cachedAt = time.Now()
	m.mu.Unlock()

	return nil
}

func parseRSAPublicKey(n, e string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(n)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(e)
	if err != nil {
		return nil, err
	}
	exponent := 0
	for _, b := range eBytes {
		exponent = exponent<<8 + int(b)
	}
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: exponent,
	}, nil
}

func parseECPublicKey(xStr, yStr string) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, err
	}
	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}