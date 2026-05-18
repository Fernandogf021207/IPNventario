package inventory

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

// RequestHandler contiene los endpoints HTTP para solicitudes de recursos.
type RequestHandler struct {
	DB   *sql.DB
	Repo *Repository
}

// NewRequestHandler crea un nuevo handler de solicitudes.
func NewRequestHandler(db *sql.DB) *RequestHandler {
	return &RequestHandler{DB: db, Repo: NewRepository(db)}
}

// RegisterRoutes registra las rutas de solicitudes en el router.
func (h *RequestHandler) RegisterRoutes(r chi.Router) {
	// Rutas scoped por sesión
	r.Route("/api/sessions/{sessionID}/requests", func(r chi.Router) {
		r.Get("/", h.HandleListForSession)
		r.Post("/", h.HandleCreateForSession)
	})

	// Rutas de acción sobre solicitudes individuales (solo teacher/admin/operator)
	r.Route("/api/requests", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireTeacher)
			r.Put("/{id}/approve", h.HandleApprove)
			r.Put("/{id}/reject", h.HandleReject)
			r.Put("/{id}/return", h.HandleReturn)
		})
	})
}

// ========================================================================
// SESSION-SCOPED HANDLERS
// ========================================================================

// HandleListForSession - GET /api/sessions/{sessionID}/requests
// Teacher/operator views all requests for a session.
func (h *RequestHandler) HandleListForSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := strconv.ParseInt(chi.URLParam(r, "sessionID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID de sesión inválido."})
		return
	}

	query := `
		SELECT rr.id, rr.session_id, rr.assignment_id, rr.student_id, rr.item_id,
		       rr.request_type, rr.quantity, rr.status, rr.requested_at,
		       rr.resolved_at, rr.resolved_by, rr.notes,
		       s.full_name AS student_name, s.student_code,
		       i.name AS item_name, i.sku AS item_sku
		FROM resource_requests rr
		JOIN students s ON s.id = rr.student_id
		JOIN items i ON i.id = rr.item_id
		WHERE rr.session_id = ?
		ORDER BY rr.requested_at DESC
	`

	rows, err := h.DB.Query(query, sessionID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando solicitudes: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	requests := h.scanRequests(rows)
	if requests == nil {
		requests = []models.ResourceRequest{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: requests})
}

// CreateRequestBody es el cuerpo para crear una solicitud.
type CreateRequestBody struct {
	SessionID    int64   `json:"session_id"`
	AssignmentID *int64  `json:"assignment_id"`
	StudentID    int64   `json:"student_id"`
	ItemID       int64   `json:"item_id"`
	RequestType  string  `json:"request_type"` // loan | consumption | machine_access
	Quantity     float64 `json:"quantity"`
	Notes        string  `json:"notes"`
}

