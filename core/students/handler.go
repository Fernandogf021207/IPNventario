package students

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/Fernandogf021207/IPNventario/core/auth"
	"github.com/Fernandogf021207/IPNventario/core/models"
	"github.com/go-chi/chi/v5"
)

// Handler contiene los endpoints HTTP para alumnos.
type Handler struct {
	Repo *Repository
}

// NewHandler crea un nuevo handler de alumnos.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{Repo: NewRepository(db)}
}

// RegisterRoutes registra las rutas de alumnos (solo teacher/admin).
func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/students", func(r chi.Router) {
		r.Use(auth.RequireTeacher)
		r.Get("/", h.HandleList)
		r.Post("/", h.HandleCreate)
		r.Post("/import-csv", h.HandleImportCSV)
		r.Get("/{id}", h.HandleGetByID)
		r.Put("/{id}", h.HandleUpdate)
		r.Delete("/{id}", h.HandleDeactivate)
	})
}

// HandleList - GET /api/students
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	groupFilter := r.URL.Query().Get("group")
	students, err := h.Repo.List(groupFilter)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error listando alumnos: " + err.Error(),
		})
		return
	}
	if students == nil {
		students = []models.Student{}
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: students})
}

// HandleGetByID - GET /api/students/{id}
func (h *Handler) HandleGetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	student, err := h.Repo.GetByID(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.APIResponse{Success: false, Error: "Alumno no encontrado."})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Data: student})
}

// CreateStudentRequest es el cuerpo para crear un alumno.
type CreateStudentRequest struct {
	StudentCode string `json:"student_code"`
	FullName    string `json:"full_name"`
	Email       string `json:"email"`
	GroupName   string `json:"group_name"`
	// Opcionales: si se proporcionan, se crea cuenta de acceso
	Username string `json:"username"`
	Password string `json:"password"`
}

// HandleCreate - POST /api/students
func (h *Handler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	var req CreateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	if req.StudentCode == "" || req.FullName == "" || req.GroupName == "" {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "student_code, full_name y group_name son requeridos.",
		})
		return
	}

	student := &models.Student{
		StudentCode: req.StudentCode,
		FullName:    req.FullName,
		GroupName:   req.GroupName,
	}
	if req.Email != "" {
		student.Email = sql.NullString{String: req.Email, Valid: true}
	}

	var id int64
	var err error

	if req.Username != "" && req.Password != "" {
		id, err = h.Repo.CreateWithAccount(student, req.Username, req.Password)
	} else {
		id, err = h.Repo.Create(student)
	}

	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Error creando alumno: " + err.Error()})
		return
	}

	created, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusCreated, models.APIResponse{Success: true, Message: "Alumno creado.", Data: created})
}

// HandleUpdate - PUT /api/students/{id}
func (h *Handler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	var req CreateStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "Cuerpo inválido."})
		return
	}

	student := &models.Student{
		StudentCode: req.StudentCode,
		FullName:    req.FullName,
		GroupName:   req.GroupName,
	}
	if req.Email != "" {
		student.Email = sql.NullString{String: req.Email, Valid: true}
	}

	if err := h.Repo.Update(id, student); err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: err.Error()})
		return
	}

	updated, _ := h.Repo.GetByID(id)
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Alumno actualizado.", Data: updated})
}

// HandleDeactivate - DELETE /api/students/{id}
func (h *Handler) HandleDeactivate(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{Success: false, Error: "ID inválido."})
		return
	}

	if err := h.Repo.Deactivate(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{Success: false, Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, models.APIResponse{Success: true, Message: "Alumno desactivado."})
}

// HandleImportCSV - POST /api/students/import-csv
// Acepta un archivo CSV con columnas: student_code, full_name, group_name, email
func (h *Handler) HandleImportCSV(w http.ResponseWriter, r *http.Request) {
	// Limitar tamaño del archivo a 5MB
	r.ParseMultipartForm(5 << 20)

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Archivo CSV requerido. Usa el campo 'file' en multipart form.",
		})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Leer header
	header, err := reader.Read()
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "Error leyendo encabezado del CSV.",
		})
		return
	}

	// Mapear columnas por nombre
	colMap := make(map[string]int)
	for i, col := range header {
		colMap[strings.TrimSpace(strings.ToLower(col))] = i
	}

	// Verificar columnas requeridas
	required := []string{"student_code", "full_name", "group_name"}
	for _, col := range required {
		if _, ok := colMap[col]; !ok {
			writeJSON(w, http.StatusBadRequest, models.APIResponse{
				Success: false,
				Error:   "Columnas requeridas: student_code, full_name, group_name, email (opcional). Falta: " + col,
			})
			return
		}
	}

	// Leer filas
	var rows []BulkImportRow
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue
		}

		row := BulkImportRow{
			StudentCode: strings.TrimSpace(record[colMap["student_code"]]),
			FullName:    strings.TrimSpace(record[colMap["full_name"]]),
			GroupName:   strings.TrimSpace(record[colMap["group_name"]]),
		}
		if emailIdx, ok := colMap["email"]; ok && emailIdx < len(record) {
			row.Email = strings.TrimSpace(record[emailIdx])
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		writeJSON(w, http.StatusBadRequest, models.APIResponse{
			Success: false, Error: "El CSV no contiene filas de datos.",
		})
		return
	}

	result, err := h.Repo.BulkImport(rows)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.APIResponse{
			Success: false, Error: "Error importando alumnos: " + err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, models.APIResponse{
		Success: true,
		Message: "Importación completada.",
		Data:    result,
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
