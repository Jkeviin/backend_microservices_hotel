package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// ClaimsKey es la clave de contexto para los claims JWT.
	ClaimsKey contextKey = "claims"
)

// JWTClaims contiene los claims extraidos del token.
type JWTClaims struct {
	UserID int    `json:"user_id"`
	Rol    string `json:"rol"`
}

// Auth retorna un middleware que valida el token JWT en el header Authorization.
// En modo desarrollo (secret = "dev-secret-change-me") acepta tokens sin
// validar firma, permitiendo el login simulado del frontend.
func Auth(secret string) func(http.Handler) http.Handler {
	devMode := secret == "dev-secret-change-me"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"token requerido"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				http.Error(w, `{"error":"formato de token invalido"}`, http.StatusUnauthorized)
				return
			}

			var claims jwt.MapClaims

			if devMode {
				parser := jwt.NewParser()
				tok, _, err := parser.ParseUnverified(parts[1], jwt.MapClaims{})
				if err != nil {
					http.Error(w, `{"error":"token invalido"}`, http.StatusUnauthorized)
					return
				}
				var ok bool
				claims, ok = tok.Claims.(jwt.MapClaims)
				if !ok {
					http.Error(w, `{"error":"claims invalidos"}`, http.StatusUnauthorized)
					return
				}
			} else {
				tok, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, jwt.ErrSignatureInvalid
					}
					return []byte(secret), nil
				})
				if err != nil || !tok.Valid {
					http.Error(w, `{"error":"token invalido"}`, http.StatusUnauthorized)
					return
				}
				var ok bool
				claims, ok = tok.Claims.(jwt.MapClaims)
				if !ok {
					http.Error(w, `{"error":"claims invalidos"}`, http.StatusUnauthorized)
					return
				}
			}

			userID := 0
			if v, exists := claims["sub"]; exists {
				switch val := v.(type) {
				case float64:
					userID = int(val)
				case string:
					// sub puede ser string en algunos JWTs
				}
			}

			rol := ""
			if v, exists := claims["role"]; exists {
				if s, ok := v.(string); ok {
					rol = s
				}
			}

			jwtClaims := JWTClaims{UserID: userID, Rol: rol}
			ctx := context.WithValue(r.Context(), ClaimsKey, jwtClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole retorna un middleware que verifica que el usuario tenga uno de los roles permitidos.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(ClaimsKey).(JWTClaims)
			if !ok {
				http.Error(w, `{"error":"no autenticado"}`, http.StatusUnauthorized)
				return
			}
			if !allowed[claims.Rol] {
				http.Error(w, `{"error":"permisos insuficientes"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
