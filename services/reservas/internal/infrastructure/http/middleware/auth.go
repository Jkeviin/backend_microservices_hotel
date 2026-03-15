package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// ContextKeyUserID es la clave de contexto para el ID del usuario autenticado.
	ContextKeyUserID contextKey = "user_id"
	// ContextKeyRole es la clave de contexto para el rol del usuario autenticado.
	ContextKeyRole contextKey = "role"
)

// UserIDFromContext extrae el user_id del contexto.
func UserIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyUserID).(string)
	return v
}

// RoleFromContext extrae el role del contexto.
func RoleFromContext(ctx context.Context) string {
	v, _ := ctx.Value(ContextKeyRole).(string)
	return v
}

// JWTAuth valida el token JWT del header Authorization y coloca sub (user_id)
// y role en el contexto de la request.
// En modo desarrollo (secret = "dev-secret-change-me") acepta tokens sin
// validar firma, permitiendo el login simulado del frontend.
func JWTAuth(secret string) func(http.Handler) http.Handler {
	devMode := secret == "dev-secret-change-me"

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"token requerido"}`, http.StatusUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				http.Error(w, `{"error":"formato de token invalido"}`, http.StatusUnauthorized)
				return
			}

			tokenStr := parts[1]
			var claims jwt.MapClaims

			if devMode {
				parser := jwt.NewParser()
				tok, _, err := parser.ParseUnverified(tokenStr, jwt.MapClaims{})
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
				tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
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

			sub, _ := claims["sub"].(string)
			role, _ := claims["role"].(string)

			ctx := context.WithValue(r.Context(), ContextKeyUserID, sub)
			ctx = context.WithValue(ctx, ContextKeyRole, role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole verifica que el usuario tenga uno de los roles permitidos.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := RoleFromContext(r.Context())
			if _, ok := allowed[role]; !ok {
				http.Error(w, `{"error":"permiso denegado"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r.WithContext(r.Context()))
		})
	}
}