// HandleCreateForSession - POST /api/sessions/{sessionID}/requests
// Student submits a request from within an open session.
func (h *RequestHandler) HandleCreateForSession(w http.ResponseWriter, r *http.Request) {
	sessionID, err := strconv.ParseInt(chi.URLParam(r, "sessionID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID de sesión inválido."})
		return
	}

	session := auth.GetSession(r.Context())

	var body CreateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	// Override session_id from URL
	body.SessionID = sessionID

	// Determine student_id: from body (for teacher creating on behalf) or from session (for student)
	studentID := body.StudentID
	if studentID == 0 {
		studentID = session.StudentID
	}

	// RB3: session_id y student_id son obligatorios
	if body.SessionID == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "session_id es requerido para crear una solicitud.",
		})
		return
	}
	if studentID == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "student_id es requerido. Especifícalo en el cuerpo o inicia sesión como alumno.",
		})
		return
	}

	// Validaciones de campos
	if body.ItemID == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "item_id es requerido.",
		})
		return
	}
	if body.RequestType != "loan" && body.RequestType != "consumption" && body.RequestType != "machine_access" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "request_type debe ser 'loan', 'consumption' o 'machine_access'.",
		})
		return
	}
	if body.Quantity <= 0 {
		body.Quantity = 1
	}

	// Verificar que la sesión exista y esté abierta
	var sessionStatus string
	err = h.DB.QueryRow(
		"SELECT status FROM lab_sessions WHERE id = ?", body.SessionID,
	).Scan(&sessionStatus)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Sesión no encontrada.",
		})
		return
	}
	if sessionStatus != "open" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Solo se pueden solicitar recursos en sesiones abiertas.",
		})
		return
	}

	// Verificar que el alumno exista
	var studentExists int
	h.DB.QueryRow("SELECT COUNT(*) FROM students WHERE id = ? AND is_active = 1", studentID).Scan(&studentExists)
	if studentExists == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Alumno no encontrado.",
		})
		return
	}

	// Verificar que el item exista y esté activo
	var itemExists int
	h.DB.QueryRow("SELECT COUNT(*) FROM items WHERE id = ? AND is_active = 1", body.ItemID).Scan(&itemExists)
	if itemExists == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Item no encontrado.",
		})
		return
	}

	// Insertar solicitud
	var notesVal interface{}
	if body.Notes != "" {
		notesVal = body.Notes
	}

	result, err := h.DB.Exec(`
		INSERT INTO resource_requests (session_id, assignment_id, student_id, item_id, request_type, quantity, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, body.SessionID, body.AssignmentID, studentID, body.ItemID, body.RequestType, body.Quantity, notesVal)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error creando solicitud: " + err.Error(),
		})
		return
	}

	id, _ := result.LastInsertId()
	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true, Message: "Solicitud creada.",
		Data: map[string]int64{"id": id},
	})
}

// ========================================================================
// APPROVE / REJECT / RETURN
// ========================================================================

// HandleApprove - PUT /api/requests/{id}/approve
// Teacher/operator approves a pending request. Triggers stock change or equipment_usage.
func (h *RequestHandler) HandleApprove(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	session := auth.GetSession(r.Context())

	// Obtener la solicitud
	var req models.ResourceRequest
	err = h.DB.QueryRow(`
		SELECT id, session_id, student_id, item_id, request_type, quantity, status
		FROM resource_requests WHERE id = ?
	`, id).Scan(&req.ID, &req.SessionID, &req.StudentID, &req.ItemID, &req.RequestType, &req.Quantity, &req.Status)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Solicitud no encontrada."})
		return
	}

	if req.Status != "pending" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: fmt.Sprintf("Solo se pueden aprobar solicitudes pendientes. Estado actual: %s", req.Status),
		})
		return
	}

	// RB3: Verificar que session_id y student_id estén presentes
	if !req.SessionID.Valid || req.SessionID.Int64 == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "No se puede aprobar: la solicitud no tiene una sesión vinculada (RB3).",
		})
		return
	}
	if req.StudentID == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "No se puede aprobar: la solicitud no tiene un alumno vinculado (RB3).",
		})
		return
	}

	if req.RequestType == "machine_access" {
		// Para machine_access: no modificar stock, abrir un equipment_usage
		err = h.createEquipmentUsage(req.SessionID.Int64, req.ItemID, req.StudentID, session.UserID)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
	} else {
		// Para loan y consumption: descontar stock
		txType := "loan_out"
		if req.RequestType == "consumption" {
			txType = "consumption"
		}

		sessionIDPtr := nilInt64(req.SessionID)
		studentIDPtr := &req.StudentID
		refType := "resource_request"
		err = h.Repo.AdjustStock(req.ItemID, -req.Quantity, txType, session.UserID,
			sessionIDPtr, studentIDPtr, &refType, &id, "Aprobación de solicitud")
		if err != nil {
			writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
			return
		}
	}

	// Actualizar estado de la solicitud
	_, err = h.DB.Exec(`
		UPDATE resource_requests
		SET status = 'approved', resolved_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), resolved_by = ?
		WHERE id = ?
	`, session.UserID, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error aprobando solicitud: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Solicitud aprobada."})
}

// createEquipmentUsage abre un registro de uso de maquinaria desde una solicitud aprobada.
// Valida que el item sea una máquina, que no esté en estado crítico/fuera de servicio (RB8),
// y que no haya un uso activo del mismo equipo en la misma sesión.
func (h *RequestHandler) createEquipmentUsage(sessionID, itemID, studentID, supervisorID int64) error {
	// Verificar que el item sea una máquina
	var itemType, maintenanceStatus string
	err := h.DB.QueryRow(
		"SELECT item_type, maintenance_status FROM items WHERE id = ? AND is_active = 1",
		itemID,
	).Scan(&itemType, &maintenanceStatus)
	if err != nil {
		return fmt.Errorf("máquina no encontrada")
	}
	if itemType != "machine" {
		return fmt.Errorf("el item seleccionado no es una máquina")
	}

	// RB8: No permitir uso si está en estado crítico o fuera de servicio
	if maintenanceStatus == "critical" || maintenanceStatus == "out_of_service" {
		return fmt.Errorf("la máquina no está disponible. Estado de mantenimiento: %s", maintenanceStatus)
	}

	// Verificar que no haya un uso activo del mismo equipo en la misma sesión
	var activeCount int
	h.DB.QueryRow(
		"SELECT COUNT(*) FROM equipment_usage WHERE item_id = ? AND session_id = ? AND status = 'active'",
		itemID, sessionID,
	).Scan(&activeCount)
	if activeCount > 0 {
		return fmt.Errorf("esta máquina ya tiene un uso activo en esta sesión. Ciérralo primero")
	}

	// Insertar registro de uso
	_, err = h.DB.Exec(`
		INSERT INTO equipment_usage (session_id, item_id, student_id, supervisor_id)
		VALUES (?, ?, ?, ?)
	`, sessionID, itemID, studentID, supervisorID)
	if err != nil {
		return fmt.Errorf("error creando registro de uso de maquinaria: %w", err)
	}

	return nil
}

// RejectBody es el cuerpo opcional para rechazar una solicitud.
type RejectBody struct {
	Notes string `json:"notes"`
}

// HandleReject - PUT /api/requests/{id}/reject
func (h *RequestHandler) HandleReject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	session := auth.GetSession(r.Context())

	// Leer notas opcionales
	var body RejectBody
	_ = json.NewDecoder(r.Body).Decode(&body)

	// Verificar que esté pendiente
	var status string
	err = h.DB.QueryRow("SELECT status FROM resource_requests WHERE id = ?", id).Scan(&status)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Solicitud no encontrada."})
		return
	}
	if status != "pending" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Solo se pueden rechazar solicitudes pendientes.",
		})
		return
	}

	// Actualizar estado y notas
	query := `
		UPDATE resource_requests
		SET status = 'rejected', resolved_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), resolved_by = ?
	`
	args := []interface{}{session.UserID}

	if body.Notes != "" {
		query += `, notes = ?`
		args = append(args, body.Notes)
	}

	query += ` WHERE id = ?`
	args = append(args, id)

	_, err = h.DB.Exec(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error rechazando solicitud: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Solicitud rechazada."})
}

// HandleReturn - PUT /api/requests/{id}/return
// Marks a loan as returned and restores stock.
func (h *RequestHandler) HandleReturn(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	session := auth.GetSession(r.Context())

	// Obtener la solicitud
	var req models.ResourceRequest
	err = h.DB.QueryRow(`
		SELECT id, session_id, student_id, item_id, request_type, quantity, status
		FROM resource_requests WHERE id = ?
	`, id).Scan(&req.ID, &req.SessionID, &req.StudentID, &req.ItemID, &req.RequestType, &req.Quantity, &req.Status)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Solicitud no encontrada."})
		return
	}

	if req.Status != "approved" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Solo se pueden devolver solicitudes aprobadas (préstamos).",
		})
		return
	}

	if req.RequestType != "loan" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Solo los préstamos pueden ser devueltos.",
		})
		return
	}

	// Restaurar stock dentro de una transacción SQL
	sessionIDPtr := nilInt64(req.SessionID)
	studentIDPtr := &req.StudentID
	refType := "resource_request"
	err = h.Repo.AdjustStock(req.ItemID, req.Quantity, "loan_return", session.UserID,
		sessionIDPtr, studentIDPtr, &refType, &id, "Devolución de préstamo")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error restaurando stock: " + err.Error(),
		})
		return
	}

	// Actualizar estado
	_, err = h.DB.Exec(`
		UPDATE resource_requests
		SET status = 'returned', resolved_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), resolved_by = ?
		WHERE id = ?
	`, session.UserID, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error actualizando solicitud: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Préstamo devuelto y stock restaurado."})
}

// ========================================================================
// HELPERS
// ========================================================================

// scanRequests escanea filas de resource_requests con JOINs.
func (h *RequestHandler) scanRequests(rows *sql.Rows) []models.ResourceRequest {
	var requests []models.ResourceRequest
	for rows.Next() {
		var req models.ResourceRequest
		err := rows.Scan(
			&req.ID, &req.SessionID, &req.AssignmentID, &req.StudentID, &req.ItemID,
			&req.RequestType, &req.Quantity, &req.Status, &req.RequestedAt,
			&req.ResolvedAt, &req.ResolvedBy, &req.Notes,
			&req.StudentName, &req.StudentCode,
			&req.ItemName, &req.ItemSKU,
		)
		if err != nil {
			continue
		}
		requests = append(requests, req)
	}
	return requests
}

func nilIfZero(v int64) interface{} {
	if v == 0 {
		return nil
	}
	return v
}

func nilInt64(v sql.NullInt64) *int64 {
	if v.Valid {
		return &v.Int64
	}
	return nil
}
