package auth

import (
	"encoding/json"
	"net/http"

	"github.com/Fernandogf021207/IPNventario/core/models"
)

// Handler contiene los endpoints HTTP de autenticación.
type Handler struct {
	AuthService *Service
}

// NewHandler crea un nuevo handler de autenticación.
func NewHandler(authService *Service) *Handler {
	return &Handler{AuthService: authService}
}

// LoginRequest es el cuerpo de la petición de login.
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleLogin maneja POST /api/auth/login
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Cuerpo de la petición inválido.",
		})
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error:   "Usuario y contraseña son requeridos.",
		})
		return
	}

	session, token, err := h.AuthService.Login(req.Username, req.Password)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "Credenciales inválidas.",
		})
		return
	}

	SetSessionCookie(w, token)

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Inicio de sesión exitoso.",
		Data: models.SessionInfo{
			UserID:    session.UserID,
			Username:  session.Username,
			FullName:  session.FullName,
			Role:      session.Role,
			StudentID: session.StudentID,
			GroupName: session.GroupName,
		},
	})
}

// HandleLogout maneja POST /api/auth/logout
func (h *Handler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(CookieName)
	if err == nil {
		h.AuthService.Logout(cookie.Value)
	}
	ClearSessionCookie(w)
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Sesión cerrada.",
	})
}

// HandleMe maneja GET /api/auth/me
func (h *Handler) HandleMe(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r.Context())
	if session == nil {
		writeJSON(w, http.StatusUnauthorized, models.APIResponse{
			Success: false,
			Error:   "No autenticado.",
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: models.SessionInfo{
			UserID:    session.UserID,
			Username:  session.Username,
			FullName:  session.FullName,
			Role:      session.Role,
			StudentID: session.StudentID,
			GroupName: session.GroupName,
		},
	})
}
