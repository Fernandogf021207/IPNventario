package main

import (
	"database/sql"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Fernandogf021207/IPNventario/core/assignments"
	"github.com/Fernandogf021207/IPNventario/core/attendance"
	"github.com/Fernandogf021207/IPNventario/core/auth"
	"github.com/Fernandogf021207/IPNventario/core/database"
	"github.com/Fernandogf021207/IPNventario/core/inventory"
	"github.com/Fernandogf021207/IPNventario/core/sessions"
	"github.com/Fernandogf021207/IPNventario/core/students"
	"github.com/Fernandogf021207/IPNventario/modules/heavylab"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed schema.sql
var schemaSQL string

func main() {
	// Flags de configuración
	port := flag.Int("port", 8080, "Puerto del servidor HTTP")
	dbPath := flag.String("db", "./ipnventario.db", "Ruta al archivo de base de datos SQLite")
	uploadsDir := flag.String("uploads", "./uploads", "Directorio para archivos subidos (manuales PDF)")
	flag.Parse()

	// Crear directorio de uploads si no existe
	if err := os.MkdirAll(*uploadsDir+"/manuals", 0755); err != nil {
		log.Fatalf("[IPNventario] Error creando directorio de uploads: %v", err)
	}

	// Abrir y configurar la base de datos
	db, err := database.Open(*dbPath)
	if err != nil {
		log.Fatalf("[IPNventario] Error abriendo base de datos: %v", err)
	}
	defer db.Close()

	// Inicializar schema y seed
	if err := database.Initialize(db, schemaSQL); err != nil {
		log.Fatalf("[IPNventario] Error inicializando schema: %v", err)
	}

	// Crear servicios
	sessionStore := auth.NewSessionStore()
	authService := auth.NewService(db, sessionStore)
	authHandler := auth.NewHandler(authService)

	// Construir router
	r := buildRouter(db, sessionStore, authHandler, *uploadsDir)

	// Iniciar servidor
	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("[IPNventario] Servidor iniciado en http://%s", addr)
	log.Printf("[IPNventario] Base de datos: %s", *dbPath)
	log.Printf("[IPNventario] Directorio de uploads: %s", *uploadsDir)

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("[IPNventario] Error del servidor: %v", err)
	}
}

func buildRouter(db *sql.DB, sessionStore *auth.SessionStore, authHandler *auth.Handler, uploadsDir string) *chi.Mux {
	r := chi.NewRouter()

	// Middleware global
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Rutas públicas de autenticación
	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/login", authHandler.HandleLogin)
		r.Post("/logout", authHandler.HandleLogout)
	})

	// Rutas protegidas
	r.Group(func(r chi.Router) {
		r.Use(auth.RequireAuth(sessionStore))

		// Auth - info del usuario actual
		r.Get("/api/auth/me", authHandler.HandleMe)

		// Montar rutas de cada módulo (se agregan en fases siguientes)
		mountAssignmentRoutes(r, db)
		mountSessionRoutes(r, db)
		mountStudentRoutes(r, db)
		mountAttendanceRoutes(r, db)
		mountInventoryRoutes(r, db)
		mountRequestRoutes(r, db)
		mountEquipmentRoutes(r, db)
		mountManualRoutes(r, db, uploadsDir)
		mountMaintenanceRoutes(r, db)
		mountIncidentRoutes(r, db)
		mountReportRoutes(r, db)
	})

	// Ruta raíz de salud
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"IPNventario"}`))
	})

	return r
}

// ========================================================================
// Funciones de montaje de rutas — se implementarán fase por fase.
// Por ahora son stubs que no registran nada.
// ========================================================================

func mountAssignmentRoutes(r chi.Router, db *sql.DB) {
	h := assignments.NewHandler(db)
	h.RegisterRoutes(r)
}

func mountSessionRoutes(r chi.Router, db *sql.DB) {
	h := sessions.NewHandler(db)
	h.RegisterRoutes(r)
}

func mountStudentRoutes(r chi.Router, db *sql.DB) {
	h := students.NewHandler(db)
	h.RegisterRoutes(r)
}

func mountAttendanceRoutes(r chi.Router, db *sql.DB) {
	h := attendance.NewHandler(db)
	h.RegisterRoutes(r)
}

func mountInventoryRoutes(r chi.Router, db *sql.DB) {
	h := inventory.NewHandler(db)
	h.RegisterRoutes(r)
}

func mountRequestRoutes(r chi.Router, db *sql.DB) {
	h := inventory.NewRequestHandler(db)
	h.RegisterRoutes(r)
}

func mountEquipmentRoutes(r chi.Router, db *sql.DB) {
	h := heavylab.NewEquipmentHandler(db)
	h.RegisterRoutes(r)
}

func mountManualRoutes(r chi.Router, db *sql.DB, uploadsDir string) {
	// Fase 4: Se implementará en modules/heavylab/manuals/
}

func mountMaintenanceRoutes(r chi.Router, db *sql.DB) {
	// Fase 4: Se implementará en modules/heavylab/maintenance/
}

func mountIncidentRoutes(r chi.Router, db *sql.DB) {
	// Fase 4: Se implementará en modules/heavylab/incidents/
}

func mountReportRoutes(r chi.Router, db *sql.DB) {
	// Fase 4: Se implementará en core/reports/
}
