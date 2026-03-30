package models

import "database/sql"

// ========================================================================
// MÓDULO: AUTENTICACIÓN Y USUARIOS
// ========================================================================

// User representa a un profesor, administrador u operador del sistema.
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // Nunca se serializa al JSON
	FullName     string `json:"full_name"`
	Role         string `json:"role"` // admin | teacher | operator
	IsActive     int    `json:"is_active"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// Student representa a un alumno registrado.
type Student struct {
	ID          int64          `json:"id"`
	StudentCode string         `json:"student_code"`
	FullName    string         `json:"full_name"`
	Email       sql.NullString `json:"email"`
	GroupName   string         `json:"group_name"`
	QRToken     sql.NullString `json:"qr_token"`
	IsActive    int            `json:"is_active"`
	CreatedAt   string         `json:"created_at"`
	UpdatedAt   string         `json:"updated_at"`
}

// StudentAccount es la cuenta de acceso de un alumno, separada de users.
type StudentAccount struct {
	ID           int64  `json:"id"`
	StudentID    int64  `json:"student_id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	IsActive     int    `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

// ========================================================================
// MÓDULO: INVENTARIO
// ========================================================================

// Category es una clasificación de recursos.
type Category struct {
	ID          int64          `json:"id"`
	Name        string         `json:"name"`
	Description sql.NullString `json:"description"`
	CreatedAt   string         `json:"created_at"`
}

// Item es un recurso inventariable (herramienta, consumible o máquina).
type Item struct {
	ID                int64          `json:"id"`
	SKU               string         `json:"sku"`
	Name              string         `json:"name"`
	ItemType          string         `json:"item_type"` // tool | consumable | machine
	CategoryID        sql.NullInt64  `json:"category_id"`
	Stock             float64        `json:"stock"`
	MinStock          float64        `json:"min_stock"`
	Unit              string         `json:"unit"`
	MaintenanceStatus string         `json:"maintenance_status"` // ok | scheduled | in_maintenance | critical | out_of_service
	Location          sql.NullString `json:"location"`
	ModuleData        string         `json:"module_data"`
	IsActive          int            `json:"is_active"`
	CreatedAt         string         `json:"created_at"`
	UpdatedAt         string         `json:"updated_at"`
	// Campos de JOIN (no en la tabla directamente)
	CategoryName string `json:"category_name,omitempty"`
}

// ========================================================================
// MÓDULO: PRÁCTICAS Y SESIONES
// ========================================================================

// Assignment es una práctica o actividad asignada por el profesor.
type Assignment struct {
	ID            int64          `json:"id"`
	Title         string         `json:"title"`
	Description   sql.NullString `json:"description"`
	Instructions  sql.NullString `json:"instructions"`
	TeacherID     int64          `json:"teacher_id"`
	GroupName     string         `json:"group_name"`
	Status        string         `json:"status"` // draft | published | closed
	ScheduledDate sql.NullString `json:"scheduled_date"`
	PublishedAt   sql.NullString `json:"published_at"`
	CreatedAt     string         `json:"created_at"`
	UpdatedAt     string         `json:"updated_at"`
	// Campo de JOIN
	TeacherName string `json:"teacher_name,omitempty"`
}

// AssignmentItem es un recurso sugerido/requerido para una práctica.
type AssignmentItem struct {
	ID           int64          `json:"id"`
	AssignmentID int64          `json:"assignment_id"`
	ItemID       int64          `json:"item_id"`
	Quantity     float64        `json:"quantity"`
	Notes        sql.NullString `json:"notes"`
	// Campos de JOIN
	ItemName string `json:"item_name,omitempty"`
	ItemSKU  string `json:"item_sku,omitempty"`
}

