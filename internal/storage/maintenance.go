package storage

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sasha/stata/internal/pagination"
)

type Component struct {
	ID                 int64     `json:"id"`
	GearID             string    `json:"gear_id"`
	Name               string    `json:"name"`
	ImageURL           string    `json:"image_url,omitempty"`
	MaintenanceHashtag string    `json:"maintenance_hashtag,omitempty"`
	CreatedAt            SQLiteTime `json:"created_at"`
	UpdatedAt            SQLiteTime `json:"updated_at"`
}

type MaintenanceRule struct {
	ID             int64     `json:"id"`
	ComponentID    int64     `json:"component_id"`
	Type           string    `json:"type"` // distance_m | time_s | days
	ThresholdValue float64   `json:"threshold_value"`
	CreatedAt            SQLiteTime `json:"created_at"`
	UpdatedAt            SQLiteTime `json:"updated_at"`
}

type MaintenanceLogEntry struct {
	ComponentID int64     `json:"component_id"`
	ActivityID  *int64    `json:"activity_id,omitempty"`
	CompletedAt          SQLiteTime `json:"completed_at"`
}

type ComponentWithRules struct {
	Component
	Rules           []MaintenanceRule `json:"rules"`
	LastCompletedAt *SQLiteTime       `json:"last_completed_at,omitempty"`
}

type RuleProgress struct {
	Type           string  `json:"type"`
	ThresholdValue float64 `json:"threshold_value"`
	CurrentValue   float64 `json:"current_value"`
	Percent        float64 `json:"percent"`
	Due            bool    `json:"due"`
}

type DueComponent struct {
	ComponentWithRules
	DistanceSince   float64        `json:"distance_since"`
	MovingTimeSince int            `json:"moving_time_since"`
	DaysSince       int            `json:"days_since"`
	Progress        []RuleProgress `json:"progress"`
	IsDue           bool           `json:"is_due"`
}

type MaintenanceRepository struct {
	db *DB
}

func NewMaintenanceRepository(db *DB) *MaintenanceRepository {
	return &MaintenanceRepository{db: db}
}

func (r *MaintenanceRepository) ensureGearOwner(ctx context.Context, athleteID int64, gearID string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `SELECT TRUE FROM gear WHERE id = ? AND athlete_id = ?`, gearID, athleteID).Scan(&exists)
	if isNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func normalizeTag(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "#")
	return strings.ToLower(s)
}

var maintTagRe = regexp.MustCompile(`#([A-Za-z0-9][A-Za-z0-9_-]{0,63})`)

func extractMaintenanceTags(text string) []string {
	matches := maintTagRe.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(matches))
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		tag := normalizeTag(m[1])
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func validateRuleType(t string) bool {
	switch t {
	case "distance_m", "time_s", "days":
		return true
	default:
		return false
	}
}

func (r *MaintenanceRepository) ListComponents(ctx context.Context, athleteID int64, gearID string) ([]ComponentWithRules, error) {
	ok, err := r.ensureGearOwner(ctx, athleteID, gearID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			c.id, c.gear_id, c.name, COALESCE(c.image_url, ''), COALESCE(c.maintenance_hashtag, ''),
			c.created_at, c.updated_at,
			(
				SELECT MAX(completed_at)
				FROM maintenance_log ml
				WHERE ml.component_id = c.id
			) AS last_completed_at
		FROM components c
		WHERE c.gear_id = ?
		ORDER BY c.created_at ASC
	`, gearID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []ComponentWithRules
	componentIDs := make([]int64, 0, 16)
	for rows.Next() {
		var it ComponentWithRules
		if err := rows.Scan(
			&it.ID,
			&it.GearID,
			&it.Name,
			&it.ImageURL,
			&it.MaintenanceHashtag,
			&it.CreatedAt,
			&it.UpdatedAt,
			&it.LastCompletedAt,
		); err != nil {
			return nil, err
		}
		if it.MaintenanceHashtag != "" {
			it.MaintenanceHashtag = normalizeTag(it.MaintenanceHashtag)
		}
		componentIDs = append(componentIDs, it.ID)
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(items) == 0 {
		return items, nil
	}

	placeholders := make([]string, 0, len(componentIDs))
	args := make([]any, 0, len(componentIDs))
	for _, id := range componentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}
	rulesRows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, component_id, type, threshold_value, created_at, updated_at
		FROM maintenance_rules
		WHERE component_id IN (%s)
		ORDER BY component_id ASC, id ASC
	`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rulesRows.Close() }()

	rulesByComponent := make(map[int64][]MaintenanceRule, len(componentIDs))
	for rulesRows.Next() {
		var mr MaintenanceRule
		if err := rulesRows.Scan(&mr.ID, &mr.ComponentID, &mr.Type, &mr.ThresholdValue, &mr.CreatedAt, &mr.UpdatedAt); err != nil {
			return nil, err
		}
		rulesByComponent[mr.ComponentID] = append(rulesByComponent[mr.ComponentID], mr)
	}
	if err := rulesRows.Err(); err != nil {
		return nil, err
	}

	for idx := range items {
		items[idx].Rules = rulesByComponent[items[idx].ID]
	}

	return items, nil
}

