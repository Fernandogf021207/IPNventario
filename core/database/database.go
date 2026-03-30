package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

// Open abre la conexión a SQLite y configura los PRAGMAs requeridos.
func Open(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("no se pudo abrir la base de datos: %w", err)
	}

	// PRAGMAs obligatorios según la documentación técnica
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA busy_timeout = 5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return nil, fmt.Errorf("error ejecutando %s: %w", p, err)
		}
	}

	// Verificar que los PRAGMAs se aplicaron
	var fk int
	db.QueryRow("PRAGMA foreign_keys").Scan(&fk)
	if fk != 1 {
		db.Close()
		return nil, fmt.Errorf("foreign_keys no se activó correctamente")
	}

	return db, nil
}

// Initialize ejecuta el schema SQL si las tablas no existen.
func Initialize(db *sql.DB, schemaSQL string) error {
	// Verificar si ya existe la tabla 'users' como indicador de esquema aplicado
	var tableName string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='users'").Scan(&tableName)
	if err == nil {
		// La tabla ya existe, verificar/corregir seed del admin
		return fixAdminSeed(db)
	}

	// Ejecutar el schema completo
	log.Println("[IPNventario] Inicializando base de datos con schema.sql...")

	// Reemplazar el hash placeholder con uno real de bcrypt
	realHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("error generando hash bcrypt para admin: %w", err)
	}

	schemaSQL = strings.Replace(
		schemaSQL,
		"$2a$10$PLACEHOLDER_REPLACE_WITH_REAL_BCRYPT_HASH",
		string(realHash),
		1,
	)

	_, err = db.Exec(schemaSQL)
	if err != nil {
		return fmt.Errorf("error ejecutando schema.sql: %w", err)
	}

	log.Println("[IPNventario] Contraseña inicial del admin: admin123 — cámbiala después del primer acceso.")
	return nil
}

// fixAdminSeed verifica si el hash del admin sigue siendo el placeholder y lo reemplaza.
func fixAdminSeed(db *sql.DB) error {
	var hash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = 'admin'").Scan(&hash)
	if err != nil {
		// No hay usuario admin, no es un error crítico
		return nil
	}

	if strings.Contains(hash, "PLACEHOLDER") {
		realHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("error generando hash bcrypt: %w", err)
		}
		_, err = db.Exec("UPDATE users SET password_hash = ? WHERE username = 'admin'", string(realHash))
		if err != nil {
			return fmt.Errorf("error actualizando hash del admin: %w", err)
		}
		log.Println("[IPNventario] Contraseña inicial del admin: admin123 — cámbiala después del primer acceso.")
	}

	return nil
}
