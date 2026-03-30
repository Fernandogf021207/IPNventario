package attendance

import (
	"database/sql"
	"fmt"

	"github.com/Fernandogf021207/IPNventario/core/models"
)

// Repository maneja las operaciones de base de datos para asistencia.
type Repository struct {
	DB *sql.DB
}

// NewRepository crea un nuevo repositorio de asistencia.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// ListBySession obtiene todos los registros de asistencia de una sesión.
func (r *Repository) ListBySession(sessionID int64) ([]models.Attendance, error) {
	rows, err := r.DB.Query(`
		SELECT a.id, a.session_id, a.student_id, a.status, a.check_in_at,
		       a.notes, a.recorded_by,
		       s.full_name AS student_name, s.student_code
		FROM attendance a
		JOIN students s ON s.id = a.student_id
		WHERE a.session_id = ?
		ORDER BY a.check_in_at ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("error listando asistencia: %w", err)
	}
	defer rows.Close()

	var records []models.Attendance
	for rows.Next() {
		var a models.Attendance
		err := rows.Scan(
			&a.ID, &a.SessionID, &a.StudentID, &a.Status, &a.CheckInAt,
			&a.Notes, &a.RecordedBy,
			&a.StudentName, &a.StudentCode,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando asistencia: %w", err)
		}
		records = append(records, a)
	}
	return records, nil
}

// CheckSessionIsOpen verifica que la sesión esté abierta (RB1).
func (r *Repository) CheckSessionIsOpen(sessionID int64) error {
	var status string
	err := r.DB.QueryRow("SELECT status FROM lab_sessions WHERE id = ?", sessionID).Scan(&status)
	if err != nil {
		return fmt.Errorf("sesión no encontrada")
	}
	if status != "open" {
		return fmt.Errorf("RB1: no se puede registrar asistencia en una sesión con estado '%s'. Solo en sesiones abiertas", status)
	}
	return nil
}

// CheckIn registra la asistencia de un alumno en una sesión.
// Enforces RB1: solo permite check-in en sesiones abiertas.
func (r *Repository) CheckIn(sessionID, studentID int64, status string, recordedBy sql.NullInt64, notes string) (int64, error) {
	// RB1: Verificar que la sesión esté abierta
	if err := r.CheckSessionIsOpen(sessionID); err != nil {
		return 0, err
	}

	var notesVal sql.NullString
	if notes != "" {
		notesVal = sql.NullString{String: notes, Valid: true}
	}

	result, err := r.DB.Exec(`
		INSERT INTO attendance (session_id, student_id, status, recorded_by, notes)
		VALUES (?, ?, ?, ?, ?)
	`, sessionID, studentID, status, recordedBy, notesVal)
	if err != nil {
		return 0, fmt.Errorf("error registrando asistencia (posiblemente ya registrado): %w", err)
	}
	return result.LastInsertId()
}

// UpdateStatus actualiza el estado de un registro de asistencia.
func (r *Repository) UpdateStatus(id int64, status, notes string) error {
	var notesVal sql.NullString
	if notes != "" {
		notesVal = sql.NullString{String: notes, Valid: true}
	}

	result, err := r.DB.Exec(`
		UPDATE attendance SET status = ?, notes = ? WHERE id = ?
	`, status, notesVal, id)
	if err != nil {
		return fmt.Errorf("error actualizando asistencia: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("registro de asistencia no encontrado")
	}
	return nil
}

// GetByID obtiene un registro de asistencia por su ID.
func (r *Repository) GetByID(id int64) (*models.Attendance, error) {
	var a models.Attendance
	err := r.DB.QueryRow(`
		SELECT a.id, a.session_id, a.student_id, a.status, a.check_in_at,
		       a.notes, a.recorded_by,
		       s.full_name AS student_name, s.student_code
		FROM attendance a
		JOIN students s ON s.id = a.student_id
		WHERE a.id = ?
	`, id).Scan(
		&a.ID, &a.SessionID, &a.StudentID, &a.Status, &a.CheckInAt,
		&a.Notes, &a.RecordedBy,
		&a.StudentName, &a.StudentCode,
	)
	if err != nil {
		return nil, fmt.Errorf("registro no encontrado: %w", err)
	}
	return &a, nil
}

// GetSessionGroupName obtiene el group_name de una sesión.
func (r *Repository) GetSessionGroupName(sessionID int64) (string, error) {
	var groupName string
	err := r.DB.QueryRow("SELECT group_name FROM lab_sessions WHERE id = ?", sessionID).Scan(&groupName)
	return groupName, err
}
