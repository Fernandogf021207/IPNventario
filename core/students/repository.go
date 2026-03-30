package students

import (
	"database/sql"
	"fmt"

	"github.com/Fernandogf021207/IPNventario/core/models"
	"golang.org/x/crypto/bcrypt"
)

// Repository maneja las operaciones de base de datos para alumnos.
type Repository struct {
	DB *sql.DB
}

// NewRepository crea un nuevo repositorio de alumnos.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

// List obtiene todos los alumnos activos. Filtra por grupo si se proporciona.
func (r *Repository) List(groupName string) ([]models.Student, error) {
	query := `
		SELECT id, student_code, full_name, email, group_name, qr_token,
		       is_active, created_at, updated_at
		FROM students
		WHERE is_active = 1
	`
	args := []interface{}{}

	if groupName != "" {
		query += " AND group_name = ?"
		args = append(args, groupName)
	}

	query += " ORDER BY full_name ASC"

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error listando alumnos: %w", err)
	}
	defer rows.Close()

	var students []models.Student
	for rows.Next() {
		var s models.Student
		err := rows.Scan(
			&s.ID, &s.StudentCode, &s.FullName, &s.Email, &s.GroupName,
			&s.QRToken, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error escaneando alumno: %w", err)
		}
		students = append(students, s)
	}
	return students, nil
}

// GetByID obtiene un alumno por su ID.
func (r *Repository) GetByID(id int64) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, student_code, full_name, email, group_name, qr_token,
		       is_active, created_at, updated_at
		FROM students
		WHERE id = ?
	`, id).Scan(
		&s.ID, &s.StudentCode, &s.FullName, &s.Email, &s.GroupName,
		&s.QRToken, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("alumno no encontrado: %w", err)
	}
	return &s, nil
}

// GetByStudentCode obtiene un alumno por su boleta.
func (r *Repository) GetByStudentCode(code string) (*models.Student, error) {
	var s models.Student
	err := r.DB.QueryRow(`
		SELECT id, student_code, full_name, email, group_name, qr_token,
		       is_active, created_at, updated_at
		FROM students
		WHERE student_code = ?
	`, code).Scan(
		&s.ID, &s.StudentCode, &s.FullName, &s.Email, &s.GroupName,
		&s.QRToken, &s.IsActive, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("alumno no encontrado: %w", err)
	}
	return &s, nil
}

// Create crea un nuevo alumno.
func (r *Repository) Create(s *models.Student) (int64, error) {
	result, err := r.DB.Exec(`
		INSERT INTO students (student_code, full_name, email, group_name)
		VALUES (?, ?, ?, ?)
	`, s.StudentCode, s.FullName, s.Email, s.GroupName)
	if err != nil {
		return 0, fmt.Errorf("error creando alumno: %w", err)
	}
	return result.LastInsertId()
}

// Update actualiza los datos de un alumno.
func (r *Repository) Update(id int64, s *models.Student) error {
	_, err := r.DB.Exec(`
		UPDATE students
		SET student_code = ?, full_name = ?, email = ?, group_name = ?
		WHERE id = ?
	`, s.StudentCode, s.FullName, s.Email, s.GroupName, id)
	if err != nil {
		return fmt.Errorf("error actualizando alumno: %w", err)
	}
	return nil
}

// Deactivate desactiva un alumno.
func (r *Repository) Deactivate(id int64) error {
	_, err := r.DB.Exec("UPDATE students SET is_active = 0 WHERE id = ?", id)
	return err
}

// CreateWithAccount crea un alumno y su cuenta de acceso.
func (r *Repository) CreateWithAccount(s *models.Student, username, password string) (int64, error) {
	tx, err := r.DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback()

	// Crear alumno
	result, err := tx.Exec(`
		INSERT INTO students (student_code, full_name, email, group_name)
		VALUES (?, ?, ?, ?)
	`, s.StudentCode, s.FullName, s.Email, s.GroupName)
	if err != nil {
		return 0, fmt.Errorf("error creando alumno: %w", err)
	}

	studentID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Crear cuenta si se proporcionó username
	if username != "" && password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return 0, fmt.Errorf("error generando hash: %w", err)
		}
		_, err = tx.Exec(`
			INSERT INTO student_accounts (student_id, username, password_hash)
			VALUES (?, ?, ?)
		`, studentID, username, string(hash))
		if err != nil {
			return 0, fmt.Errorf("error creando cuenta de alumno: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("error al confirmar transacción: %w", err)
	}

	return studentID, nil
}

// BulkImportResult contiene los resultados de una importación masiva.
type BulkImportResult struct {
	Created  int      `json:"created"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

// BulkImportRow es una fila del CSV de importación.
type BulkImportRow struct {
	StudentCode string
	FullName    string
	GroupName   string
	Email       string
}

// BulkImport importa múltiples alumnos desde filas CSV.
func (r *Repository) BulkImport(rows []BulkImportRow) (*BulkImportResult, error) {
	result := &BulkImportResult{}

	tx, err := r.DB.Begin()
	if err != nil {
		return nil, fmt.Errorf("error iniciando transacción: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO students (student_code, full_name, email, group_name)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return nil, fmt.Errorf("error preparando statement: %w", err)
	}
	defer stmt.Close()

	for i, row := range rows {
		if row.StudentCode == "" || row.FullName == "" || row.GroupName == "" {
			result.Errors = append(result.Errors, fmt.Sprintf("Fila %d: student_code, full_name y group_name son requeridos", i+1))
			result.Skipped++
			continue
		}

		var email sql.NullString
		if row.Email != "" {
			email = sql.NullString{String: row.Email, Valid: true}
		}

		_, err := stmt.Exec(row.StudentCode, row.FullName, email, row.GroupName)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Fila %d (%s): %s", i+1, row.StudentCode, err.Error()))
			result.Skipped++
			continue
		}
		result.Created++
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("error al confirmar transacción: %w", err)
	}

	return result, nil
}