// LabSession es la ejecución operativa de una práctica.
type LabSession struct {
	ID             int64          `json:"id"`
	AssignmentID   int64          `json:"assignment_id"`
	Title          string         `json:"title"`
	TeacherID      int64          `json:"teacher_id"`
	GroupName      string         `json:"group_name"`
	ScheduledStart string         `json:"scheduled_start"`
	ScheduledEnd   string         `json:"scheduled_end"`
	Status         string         `json:"status"` // planned | open | closed | cancelled
	Notes          sql.NullString `json:"notes"`
	CreatedAt      string         `json:"created_at"`
	UpdatedAt      string         `json:"updated_at"`
	// Campos de JOIN
	AssignmentTitle string `json:"assignment_title,omitempty"`
	TeacherName     string `json:"teacher_name,omitempty"`
}

// ========================================================================
// MÓDULO: ASISTENCIA
// ========================================================================

// Attendance es el registro de asistencia de un alumno a una sesión.
type Attendance struct {
	ID         int64          `json:"id"`
	SessionID  int64          `json:"session_id"`
	StudentID  int64          `json:"student_id"`
	Status     string         `json:"status"` // present | late | absent | excused
	CheckInAt  string         `json:"check_in_at"`
	Notes      sql.NullString `json:"notes"`
	RecordedBy sql.NullInt64  `json:"recorded_by"`
	// Campos de JOIN
	StudentName string `json:"student_name,omitempty"`
	StudentCode string `json:"student_code,omitempty"`
}

// ========================================================================
// MÓDULO: SOLICITUDES Y PRÉSTAMOS
// ========================================================================

// ResourceRequest es una solicitud de recurso por parte de un alumno.
type ResourceRequest struct {
	ID           int64          `json:"id"`
	SessionID    sql.NullInt64  `json:"session_id"`
	AssignmentID sql.NullInt64  `json:"assignment_id"`
	StudentID    int64          `json:"student_id"`
	ItemID       int64          `json:"item_id"`
	RequestType  string         `json:"request_type"` // loan | consumption | machine_access
	Quantity     float64        `json:"quantity"`
	Status       string         `json:"status"` // pending | approved | rejected | returned
	RequestedAt  string         `json:"requested_at"`
	ResolvedAt   sql.NullString `json:"resolved_at"`
	ResolvedBy   sql.NullInt64  `json:"resolved_by"`
	Notes        sql.NullString `json:"notes"`
	// Campos de JOIN
	StudentName string `json:"student_name,omitempty"`
	StudentCode string `json:"student_code,omitempty"`
	ItemName    string `json:"item_name,omitempty"`
	ItemSKU     string `json:"item_sku,omitempty"`
}

// Transaction es un registro de movimiento de inventario (auditoría).
type Transaction struct {
	ID            int64          `json:"id"`
	ItemID        int64          `json:"item_id"`
	SessionID     sql.NullInt64  `json:"session_id"`
	StudentID     sql.NullInt64  `json:"student_id"`
	UserID        sql.NullInt64  `json:"user_id"`
	Type          string         `json:"type"` // loan_out | loan_return | consumption | adjustment | maintenance_hold | initial_stock
	Quantity      float64        `json:"quantity"`
	StockAfter    float64        `json:"stock_after"`
	ReferenceType sql.NullString `json:"reference_type"`
	ReferenceID   sql.NullInt64  `json:"reference_id"`
	Notes         sql.NullString `json:"notes"`
	CreatedAt     string         `json:"created_at"`
	// Campos de JOIN
	ItemName string `json:"item_name,omitempty"`
}

// ========================================================================
// MÓDULO: USO DE MAQUINARIA
// ========================================================================

// EquipmentUsage es un registro de uso de maquinaria fija.
type EquipmentUsage struct {
	ID           int64          `json:"id"`
	SessionID    int64          `json:"session_id"`
	ItemID       int64          `json:"item_id"`
	StudentID    int64          `json:"student_id"`
	SupervisorID int64          `json:"supervisor_id"`
	StartedAt    string         `json:"started_at"`
	EndedAt      sql.NullString `json:"ended_at"`
	Status       string         `json:"status"` // active | completed | interrupted
	Notes        sql.NullString `json:"notes"`
	CreatedAt    string         `json:"created_at"`
	// Campos de JOIN
	ItemName       string `json:"item_name,omitempty"`
	StudentName    string `json:"student_name,omitempty"`
	SupervisorName string `json:"supervisor_name,omitempty"`
}

