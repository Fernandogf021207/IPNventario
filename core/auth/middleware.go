package auth

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Fernandogf021207/IPNventario/core/models"
)

// Claves de contexto para la sesión del usuario.
type contextKey string

const sessionContextKey contextKey = "session"

// RequireAuth es un middleware que verifica que el usuario esté autenticado.
func RequireAuth(store *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(CookieName)
			if err != nil {
				writeJSON(w, http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Error:   "Sesión no encontrada. Inicia sesión.",
				})
				return
			}

			session, ok := store.Get(cookie.Value)
			if !ok {
				ClearSessionCookie(w)
				writeJSON(w, http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Error:   "Sesión expirada. Inicia sesión nuevamente.",
				})
				return
			}

			// Guardar la sesión en el contexto
			ctx := context.WithValue(r.Context(), sessionContextKey, session)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireRole es un middleware que verifica que el usuario tenga uno de los roles permitidos.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session := GetSession(r.Context())
			if session == nil {
				writeJSON(w, http.StatusUnauthorized, models.APIResponse{
					Success: false,
					Error:   "No autenticado.",
				})
				return
			}

			for _, role := range roles {
				if session.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			writeJSON(w, http.StatusForbidden, models.APIResponse{
				Success: false,
				Error:   "No tienes permisos para realizar esta acción.",
			})
		})
	}
}

// RequireTeacher es un middleware que permite acceso solo a admin, teacher y operator.
func RequireTeacher(next http.Handler) http.Handler {
	return RequireRole("admin", "teacher", "operator")(next)
}

// RequireStudent es un middleware que permite acceso solo a estudiantes.
func RequireStudent(next http.Handler) http.Handler {
	return RequireRole("student")(next)
}

// GetSession obtiene los datos de sesión del contexto.
func GetSession(ctx context.Context) *SessionData {
	session, ok := ctx.Value(sessionContextKey).(*SessionData)
	if !ok {
		return nil
	}
	return session
}

// writeJSON es un helper para escribir respuestas JSON.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