// ComponentFilters contains pagination parameters for components.
type ComponentFilters struct {
	pagination.QueryParams
}

// ComponentListResult contains paginated component results.
type ComponentListResult struct {
	Items      []ComponentWithRules
	Total      int
	Page       int
	PerPage    int
	TotalPages int
}

// ListComponentsPaginated returns a paginated list of components for a gear item.
func (r *MaintenanceRepository) ListComponentsPaginated(ctx context.Context, athleteID int64, gearID string, f ComponentFilters) (ComponentListResult, error) {
	ok, err := r.ensureGearOwner(ctx, athleteID, gearID)
	if err != nil {
		return ComponentListResult{}, err
	}
	if !ok {
		return ComponentListResult{}, nil
	}

	// Count total
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM components WHERE gear_id = ?`, gearID,
	).Scan(&total); err != nil {
		return ComponentListResult{}, fmt.Errorf("counting components: %w", err)
	}

	// Normalize pagination params
	p := pagination.NewParams(f.Page, f.PerPage)

	// Build ORDER BY with validation
	validOrderBy := map[string]string{
		"name":       "name",
		"created_at": "created_at",
	}
	orderBy := pagination.BuildOrderClause(f.OrderBy, f.OrderDir, validOrderBy, "created_at ASC")
	query := fmt.Sprintf(`
		SELECT
			c.id, c.gear_id, c.name, COALESCE(c.image_url, ''), COALESCE(c.maintenance_hashtag, ''),
			c.created_at, c.updated_at,
			(
				SELECT MAX(completed_at)
				FROM maintenance_log ml
				WHERE ml.component_id = c.id
			) AS last_completed_at
		FROM components c
		WHERE c.gear_id = ?
		ORDER BY %s
		LIMIT ? OFFSET ?
	`, orderBy)

	rows, err := r.db.QueryContext(ctx, query, gearID, p.PerPage, p.Offset())
	if err != nil {
		return ComponentListResult{}, err
	}
	defer func() { _ = rows.Close() }()

	var items []ComponentWithRules
	componentIDs := make([]int64, 0, 16)
	for rows.Next() {
		var it ComponentWithRules
		if err := rows.Scan(
			&it.ID,
			&it.GearID,
			&it.Name,
			&it.ImageURL,
			&it.MaintenanceHashtag,
			&it.CreatedAt,
			&it.UpdatedAt,
			&it.LastCompletedAt,
		); err != nil {
			return ComponentListResult{}, err
		}
		if it.MaintenanceHashtag != "" {
			it.MaintenanceHashtag = normalizeTag(it.MaintenanceHashtag)
		}
		componentIDs = append(componentIDs, it.ID)
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return ComponentListResult{}, err
	}

	// Fetch rules for all components
	if len(items) > 0 {
		placeholders := make([]string, 0, len(componentIDs))
		args := make([]any, 0, len(componentIDs))
		for _, id := range componentIDs {
			placeholders = append(placeholders, "?")
			args = append(args, id)
		}
		rulesRows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
			SELECT id, component_id, type, threshold_value, created_at, updated_at
			FROM maintenance_rules
			WHERE component_id IN (%s)
			ORDER BY component_id ASC, id ASC
		`, strings.Join(placeholders, ",")), args...)
		if err != nil {
			return ComponentListResult{}, err
		}
		defer func() { _ = rulesRows.Close() }()

		rulesByComponent := make(map[int64][]MaintenanceRule, len(componentIDs))
		for rulesRows.Next() {
			var mr MaintenanceRule
			if err := rulesRows.Scan(&mr.ID, &mr.ComponentID, &mr.Type, &mr.ThresholdValue, &mr.CreatedAt, &mr.UpdatedAt); err != nil {
				return ComponentListResult{}, err
			}
			rulesByComponent[mr.ComponentID] = append(rulesByComponent[mr.ComponentID], mr)
		}
		if err := rulesRows.Err(); err != nil {
			return ComponentListResult{}, err
		}

		for idx := range items {
			items[idx].Rules = rulesByComponent[items[idx].ID]
		}
	}

	return ComponentListResult{
		Items:      items,
		Total:      total,
		Page:       p.Page,
		PerPage:    p.PerPage,
		TotalPages: p.TotalPages(total),
	}, nil
}

