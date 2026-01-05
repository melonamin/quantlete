package scheduler

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/melonamin/quantlete/internal/importer"
	notifcfg "github.com/melonamin/quantlete/internal/notifications"
	"github.com/melonamin/quantlete/internal/services"
	"github.com/melonamin/quantlete/internal/storage"
	"github.com/melonamin/quantlete/internal/strava"
)

type slogCronLogger struct {
	logger *slog.Logger
}

func (l slogCronLogger) Info(msg string, keysAndValues ...any) {
	if l.logger == nil {
		return
	}
	l.logger.Info(msg, keysAndValues...)
}

func (l slogCronLogger) Error(err error, msg string, keysAndValues ...any) {
	if l.logger == nil {
		return
	}
	l.logger.Error(msg, append(keysAndValues, "error", err)...)
}

// Scheduler manages periodic jobs for a single authenticated athlete.
// Jobs are keyed by athlete ID to allow proper cleanup when the athlete changes.
// Note: Currently the scheduler handles one athlete at a time (via strava.GetAthlete()),
// but athlete-specific job keys ensure correct behavior if this assumption changes.
type Scheduler struct {
	logger        *slog.Logger
	strava        *strava.Client
	settingsRepo  *storage.SettingsRepository
	importer      *importer.Importer
	maintenance   *services.MaintenanceService
	notifications *services.NotificationService
	stats         *storage.StatsRepository
	reconcileTick time.Duration

	mu         sync.Mutex
	cron       *cron.Cron
	entryIDs   map[string]cron.EntryID
	lastConfig *schedulerConfigSnapshot
	cancel     context.CancelFunc
}

// schedulerConfigSnapshot holds the last applied config for change detection.
type schedulerConfigSnapshot struct {
	Scheduler           storage.SchedulerSettings
	MaintenanceSchedule string
	MaintenanceEnabled  bool
	WeeklyDigest        bool
	MonthlyDigest       bool
}

func New(
	logger *slog.Logger,
	stravaClient *strava.Client,
	settingsRepo *storage.SettingsRepository,
	imp *importer.Importer,
	maintenance *services.MaintenanceService,
	notifications *services.NotificationService,
	stats *storage.StatsRepository,
) *Scheduler {
	return &Scheduler{
		logger:        logger,
		strava:        stravaClient,
		settingsRepo:  settingsRepo,
		importer:      imp,
		maintenance:   maintenance,
		notifications: notifications,
		stats:         stats,
		reconcileTick: 30 * time.Second,
		entryIDs:      map[string]cron.EntryID{},
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron != nil {
		return nil
	}

	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	s.cron = cron.New(
		cron.WithParser(parser),
		cron.WithLogger(slogCronLogger{logger: s.logger}),
		cron.WithChain(
			cron.SkipIfStillRunning(cron.DiscardLogger),
			cron.Recover(cron.DiscardLogger),
		),
	)
	s.cron.Start()

	loopCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	go s.reconcileLoop(loopCtx)

	s.logger.Info("scheduler started")
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	s.mu.Lock()
	c := s.cron
	cancel := s.cancel
	s.cron = nil
	s.cancel = nil
	s.entryIDs = map[string]cron.EntryID{}
	s.lastConfig = nil
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
	if c == nil {
		return nil
	}

	stopped := c.Stop()
	select {
	case <-stopped.Done():
		s.logger.Info("scheduler stopped")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Scheduler) reconcileLoop(ctx context.Context) {
	ticker := time.NewTicker(s.reconcileTick)
	defer ticker.Stop()

	s.reconcileOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.reconcileOnce(ctx)
		}
	}
}

func (s *Scheduler) reconcileOnce(ctx context.Context) {
	if s.strava == nil {
		s.clearJobs()
		return
	}

	athlete := s.strava.GetAthlete()
	if athlete == nil || s.strava.GetToken() == nil {
		s.clearJobs()
		return
	}

	settings, err := s.settingsRepo.Get(ctx, athlete.ID)
	if err != nil {
		s.logger.Warn("scheduler: failed to load settings", "error", err)
		return
	}

	s.applyConfig(ctx, athlete.ID, settings)
}

func (s *Scheduler) clearJobs() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron == nil {
		return
	}

	for key, id := range s.entryIDs {
		s.cron.Remove(id)
		delete(s.entryIDs, key)
	}
	s.lastConfig = nil
}