// ========================================================================
// MÓDULO: MANUALES PDF
// ========================================================================

// Manual es un documento PDF vinculado a un item o práctica.
type Manual struct {
	ID           int64          `json:"id"`
	Title        string         `json:"title"`
	Description  sql.NullString `json:"description"`
	ItemID       sql.NullInt64  `json:"item_id"`
	AssignmentID sql.NullInt64  `json:"assignment_id"`
	FilePath     string         `json:"file_path"`
	FileSizeKB   sql.NullInt64  `json:"file_size_kb"`
	UploadedBy   sql.NullInt64  `json:"uploaded_by"`
	Version      string         `json:"version"`
	IsActive     int            `json:"is_active"`
	CreatedAt    string         `json:"created_at"`
	// Campos de JOIN
	ItemName       string `json:"item_name,omitempty"`
	AssignmentName string `json:"assignment_name,omitempty"`
}

// ========================================================================
// MÓDULO: MANTENIMIENTO
// ========================================================================

// MaintenanceLog es una entrada de la bitácora de mantenimiento.
type MaintenanceLog struct {
	ID                 int64          `json:"id"`
	ItemID             int64          `json:"item_id"`
	UserID             int64          `json:"user_id"`
	EntryDate          string         `json:"entry_date"`
	MaintenanceType    string         `json:"maintenance_type"` // preventive | corrective | inspection | calibration | other
	Description        string         `json:"description"`
	StatusAfter        string         `json:"status_after"` // ok | scheduled | in_maintenance | critical | out_of_service
	NextMaintenanceDue sql.NullString `json:"next_maintenance_due"`
	CostEstimate       sql.NullFloat64 `json:"cost_estimate"`
	CreatedAt          string         `json:"created_at"`
	// Campos de JOIN
	ItemName string `json:"item_name,omitempty"`
	UserName string `json:"user_name,omitempty"`
}

// ========================================================================
// MÓDULO: INCIDENCIAS
// ========================================================================

// IncidentReport es un reporte de fallo o incidencia.
type IncidentReport struct {
	ID                       int64          `json:"id"`
	SessionID                sql.NullInt64  `json:"session_id"`
	ItemID                   int64          `json:"item_id"`
	ReportedByStudentID      sql.NullInt64  `json:"reported_by_student_id"`
	RelatedPreviousStudentID sql.NullInt64  `json:"related_previous_student_id"`
	SupervisorID             sql.NullInt64  `json:"supervisor_id"`
	Description              string         `json:"description"`
	Severity                 string         `json:"severity"` // low | medium | high | critical
	Status                   string         `json:"status"`   // open | in_review | resolved | dismissed
	ResolutionNotes          sql.NullString `json:"resolution_notes"`
	CreatedAt                string         `json:"created_at"`
	ResolvedAt               sql.NullString `json:"resolved_at"`
	// Campos de JOIN
	ItemName                   string `json:"item_name,omitempty"`
	ReporterName               string `json:"reporter_name,omitempty"`
	RelatedPreviousStudentName string `json:"related_previous_student_name,omitempty"`
	SupervisorName             string `json:"supervisor_name,omitempty"`
}

// ========================================================================
// HELPERS: Respuestas API comunes
// ========================================================================

// APIResponse es la estructura estándar para respuestas JSON.
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// SessionInfo contiene la info de sesión del usuario activo (para /api/auth/me).
type SessionInfo struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	StudentID int64  `json:"student_id,omitempty"` // Solo si es alumno
	GroupName string `json:"group_name,omitempty"` // Solo si es alumno
}
