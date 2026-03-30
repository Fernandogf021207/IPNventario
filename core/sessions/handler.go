package sessions

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Fernandogf021207/IPNventario/core/auth"
	"github.com/Fernandogf021207/IPNventario/core/models"
	"github.com/go-chi/chi/v5"
)

// Handler contiene los endpoints HTTP para sesiones de laboratorio.
type Handler struct {
	Repo *Repository
}

// NewHandler crea un nuevo handler de sesiones.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{Repo: NewRepository(db)}
}

// RegisterRoutes registra las rutas de sesiones en el router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/sessions", func(r chi.Router) {
		r.Get("/", h.HandleList)
		r.Get("/{id}", h.HandleGetByID)

		// Solo teacher/admin/operator
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireTeacher)
			r.Post("/", h.HandleCreate)
			r.Put("/{id}", h.HandleUpdate)
			r.Put("/{id}/open", h.HandleOpen)
			r.Put("/{id}/close", h.HandleClose)
			r.Put("/{id}/cancel", h.HandleCancel)
		})
	})
}

// HandleList - GET /api/sessions
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())

	var sessions []models.LabSession
	var err error

	if session.Role == "student" {
		// Alumnos ven sesiones abiertas de su grupo
		statusFilter := r.URL.Query().Get("status")
		if statusFilter == "" {
			statusFilter = "open"
		}
		sessions, err = h.Repo.List(session.GroupName, statusFilter)
	} else {
		groupFilter := r.URL.Query().Get("group")
		statusFilter := r.URL.Query().Get("status")
		sessions, err = h.Repo.List(groupFilter, statusFilter)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando sesiones: " + err.Error(),
		})
		return
	}

	if sessions == nil {
		sessions = []models.LabSession{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: sessions})
}

// HandleGetByID - GET /api/sessions/{id}
func (h *Handler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	labSession, err := h.Repo.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Sesión no encontrada."})
		return
	}

	// Si es alumno, verificar que la sesión sea de su grupo
	session := auth.GetSession(r.Context())
	if session.Role == "student" && labSession.GroupName != session.GroupName {
		writeJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Error: "No tienes acceso a esta sesión."})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: labSession})
}

// CreateSessionRequest es el cuerpo para crear una sesión.
type CreateSessionRequest struct {
	AssignmentID   int64  `json:"assignment_id"`
	Title          string `json:"title"`
	GroupName      string `json:"group_name"`
	ScheduledStart string `json:"scheduled_start"`
	ScheduledEnd   string `json:"scheduled_end"`
	Notes          string `json:"notes"`
}

// HandleCreate - POST /api/sessions
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	if req.AssignmentID == 0 || req.Title == "" || req.GroupName == "" ||
		req.ScheduledStart == "" || req.ScheduledEnd == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "assignment_id, title, group_name, scheduled_start y scheduled_end son requeridos.",
		})
		return
	}

	session := auth.GetSession(r.Context())
	labSession := &models.LabSession{
		AssignmentID:   req.AssignmentID,
		Title:          req.Title,
		TeacherID:      session.UserID,
		GroupName:      req.GroupName,
		ScheduledStart: req.ScheduledStart,
		ScheduledEnd:   req.ScheduledEnd,
	}
	labSession.Notes.String = req.Notes
	labSession.Notes.Valid = req.Notes != ""

	id, err := h.Repo.Create(labSession)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Error creando sesión: " + err.Error()})
		return
	}

	created, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Message: "Sesión creada.", Data: created})
}

// HandleUpdate - PUT /api/sessions/{id}
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	labSession := &models.LabSession{
		AssignmentID:   req.AssignmentID,
		Title:          req.Title,
		GroupName:      req.GroupName,
		ScheduledStart: req.ScheduledStart,
		ScheduledEnd:   req.ScheduledEnd,
	}
	labSession.Notes.String = req.Notes
	labSession.Notes.Valid = req.Notes != ""

	if err := h.Repo.Update(id, labSession); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Error actualizando sesión: " + err.Error()})
		return
	}

	updated, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Sesión actualizada.", Data: updated})
}

// HandleOpen - PUT /api/sessions/{id}/open
func (h *Handler) HandleOpen(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	if err := h.Repo.Open(id); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Sesión abierta.", Data: updated})
}

// HandleClose - PUT /api/sessions/{id}/close
func (h *Handler) HandleClose(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	if err := h.Repo.Close(id); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Sesión cerrada.", Data: updated})
}

// HandleCancel - PUT /api/sessions/{id}/cancel
func (h *Handler) HandleCancel(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	if err := h.Repo.Cancel(id); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Sesión cancelada.", Data: updated})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
