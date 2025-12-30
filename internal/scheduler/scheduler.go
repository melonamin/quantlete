package scheduler

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/melonamin/quantlete/internal/importer"
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

type Scheduler struct {
	logger        *slog.Logger
	strava        *strava.Client
	settingsRepo  *storage.SettingsRepository
	importer      *importer.Importer
	reconcileTick time.Duration

	mu         sync.Mutex
	cron       *cron.Cron
	entryIDs   map[string]cron.EntryID
	lastConfig *storage.SchedulerSettings
	cancel     context.CancelFunc
}

func New(
	logger *slog.Logger,
	stravaClient *strava.Client,
	settingsRepo *storage.SettingsRepository,
	imp *importer.Importer,
) *Scheduler {
	return &Scheduler{
		logger:        logger,
		strava:        stravaClient,
		settingsRepo:  settingsRepo,
		importer:      imp,
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

	s.applyConfig(ctx, athlete.ID, &settings.Scheduler)
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

func (s *Scheduler) applyConfig(_ context.Context, athleteID int64, cfg *storage.SchedulerSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cron == nil || cfg == nil {
		return
	}

	// Avoid churn if unchanged.
	if s.lastConfig != nil && *s.lastConfig == *cfg {
		return
	}
	copied := *cfg
	s.lastConfig = &copied

	s.configurePullSyncLocked(cfg.Pull)
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
