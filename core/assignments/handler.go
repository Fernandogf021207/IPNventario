package assignments

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Fernandogf021207/IPNventario/core/auth"
	"github.com/Fernandogf021207/IPNventario/core/models"
	"github.com/go-chi/chi/v5"
)

// Handler contiene los endpoints HTTP para prácticas.
type Handler struct {
	Repo *Repository
}

// NewHandler crea un nuevo handler de prácticas.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{Repo: NewRepository(db)}
}

// RegisterRoutes registra las rutas de prácticas en el router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/assignments", func(r chi.Router) {
		r.Get("/", h.HandleList)
		r.Get("/{id}", h.HandleGetByID)

		// Solo teacher/admin/operator
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireTeacher)
			r.Post("/", h.HandleCreate)
			r.Put("/{id}", h.HandleUpdate)
			r.Put("/{id}/publish", h.HandlePublish)
			r.Put("/{id}/close", h.HandleClose)
			// Assignment items
			r.Get("/{id}/items", h.HandleListItems)
			r.Post("/{id}/items", h.HandleAddItem)
			r.Delete("/{id}/items/{itemID}", h.HandleRemoveItem)
		})
	})
}

// HandleList - GET /api/assignments
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	session := auth.GetSession(r.Context())

	var assignments []models.Assignment
	var err error

	if session.Role == "student" {
		// Alumnos solo ven prácticas publicadas de su grupo
		assignments, err = h.Repo.List(session.GroupName, true)
	} else {
		// Profesores ven todas
		groupFilter := r.URL.Query().Get("group")
		assignments, err = h.Repo.List(groupFilter, false)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando prácticas: " + err.Error(),
		})
		return
	}

	if assignments == nil {
		assignments = []models.Assignment{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: assignments})
}

// HandleGetByID - GET /api/assignments/{id}
func (h *Handler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	assignment, err := h.Repo.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Práctica no encontrada."})
		return
	}

	// Si es alumno, verificar que la práctica sea de su grupo y esté publicada
	session := auth.GetSession(r.Context())
	if session.Role == "student" {
		if assignment.GroupName != session.GroupName || assignment.Status != "published" {
			writeJSON(w, http.StatusForbidden, models.APIResponse{Success: false, Error: "No tienes acceso a esta práctica."})
			return
		}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: assignment})
}

// CreateAssignmentRequest es el cuerpo para crear una práctica.
type CreateAssignmentRequest struct {
	Title         string `json:"title"`
	Description   string `json:"description"`
	Instructions  string `json:"instructions"`
	GroupName     string `json:"group_name"`
	ScheduledDate string `json:"scheduled_date"`
}

// HandleCreate - POST /api/assignments
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	if req.Title == "" || req.GroupName == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Título y grupo son requeridos."})
		return
	}

	session := auth.GetSession(r.Context())
	assignment := &models.Assignment{
		Title:       req.Title,
		GroupName:   req.GroupName,
		TeacherID:   session.UserID,
	}
	assignment.Description.String = req.Description
	assignment.Description.Valid = req.Description != ""
	assignment.Instructions.String = req.Instructions
	assignment.Instructions.Valid = req.Instructions != ""
	assignment.ScheduledDate.String = req.ScheduledDate
	assignment.ScheduledDate.Valid = req.ScheduledDate != ""

	id, err := h.Repo.Create(assignment)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: "Error creando práctica: " + err.Error()})
		return
	}

	created, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Message: "Práctica creada.", Data: created})
}

// HandleUpdate - PUT /api/assignments/{id}
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var req CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	assignment := &models.Assignment{
		Title:     req.Title,
		GroupName: req.GroupName,
	}
	assignment.Description.String = req.Description
	assignment.Description.Valid = req.Description != ""
	assignment.Instructions.String = req.Instructions
	assignment.Instructions.Valid = req.Instructions != ""
	assignment.ScheduledDate.String = req.ScheduledDate
	assignment.ScheduledDate.Valid = req.ScheduledDate != ""

	if err := h.Repo.Update(id, assignment); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Práctica actualizada.", Data: updated})
}

// HandlePublish - PUT /api/assignments/{id}/publish
func (h *Handler) HandlePublish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	if err := h.Repo.Publish(id); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Práctica publicada.", Data: updated})
}

// HandleClose - PUT /api/assignments/{id}/close
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
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Práctica cerrada.", Data: updated})
}

// HandleListItems - GET /api/assignments/{id}/items
func (h *Handler) HandleListItems(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	items, err := h.Repo.ListItems(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}
	if items == nil {
		items = []models.AssignmentItem{}
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: items})
}

// AddItemRequest es el cuerpo para vincular un recurso.
type AddItemRequest struct {
	ItemID   int64   `json:"item_id"`
	Quantity float64 `json:"quantity"`
	Notes    string  `json:"notes"`
}

// HandleAddItem - POST /api/assignments/{id}/items
func (h *Handler) HandleAddItem(w http.ResponseWriter, r *http.Request) {
	assignmentID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var req AddItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	ai := &models.AssignmentItem{
		AssignmentID: assignmentID,
		ItemID:       req.ItemID,
		Quantity:     req.Quantity,
	}
	ai.Notes.String = req.Notes
	ai.Notes.Valid = req.Notes != ""

	id, err := h.Repo.AddItem(ai)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Message: "Recurso vinculado.", Data: map[string]int64{"id": id}})
}

// HandleRemoveItem - DELETE /api/assignments/{id}/items/{itemID}
func (h *Handler) HandleRemoveItem(w http.ResponseWriter, r *http.Request) {
	itemID, err := strconv.ParseInt(chi.URLParam(r, "itemID"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	if err := h.Repo.RemoveItem(itemID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Recurso desvinculado."})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
