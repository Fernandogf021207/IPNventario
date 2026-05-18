package inventory

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Fernandogf021207/IPNventario/core/auth"
	"github.com/Fernandogf021207/IPNventario/core/models"
	"github.com/go-chi/chi/v5"
)

// Handler contiene los endpoints HTTP para inventario.
type Handler struct {
	Repo *Repository
}

// NewHandler crea un nuevo handler de inventario.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{Repo: NewRepository(db)}
}

// RegisterRoutes registra las rutas de inventario en el router.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Items
	r.Route("/api/items", func(r chi.Router) {
		r.Get("/", h.HandleListItems)
		r.Get("/{id}", h.HandleGetItem)
		r.Get("/{id}/available", h.HandleCheckAvailable)
		r.Get("/{id}/transactions", h.HandleGetTransactions)
		r.Get("/{id}/last-usage", h.HandleGetLastUsage)

		// Solo teacher/admin/operator
		r.Group(func(r chi.Router) {
			r.Use(auth.RequireTeacher)
			r.Post("/", h.HandleCreateItem)
			r.Put("/{id}", h.HandleUpdateItem)
			r.Delete("/{id}", h.HandleDeleteItem)
			r.Post("/{id}/adjust-stock", h.HandleAdjustStock)
		})
	})

	// Categories
	r.Route("/api/categories", func(r chi.Router) {
		r.Get("/", h.HandleListCategories)
		r.Get("/{id}", h.HandleGetCategory)

		r.Group(func(r chi.Router) {
			r.Use(auth.RequireTeacher)
			r.Post("/", h.HandleCreateCategory)
			r.Put("/{id}", h.HandleUpdateCategory)
			r.Delete("/{id}", h.HandleDeleteCategory)
		})
	})
}

// ========================================================================
// ITEMS HANDLERS
// ========================================================================

// HandleListItems - GET /api/items
// Soporta ?type=tool|consumable|machine, ?available=true (RB8), ?category_id=, ?search=
func (h *Handler) HandleListItems(w http.ResponseWriter, r *http.Request) {
	itemType := r.URL.Query().Get("type")
	categoryID := r.URL.Query().Get("category_id")
	search := r.URL.Query().Get("search")
	activeOnly := r.URL.Query().Get("active_only") != "false"
	availableOnly := r.URL.Query().Get("available") == "true"

	items, err := h.Repo.ListItems(itemType, categoryID, search, activeOnly, availableOnly)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando items: " + err.Error(),
		})
		return
	}

	if items == nil {
		items = []models.Item{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: items})
}

// HandleGetItem - GET /api/items/{id}
func (h *Handler) HandleGetItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	item, err := h.Repo.GetItemByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Item no encontrado."})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: item})
}

// CreateItemRequest es el cuerpo para crear un item.
type CreateItemRequest struct {
	SKU               string  `json:"sku"`
	Name              string  `json:"name"`
	ItemType          string  `json:"item_type"`
	CategoryID        *int64  `json:"category_id"`
	Stock             float64 `json:"stock"`
	MinStock          float64 `json:"min_stock"`
	Unit              string  `json:"unit"`
	MaintenanceStatus string  `json:"maintenance_status"`
	Location          string  `json:"location"`
	ModuleData        string  `json:"module_data"`
}

// HandleCreateItem - POST /api/items
func (h *Handler) HandleCreateItem(w http.ResponseWriter, r *http.Request) {
	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	if req.SKU == "" || req.Name == "" || req.ItemType == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "SKU, nombre y tipo son requeridos.",
		})
		return
	}

	// Validar tipo
	if req.ItemType != "tool" && req.ItemType != "consumable" && req.ItemType != "machine" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Tipo debe ser 'tool', 'consumable' o 'machine'.",
		})
		return
	}

	if req.Unit == "" {
		req.Unit = "pza"
	}
	if req.MaintenanceStatus == "" {
		req.MaintenanceStatus = "ok"
	}
	if req.ModuleData == "" {
		req.ModuleData = "{}"
	}

	item := &models.Item{
		SKU:               req.SKU,
		Name:              req.Name,
		ItemType:          req.ItemType,
		Stock:             req.Stock,
		MinStock:          req.MinStock,
		Unit:              req.Unit,
		MaintenanceStatus: req.MaintenanceStatus,
		ModuleData:        req.ModuleData,
	}
	if req.CategoryID != nil {
		item.CategoryID.Int64 = *req.CategoryID
		item.CategoryID.Valid = true
	}
	item.Location.String = req.Location
	item.Location.Valid = req.Location != ""

	id, err := h.Repo.CreateItem(item)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error creando item: " + err.Error(),
		})
		return
	}

	// Si hay stock inicial, registrar transacción
	if req.Stock > 0 {
		session := auth.GetSession(r.Context())
		_ = h.Repo.AdjustStock(id, 0, "initial_stock", session.UserID, nil, nil, nil, nil,
			"Stock inicial al crear item")
		// El stock ya está establecido en el CREATE, no necesitamos ajustar realmente.
		// Solo registramos la transacción de stock inicial.
		h.Repo.DB.Exec(`
			INSERT INTO transactions (item_id, user_id, type, quantity, stock_after, notes)
			VALUES (?, ?, 'initial_stock', ?, ?, 'Stock inicial al crear item')
		`, id, session.UserID, req.Stock, req.Stock)
	}

	created, _ := h.Repo.GetItemByID(id)
	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true, Message: "Item creado.", Data: created,
	})
}