func (s *Scheduler) applyConfig(_ context.Context, athleteID int64, settings *storage.AthleteSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron == nil || settings == nil {
		return
	}

	// Build current config snapshot.
	// Check for nil at each level to avoid panics on partially initialized settings.
	var maintenanceSchedule string
	var maintenanceEnabled bool
	var weeklyDigest bool
	var monthlyDigest bool
	if settings.Notifications != nil && settings.Notifications.Enabled {
		if settings.Notifications.Events.MaintenanceDue {
			maintenanceEnabled = true
			maintenanceSchedule = settings.Notifications.Events.MaintenanceSchedule
		}
		weeklyDigest = settings.Notifications.Events.WeeklyDigest
		monthlyDigest = settings.Notifications.Events.MonthlyDigest
	}

	current := schedulerConfigSnapshot{
		Scheduler:           settings.Scheduler,
		MaintenanceSchedule: maintenanceSchedule,
		MaintenanceEnabled:  maintenanceEnabled,
		WeeklyDigest:        weeklyDigest,
		MonthlyDigest:       monthlyDigest,
	}

	// Avoid churn if unchanged.
	if s.lastConfig != nil && *s.lastConfig == current {
		return
	}
	s.lastConfig = &current

	s.configurePullSyncLocked(settings.Scheduler.Pull)
	s.configureMaintenanceCheckLocked(athleteID, maintenanceEnabled, maintenanceSchedule)
	s.configureWeeklyDigestLocked(athleteID, weeklyDigest)
	s.configureMonthlyDigestLocked(athleteID, monthlyDigest)
}

func (s *Scheduler) setJobLocked(key string, enabled bool, spec string, fn func()) {
	if id, ok := s.entryIDs[key]; ok {
		s.cron.Remove(id)
		delete(s.entryIDs, key)
	}

	if !enabled || spec == "" {
		return
	}

	id, err := s.cron.AddFunc(spec, fn)
	if err != nil {
		s.logger.Warn("scheduler: invalid cron spec", "job", key, "spec", spec, "error", err)
		return
	}

	s.entryIDs[key] = id
	s.logger.Info("scheduler: job scheduled", "job", key, "spec", spec)
}

func (s *Scheduler) configurePullSyncLocked(cfg storage.PullSettings) {
	spec := cronSpecForPullSchedule(cfg.Schedule)
	s.setJobLocked("pull_sync", cfg.Enabled, spec, func() {
		if err := s.importer.Start(context.Background(), importer.ImportOptions{}); err != nil {
			s.logger.Info("scheduler: pull sync not started", "error", err)
		}
	})
}

func (s *Scheduler) configureMaintenanceCheckLocked(athleteID int64, enabled bool, schedule string) {
	if s.maintenance == nil || s.notifications == nil {
		return
	}

	spec := cronSpecForMaintenanceSchedule(schedule)
	job := NewMaintenanceCheckJob(s.logger, s.maintenance, s.notifications, s.settingsRepo, athleteID)
	key := fmt.Sprintf("maintenance_check_%d", athleteID)
	s.setJobLocked(key, enabled, spec, job.Run)
}

func (s *Scheduler) configureWeeklyDigestLocked(athleteID int64, enabled bool) {
	if s.stats == nil || s.notifications == nil {
		return
	}

	// Monday at 9:00 AM
	spec := "0 9 * * 1"
	job := NewWeeklyDigestJob(s.logger, s.stats, s.notifications, s.settingsRepo, athleteID)
	key := fmt.Sprintf("weekly_digest_%d", athleteID)
	s.setJobLocked(key, enabled, spec, job.Run)
}

func (s *Scheduler) configureMonthlyDigestLocked(athleteID int64, enabled bool) {
	if s.stats == nil || s.notifications == nil {
		return
	}

	// 1st of month at 9:00 AM
	spec := "0 9 1 * *"
	job := NewMonthlyDigestJob(s.logger, s.stats, s.notifications, s.settingsRepo, athleteID)
	key := fmt.Sprintf("monthly_digest_%d", athleteID)
	s.setJobLocked(key, enabled, spec, job.Run)
}

// cronSpecForMaintenanceSchedule converts a maintenance schedule string to a cron spec.
// Returns a weekly cron spec (Sunday at 9:00 AM) by default.
func cronSpecForMaintenanceSchedule(schedule string) string {
	switch schedule {
	case notifcfg.MaintenanceScheduleWeekly:
		return "0 9 * * 0" // Sunday at 9:00 AM
	case notifcfg.MaintenanceScheduleMonthly:
		return "0 9 1 * *" // 1st of month at 9:00 AM
	default:
		return "0 9 * * 0" // Default to weekly
	}
}

func cronSpecForPullSchedule(sched storage.PullSchedule) string {
	switch sched {
	case storage.PullScheduleHourly:
		return "0 * * * *"
	case storage.PullScheduleEvery6Hours:
		return "0 */6 * * *"
	case storage.PullScheduleMidnight:
		return "0 0 * * *"
	default:
		return "0 0 * * *"
	}
}

var ErrSchedulerNotRunning = errors.New("scheduler not running")
