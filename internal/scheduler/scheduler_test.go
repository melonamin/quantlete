package scheduler

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/sasha/stata/internal/storage"
)

func TestCronSpecForPullSchedule(t *testing.T) {
	tests := []struct {
		schedule storage.PullSchedule
		wantSpec string
	}{
		{storage.PullScheduleHourly, "0 * * * *"},
		{storage.PullScheduleEvery6Hours, "0 */6 * * *"},
		{storage.PullScheduleMidnight, "0 0 * * *"},
		{"unknown", "0 0 * * *"}, // Falls back to midnight
		{"", "0 0 * * *"},        // Falls back to midnight
	}

	for _, tt := range tests {
		t.Run(string(tt.schedule), func(t *testing.T) {
			got := cronSpecForPullSchedule(tt.schedule)
			if got != tt.wantSpec {
				t.Errorf("cronSpecForPullSchedule(%q) = %q, want %q", tt.schedule, got, tt.wantSpec)
			}
		})
	}
}

func TestScheduler_StartStop(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	sched := New(logger, nil, nil, nil)

	ctx := context.Background()

	// Start the scheduler
	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Starting again should be a no-op
	if err := sched.Start(ctx); err != nil {
		t.Errorf("Start() second call error = %v, want nil", err)
	}

	// Verify scheduler is running
	sched.mu.Lock()
	running := sched.cron != nil
	sched.mu.Unlock()

	if !running {
		t.Error("scheduler should be running after Start()")
	}

	// Stop the scheduler
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sched.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error = %v", err)
	}

	// Verify scheduler is stopped
	sched.mu.Lock()
	stopped := sched.cron == nil
	sched.mu.Unlock()

	if !stopped {
		t.Error("scheduler should be stopped after Stop()")
	}
}

func TestScheduler_StopWhenNotRunning(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	sched := New(logger, nil, nil, nil)

	ctx := context.Background()

	// Stopping a scheduler that was never started should be a no-op
	if err := sched.Stop(ctx); err != nil {
		t.Errorf("Stop() on non-running scheduler error = %v, want nil", err)
	}
}

func TestScheduler_ClearJobs(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	sched := New(logger, nil, nil, nil)

	ctx := context.Background()

	// Start the scheduler
	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Manually add a job for testing
	sched.mu.Lock()
	sched.setJobLocked("test_job", true, "0 * * * *", func() {})
	jobCount := len(sched.entryIDs)
	sched.mu.Unlock()

	if jobCount != 1 {
		t.Errorf("expected 1 job after setJobLocked, got %d", jobCount)
	}

	// Clear all jobs
	sched.clearJobs()

	sched.mu.Lock()
	jobCountAfterClear := len(sched.entryIDs)
	lastConfig := sched.lastConfig
	sched.mu.Unlock()

	if jobCountAfterClear != 0 {
		t.Errorf("expected 0 jobs after clearJobs, got %d", jobCountAfterClear)
	}

	if lastConfig != nil {
		t.Error("lastConfig should be nil after clearJobs")
	}

	// Cleanup
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	sched.Stop(stopCtx)
}

func TestScheduler_SetJobLocked(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	sched := New(logger, nil, nil, nil)

	ctx := context.Background()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	defer func() {
		stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		sched.Stop(stopCtx)
	}()

	tests := []struct {
		name      string
		key       string
		enabled   bool
		spec      string
		wantAdded bool
	}{
		{
			name:      "valid job enabled",
			key:       "test_job",
			enabled:   true,
			spec:      "0 * * * *",
			wantAdded: true,
		},
		{
			name:      "job disabled",
			key:       "disabled_job",
			enabled:   false,
			spec:      "0 * * * *",
			wantAdded: false,
		},
		{
			name:      "empty spec",
			key:       "empty_spec_job",
			enabled:   true,
			spec:      "",
			wantAdded: false,
		},
		{
			name:      "invalid spec",
			key:       "invalid_spec_job",
			enabled:   true,
			spec:      "invalid cron spec",
			wantAdded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sched.mu.Lock()
			sched.setJobLocked(tt.key, tt.enabled, tt.spec, func() {})
			_, exists := sched.entryIDs[tt.key]
			sched.mu.Unlock()

			if exists != tt.wantAdded {
				t.Errorf("setJobLocked(%q, %v, %q) job exists = %v, want %v",
					tt.key, tt.enabled, tt.spec, exists, tt.wantAdded)
			}
		})
	}
}

func TestScheduler_ConfigurePullSyncLocked(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	sched := New(logger, nil, nil, nil)

	ctx := context.Background()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	defer func() {
		stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		sched.Stop(stopCtx)
	}()

	tests := []struct {
		name      string
		cfg       storage.PullSettings
		wantAdded bool
	}{
		{
			name: "enabled hourly",
			cfg: storage.PullSettings{
				Enabled:  true,
				Schedule: storage.PullScheduleHourly,
			},
			wantAdded: true,
		},
		{
			name: "enabled every 6 hours",
			cfg: storage.PullSettings{
				Enabled:  true,
				Schedule: storage.PullScheduleEvery6Hours,
			},
			wantAdded: true,
		},
		{
			name: "disabled",
			cfg: storage.PullSettings{
				Enabled:  false,
				Schedule: storage.PullScheduleHourly,
			},
			wantAdded: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sched.mu.Lock()
			sched.configurePullSyncLocked(tt.cfg)
			_, exists := sched.entryIDs["pull_sync"]
			sched.mu.Unlock()

			if exists != tt.wantAdded {
				t.Errorf("configurePullSyncLocked(%+v) pull_sync exists = %v, want %v",
					tt.cfg, exists, tt.wantAdded)
			}
		})
	}
}

func TestScheduler_ApplyConfig_NoChurn(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	sched := New(logger, nil, nil, nil)

	ctx := context.Background()

	if err := sched.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	defer func() {
		stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		sched.Stop(stopCtx)
	}()

	cfg := &storage.SchedulerSettings{
		Pull: storage.PullSettings{
			Enabled:  true,
			Schedule: storage.PullScheduleHourly,
		},
	}

	// Apply config first time
	sched.applyConfig(ctx, 12345, cfg)

	sched.mu.Lock()
	firstConfig := sched.lastConfig
	firstEntryID := sched.entryIDs["pull_sync"]
	sched.mu.Unlock()

	if firstConfig == nil {
		t.Fatal("lastConfig should be set after applyConfig")
	}

	// Apply same config again - should be a no-op (no churn)
	sched.applyConfig(ctx, 12345, cfg)

	sched.mu.Lock()
	secondEntryID := sched.entryIDs["pull_sync"]
	sched.mu.Unlock()

	if secondEntryID != firstEntryID {
		t.Error("applying same config should not recreate jobs (no churn)")
	}

	// Apply different config - should update
	cfg2 := &storage.SchedulerSettings{
		Pull: storage.PullSettings{
			Enabled:  true,
			Schedule: storage.PullScheduleMidnight,
		},
	}

	sched.applyConfig(ctx, 12345, cfg2)

	sched.mu.Lock()
	thirdEntryID := sched.entryIDs["pull_sync"]
	sched.mu.Unlock()

	if thirdEntryID == firstEntryID {
		t.Error("applying different config should recreate jobs")
	}
}

func TestSlogCronLogger(t *testing.T) {
	// Test that nil logger doesn't panic
	logger := slogCronLogger{logger: nil}

	// These should not panic
	logger.Info("test message", "key", "value")
	logger.Error(nil, "test error", "key", "value")
}