type CreateComponentInput struct {
	Name               string
	ImageURL           string
	MaintenanceHashtag string
	Rules              []CreateRuleInput
}

type CreateRuleInput struct {
	Type           string
	ThresholdValue float64
}

func (r *MaintenanceRepository) CreateComponent(ctx context.Context, athleteID int64, gearID string, in CreateComponentInput) (*ComponentWithRules, error) {
	ok, err := r.ensureGearOwner(ctx, athleteID, gearID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	name := strings.TrimSpace(in.Name)
	if name == "" {
		return nil, fmt.Errorf("name is required")
	}
	tag := normalizeTag(in.MaintenanceHashtag)

	now := SQLiteTime{Time: time.Now()}
	var id int64
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO components (gear_id, name, image_url, maintenance_hashtag, created_at, updated_at)
		VALUES (?, ?, NULLIF(?, ''), NULLIF(?, ''), ?, ?)
		RETURNING id
	`, gearID, name, strings.TrimSpace(in.ImageURL), tag, now, now).Scan(&id)
	if err != nil {
		return nil, err
	}

	for _, rule := range in.Rules {
		if !validateRuleType(rule.Type) {
			return nil, fmt.Errorf("invalid rule type: %s", rule.Type)
		}
		if rule.ThresholdValue <= 0 {
			return nil, fmt.Errorf("threshold_value must be > 0")
		}
		_, err := r.db.ExecContext(ctx, `
			INSERT INTO maintenance_rules (component_id, type, threshold_value, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?)
		`, id, rule.Type, rule.ThresholdValue, now, now)
		if err != nil {
			return nil, err
		}
	}

	components, err := r.ListComponents(ctx, athleteID, gearID)
	if err != nil {
		return nil, err
	}
	for _, c := range components {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

type UpdateComponentInput struct {
	Name               *string
	ImageURL           *string
	MaintenanceHashtag *string
	Rules              *[]CreateRuleInput
}

func (r *MaintenanceRepository) UpdateComponent(ctx context.Context, athleteID int64, componentID int64, in UpdateComponentInput) (*ComponentWithRules, error) {
	var gearID string
	err := r.db.QueryRowContext(ctx, `
		SELECT c.gear_id
		FROM components c
		JOIN gear g ON g.id = c.gear_id
		WHERE c.id = ? AND g.athlete_id = ?
	`, componentID, athleteID).Scan(&gearID)
	if isNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	now := SQLiteTime{Time: time.Now()}

	if in.Name != nil || in.ImageURL != nil || in.MaintenanceHashtag != nil {
		name := ""
		image := ""
		tag := ""
		if in.Name != nil {
			name = strings.TrimSpace(*in.Name)
		}
		if in.ImageURL != nil {
			image = strings.TrimSpace(*in.ImageURL)
		}
		if in.MaintenanceHashtag != nil {
			tag = normalizeTag(*in.MaintenanceHashtag)
		}
		_, err := r.db.ExecContext(ctx, `
			UPDATE components
			SET
				name = COALESCE(NULLIF(?, ''), name),
				image_url = CASE WHEN ? THEN NULLIF(?, '') ELSE image_url END,
				maintenance_hashtag = CASE WHEN ? THEN NULLIF(?, '') ELSE maintenance_hashtag END,
				updated_at = ?
			WHERE id = ?
		`,
			name,
			in.ImageURL != nil, image,
			in.MaintenanceHashtag != nil, tag,
			now,
			componentID,
		)
		if err != nil {
			return nil, err
		}
	}

	if in.Rules != nil {
		_, err := r.db.ExecContext(ctx, `DELETE FROM maintenance_rules WHERE component_id = ?`, componentID)
		if err != nil {
			return nil, err
		}
		for _, rule := range *in.Rules {
			if !validateRuleType(rule.Type) {
				return nil, fmt.Errorf("invalid rule type: %s", rule.Type)
			}
			if rule.ThresholdValue <= 0 {
				return nil, fmt.Errorf("threshold_value must be > 0")
			}
			_, err := r.db.ExecContext(ctx, `
				INSERT INTO maintenance_rules (component_id, type, threshold_value, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?)
			`, componentID, rule.Type, rule.ThresholdValue, now, now)
			if err != nil {
				return nil, err
			}
		}
	}

	items, err := r.ListComponents(ctx, athleteID, gearID)
	if err != nil {
		return nil, err
	}
	for _, c := range items {
		if c.ID == componentID {
			return &c, nil
		}
	}
	return nil, nil
}

func (r *MaintenanceRepository) DeleteComponent(ctx context.Context, athleteID int64, componentID int64) error {
	var gearID string
	err := r.db.QueryRowContext(ctx, `
		SELECT c.gear_id
		FROM components c
		JOIN gear g ON g.id = c.gear_id
		WHERE c.id = ? AND g.athlete_id = ?
	`, componentID, athleteID).Scan(&gearID)
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// Use transaction to ensure atomicity
	tx, err := r.db.Conn().BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Delete in order respecting foreign key constraints
	if _, err := tx.ExecContext(ctx, `DELETE FROM maintenance_log WHERE component_id = ?`, componentID); err != nil {
		return fmt.Errorf("deleting maintenance log: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM maintenance_rules WHERE component_id = ?`, componentID); err != nil {
		return fmt.Errorf("deleting maintenance rules: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM components WHERE id = ?`, componentID); err != nil {
		return fmt.Errorf("deleting component: %w", err)
	}

	return tx.Commit()
}

func (r *MaintenanceRepository) LogMaintenance(ctx context.Context, athleteID int64, componentID int64, activityID *int64, completedAt time.Time) error {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
		SELECT TRUE
		FROM components c
		JOIN gear g ON g.id = c.gear_id
		WHERE c.id = ? AND g.athlete_id = ?
	`, componentID, athleteID).Scan(&exists)
	if isNotFound(err) {
		return nil
	}
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO maintenance_log (component_id, activity_id, completed_at, created_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (component_id, completed_at) DO NOTHING
	`, componentID, activityID, completedAt, SQLiteTime{Time: time.Now()})
	return err
}

