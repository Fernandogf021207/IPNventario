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
	r.Route("/api/requests", func(r chi.Router) {
		r.Get("/", h.HandleList)
		r.Post("/", h.HandleCreate)

		// Solo teacher/admin/operator para aprobar/rechazar/devolver
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireTeacher)
			r.Put("/{id}/approve", h.HandleApprove)
			r.Put("/{id}/reject", h.HandleReject)
			r.Put("/{id}/return", h.HandleReturn)
		})
	})
}

// HandleList - GET /api/requests
func (h *RequestHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())

	query := `
		SELECT rr.id, rr.session_id, rr.assignment_id, rr.student_id, rr.item_id,
		       rr.request_type, rr.quantity, rr.status, rr.requested_at,
		       rr.resolved_at, rr.resolved_by, rr.notes,
		       s.full_name AS student_name, s.student_code,
		       i.name AS item_name, i.sku AS item_sku
		FROM resource_requests rr
		JOIN students s ON s.id = rr.student_id
		JOIN items i ON i.id = rr.item_id
		WHERE 1=1
	`
	args := []interface{}{}

	// Alumnos solo ven sus propias solicitudes
	if session.Role == "student" {
		query += " AND rr.student_id = ?"
		args = append(args, session.StudentID)
	}

	// Filtrar por estado si se especifica
	statusFilter := r.URL.Query().Get("status")
	if statusFilter != "" {
		query += " AND rr.status = ?"
		args = append(args, statusFilter)
	}

	query += " ORDER BY rr.requested_at DESC"

	rows, err := h.DB.Query(query, args...)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando solicitudes: " + err.Error(),
		})
		return
	}
	defer rows.Close()

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
			writeJSON(w, http.StatusInternalServerError, models.APIResponse{
				Success: false, Error: "Error escaneando solicitud: " + err.Error(),
			})
			return
		}
		requests = append(requests, req)
	}

	if requests == nil {
		requests = []models.ResourceRequest{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: requests})
}

// CreateRequestBody es el cuerpo para crear una solicitud.
type CreateRequestBody struct {
	SessionID    int64   `json:"session_id"`
	AssignmentID *int64  `json:"assignment_id"`
	ItemID       int64   `json:"item_id"`
	RequestType  string  `json:"request_type"` // loan | consumption | machine_access
	Quantity     float64 `json:"quantity"`
	Notes        string  `json:"notes"`
}

// HandleCreate - POST /api/requests
func (h *RequestHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())

	var body CreateRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	// Validaciones
	if body.ItemID == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "item_id es requerido.",
		})
		return
	}
	if body.RequestType == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "request_type es requerido (loan, consumption, machine_access).",
		})
		return
	}
	if body.Quantity <= 0 {
		body.Quantity = 1
	}

	// RB3: Toda entrega vinculada a sesión + alumno
	// Verificar que la sesión exista y esté abierta
	if body.SessionID > 0 {
		var sessionStatus string
		err := h.DB.QueryRow(
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
	}

	// Determinar student_id
	studentID := session.StudentID
	if studentID == 0 {
		// Si es profesor creando para un alumno, necesitaría el student_id en el body
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Solo los alumnos pueden crear solicitudes de recursos.",
		})
		return
	}

	// Insertar solicitud
	result, err := h.DB.Exec(`
		INSERT INTO resource_requests (session_id, assignment_id, student_id, item_id, request_type, quantity, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, nilIfZero(body.SessionID), body.AssignmentID, studentID, body.ItemID, body.RequestType, body.Quantity, body.Notes)
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

// HandleApprove - PUT /api/requests/{id}/approve
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

	// Determinar tipo de transacción
	txType := "loan_out"
	if req.RequestType == "consumption" {
		txType = "consumption"
	}

	// Solo descontar stock para loan y consumption (no machine_access)
	if req.RequestType != "machine_access" {
		// RB2: Verificar stock suficiente y descontar
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

// HandleReject - PUT /api/requests/{id}/reject
func (h *RequestHandler) HandleReject(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	session := auth.GetSession(r.Context())

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

	_, err = h.DB.Exec(`
		UPDATE resource_requests
		SET status = 'rejected', resolved_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now'), resolved_by = ?
		WHERE id = ?
	`, session.UserID, id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error rechazando solicitud: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Solicitud rechazada."})
}

// HandleReturn - PUT /api/requests/{id}/return
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

	// Restaurar stock
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
