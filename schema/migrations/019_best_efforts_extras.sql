-- Best efforts extras (name/pr_rank/moving_time)

ALTER TABLE best_efforts ADD COLUMN name TEXT;
ALTER TABLE best_efforts ADD COLUMN pr_rank INTEGER;
ALTER TABLE best_efforts ADD COLUMN moving_time INTEGER;

CREATE INDEX IF NOT EXISTS idx_best_efforts_pr_rank ON best_efforts(athlete_id, pr_rank);

