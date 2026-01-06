package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type HRZoneDefinition struct {
	AthleteID     int64           `json:"-"`
	SportType     string          `json:"sport_type"`
	EffectiveFrom string          `json:"effective_from"` // YYYY-MM-DD
	Method        string          `json:"method"`         // 'absolute_bpm' | 'percent_hrmax'
	Zones         json.RawMessage `json:"zones"`          // {"bounds":[...], "hr_max":...}
}

type ZonesRepository struct {
	db *DB
}

func NewZonesRepository(db *DB) *ZonesRepository {
	return &ZonesRepository{db: db}
}

const hrZoneDefinitionsTable = "hr_zone_definitions"

func (r *ZonesRepository) ensureHRZoneDefinitionsTable(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS hr_zone_definitions (
			athlete_id INTEGER NOT NULL REFERENCES athletes(id),
			sport_type TEXT NOT NULL,
			effective_from TEXT NOT NULL,
			method TEXT NOT NULL,
			zones TEXT NOT NULL,
			PRIMARY KEY (athlete_id, sport_type, effective_from)
		)
	`)
	if err != nil {
		return err
	}
	_, _ = r.db.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS idx_hr_zones_athlete_sport ON hr_zone_definitions(athlete_id, sport_type)`)
	return nil
}

func (r *ZonesRepository) ListHR(ctx context.Context, athleteID int64) ([]HRZoneDefinition, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT athlete_id, sport_type, effective_from, method, zones
		FROM hr_zone_definitions
		WHERE athlete_id = ?
		ORDER BY sport_type ASC, effective_from DESC
	`, athleteID)
	if err != nil {
		if isMissingTable(err, hrZoneDefinitionsTable) {
			if err = r.ensureHRZoneDefinitionsTable(ctx); err != nil { //nolint:gocritic // sloppyReassign: using = to avoid shadow
				return nil, err
			}
			rows, err = r.db.QueryContext(ctx, `
				SELECT athlete_id, sport_type, effective_from, method, zones
				FROM hr_zone_definitions
				WHERE athlete_id = ?
				ORDER BY sport_type ASC, effective_from DESC
			`, athleteID)
		}
		if err != nil {
			return nil, err
		}
	}
	defer func() { _ = rows.Close() }()

	defs := make([]HRZoneDefinition, 0)
	for rows.Next() {
		var d HRZoneDefinition
		var zonesStr string
		if err := rows.Scan(&d.AthleteID, &d.SportType, &d.EffectiveFrom, &d.Method, &zonesStr); err != nil {
			return nil, err
		}
		d.Zones = json.RawMessage(zonesStr)
		defs = append(defs, d)
	}
	return defs, rows.Err()
}

func (r *ZonesRepository) UpsertHR(ctx context.Context, athleteID int64, def HRZoneDefinition) error {
	if def.SportType == "" {
		return fmt.Errorf("sport_type is required")
	}
	if def.EffectiveFrom == "" {
		def.EffectiveFrom = "1970-01-01"
	}
	eff, err := time.Parse("2006-01-02", def.EffectiveFrom)
	if err != nil {
		return fmt.Errorf("invalid effective_from")
	}
	if def.Method != "absolute_bpm" && def.Method != "percent_hrmax" {
		return fmt.Errorf("invalid method")
	}
	if len(def.Zones) == 0 {
		return fmt.Errorf("zones is required")
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO hr_zone_definitions (athlete_id, sport_type, effective_from, method, zones)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT (athlete_id, sport_type, effective_from) DO UPDATE SET
			method = EXCLUDED.method,
			zones = EXCLUDED.zones
	`, athleteID, def.SportType, eff, def.Method, def.Zones)
	if err != nil && isMissingTable(err, hrZoneDefinitionsTable) {
		if err2 := r.ensureHRZoneDefinitionsTable(ctx); err2 != nil {
			return err2
		}
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO hr_zone_definitions (athlete_id, sport_type, effective_from, method, zones)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT (athlete_id, sport_type, effective_from) DO UPDATE SET
				method = EXCLUDED.method,
				zones = EXCLUDED.zones
		`, athleteID, def.SportType, eff, def.Method, def.Zones)
	}
	if err != nil {
		return err
	}

	// Invalidate zone distributions for affected activities.
	// This triggers recomputation on next zone trend access.
	q := NewQueries(r.db.Conn())
	if err := q.DeleteZoneDistributionsForSportType(ctx, athleteID, def.SportType, eff.Format("2006-01-02")); err != nil {
		return fmt.Errorf("invalidate zone distributions: %w", err)
	}

	return nil
}

func (r *ZonesRepository) DeleteHR(ctx context.Context, athleteID int64, sportType, effectiveFrom string) error {
	if sportType == "" || effectiveFrom == "" {
		return fmt.Errorf("sport_type and effective_from are required")
	}
	eff, err := time.Parse("2006-01-02", effectiveFrom)
	if err != nil {
		return fmt.Errorf("invalid effective_from")
	}
	_, err = r.db.ExecContext(ctx, `
		DELETE FROM hr_zone_definitions
		WHERE athlete_id = ? AND sport_type = ? AND effective_from = ?
	`, athleteID, sportType, eff)
	if err != nil && isMissingTable(err, hrZoneDefinitionsTable) {
		if err2 := r.ensureHRZoneDefinitionsTable(ctx); err2 != nil {
			return err2
		}
		_, err = r.db.ExecContext(ctx, `
			DELETE FROM hr_zone_definitions
			WHERE athlete_id = ? AND sport_type = ? AND effective_from = ?
		`, athleteID, sportType, eff)
	}
	if err != nil {
		return err
	}

	// Invalidate zone distributions for affected activities.
	// Activities from effective_from onwards may now use a different zone definition.
	q := NewQueries(r.db.Conn())
	if err := q.DeleteZoneDistributionsForSportType(ctx, athleteID, sportType, eff.Format("2006-01-02")); err != nil {
		return fmt.Errorf("invalidate zone distributions: %w", err)
	}

	return nil
}

type HRZoneConfig struct {
	Bounds []float64 `json:"bounds"`
	HRMax  float64   `json:"hr_max,omitempty"`
}

func (r *ZonesRepository) GetApplicableHR(ctx context.Context, athleteID int64, sportType string, at time.Time) (*HRZoneDefinition, *HRZoneConfig, error) {
	// Prefer sport-specific, fall back to "All".
	for _, st := range []string{sportType, "All"} {
		var def HRZoneDefinition
		var effective string
		var zonesStr string
		err := r.db.QueryRowContext(ctx, `
			SELECT sport_type, effective_from, method, zones
			FROM hr_zone_definitions
			WHERE athlete_id = ? AND sport_type = ? AND effective_from <= date(?)
			ORDER BY effective_from DESC
			LIMIT 1
		`, athleteID, st, at.Format("2006-01-02")).Scan(&def.SportType, &effective, &def.Method, &zonesStr)
		if err != nil && isMissingTable(err, hrZoneDefinitionsTable) {
			if err2 := r.ensureHRZoneDefinitionsTable(ctx); err2 != nil {
				return nil, nil, err2
			}
			err = r.db.QueryRowContext(ctx, `
				SELECT sport_type, effective_from, method, zones
				FROM hr_zone_definitions
				WHERE athlete_id = ? AND sport_type = ? AND effective_from <= date(?)
				ORDER BY effective_from DESC
				LIMIT 1
			`, athleteID, st, at.Format("2006-01-02")).Scan(&def.SportType, &effective, &def.Method, &zonesStr)
		}
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return nil, nil, err
		}
		def.AthleteID = athleteID
		def.EffectiveFrom = effective
		def.Zones = json.RawMessage(zonesStr)
		var cfg HRZoneConfig
		if err := json.Unmarshal(def.Zones, &cfg); err != nil {
			return &def, nil, fmt.Errorf("invalid zones config")
		}
		return &def, &cfg, nil
	}
	return nil, nil, nil
}

func HRZoneIndex(method string, cfg *HRZoneConfig, hr float64) int {
	if cfg == nil {
		return -1
	}
	if hr <= 0 {
		return -1
	}
	switch method {
	case "absolute_bpm":
		for i, bound := range cfg.Bounds {
			if hr <= bound {
				return i
			}
		}
		return len(cfg.Bounds) - 1
	case "percent_hrmax":
		if cfg.HRMax <= 0 {
			return -1
		}
		p := hr / cfg.HRMax
		for i, bound := range cfg.Bounds {
			if p <= bound {
				return i
			}
		}
		return len(cfg.Bounds) - 1
	default:
		return -1
	}
}