// HandleUpdateItem - PUT /api/items/{id}
func (h *Handler) HandleUpdateItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	if req.ModuleData == "" {
		req.ModuleData = "{}"
	}

	item := &models.Item{
		SKU:      req.SKU,
		Name:     req.Name,
		ItemType: req.ItemType,
		MinStock: req.MinStock,
		Unit:     req.Unit,
		ModuleData: req.ModuleData,
	}
	if req.CategoryID != nil {
		item.CategoryID.Int64 = *req.CategoryID
		item.CategoryID.Valid = true
	}
	item.Location.String = req.Location
	item.Location.Valid = req.Location != ""

	if err := h.Repo.UpdateItem(id, item); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetItemByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true, Message: "Item actualizado.", Data: updated,
	})
}

// HandleDeleteItem - DELETE /api/items/{id}
func (h *Handler) HandleDeleteItem(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	if err := h.Repo.DeleteItem(id); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Item desactivado."})
}

// AdjustStockRequest es el cuerpo para ajustar stock.
type AdjustStockRequest struct {
	Quantity float64 `json:"quantity"` // Positivo = entrada, Negativo = salida
	Notes    string  `json:"notes"`
}

// HandleAdjustStock - POST /api/items/{id}/adjust-stock
func (h *Handler) HandleAdjustStock(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var req AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	if req.Quantity == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "La cantidad no puede ser cero.",
		})
		return
	}

	session := auth.GetSession(r.Context())
	err = h.Repo.AdjustStock(id, req.Quantity, "adjustment", session.UserID, nil, nil, nil, nil, req.Notes)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetItemByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true, Message: "Stock ajustado.", Data: updated,
	})
}

// HandleCheckAvailable - GET /api/items/{id}/available
func (h *Handler) HandleCheckAvailable(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	available, status, err := h.Repo.CheckMachineAvailable(id)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"available":          available,
			"maintenance_status": status,
		},
	})
}

// ========================================================================
// CATEGORIES HANDLERS
// ========================================================================

// HandleListCategories - GET /api/categories
func (h *Handler) HandleListCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := h.Repo.ListCategories()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando categorías: " + err.Error(),
		})
		return
	}
	if categories == nil {
		categories = []models.Category{}
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: categories})
}

// HandleGetCategory - GET /api/categories/{id}
func (h *Handler) HandleGetCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	category, err := h.Repo.GetCategoryByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Categoría no encontrada."})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: category})
}

// CategoryRequest es el cuerpo para crear/editar una categoría.
type CategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// HandleCreateCategory - POST /api/categories
func (h *Handler) HandleCreateCategory(w http.ResponseWriter, r *http.Request) {
	var req CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Nombre es requerido.",
		})
		return
	}

	id, err := h.Repo.CreateCategory(req.Name, req.Description)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error creando categoría: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusCreated, models.APIResponse{
		Success: true, Message: "Categoría creada.",
		Data: map[string]int64{"id": id},
	})
}

// HandleUpdateCategory - PUT /api/categories/{id}
func (h *Handler) HandleUpdateCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var req CategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	if err := h.Repo.UpdateCategory(id, req.Name, req.Description); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Categoría actualizada."})
}

// HandleDeleteCategory - DELETE /api/categories/{id}
func (h *Handler) HandleDeleteCategory(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	if err := h.Repo.DeleteCategory(id); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Categoría eliminada."})
}

// ========================================================================
// TRANSACTION HISTORY
// ========================================================================

// HandleGetTransactions - GET /api/items/{id}/transactions
func (h *Handler) HandleGetTransactions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	// Verificar que el item existe
	_, err = h.Repo.GetItemByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Item no encontrado."})
		return
	}

	transactions, err := h.Repo.GetTransactionsByItemID(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error obteniendo transacciones: " + err.Error(),
		})
		return
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: transactions})
}

// ========================================================================
// LAST USAGE (RF16)
// ========================================================================

// LastUsageView representa una fila de la vista v_last_equipment_usage.
type LastUsageView struct {
	ItemID      int64          `json:"item_id"`
	ItemName    string         `json:"item_name"`
	SKU         string         `json:"sku"`
	StudentID   sql.NullInt64  `json:"student_id"`
	StudentName sql.NullString `json:"student_name"`
	StudentCode sql.NullString `json:"student_code"`
	StartedAt   sql.NullString `json:"started_at"`
	EndedAt     sql.NullString `json:"ended_at"`
	SessionID   sql.NullInt64  `json:"session_id"`
	UsageStatus sql.NullString `json:"usage_status"`
}

// HandleGetLastUsage - GET /api/items/{id}/last-usage
// Usa la vista v_last_equipment_usage para trazabilidad (RF16).
func (h *Handler) HandleGetLastUsage(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var v LastUsageView
	err = h.Repo.DB.QueryRow(`
		SELECT item_id, item_name, sku, student_id, student_name,
		       student_code, started_at, ended_at, session_id, usage_status
		FROM v_last_equipment_usage
		WHERE item_id = ?
	`, id).Scan(
		&v.ItemID, &v.ItemName, &v.SKU, &v.StudentID, &v.StudentName,
		&v.StudentCode, &v.StartedAt, &v.EndedAt, &v.SessionID, &v.UsageStatus,
	)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{
			Success: false, Error: "No se encontró registro de uso para este item.",
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: v})
}

// ========================================================================
// HELPERS
// ========================================================================

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
