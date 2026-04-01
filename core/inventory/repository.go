package inventory

import (
	"database/sql"
	"fmt"

	"github.com/Fernandogf021207/IPNventario/core/models"
)

// Repository maneja las operaciones de base de datos para inventario.
type Repository struct {
	DB *sql.DB
}

// NewRepository crea un nuevo repositorio de inventario.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// ========================================================================
// ITEMS
// ========================================================================

// ListItems obtiene los items con filtros opcionales.
func (r *Repository) ListItems(itemType, categoryID, search string, activeOnly bool) ([]models.Item, error) {
	query := `
		SELECT i.id, i.sku, i.name, i.item_type, i.category_id, i.stock, i.min_stock,
		       i.unit, i.maintenance_status, i.location, i.module_data, i.is_active,
		       i.created_at, i.updated_at,
		       COALESCE(c.name, '') AS category_name
		FROM items i
		LEFT JOIN categories c ON c.id = i.category_id
		WHERE 1=1
	`
	args := []interface{}{}

	if activeOnly {
		query += " AND i.is_active = 1"
	}
	if itemType != "" {
		query += " AND i.item_type = ?"
		args = append(args, itemType)
	}
	if categoryID != "" {
		query += " AND i.category_id = ?"
		args = append(args, categoryID)
	}
	if search != "" {
		query += " AND (i.name LIKE ? OR i.sku LIKE ?)"
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm)
	}

	query += " ORDER BY i.name ASC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listando items: %w", err)
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		err := rows.Scan(
			&item.ID, &item.SKU, &item.Name, &item.ItemType, &item.CategoryID,
			&item.Stock, &item.MinStock, &item.Unit, &item.MaintenanceStatus,
			&item.Location, &item.ModuleData, &item.IsActive,
			&item.CreatedAt, &item.UpdatedAt,
			&item.CategoryName,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando item: %w", err)
		}
		items = append(items, item)
	}
	return items, nil
}

// GetItemByID obtiene un item por su ID.
func (r *Repository) GetItemByID(id int64) (*models.Item, error) {
	var item models.Item
	err := r.DB.QueryRow(`
		SELECT i.id, i.sku, i.name, i.item_type, i.category_id, i.stock, i.min_stock,
		       i.unit, i.maintenance_status, i.location, i.module_data, i.is_active,
		       i.created_at, i.updated_at,
		       COALESCE(c.name, '') AS category_name
		FROM items i
		LEFT JOIN categories c ON c.id = i.category_id
		WHERE i.id = ?
	`, id).Scan(
		&item.ID, &item.SKU, &item.Name, &item.ItemType, &item.CategoryID,
		&item.Stock, &item.MinStock, &item.Unit, &item.MaintenanceStatus,
		&item.Location, &item.ModuleData, &item.IsActive,
		&item.CreatedAt, &item.UpdatedAt,
		&item.CategoryName,
	)
	if err != nil {
		return nil, fmt.Errorf("item no encontrado: %w", err)
	}
	return &item, nil
}

