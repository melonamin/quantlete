-- Athlete segment stats and map polyline support

ALTER TABLE segments ADD COLUMN athlete_kom_rank INTEGER;
ALTER TABLE segments ADD COLUMN athlete_effort_count INTEGER;
ALTER TABLE segments ADD COLUMN athlete_pr_elapsed_time INTEGER;
ALTER TABLE segments ADD COLUMN athlete_pr_date TIMESTAMP;