func (r *MaintenanceRepository) LogFromActivityHashtags(ctx context.Context, athleteID int64, activityID int64, completedAt time.Time, activityName string) (int, error) {
	tags := extractMaintenanceTags(activityName)
	if len(tags) == 0 {
		return 0, nil
	}

	placeholders := make([]string, 0, len(tags))
	tagArgs := make([]any, 0, len(tags))
	for _, t := range tags {
		placeholders = append(placeholders, "?")
		tagArgs = append(tagArgs, t)
	}
	createdAt := SQLiteTime{Time: time.Now()}

	query := fmt.Sprintf(`
		INSERT INTO maintenance_log (component_id, activity_id, completed_at, created_at)
		SELECT c.id, ?, ?, ?
		FROM components c
		JOIN gear g ON g.id = c.gear_id
		WHERE g.athlete_id = ?
		  AND COALESCE(c.maintenance_hashtag, '') != ''
		  AND lower(COALESCE(c.maintenance_hashtag, '')) IN (%s)
		ON CONFLICT (component_id, completed_at) DO NOTHING
	`, strings.Join(placeholders, ","))

	params := make([]any, 0, 4+len(tagArgs))
	params = append(params, activityID, completedAt, createdAt, athleteID)
	params = append(params, tagArgs...)

	res, err := r.db.ExecContext(ctx, query, params...)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

func (r *MaintenanceRepository) Due(ctx context.Context, athleteID int64) ([]DueComponent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT
			c.id, c.gear_id, c.name, COALESCE(c.image_url, ''), COALESCE(c.maintenance_hashtag, ''),
			c.created_at, c.updated_at,
			(
				SELECT MAX(completed_at)
				FROM maintenance_log ml
				WHERE ml.component_id = c.id
			) AS last_completed_at
		FROM components c
		JOIN gear g ON g.id = c.gear_id
		WHERE g.athlete_id = ?
		ORDER BY c.created_at ASC
	`, athleteID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var components []ComponentWithRules
	componentIDs := make([]int64, 0, 64)
	for rows.Next() {
		var it ComponentWithRules
		if err := rows.Scan(
			&it.ID,
			&it.GearID,
			&it.Name,
			&it.ImageURL,
			&it.MaintenanceHashtag,
			&it.CreatedAt,
			&it.UpdatedAt,
			&it.LastCompletedAt,
		); err != nil {
			return nil, err
		}
		if it.MaintenanceHashtag != "" {
			it.MaintenanceHashtag = normalizeTag(it.MaintenanceHashtag)
		}
		components = append(components, it)
		componentIDs = append(componentIDs, it.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(components) == 0 {
		return nil, nil
	}

	placeholders := make([]string, 0, len(componentIDs))
	args := make([]any, 0, len(componentIDs))
	for _, id := range componentIDs {
		placeholders = append(placeholders, "?")
		args = append(args, id)
	}

	rulesRows, err := r.db.QueryContext(ctx, fmt.Sprintf(`
		SELECT id, component_id, type, threshold_value, created_at, updated_at
		FROM maintenance_rules
		WHERE component_id IN (%s)
		ORDER BY component_id ASC, id ASC
	`, strings.Join(placeholders, ",")), args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rulesRows.Close() }()

	rulesByComponent := make(map[int64][]MaintenanceRule, len(componentIDs))
	for rulesRows.Next() {
		var mr MaintenanceRule
		if err := rulesRows.Scan(&mr.ID, &mr.ComponentID, &mr.Type, &mr.ThresholdValue, &mr.CreatedAt, &mr.UpdatedAt); err != nil {
			return nil, err
		}
		rulesByComponent[mr.ComponentID] = append(rulesByComponent[mr.ComponentID], mr)
	}
	if err := rulesRows.Err(); err != nil {
		return nil, err
	}

	now := time.Now()

	// Build a map of gear_id -> since date for batched activity stats query
	type gearSince struct {
		gearID string
		since  time.Time
	}
	componentGearSince := make(map[int64]gearSince, len(components))
	gearIDs := make(map[string]bool)
	for _, c := range components {
		since := c.CreatedAt.Time
		if c.LastCompletedAt != nil && !c.LastCompletedAt.IsZero() {
			since = c.LastCompletedAt.Time
		}
		componentGearSince[c.ID] = gearSince{gearID: c.GearID, since: since}
		gearIDs[c.GearID] = true
	}

	// Batch query: get activity stats per gear since the earliest date we care about
	// We need per-component since dates, but we can optimize by fetching all activities
	// for the athlete's gear and computing per-component in memory
	type activityStats struct {
		distance   float64
		movingTime int
	}
	gearStats := make(map[int64]activityStats, len(components))

	for compID, gs := range componentGearSince {
		var dist float64
		var moving int
		if err := r.db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(distance), 0), COALESCE(SUM(moving_time), 0)
			FROM activities
			WHERE athlete_id = ? AND gear_id = ? AND start_date_local >= ?
		`, athleteID, gs.gearID, gs.since).Scan(&dist, &moving); err != nil {
			return nil, err
		}
		gearStats[compID] = activityStats{distance: dist, movingTime: moving}
	}

	out := make([]DueComponent, 0, len(components))
	for _, c := range components {
		c.Rules = rulesByComponent[c.ID]
		gs := componentGearSince[c.ID]
		stats := gearStats[c.ID]

		daysSince := int(now.Sub(gs.since).Hours() / 24)
		var progress []RuleProgress
		isDue := false
		for _, rule := range c.Rules {
			var current float64
			switch rule.Type {
			case "distance_m":
				current = stats.distance
			case "time_s":
				current = float64(stats.movingTime)
			case "days":
				current = float64(daysSince)
			default:
				continue
			}
			pct := 0.0
			if rule.ThresholdValue > 0 {
				pct = (current / rule.ThresholdValue) * 100
			}
			due := rule.ThresholdValue > 0 && current >= rule.ThresholdValue
			if due {
				isDue = true
			}
			progress = append(progress, RuleProgress{
				Type:           rule.Type,
				ThresholdValue: rule.ThresholdValue,
				CurrentValue:   current,
				Percent:        pct,
				Due:            due,
			})
		}
		out = append(out, DueComponent{
			ComponentWithRules: c,
			DistanceSince:      stats.distance,
			MovingTimeSince:    stats.movingTime,
			DaysSince:          daysSince,
			Progress:           progress,
			IsDue:              isDue,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].IsDue != out[j].IsDue {
			return out[i].IsDue
		}
		maxPct := func(p []RuleProgress) float64 {
			m := 0.0
			for _, x := range p {
				if x.Percent > m {
					m = x.Percent
				}
			}
			return m
		}
		pi := maxPct(out[i].Progress)
		pj := maxPct(out[j].Progress)
		if pi != pj {
			return pi > pj
		}
		return out[i].ID < out[j].ID
	})

	return out, nil
}