// CreateItem crea un nuevo item.
func (r *Repository) CreateItem(item *models.Item) (int64, error) {
	result, err := r.DB.Exec(`
		INSERT INTO items (sku, name, item_type, category_id, stock, min_stock, unit,
		                   maintenance_status, location, module_data)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.SKU, item.Name, item.ItemType, item.CategoryID, item.Stock,
		item.MinStock, item.Unit, item.MaintenanceStatus, item.Location, item.ModuleData)
	if err != nil {
		return 0, fmt.Errorf("error creando item: %w", err)
	}
	return result.LastInsertId()
}

// UpdateItem actualiza un item existente.
func (r *Repository) UpdateItem(id int64, item *models.Item) error {
	result, err := r.DB.Exec(`
		UPDATE items
		SET sku = ?, name = ?, item_type = ?, category_id = ?, min_stock = ?,
		    unit = ?, location = ?, module_data = ?
		WHERE id = ? AND is_active = 1
	`, item.SKU, item.Name, item.ItemType, item.CategoryID, item.MinStock,
		item.Unit, item.Location, item.ModuleData, id)
	if err != nil {
		return fmt.Errorf("error actualizando item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item no encontrado o está desactivado")
	}
	return nil
}

// DeleteItem desactiva un item (soft-delete).
func (r *Repository) DeleteItem(id int64) error {
	result, err := r.DB.Exec("UPDATE items SET is_active = 0 WHERE id = ? AND is_active = 1", id)
	if err != nil {
		return fmt.Errorf("error desactivando item: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("item no encontrado o ya está desactivado")
	}
	return nil
}

// AdjustStock ajusta el stock de un item y registra la transacción.
// RB2: Valida que stock + quantity >= 0
// RB5: Toda modificación de stock genera una transacción
func (r *Repository) AdjustStock(itemID int64, quantity float64, txType string,
	userID int64, sessionID, studentID *int64, refType *string, refID *int64, notes string) error {

	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback()

	// Obtener stock actual
	var currentStock float64
	err = tx.QueryRow("SELECT stock FROM items WHERE id = ? AND is_active = 1", itemID).Scan(&currentStock)
	if err != nil {
		return fmt.Errorf("item no encontrado: %w", err)
	}

	newStock := currentStock + quantity
	// RB2: Stock no puede ser negativo
	if newStock < 0 {
		return fmt.Errorf("stock insuficiente: actual=%.2f, ajuste=%.2f, resultante=%.2f", currentStock, quantity, newStock)
	}

	// Actualizar stock
	_, err = tx.Exec("UPDATE items SET stock = ? WHERE id = ?", newStock, itemID)
	if err != nil {
		return fmt.Errorf("error actualizando stock: %w", err)
	}

	// RB5: Registrar transacción
	_, err = tx.Exec(`
		INSERT INTO transactions (item_id, session_id, student_id, user_id, type, quantity, stock_after, reference_type, reference_id, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, itemID, sessionID, studentID, userID, txType, quantity, newStock, refType, refID, notes)
	if err != nil {
		return fmt.Errorf("error registrando transacción: %w", err)
	}

	return tx.Commit()
}

// CheckMachineAvailable verifica si una máquina está disponible para uso.
// RB8: No disponible si maintenance_status IN ('critical', 'out_of_service')
func (r *Repository) CheckMachineAvailable(itemID int64) (bool, string, error) {
	var maintenanceStatus string
	var itemType string
	err := r.DB.QueryRow(
		"SELECT item_type, maintenance_status FROM items WHERE id = ? AND is_active = 1",
		itemID,
	).Scan(&itemType, &maintenanceStatus)
	if err != nil {
		return false, "", fmt.Errorf("item no encontrado: %w", err)
	}
	if itemType != "machine" {
		return false, maintenanceStatus, fmt.Errorf("el item no es una máquina")
	}
	if maintenanceStatus == "critical" || maintenanceStatus == "out_of_service" {
		return false, maintenanceStatus, nil
	}
	return true, maintenanceStatus, nil
}

// ========================================================================
// CATEGORIES
// ========================================================================

// ListCategories obtiene todas las categorías.
func (r *Repository) ListCategories() ([]models.Category, error) {
	rows, err := r.DB.Query("SELECT id, name, description, created_at FROM categories ORDER BY name ASC")
	if err != nil {
		return nil, fmt.Errorf("error listando categorías: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("error escaneando categoría: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, nil
}

// CreateCategory crea una nueva categoría.
func (r *Repository) CreateCategory(name, description string) (int64, error) {
	result, err := r.DB.Exec(
		"INSERT INTO categories (name, description) VALUES (?, ?)",
		name, description,
	)
	if err != nil {
		return 0, fmt.Errorf("error creando categoría: %w", err)
	}
	return result.LastInsertId()
}

// UpdateCategory actualiza una categoría existente.
func (r *Repository) UpdateCategory(id int64, name, description string) error {
	result, err := r.DB.Exec(
		"UPDATE categories SET name = ?, description = ? WHERE id = ?",
		name, description, id,
	)
	if err != nil {
		return fmt.Errorf("error actualizando categoría: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("categoría no encontrada")
	}
	return nil
}

// DeleteCategory elimina una categoría.
func (r *Repository) DeleteCategory(id int64) error {
	result, err := r.DB.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("error eliminando categoría: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("categoría no encontrada")
	}
	return nil
}
