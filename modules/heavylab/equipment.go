package heavylab

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/Fernandogf021207/IPNventario/core/auth"
	"github.com/Fernandogf021207/IPNventario/core/models"
	"github.com/go-chi/chi/v5"
)

// EquipmentHandler contiene los endpoints HTTP para uso de maquinaria.
type EquipmentHandler struct {
	DB *sql.DB
}

// NewEquipmentHandler crea un nuevo handler de uso de maquinaria.
func NewEquipmentHandler(db *sql.DB) *EquipmentHandler {
	return &EquipmentHandler{DB: db}
}

// RegisterRoutes registra las rutas de uso de maquinaria en el router.
func (h *EquipmentHandler) RegisterRoutes(r chi.Router) {
	r.Route("/api/equipment-usage", func(r chi.Router) {
		r.Get("/", h.HandleList)

		// Solo teacher/admin/operator para crear y cerrar
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireTeacher)
			r.Post("/", h.HandleCreate)
			r.Put("/{id}/end", h.HandleEnd)
		})
	})
}

// HandleList - GET /api/equipment-usage
func (h *EquipmentHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())

	query := `
		SELECT eu.id, eu.session_id, eu.item_id, eu.student_id, eu.supervisor_id,
		       eu.started_at, eu.ended_at, eu.status, eu.notes, eu.created_at,
		       i.name AS item_name,
		       s.full_name AS student_name,
		       u.full_name AS supervisor_name
		FROM equipment_usage eu
		JOIN items i ON i.id = eu.item_id
		JOIN students s ON s.id = eu.student_id
		JOIN users u ON u.id = eu.supervisor_id
		WHERE 1=1
	`
	args := []interface{}{}

	// Alumnos solo ven sus propios registros activos
	if session.Role == "student" {
		query += " AND eu.student_id = ?"
		args = append(args, session.StudentID)
	}

	// Filtros opcionales
	statusFilter := r.URL.Query().Get("status")
	if statusFilter != "" {
		query += " AND eu.status = ?"
		args = append(args, statusFilter)
	}

	sessionFilter := r.URL.Query().Get("session_id")
	if sessionFilter != "" {
		query += " AND eu.session_id = ?"
		args = append(args, sessionFilter)
	}

	query += " ORDER BY eu.started_at DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando uso de maquinaria: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var usages []models.EquipmentUsage
	for rows.Next() {
		var u models.EquipmentUsage
		err := rows.Scan(
			&u.ID, &u.SessionID, &u.ItemID, &u.StudentID, &u.SupervisorID,
			&u.StartedAt, &u.EndedAt, &u.Status, &u.Notes, &u.CreatedAt,
			&u.ItemName, &u.StudentName, &u.SupervisorName,
		)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.APIResponse{
				Success: false, Error: "Error escaneando registro: " + err.Error(),
			})
			return
		}
		usages = append(usages, u)
	}

	if usages == nil {
		usages = []models.EquipmentUsage{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: usages})
}

// CreateUsageRequest es el cuerpo para abrir un registro de uso.
type CreateUsageRequest struct {
	SessionID int64  `json:"session_id"`
	ItemID    int64  `json:"item_id"`
	StudentID int64  `json:"student_id"`
	Notes     string `json:"notes"`
}

// HandleCreate - POST /api/equipment-usage
func (h *EquipmentHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())

	var body CreateUsageRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	// Validaciones
	if body.SessionID == 0 || body.ItemID == 0 || body.StudentID == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "session_id, item_id y student_id son requeridos.",
		})
		return
	}

	// Verificar que la sesión esté abierta (RB4)
	var sessionStatus string
	err := h.DB.QueryRow("SELECT status FROM lab_sessions WHERE id = ?", body.SessionID).Scan(&sessionStatus)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Sesión no encontrada.",
		})
		return
	}
	if sessionStatus != "open" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Solo se puede registrar uso de maquinaria en sesiones abiertas.",
		})
		return
	}

	// Verificar que el item sea una máquina
	var itemType string
	var maintenanceStatus string
	err = h.DB.QueryRow(
		"SELECT item_type, maintenance_status FROM items WHERE id = ? AND is_active = 1",
		body.ItemID,
	).Scan(&itemType, &maintenanceStatus)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Máquina no encontrada.",
		})
		return
	}
	if itemType != "machine" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "El item seleccionado no es una máquina.",
		})
		return
	}

	// RB8: No permitir uso si está en estado crítico o fuera de servicio
	if maintenanceStatus == "critical" || maintenanceStatus == "out_of_service" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false,
			Error: fmt.Sprintf(
				"La máquina no está disponible. Estado de mantenimiento: %s",
				maintenanceStatus,
			),
		})
		return
	}

	// Verificar que no haya un uso activo del mismo equipo
	var activeCount int
	h.DB.QueryRow(
		"SELECT COUNT(*) FROM equipment_usage WHERE item_id = ? AND status = 'active'",
		body.ItemID,
	).Scan(&activeCount)
	if activeCount > 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Esta máquina ya tiene un uso activo. Ciérralo primero.",
		})
		return
	}

	// Verificar que el alumno exista
	var studentExists int
	h.DB.QueryRow("SELECT COUNT(*) FROM students WHERE id = ? AND is_active = 1", body.StudentID).Scan(&studentExists)
	if studentExists == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Alumno no encontrado.",
		})
		return
	}

	// Insertar registro de uso
	result, err := h.DB.Exec(`
		INSERT INTO equipment_usage (session_id, item_id, student_id, supervisor_id, notes)
		VALUES (?, ?, ?, ?, ?)
	`, body.SessionID, body.ItemID, body.StudentID, session.UserID, body.Notes)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error registrando uso: " + err.Error(),
		})
		return
	}

	id, _ := result.LastInsertId()
	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true, Message: "Uso de maquinaria registrado.",
		Data: map[string]int64{"id": id},
	})
}

// HandleEnd - PUT /api/equipment-usage/{id}/end
func (h *EquipmentHandler) HandleEnd(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	// Verificar que esté activo
	var status string
	err = h.DB.QueryRow("SELECT status FROM equipment_usage WHERE id = ?", id).Scan(&status)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Registro de uso no encontrado."})
		return
	}
	if status != "active" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Solo se pueden finalizar usos activos.",
		})
		return
	}

	// Leer body opcional para status alternativo (interrupted)
	var body struct {
		Status string `json:"status"` // completed | interrupted
		Notes  string `json:"notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	endStatus := "completed"
	if body.Status == "interrupted" {
		endStatus = "interrupted"
	}

	_, err = h.DB.Exec(`
		UPDATE equipment_usage
		SET ended_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'),
		    status = ?,
		    notes = CASE WHEN ? != '' THEN ? ELSE notes END
		WHERE id = ?
	`, endStatus, body.Notes, body.Notes, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error finalizando uso: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true, Message: "Uso de maquinaria finalizado.",
	})
}

// ========================================================================
// HELPERS
// ========================================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
