package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"mapup-geofence/internal/domain"
	"strings"
	"time"

	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository connects to PostgreSQL and applies schema migrations
func NewPostgresRepository(connStr string) (*PostgresRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("unable to open database connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	repo := &PostgresRepository{db: db}
	if err := repo.initSchema(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

func (p *PostgresRepository) Close() error {
	return p.db.Close()
}

func (p *PostgresRepository) initSchema(ctx context.Context) error {
	queries := []string{
		`CREATE EXTENSION IF NOT EXISTS postgis;`,
		`CREATE TABLE IF NOT EXISTS geofences (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			category VARCHAR(64) NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			geom GEOMETRY(Polygon, 4326) NOT NULL,
			coordinates_json TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_geofences_geom ON geofences USING GIST (geom);`,
		`CREATE INDEX IF NOT EXISTS idx_geofences_category ON geofences (category);`,

		`CREATE TABLE IF NOT EXISTS vehicles (
			id VARCHAR(64) PRIMARY KEY,
			vehicle_number VARCHAR(64) UNIQUE NOT NULL,
			driver_name VARCHAR(255) NOT NULL,
			vehicle_type VARCHAR(64) NOT NULL,
			phone VARCHAR(64) NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS vehicle_locations (
			id BIGSERIAL PRIMARY KEY,
			vehicle_id VARCHAR(64) NOT NULL REFERENCES vehicles(id) ON DELETE CASCADE,
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			geom GEOMETRY(Point, 4326) NOT NULL,
			timestamp TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_locations_vehicle_time ON vehicle_locations (vehicle_id, timestamp DESC);`,

		`CREATE TABLE IF NOT EXISTS vehicle_current_geofences (
			vehicle_id VARCHAR(64) PRIMARY KEY REFERENCES vehicles(id) ON DELETE CASCADE,
			geofence_ids TEXT[] NOT NULL DEFAULT '{}',
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS alert_rules (
			alert_id VARCHAR(64) PRIMARY KEY,
			geofence_id VARCHAR(64) NOT NULL REFERENCES geofences(id) ON DELETE CASCADE,
			vehicle_id VARCHAR(64) REFERENCES vehicles(id) ON DELETE CASCADE,
			event_type VARCHAR(32) NOT NULL,
			status VARCHAR(32) NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,

		`CREATE TABLE IF NOT EXISTS violations (
			id VARCHAR(64) PRIMARY KEY,
			vehicle_id VARCHAR(64) NOT NULL,
			vehicle_number VARCHAR(64) NOT NULL,
			geofence_id VARCHAR(64) NOT NULL,
			geofence_name VARCHAR(255) NOT NULL,
			event_type VARCHAR(32) NOT NULL,
			latitude DOUBLE PRECISION NOT NULL,
			longitude DOUBLE PRECISION NOT NULL,
			timestamp TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);`,
		`CREATE INDEX IF NOT EXISTS idx_violations_vehicle_geofence ON violations (vehicle_id, geofence_id, timestamp DESC);`,
	}

	for _, q := range queries {
		if _, err := p.db.ExecContext(ctx, q); err != nil {
			return fmt.Errorf("error executing schema statement: %w", err)
		}
	}
	return nil
}

// buildPolygonWKT formats domain Coordinates [lat, lon] to PostGIS WKT POLYGON((lon lat, ...))
func buildPolygonWKT(coords []domain.Coordinate) string {
	parts := make([]string, len(coords))
	for i, c := range coords {
		// Notice: PostGIS uses Longitude Latitude (X Y)
		parts[i] = fmt.Sprintf("%f %f", c.Lon(), c.Lat())
	}
	return fmt.Sprintf("SRID=4326;POLYGON((%s))", strings.Join(parts, ", "))
}

func (p *PostgresRepository) CreateGeofence(ctx context.Context, g *domain.Geofence) error {
	coordsBytes, err := json.Marshal(g.Coordinates)
	if err != nil {
		return err
	}
	wkt := buildPolygonWKT(g.Coordinates)

	query := `INSERT INTO geofences (id, name, description, category, status, geom, coordinates_json, created_at)
		VALUES ($1, $2, $3, $4, $5, ST_GeomFromEWKT($6), $7, $8)`

	_, err = p.db.ExecContext(ctx, query, g.ID, g.Name, g.Description, g.Category, g.Status, wkt, string(coordsBytes), g.CreatedAt)
	return err
}

func (p *PostgresRepository) GetGeofences(ctx context.Context, category string) ([]domain.Geofence, error) {
	query := `SELECT id, name, description, category, status, coordinates_json, created_at FROM geofences`
	args := []interface{}{}
	if category != "" {
		query += ` WHERE LOWER(category) = LOWER($1)`
		args = append(args, category)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Geofence, 0)
	for rows.Next() {
		var g domain.Geofence
		var coordsJSON string
		var desc sql.NullString
		if err := rows.Scan(&g.ID, &g.Name, &desc, &g.Category, &g.Status, &coordsJSON, &g.CreatedAt); err != nil {
			return nil, err
		}
		if desc.Valid {
			g.Description = desc.String
		}
		_ = json.Unmarshal([]byte(coordsJSON), &g.Coordinates)
		result = append(result, g)
	}
	return result, nil
}

func (p *PostgresRepository) GetGeofenceByID(ctx context.Context, id string) (*domain.Geofence, error) {
	query := `SELECT id, name, description, category, status, coordinates_json, created_at FROM geofences WHERE id = $1`
	row := p.db.QueryRowContext(ctx, query, id)

	var g domain.Geofence
	var coordsJSON string
	var desc sql.NullString
	if err := row.Scan(&g.ID, &g.Name, &desc, &g.Category, &g.Status, &coordsJSON, &g.CreatedAt); err != nil {
		return nil, err
	}
	if desc.Valid {
		g.Description = desc.String
	}
	_ = json.Unmarshal([]byte(coordsJSON), &g.Coordinates)
	return &g, nil
}

func (p *PostgresRepository) GetAllActiveGeofences(ctx context.Context) ([]domain.Geofence, error) {
	return p.GetGeofences(ctx, "")
}

func (p *PostgresRepository) CreateVehicle(ctx context.Context, v *domain.Vehicle) error {
	query := `INSERT INTO vehicles (id, vehicle_number, driver_name, vehicle_type, phone, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := p.db.ExecContext(ctx, query, v.ID, v.VehicleNumber, v.DriverName, v.VehicleType, v.Phone, v.Status, v.CreatedAt)
	return err
}

func (p *PostgresRepository) GetVehicles(ctx context.Context) ([]domain.Vehicle, error) {
	query := `SELECT id, vehicle_number, driver_name, vehicle_type, phone, status, created_at FROM vehicles ORDER BY created_at DESC`
	rows, err := p.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]domain.Vehicle, 0)
	for rows.Next() {
		var v domain.Vehicle
		if err := rows.Scan(&v.ID, &v.VehicleNumber, &v.DriverName, &v.VehicleType, &v.Phone, &v.Status, &v.CreatedAt); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (p *PostgresRepository) GetVehicleByID(ctx context.Context, id string) (*domain.Vehicle, error) {
	query := `SELECT id, vehicle_number, driver_name, vehicle_type, phone, status, created_at FROM vehicles WHERE id = $1`
	row := p.db.QueryRowContext(ctx, query, id)

	var v domain.Vehicle
	if err := row.Scan(&v.ID, &v.VehicleNumber, &v.DriverName, &v.VehicleType, &v.Phone, &v.Status, &v.CreatedAt); err != nil {
		return nil, err
	}
	return &v, nil
}

func (p *PostgresRepository) RecordLocation(ctx context.Context, loc *domain.LocationRecord) error {
	query := `INSERT INTO vehicle_locations (vehicle_id, latitude, longitude, geom, timestamp)
		VALUES ($1, $2, $3, ST_SetSRID(ST_MakePoint($3, $2), 4326), $4)`
	_, err := p.db.ExecContext(ctx, query, loc.VehicleID, loc.Latitude, loc.Longitude, loc.Timestamp)
	return err
}

func (p *PostgresRepository) GetLatestLocation(ctx context.Context, vehicleID string) (*domain.LocationRecord, error) {
	query := `SELECT vehicle_id, latitude, longitude, timestamp FROM vehicle_locations
		WHERE vehicle_id = $1 ORDER BY timestamp DESC LIMIT 1`
	row := p.db.QueryRowContext(ctx, query, vehicleID)

	var loc domain.LocationRecord
	if err := row.Scan(&loc.VehicleID, &loc.Latitude, &loc.Longitude, &loc.Timestamp); err != nil {
		return nil, err
	}
	return &loc, nil
}

func (p *PostgresRepository) GetVehicleCurrentGeofenceIDs(ctx context.Context, vehicleID string) ([]string, error) {
	query := `SELECT geofence_ids FROM vehicle_current_geofences WHERE vehicle_id = $1`
	row := p.db.QueryRowContext(ctx, query, vehicleID)

	var ids []string
	if err := row.Scan(pq.Array(&ids)); err != nil {
		if err == sql.ErrNoRows {
			return []string{}, nil
		}
		return nil, err
	}
	return ids, nil
}

func (p *PostgresRepository) UpdateVehicleCurrentGeofenceIDs(ctx context.Context, vehicleID string, geofenceIDs []string) error {
	query := `INSERT INTO vehicle_current_geofences (vehicle_id, geofence_ids, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (vehicle_id) DO UPDATE SET geofence_ids = EXCLUDED.geofence_ids, updated_at = NOW()`
	_, err := p.db.ExecContext(ctx, query, vehicleID, pq.Array(geofenceIDs))
	return err
}

func (p *PostgresRepository) CreateAlertRule(ctx context.Context, rule *domain.AlertRule) error {
	query := `INSERT INTO alert_rules (alert_id, geofence_id, vehicle_id, event_type, status, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := p.db.ExecContext(ctx, query, rule.AlertID, rule.GeofenceID, rule.VehicleID, rule.EventType, rule.Status, rule.CreatedAt)
	return err
}

func (p *PostgresRepository) GetAlertRules(ctx context.Context, geofenceID, vehicleID string) ([]domain.AlertRule, error) {
	query := `SELECT r.alert_id, r.geofence_id, g.name, r.vehicle_id, v.vehicle_number, r.event_type, r.status, r.created_at
		FROM alert_rules r
		JOIN geofences g ON r.geofence_id = g.id
		LEFT JOIN vehicles v ON r.vehicle_id = v.id
		WHERE 1=1`
	args := []interface{}{}
	idx := 1

	if geofenceID != "" {
		query += fmt.Sprintf(" AND r.geofence_id = $%d", idx)
		args = append(args, geofenceID)
		idx++
	}
	if vehicleID != "" {
		query += fmt.Sprintf(" AND (r.vehicle_id = $%d OR r.vehicle_id IS NULL)", idx)
		args = append(args, vehicleID)
		idx++
	}
	query += ` ORDER BY r.created_at DESC`

	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	rules := make([]domain.AlertRule, 0)
	for rows.Next() {
		var r domain.AlertRule
		var gName sql.NullString
		var vID sql.NullString
		var vNum sql.NullString

		if err := rows.Scan(&r.AlertID, &r.GeofenceID, &gName, &vID, &vNum, &r.EventType, &r.Status, &r.CreatedAt); err != nil {
			return nil, err
		}
		if gName.Valid {
			r.GeofenceName = gName.String
		}
		if vID.Valid {
			r.VehicleID = &vID.String
		}
		if vNum.Valid {
			r.VehicleNumber = &vNum.String
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func (p *PostgresRepository) GetAllAlertRules(ctx context.Context) ([]domain.AlertRule, error) {
	return p.GetAlertRules(ctx, "", "")
}

func (p *PostgresRepository) RecordViolation(ctx context.Context, v *domain.Violation) error {
	query := `INSERT INTO violations (id, vehicle_id, vehicle_number, geofence_id, geofence_name, event_type, latitude, longitude, timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())`
	_, err := p.db.ExecContext(ctx, query, v.ID, v.VehicleID, v.VehicleNumber, v.GeofenceID, v.GeofenceName, v.EventType, v.Latitude, v.Longitude, v.Timestamp)
	return err
}

func (p *PostgresRepository) GetViolations(ctx context.Context, vehicleID, geofenceID string, startDate, endDate *time.Time, limit int) ([]domain.Violation, int, error) {
	whereClauses := []string{"1=1"}
	args := []interface{}{}
	idx := 1

	if vehicleID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("vehicle_id = $%d", idx))
		args = append(args, vehicleID)
		idx++
	}
	if geofenceID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("geofence_id = $%d", idx))
		args = append(args, geofenceID)
		idx++
	}
	if startDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp >= $%d", idx))
		args = append(args, *startDate)
		idx++
	}
	if endDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("timestamp <= $%d", idx))
		args = append(args, *endDate)
		idx++
	}

	whereSQL := strings.Join(whereClauses, " AND ")

	// Total count query
	countQuery := "SELECT COUNT(*) FROM violations WHERE " + whereSQL
	var totalCount int
	if err := p.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	dataQuery := fmt.Sprintf("SELECT id, vehicle_id, vehicle_number, geofence_id, geofence_name, event_type, latitude, longitude, timestamp FROM violations WHERE %s ORDER BY timestamp DESC LIMIT $%d", whereSQL, idx)
	args = append(args, limit)

	rows, err := p.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	result := make([]domain.Violation, 0)
	for rows.Next() {
		var v domain.Violation
		if err := rows.Scan(&v.ID, &v.VehicleID, &v.VehicleNumber, &v.GeofenceID, &v.GeofenceName, &v.EventType, &v.Latitude, &v.Longitude, &v.Timestamp); err != nil {
			return nil, 0, err
		}
		result = append(result, v)
	}
	return result, totalCount, nil
}

// pqStringArray handles PostgreSQL text[] scanning/valuing
type pqStringArray []string

func (a *pqStringArray) Scan(src interface{}) error {
	if src == nil {
		*a = []string{}
		return nil
	}
	switch s := src.(type) {
	case string:
		*a = parsePostgresArray(s)
		return nil
	case []byte:
		*a = parsePostgresArray(string(s))
		return nil
	default:
		return fmt.Errorf("cannot scan %T into pqStringArray", src)
	}
}

func (a pqStringArray) Value() (interface{}, error) {
	if len(a) == 0 {
		return "{}", nil
	}
	escaped := make([]string, len(a))
	for i, val := range a {
		escaped[i] = fmt.Sprintf(`"%s"`, strings.ReplaceAll(val, `"`, `\"`))
	}
	return fmt.Sprintf("{%s}", strings.Join(escaped, ",")), nil
}

func parsePostgresArray(src string) []string {
	src = strings.Trim(src, "{}")
	if src == "" {
		return []string{}
	}
	items := strings.Split(src, ",")
	res := make([]string, len(items))
	for i, it := range items {
		res[i] = strings.Trim(strings.TrimSpace(it), `"`)
	}
	return res
}
