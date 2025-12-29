-- Add local_badge_url column to store locally downloaded badge images
ALTER TABLE challenges ADD COLUMN local_badge_url TEXT;
