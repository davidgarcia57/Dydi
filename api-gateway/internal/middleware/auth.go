package middleware

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type jwksResponse struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

var jwksCache = struct {
	sync.RWMutex
	keys      map[string]*ecdsa.PublicKey
	expiresAt time.Time
}{keys: map[string]*ecdsa.PublicKey{}}

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenStr string
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			tokenStr = strings.TrimPrefix(header, "Bearer ")
		} else if strings.HasPrefix(r.URL.Path, "/ws") {
			// Browsers can't set headers on a WebSocket handshake, so the token
			// is only accepted via query string on the /ws route — never on REST
			// routes, where it would leak into access logs and history.
			tokenStr = r.URL.Query().Get("token")
		}

		// Supabase signs with ES256 (verified via JWKS). Restricting to ES256
		// closes the algorithm-confusion surface that HS256 would open.
		token, err := jwt.Parse(tokenStr, signingKey, jwt.WithValidMethods([]string{"ES256"}))
		if err != nil || !token.Valid {
			http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error":"invalid claims"}`, http.StatusUnauthorized)
			return
		}

		userID, _ := claims["sub"].(string)
		ctx := context.WithValue(r.Context(), "userID", userID)
		r.Header.Set("X-User-ID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func signingKey(token *jwt.Token) (interface{}, error) {
	switch token.Method.Alg() {
	case jwt.SigningMethodHS256.Alg():
		secret := os.Getenv("SUPABASE_JWT_SECRET")
		if secret == "" {
			return nil, errors.New("missing SUPABASE_JWT_SECRET")
		}
		return []byte(secret), nil

	case jwt.SigningMethodES256.Alg():
		kid, _ := token.Header["kid"].(string)
		if kid == "" {
			return nil, errors.New("missing JWT kid")
		}
		return jwkPublicKey(kid)

	default:
		return nil, errors.New("unsupported JWT signing method")
	}
}

func jwkPublicKey(kid string) (*ecdsa.PublicKey, error) {
	jwksCache.RLock()
	key, ok := jwksCache.keys[kid]
	expiresAt := jwksCache.expiresAt
	jwksCache.RUnlock()
	if ok && time.Now().Before(expiresAt) {
		return key, nil
	}

	return refreshJwks(kid)
}

func refreshJwks(kid string) (*ecdsa.PublicKey, error) {
	jwksURL := os.Getenv("SUPABASE_JWKS_URL")
	if jwksURL == "" {
		return nil, errors.New("missing SUPABASE_JWKS_URL")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("could not fetch JWKS")
	}

	var payload jwksResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		return nil, err
	}

	keys := make(map[string]*ecdsa.PublicKey, len(payload.Keys))
	for _, jwk := range payload.Keys {
		key, err := jwk.toECDSAPublicKey()
		if err != nil {
			continue
		}
		keys[jwk.Kid] = key
	}

	jwksCache.Lock()
	jwksCache.keys = keys
	jwksCache.expiresAt = time.Now().Add(5 * time.Minute)
	key := jwksCache.keys[kid]
	jwksCache.Unlock()

	if key == nil {
		return nil, errors.New("JWT kid not found in JWKS")
	}
	return key, nil
}

func (j jwkKey) toECDSAPublicKey() (*ecdsa.PublicKey, error) {
	if j.Kty != "EC" || j.Crv != "P-256" || j.X == "" || j.Y == "" {
		return nil, errors.New("unsupported JWK")
	}

	xBytes, err := base64.RawURLEncoding.DecodeString(j.X)
	if err != nil {
		return nil, err
	}
	yBytes, err := base64.RawURLEncoding.DecodeString(j.Y)
	if err != nil {
		return nil, err
	}

	return &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
	}, nil
}
