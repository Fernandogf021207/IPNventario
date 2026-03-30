package sessions

import (
	"database/sql"
	"fmt"

	"github.com/Fernandogf021207/IPNventario/core/models"
)

// Repository maneja las operaciones de base de datos para sesiones de laboratorio.
type Repository struct {
	DB *sql.DB
}

// NewRepository crea un nuevo repositorio de sesiones.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// List obtiene sesiones. Filtra por grupo y/o estado si se proporcionan.
func (r *Repository) List(groupName, status string) ([]models.LabSession, error) {
	query := `
		SELECT ls.id, ls.assignment_id, ls.title, ls.teacher_id, ls.group_name,
		       ls.scheduled_start, ls.scheduled_end, ls.status, ls.notes,
		       ls.created_at, ls.updated_at,
		       a.title AS assignment_title, u.full_name AS teacher_name
		FROM lab_sessions ls
		JOIN assignments a ON a.id = ls.assignment_id
		JOIN users u ON u.id = ls.teacher_id
		WHERE 1=1
	`
	args := []interface{}{}

	if groupName != "" {
		query += " AND ls.group_name = ?"
		args = append(args, groupName)
	}
	if status != "" {
		query += " AND ls.status = ?"
		args = append(args, status)
	}

	query += " ORDER BY ls.scheduled_start DESC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listando sesiones: %w", err)
	}
	defer rows.Close()

	var sessions []models.LabSession
	for rows.Next() {
		var s models.LabSession
		err := rows.Scan(
			&s.ID, &s.AssignmentID, &s.Title, &s.TeacherID, &s.GroupName,
			&s.ScheduledStart, &s.ScheduledEnd, &s.Status, &s.Notes,
			&s.CreatedAt, &s.UpdatedAt,
			&s.AssignmentTitle, &s.TeacherName,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando sesión: %w", err)
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

// GetByID obtiene una sesión por su ID.
func (r *Repository) GetByID(id int64) (*models.LabSession, error) {
	var s models.LabSession
	err := r.DB.QueryRow(`
		SELECT ls.id, ls.assignment_id, ls.title, ls.teacher_id, ls.group_name,
		       ls.scheduled_start, ls.scheduled_end, ls.status, ls.notes,
		       ls.created_at, ls.updated_at,
		       a.title AS assignment_title, u.full_name AS teacher_name
		FROM lab_sessions ls
		JOIN assignments a ON a.id = ls.assignment_id
		JOIN users u ON u.id = ls.teacher_id
		WHERE ls.id = ?
	`, id).Scan(
		&s.ID, &s.AssignmentID, &s.Title, &s.TeacherID, &s.GroupName,
		&s.ScheduledStart, &s.ScheduledEnd, &s.Status, &s.Notes,
		&s.CreatedAt, &s.UpdatedAt,
		&s.AssignmentTitle, &s.TeacherName,
	)
	if err != nil {
		return nil, fmt.Errorf("sesión no encontrada: %w", err)
	}
	return &s, nil
}

// Create crea una nueva sesión de laboratorio.
func (r *Repository) Create(s *models.LabSession) (int64, error) {
	result, err := r.DB.Exec(`
		INSERT INTO lab_sessions (assignment_id, title, teacher_id, group_name, scheduled_start, scheduled_end, notes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, s.AssignmentID, s.Title, s.TeacherID, s.GroupName, s.ScheduledStart, s.ScheduledEnd, s.Notes)
	if err != nil {
		return 0, fmt.Errorf("error creando sesión: %w", err)
	}
	return result.LastInsertId()
}

// Update actualiza los datos de una sesión.
func (r *Repository) Update(id int64, s *models.LabSession) error {
	_, err := r.DB.Exec(`
		UPDATE lab_sessions
		SET assignment_id = ?, title = ?, group_name = ?, scheduled_start = ?, scheduled_end = ?, notes = ?
		WHERE id = ? AND status = 'planned'
	`, s.AssignmentID, s.Title, s.GroupName, s.ScheduledStart, s.ScheduledEnd, s.Notes, id)
	return err
}

// UpdateStatus cambia el estado de una sesión.
func (r *Repository) UpdateStatus(id int64, newStatus string, allowedFrom []string) error {
	query := `UPDATE lab_sessions SET status = ? WHERE id = ? AND status IN (`
	args := []interface{}{newStatus, id}
	for i, s := range allowedFrom {
		if i > 0 {
			query += ", "
		}
		query += "?"
		args = append(args, s)
	}
	query += ")"

	result, err := r.DB.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error actualizando estado de sesión: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("sesión no encontrada o transición de estado inválida")
	}
	return nil
}

// Open abre una sesión (planned → open).
func (r *Repository) Open(id int64) error {
	return r.UpdateStatus(id, "open", []string{"planned"})
}

// Close cierra una sesión (open → closed).
func (r *Repository) Close(id int64) error {
	return r.UpdateStatus(id, "closed", []string{"open"})
}

// Cancel cancela una sesión (planned|open → cancelled).
func (r *Repository) Cancel(id int64) error {
	return r.UpdateStatus(id, "cancelled", []string{"planned", "open"})
}
