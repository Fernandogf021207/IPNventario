package attendance

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Fernandogf021207/IPNventario/core/auth"
	"github.com/Fernandogf021207/IPNventario/core/models"
	"github.com/go-chi/chi/v5"
)

// Handler contiene los endpoints HTTP para asistencia.
type Handler struct {
	Repo *Repository
}

// NewHandler crea un nuevo handler de asistencia.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{Repo: NewRepository(db)}
}

// RegisterRoutes registra las rutas de asistencia en el router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Listar asistencia de una sesión (teacher)
	r.Route("/api/sessions/{sessionID}/attendance", func(r chi.Router) {
		r.Get("/", h.HandleListBySession)
		r.Post("/", h.HandleCheckIn) // Student o Teacher pueden registrar
	})

	// Corregir asistencia (solo teacher)
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireTeacher)
		r.Put("/api/attendance/{id}", h.HandleUpdateStatus)
	})
}

// HandleListBySession - GET /api/sessions/{sessionID}/attendance
func (h *Handler) HandleListBySession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := strconv.ParseInt(chi.URLParam(r, "sessionID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID de sesión inválido."})
		return
	}

	// Si es alumno, verificar que la sesión sea de su grupo
	session := auth.GetSession(r.Context())
	if session.Role == "student" {
		groupName, err := h.Repo.GetSessionGroupName(sessionID)
		if err != nil {
			writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Sesión no encontrada."})
			return
		}
		if groupName != session.GroupName {
			writeJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Error: "No tienes acceso a esta sesión."})
			return
		}
	}

	records, err := h.Repo.ListBySession(sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando asistencia: " + err.Error(),
		})
		return
	}

	if records == nil {
		records = []models.Attendance{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: records})
}

// CheckInRequest es el cuerpo para registrar asistencia.
type CheckInRequest struct {
	StudentID int64  `json:"student_id"`
	Status    string `json:"status"` // present | late | absent | excused
	Notes     string `json:"notes"`
}

// HandleCheckIn - POST /api/sessions/{sessionID}/attendance
// RB1: Solo permite check-in en sesiones abiertas.
func (h *Handler) HandleCheckIn(w http.ResponseWriter, r *http.Request) {
	sessionID, err := strconv.ParseInt(chi.URLParam(r, "sessionID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID de sesión inválido."})
		return
	}

	var req CheckInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	session := auth.GetSession(r.Context())

	// Si es alumno, solo puede registrar su propia asistencia
	studentID := req.StudentID
	if session.Role == "student" {
		studentID = session.StudentID
		if studentID == 0 {
			writeJSON(w, http.StatusBadRequest, models.APIResponse{
				Success: false, Error: "No se pudo identificar tu registro de alumno.",
			})
			return
		}
	}

	if studentID == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "student_id es requerido.",
		})
		return
	}

	// Default status
	status := req.Status
	if status == "" {
		status = "present"
	}

	// Validar status
	validStatuses := map[string]bool{
		"present": true, "late": true, "absent": true, "excused": true,
	}
	if !validStatuses[status] {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Status inválido. Valores permitidos: present, late, absent, excused.",
		})
		return
	}

	var recordedBy sql.NullInt64
	if session.Role != "student" {
		recordedBy = sql.NullInt64{Int64: session.UserID, Valid: true}
	}

	id, err := h.Repo.CheckIn(sessionID, studentID, status, recordedBy, req.Notes)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: err.Error(),
		})
		return
	}

	record, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true, Message: "Asistencia registrada.", Data: record,
	})
}

// UpdateStatusRequest es el cuerpo para corregir asistencia.
type UpdateStatusRequest struct {
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

// HandleUpdateStatus - PUT /api/attendance/{id}
func (h *Handler) HandleUpdateStatus(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	validStatuses := map[string]bool{
		"present": true, "late": true, "absent": true, "excused": true,
	}
	if !validStatuses[req.Status] {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Status inválido. Valores permitidos: present, late, absent, excused.",
		})
		return
	}

	if err := h.Repo.UpdateStatus(id, req.Status, req.Notes); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true, Message: "Asistencia actualizada.", Data: updated,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
