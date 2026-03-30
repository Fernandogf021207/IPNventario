package assignments

import (
	"database/sql"
	"fmt"

	"github.com/Fernandogf021207/IPNventario/core/models"
)

// Repository maneja las operaciones de base de datos para prácticas.
type Repository struct {
	DB *sql.DB
}

// NewRepository crea un nuevo repositorio de prácticas.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// List obtiene todas las prácticas. Si groupName no está vacío, filtra por grupo.
// Si onlyPublished es true, solo devuelve las publicadas.
func (r *Repository) List(groupName string, onlyPublished bool) ([]models.Assignment, error) {
	query := `
		SELECT a.id, a.title, a.description, a.instructions, a.teacher_id, a.group_name,
		       a.status, a.scheduled_date, a.published_at, a.created_at, a.updated_at,
		       u.full_name AS teacher_name
		FROM assignments a
		JOIN users u ON u.id = a.teacher_id
		WHERE 1=1
	`
	args := []interface{}{}

	if groupName != "" {
		query += " AND a.group_name = ?"
		args = append(args, groupName)
	}
	if onlyPublished {
		query += " AND a.status = 'published'"
	}

	query += " ORDER BY a.created_at DESC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listando prácticas: %w", err)
	}
	defer rows.Close()

	var assignments []models.Assignment
	for rows.Next() {
		var a models.Assignment
		err := rows.Scan(
			&a.ID, &a.Title, &a.Description, &a.Instructions, &a.TeacherID, &a.GroupName,
			&a.Status, &a.ScheduledDate, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt,
			&a.TeacherName,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando práctica: %w", err)
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

// GetByID obtiene una práctica por su ID.
func (r *Repository) GetByID(id int64) (*models.Assignment, error) {
	var a models.Assignment
	err := r.DB.QueryRow(`
		SELECT a.id, a.title, a.description, a.instructions, a.teacher_id, a.group_name,
		       a.status, a.scheduled_date, a.published_at, a.created_at, a.updated_at,
		       u.full_name AS teacher_name
		FROM assignments a
		JOIN users u ON u.id = a.teacher_id
		WHERE a.id = ?
	`, id).Scan(
		&a.ID, &a.Title, &a.Description, &a.Instructions, &a.TeacherID, &a.GroupName,
		&a.Status, &a.ScheduledDate, &a.PublishedAt, &a.CreatedAt, &a.UpdatedAt,
		&a.TeacherName,
	)
	if err != nil {
		return nil, fmt.Errorf("práctica no encontrada: %w", err)
	}
	return &a, nil
}

// Create crea una nueva práctica.
func (r *Repository) Create(a *models.Assignment) (int64, error) {
	result, err := r.DB.Exec(`
		INSERT INTO assignments (title, description, instructions, teacher_id, group_name, status, scheduled_date)
		VALUES (?, ?, ?, ?, ?, 'draft', ?)
	`, a.Title, a.Description, a.Instructions, a.TeacherID, a.GroupName, a.ScheduledDate)
	if err != nil {
		return 0, fmt.Errorf("error creando práctica: %w", err)
	}
	return result.LastInsertId()
}

// Update actualiza una práctica existente (solo si está en draft).
func (r *Repository) Update(id int64, a *models.Assignment) error {
	result, err := r.DB.Exec(`
		UPDATE assignments
		SET title = ?, description = ?, instructions = ?, group_name = ?, scheduled_date = ?
		WHERE id = ? AND status = 'draft'
	`, a.Title, a.Description, a.Instructions, a.GroupName, a.ScheduledDate, id)
	if err != nil {
		return fmt.Errorf("error actualizando práctica: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("práctica no encontrada o no está en borrador")
	}
	return nil
}

// Publish cambia el estado de una práctica a 'published'.
func (r *Repository) Publish(id int64) error {
	result, err := r.DB.Exec(`
		UPDATE assignments
		SET status = 'published', published_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
		WHERE id = ? AND status = 'draft'
	`, id)
	if err != nil {
		return fmt.Errorf("error publicando práctica: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("práctica no encontrada o ya está publicada/cerrada")
	}
	return nil
}

// Close cambia el estado de una práctica a 'closed'.
func (r *Repository) Close(id int64) error {
	result, err := r.DB.Exec(`
		UPDATE assignments
		SET status = 'closed'
		WHERE id = ? AND status IN ('draft', 'published')
	`, id)
	if err != nil {
		return fmt.Errorf("error cerrando práctica: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("práctica no encontrada o ya está cerrada")
	}
	return nil
}

// --- Assignment Items (recursos sugeridos) ---

// ListItems obtiene los recursos vinculados a una práctica.
func (r *Repository) ListItems(assignmentID int64) ([]models.AssignmentItem, error) {
	rows, err := r.DB.Query(`
		SELECT ai.id, ai.assignment_id, ai.item_id, ai.quantity, ai.notes,
		       i.name AS item_name, i.sku AS item_sku
		FROM assignment_items ai
		JOIN items i ON i.id = ai.item_id
		WHERE ai.assignment_id = ?
	`, assignmentID)
	if err != nil {
		return nil, fmt.Errorf("error listando recursos de práctica: %w", err)
	}
	defer rows.Close()

	var items []models.AssignmentItem
	for rows.Next() {
		var ai models.AssignmentItem
		if err := rows.Scan(&ai.ID, &ai.AssignmentID, &ai.ItemID, &ai.Quantity, &ai.Notes, &ai.ItemName, &ai.ItemSKU); err != nil {
			return nil, err
		}
		items = append(items, ai)
	}
	return items, nil
}

// AddItem vincula un recurso a una práctica.
func (r *Repository) AddItem(ai *models.AssignmentItem) (int64, error) {
	result, err := r.DB.Exec(`
		INSERT INTO assignment_items (assignment_id, item_id, quantity, notes)
		VALUES (?, ?, ?, ?)
	`, ai.AssignmentID, ai.ItemID, ai.Quantity, ai.Notes)
	if err != nil {
		return 0, fmt.Errorf("error vinculando recurso a práctica: %w", err)
	}
	return result.LastInsertId()
}

// RemoveItem desvincula un recurso de una práctica.
func (r *Repository) RemoveItem(id int64) error {
	_, err := r.DB.Exec("DELETE FROM assignment_items WHERE id = ?", id)
	return err
}
