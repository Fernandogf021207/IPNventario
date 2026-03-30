-- =============================================================================
-- LABORATORIO DE PESADOS - UPIITA IPN
-- Esquema SQL para SQLite (MVP v2.0)
-- Configuración recomendada al abrir la conexión:
--   PRAGMA foreign_keys = ON;
--   PRAGMA journal_mode = WAL;
--   PRAGMA synchronous = NORMAL;
--   PRAGMA busy_timeout = 5000;
-- =============================================================================


-- =============================================================================
-- MÓDULO: AUTENTICACIÓN Y USUARIOS
-- =============================================================================

CREATE TABLE IF NOT EXISTS users (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT    NOT NULL UNIQUE,
    password_hash   TEXT    NOT NULL,
    full_name       TEXT    NOT NULL,
    -- 'admin' | 'teacher' | 'operator'
    -- 'operator' cubre futuros roles de técnico de laboratorio
    role            TEXT    NOT NULL CHECK (role IN ('admin', 'teacher', 'operator')),
    is_active       INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

-- Trigger para mantener updated_at automáticamente
CREATE TRIGGER IF NOT EXISTS users_updated_at
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE id = OLD.id;
END;


-- =============================================================================
-- MÓDULO: ALUMNOS
-- =============================================================================

CREATE TABLE IF NOT EXISTS students (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_code    TEXT    NOT NULL UNIQUE,   -- Número de boleta / matrícula
    full_name       TEXT    NOT NULL,
    email           TEXT,
    -- Ej: '6CM11', '7BM12' — flexible para distintos grupos
    group_name      TEXT    NOT NULL,
    -- QR o código interno para identificación rápida en asistencia
    qr_token        TEXT    UNIQUE,
    is_active       INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TRIGGER IF NOT EXISTS students_updated_at
    AFTER UPDATE ON students
    FOR EACH ROW
BEGIN
    UPDATE students SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE id = OLD.id;
END;

-- Cuenta de acceso opcional del alumno (puede existir sin cuenta al inicio)
-- Separado de 'users' para no mezclar roles operativos con usuarios académicos
CREATE TABLE IF NOT EXISTS student_accounts (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id      INTEGER NOT NULL UNIQUE REFERENCES students(id) ON DELETE CASCADE,
    username        TEXT    NOT NULL UNIQUE,
    password_hash   TEXT    NOT NULL,
    is_active       INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);


-- =============================================================================
-- MÓDULO: INVENTARIO
-- =============================================================================

CREATE TABLE IF NOT EXISTS categories (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT    NOT NULL UNIQUE,
    description     TEXT,
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TABLE IF NOT EXISTS items (
    id                  INTEGER PRIMARY KEY AUTOINCREMENT,
    sku                 TEXT    NOT NULL UNIQUE,
    name                TEXT    NOT NULL,
    -- 'tool' | 'consumable' | 'machine'
    item_type           TEXT    NOT NULL CHECK (item_type IN ('tool', 'consumable', 'machine')),
    category_id         INTEGER REFERENCES categories(id) ON DELETE SET NULL,
    -- Stock actual (para herramientas y consumibles)
    stock               REAL    NOT NULL DEFAULT 0 CHECK (stock >= 0),
    min_stock           REAL    NOT NULL DEFAULT 0 CHECK (min_stock >= 0),
    unit                TEXT    NOT NULL DEFAULT 'pza',   -- pza, m, kg, lt, etc.
    -- Estado de mantenimiento: 'ok' | 'scheduled' | 'in_maintenance' | 'critical' | 'out_of_service'
    maintenance_status  TEXT    NOT NULL DEFAULT 'ok'
                            CHECK (maintenance_status IN ('ok', 'scheduled', 'in_maintenance', 'critical', 'out_of_service')),
    location            TEXT,   -- Ej: 'Estante A-3', 'Área de fresadoras'
    -- JSON flexible para datos específicos del módulo Pesados:
    -- número de serie, potencia, modelo, restricciones de uso, observaciones de seguridad, etc.
    module_data         TEXT    DEFAULT '{}',
    is_active           INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at          TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at          TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TRIGGER IF NOT EXISTS items_updated_at
    AFTER UPDATE ON items
    FOR EACH ROW
BEGIN
    UPDATE items SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE id = OLD.id;
END;

-- RB8: No debe aparecer disponible si está en estado crítico
-- Se implementa a nivel de query, pero este índice ayuda al filtrado
CREATE INDEX IF NOT EXISTS idx_items_maintenance_status ON items(maintenance_status);
CREATE INDEX IF NOT EXISTS idx_items_item_type ON items(item_type);
CREATE INDEX IF NOT EXISTS idx_items_category ON items(category_id);


-- =============================================================================
-- MÓDULO: PRÁCTICAS Y SESIONES
-- =============================================================================

CREATE TABLE IF NOT EXISTS assignments (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT    NOT NULL,
    description     TEXT,
    instructions    TEXT,   -- Puede ser texto largo o markdown
    teacher_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    group_name      TEXT    NOT NULL,
    -- 'draft' | 'published' | 'closed'
    status          TEXT    NOT NULL DEFAULT 'draft'
                        CHECK (status IN ('draft', 'published', 'closed')),
    scheduled_date  TEXT,   -- Fecha esperada de realización (ISO 8601)
    published_at    TEXT,
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE TRIGGER IF NOT EXISTS assignments_updated_at
    AFTER UPDATE ON assignments
    FOR EACH ROW
BEGIN
    UPDATE assignments SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE id = OLD.id;
END;

CREATE INDEX IF NOT EXISTS idx_assignments_teacher ON assignments(teacher_id);
CREATE INDEX IF NOT EXISTS idx_assignments_group ON assignments(group_name);
CREATE INDEX IF NOT EXISTS idx_assignments_status ON assignments(status);

-- Recursos sugeridos / requeridos para una práctica (referencia informativa al alumno)
CREATE TABLE IF NOT EXISTS assignment_items (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id   INTEGER NOT NULL REFERENCES assignments(id) ON DELETE CASCADE,
    item_id         INTEGER NOT NULL REFERENCES items(id) ON DELETE CASCADE,
    quantity        REAL    NOT NULL DEFAULT 1 CHECK (quantity > 0),
    notes           TEXT,
    UNIQUE(assignment_id, item_id)
);

CREATE TABLE IF NOT EXISTS lab_sessions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id   INTEGER NOT NULL REFERENCES assignments(id) ON DELETE RESTRICT,
    title           TEXT    NOT NULL,
    teacher_id      INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    group_name      TEXT    NOT NULL,
    scheduled_start TEXT    NOT NULL,  -- ISO 8601
    scheduled_end   TEXT    NOT NULL,  -- ISO 8601
    -- 'planned' | 'open' | 'closed' | 'cancelled'
    status          TEXT    NOT NULL DEFAULT 'planned'
                        CHECK (status IN ('planned', 'open', 'closed', 'cancelled')),
    notes           TEXT,
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    updated_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    CHECK (scheduled_end > scheduled_start)
);

CREATE TRIGGER IF NOT EXISTS lab_sessions_updated_at
    AFTER UPDATE ON lab_sessions
    FOR EACH ROW
BEGIN
    UPDATE lab_sessions SET updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE id = OLD.id;
END;

CREATE INDEX IF NOT EXISTS idx_sessions_assignment ON lab_sessions(assignment_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON lab_sessions(status);


-- =============================================================================
-- MÓDULO: ASISTENCIA
-- =============================================================================

CREATE TABLE IF NOT EXISTS attendance (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      INTEGER NOT NULL REFERENCES lab_sessions(id) ON DELETE RESTRICT,
    student_id      INTEGER NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
    -- 'present' | 'late' | 'absent' | 'excused'
    status          TEXT    NOT NULL DEFAULT 'present'
                        CHECK (status IN ('present', 'late', 'absent', 'excused')),
    check_in_at     TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    notes           TEXT,
    recorded_by     INTEGER REFERENCES users(id) ON DELETE SET NULL,  -- Quién registró (prof u operador)
    UNIQUE(session_id, student_id)  -- Un alumno, un registro por sesión
);

CREATE INDEX IF NOT EXISTS idx_attendance_session ON attendance(session_id);
CREATE INDEX IF NOT EXISTS idx_attendance_student ON attendance(student_id);


-- =============================================================================
-- MÓDULO: SOLICITUDES Y PRÉSTAMOS DE RECURSOS
-- =============================================================================

CREATE TABLE IF NOT EXISTS resource_requests (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      INTEGER REFERENCES lab_sessions(id) ON DELETE SET NULL,
    assignment_id   INTEGER REFERENCES assignments(id) ON DELETE SET NULL,
    student_id      INTEGER NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
    item_id         INTEGER NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    -- 'loan' = préstamo de herramienta | 'consumption' = consumible | 'machine_access' = uso de máquina
    request_type    TEXT    NOT NULL CHECK (request_type IN ('loan', 'consumption', 'machine_access')),
    quantity        REAL    NOT NULL DEFAULT 1 CHECK (quantity > 0),
    -- 'pending' | 'approved' | 'rejected' | 'returned'
    status          TEXT    NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending', 'approved', 'rejected', 'returned')),
    requested_at    TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    resolved_at     TEXT,
    resolved_by     INTEGER REFERENCES users(id) ON DELETE SET NULL,
    notes           TEXT
);

CREATE INDEX IF NOT EXISTS idx_requests_session ON resource_requests(session_id);
CREATE INDEX IF NOT EXISTS idx_requests_student ON resource_requests(student_id);
CREATE INDEX IF NOT EXISTS idx_requests_item ON resource_requests(item_id);
CREATE INDEX IF NOT EXISTS idx_requests_status ON resource_requests(status);


-- =============================================================================
-- MÓDULO: TRANSACCIONES DE INVENTARIO (auditoría histórica)
-- RB5: Toda modificación de stock genera una transacción
-- =============================================================================

CREATE TABLE IF NOT EXISTS transactions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id         INTEGER NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    session_id      INTEGER REFERENCES lab_sessions(id) ON DELETE SET NULL,
    student_id      INTEGER REFERENCES students(id) ON DELETE SET NULL,
    user_id         INTEGER REFERENCES users(id) ON DELETE SET NULL,  -- Quien autorizó
    -- 'loan_out' | 'loan_return' | 'consumption' | 'adjustment' | 'maintenance_hold' | 'initial_stock'
    type            TEXT    NOT NULL
                        CHECK (type IN ('loan_out', 'loan_return', 'consumption', 'adjustment', 'maintenance_hold', 'initial_stock')),
    -- Positivo = entrada, Negativo = salida
    quantity        REAL    NOT NULL,
    -- Stock resultante después de la transacción (snapshot para auditoría rápida)
    stock_after     REAL    NOT NULL,
    -- Referencia al origen (resource_requests, maintenance_logs, etc.)
    reference_type  TEXT,
    reference_id    INTEGER,
    notes           TEXT,
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_item ON transactions(item_id);
CREATE INDEX IF NOT EXISTS idx_transactions_session ON transactions(session_id);
CREATE INDEX IF NOT EXISTS idx_transactions_student ON transactions(student_id);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);
CREATE INDEX IF NOT EXISTS idx_transactions_created ON transactions(created_at);


-- =============================================================================
-- MÓDULO: USO DE MAQUINARIA FIJA
-- RB4: Siempre vinculado a sesión, alumno y supervisor
-- =============================================================================

CREATE TABLE IF NOT EXISTS equipment_usage (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id      INTEGER NOT NULL REFERENCES lab_sessions(id) ON DELETE RESTRICT,
    item_id         INTEGER NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    student_id      INTEGER NOT NULL REFERENCES students(id) ON DELETE RESTRICT,
    supervisor_id   INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    started_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    ended_at        TEXT,
    -- 'active' | 'completed' | 'interrupted'
    status          TEXT    NOT NULL DEFAULT 'active'
                        CHECK (status IN ('active', 'completed', 'interrupted')),
    notes           TEXT,
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_usage_session ON equipment_usage(session_id);
CREATE INDEX IF NOT EXISTS idx_usage_item ON equipment_usage(item_id);
CREATE INDEX IF NOT EXISTS idx_usage_student ON equipment_usage(student_id);
-- Para trazabilidad rápida del último uso (RF16)
CREATE INDEX IF NOT EXISTS idx_usage_item_started ON equipment_usage(item_id, started_at DESC);


-- =============================================================================
-- MÓDULO: MANUALES PDF
-- =============================================================================

CREATE TABLE IF NOT EXISTS manuals (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    title           TEXT    NOT NULL,
    description     TEXT,
    -- Puede vincularse a una máquina/herramienta, a una práctica, o a ambas
    item_id         INTEGER REFERENCES items(id) ON DELETE SET NULL,
    assignment_id   INTEGER REFERENCES assignments(id) ON DELETE SET NULL,
    -- Ruta local relativa al directorio de almacenamiento de la app
    file_path       TEXT    NOT NULL,
    file_size_kb    INTEGER,
    uploaded_by     INTEGER REFERENCES users(id) ON DELETE SET NULL,
    version         TEXT    DEFAULT '1.0',
    is_active       INTEGER NOT NULL DEFAULT 1 CHECK (is_active IN (0, 1)),
    created_at      TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_manuals_item ON manuals(item_id);
CREATE INDEX IF NOT EXISTS idx_manuals_assignment ON manuals(assignment_id);


-- =============================================================================
-- MÓDULO: BITÁCORA DE MANTENIMIENTO
-- =============================================================================

CREATE TABLE IF NOT EXISTS maintenance_logs (
    id                      INTEGER PRIMARY KEY AUTOINCREMENT,
    item_id                 INTEGER NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    user_id                 INTEGER NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    entry_date              TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    -- 'preventive' | 'corrective' | 'inspection' | 'calibration' | 'other'
    maintenance_type        TEXT    NOT NULL
                                CHECK (maintenance_type IN ('preventive', 'corrective', 'inspection', 'calibration', 'other')),
    description             TEXT    NOT NULL,
    -- Estado del equipo DESPUÉS de este mantenimiento
    status_after            TEXT    NOT NULL
                                CHECK (status_after IN ('ok', 'scheduled', 'in_maintenance', 'critical', 'out_of_service')),
    next_maintenance_due    TEXT,   -- Fecha estimada del próximo mantenimiento (ISO 8601)
    -- Costo aproximado si aplica (útil para reportes futuros)
    cost_estimate           REAL,
    created_at              TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_maintenance_item ON maintenance_logs(item_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_date ON maintenance_logs(entry_date);

-- Trigger: al registrar mantenimiento, actualizar estado del item automáticamente
CREATE TRIGGER IF NOT EXISTS sync_item_maintenance_status
    AFTER INSERT ON maintenance_logs
    FOR EACH ROW
BEGIN
    UPDATE items
    SET maintenance_status = NEW.status_after,
        updated_at = strftime('%Y-%m-%dT%H:%M:%SZ', 'now')
    WHERE id = NEW.item_id;
END;


-- =============================================================================
-- MÓDULO: INCIDENCIAS Y FALLOS
-- RB6: Toda incidencia vinculada al recurso afectado
-- RB7: related_previous_student_id es informativo, no implica responsabilidad
-- =============================================================================

CREATE TABLE IF NOT EXISTS incident_reports (
    id                          INTEGER PRIMARY KEY AUTOINCREMENT,
    session_id                  INTEGER REFERENCES lab_sessions(id) ON DELETE SET NULL,
    item_id                     INTEGER NOT NULL REFERENCES items(id) ON DELETE RESTRICT,
    -- Alumno que reporta
    reported_by_student_id      INTEGER REFERENCES students(id) ON DELETE SET NULL,
    -- Último usuario registrado del recurso (trazabilidad, no culpabilidad — RB7)
    related_previous_student_id INTEGER REFERENCES students(id) ON DELETE SET NULL,
    -- Profesor o técnico asignado al seguimiento
    supervisor_id               INTEGER REFERENCES users(id) ON DELETE SET NULL,
    description                 TEXT    NOT NULL,
    -- 'low' | 'medium' | 'high' | 'critical'
    severity                    TEXT    NOT NULL DEFAULT 'medium'
                                    CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    -- 'open' | 'in_review' | 'resolved' | 'dismissed'
    status                      TEXT    NOT NULL DEFAULT 'open'
                                    CHECK (status IN ('open', 'in_review', 'resolved', 'dismissed')),
    resolution_notes            TEXT,
    created_at                  TEXT    NOT NULL DEFAULT (strftime('%Y-%m-%dT%H:%M:%SZ', 'now')),
    resolved_at                 TEXT
);

CREATE INDEX IF NOT EXISTS idx_incidents_item ON incident_reports(item_id);
CREATE INDEX IF NOT EXISTS idx_incidents_session ON incident_reports(session_id);
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incident_reports(status);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incident_reports(severity);


-- =============================================================================
-- VISTAS ÚTILES (no afectan el esquema, facilitan queries comunes)
-- =============================================================================

-- Vista: último uso de cada máquina (para trazabilidad RF16)
CREATE VIEW IF NOT EXISTS v_last_equipment_usage AS
SELECT
    i.id            AS item_id,
    i.name          AS item_name,
    i.sku,
    eu.student_id,
    s.full_name     AS student_name,
    s.student_code,
    eu.started_at,
    eu.ended_at,
    eu.session_id,
    eu.status       AS usage_status
FROM items i
LEFT JOIN equipment_usage eu ON eu.id = (
    SELECT id FROM equipment_usage
    WHERE item_id = i.id
    ORDER BY started_at DESC
    LIMIT 1
)
LEFT JOIN students s ON s.id = eu.student_id
WHERE i.item_type = 'machine';

-- Vista: inventario con alerta de stock mínimo
CREATE VIEW IF NOT EXISTS v_inventory_alerts AS
SELECT
    i.id,
    i.sku,
    i.name,
    i.item_type,
    c.name          AS category,
    i.stock,
    i.min_stock,
    i.unit,
    i.maintenance_status,
    i.location,
    CASE WHEN i.stock <= i.min_stock THEN 1 ELSE 0 END AS below_min_stock
FROM items i
LEFT JOIN categories c ON c.id = i.category_id
WHERE i.is_active = 1;

-- Vista: resumen de asistencia por sesión
CREATE VIEW IF NOT EXISTS v_session_attendance_summary AS
SELECT
    ls.id           AS session_id,
    ls.title        AS session_title,
    ls.group_name,
    ls.scheduled_start,
    COUNT(a.id)                                         AS total_registered,
    SUM(CASE WHEN a.status = 'present' THEN 1 ELSE 0 END) AS present,
    SUM(CASE WHEN a.status = 'late'    THEN 1 ELSE 0 END) AS late,
    SUM(CASE WHEN a.status = 'absent'  THEN 1 ELSE 0 END) AS absent,
    SUM(CASE WHEN a.status = 'excused' THEN 1 ELSE 0 END) AS excused
FROM lab_sessions ls
LEFT JOIN attendance a ON a.session_id = ls.id
GROUP BY ls.id;


-- =============================================================================
-- DATOS SEMILLA MÍNIMOS (categorías base y usuario admin)
-- =============================================================================

INSERT OR IGNORE INTO categories (name, description) VALUES
    ('Herramienta de corte',    'Fresas, brocas, cuchillas y similares'),
    ('Herramienta de medición', 'Micrómetros, calibradores, comparadores'),
    ('Consumibles',             'Refrigerantes, lubricantes, material de limpieza'),
    ('Maquinaria CNC',          'Fresadoras, tornos y centros de maquinado CNC'),
    ('Maquinaria convencional', 'Tornos y fresadoras manuales'),
    ('Equipo de seguridad',     'Lentes, guantes, caretas y protección personal');

-- Usuario administrador por defecto (cambiar contraseña en primer uso)
-- password: 'admin123' — REEMPLAZAR con hash bcrypt real en producción
INSERT OR IGNORE INTO users (username, password_hash, full_name, role) VALUES
    ('admin', '$2a$10$PLACEHOLDER_REPLACE_WITH_REAL_BCRYPT_HASH', 'Administrador del Sistema', 'admin');
