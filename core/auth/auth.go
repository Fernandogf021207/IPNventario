package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Fernandogf021207/IPNventario/core/models"
	"golang.org/x/crypto/bcrypt"
)

// SessionData almacena la información de una sesión activa.
type SessionData struct {
	UserID    int64
	Username  string
	FullName  string
	Role      string // admin | teacher | operator | student
	StudentID int64  // Solo para alumnos
	GroupName string // Solo para alumnos
	CreatedAt time.Time
}

// SessionStore gestiona las sesiones en memoria (suficiente para servidor local).
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*SessionData
}

// NewSessionStore crea un nuevo almacén de sesiones.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*SessionData),
	}
}

// Create genera un nuevo token de sesión y almacena los datos.
func (s *SessionStore) Create(data *SessionData) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	data.CreatedAt = time.Now()
	s.mu.Lock()
	s.sessions[token] = data
	s.mu.Unlock()
	return token, nil
}

// Get obtiene los datos de una sesión por su token.
func (s *SessionStore) Get(token string) (*SessionData, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.sessions[token]
	return data, ok
}

// Delete elimina una sesión.
func (s *SessionStore) Delete(token string) {
	s.mu.Lock()
	delete(s.sessions, token)
	s.mu.Unlock()
}

// generateToken genera un token seguro de 32 bytes.
func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("error generando token: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

// CookieName es el nombre de la cookie de sesión.
const CookieName = "ipnventario_session"

// Service contiene la lógica de autenticación.
type Service struct {
	DB    *sql.DB
	Store *SessionStore
}

// NewService crea un nuevo servicio de autenticación.
func NewService(db *sql.DB, store *SessionStore) *Service {
	return &Service{DB: db, Store: store}
}

// Login valida credenciales. Primero busca en users (admin/teacher/operator),
// luego en student_accounts (student).
func (s *Service) Login(username, password string) (*SessionData, string, error) {
	// 1. Intentar como usuario del sistema (admin/teacher/operator)
	var user models.User
	err := s.DB.QueryRow(
		"SELECT id, username, password_hash, full_name, role, is_active FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.FullName, &user.Role, &user.IsActive)

	if err == nil {
		// Encontrado en users
		if user.IsActive != 1 {
			return nil, "", fmt.Errorf("cuenta desactivada")
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			return nil, "", fmt.Errorf("contraseña incorrecta")
		}

		session := &SessionData{
			UserID:   user.ID,
			Username: user.Username,
			FullName: user.FullName,
			Role:     user.Role,
		}
		token, err := s.Store.Create(session)
		if err != nil {
			return nil, "", err
		}
		return session, token, nil
	}

	// 2. Intentar como alumno (student_accounts)
	var sa models.StudentAccount
	var student models.Student
	err = s.DB.QueryRow(`
		SELECT sa.id, sa.student_id, sa.username, sa.password_hash, sa.is_active,
		       st.full_name, st.group_name, st.is_active
		FROM student_accounts sa
		JOIN students st ON st.id = sa.student_id
		WHERE sa.username = ?
	`, username).Scan(
		&sa.ID, &sa.StudentID, &sa.Username, &sa.PasswordHash, &sa.IsActive,
		&student.FullName, &student.GroupName, &student.IsActive,
	)

	if err != nil {
		return nil, "", fmt.Errorf("usuario no encontrado")
	}

	if sa.IsActive != 1 || student.IsActive != 1 {
		return nil, "", fmt.Errorf("cuenta desactivada")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(sa.PasswordHash), []byte(password)); err != nil {
		return nil, "", fmt.Errorf("contraseña incorrecta")
	}

	session := &SessionData{
		UserID:    sa.ID,
		Username:  sa.Username,
		FullName:  student.FullName,
		Role:      "student",
		StudentID: sa.StudentID,
		GroupName: student.GroupName,
	}
	token, err := s.Store.Create(session)
	if err != nil {
		return nil, "", err
	}
	return session, token, nil
}

// Logout destruye la sesión.
func (s *Service) Logout(token string) {
	s.Store.Delete(token)
}

// SetSessionCookie establece la cookie de sesión en la respuesta HTTP.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400 * 7, // 7 días
	})
}

// ClearSessionCookie elimina la cookie de sesión.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// HashPassword genera un hash bcrypt de la contraseña.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// CheckPassword verifica una contraseña contra un hash bcrypt.
func CheckPassword(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
