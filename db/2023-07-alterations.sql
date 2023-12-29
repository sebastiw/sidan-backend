ALTER TABLE cl2003_msgs ALTER COLUMN `date` SET DEFAULT '1970-01-01T00:00:00';
ALTER TABLE cl2003_msgs ADD COLUMN datetime datetime NOT NULL DEFAULT (now());
UPDATE cl2003_msgs SET datetime = concat(date, " ", time);
ALTER TABLE cl_news ADD COLUMN datetime datetime NOT NULL DEFAULT (now());
UPDATE cl2003_msgs SET datetime = concat(date, " ", time);
UPDATE cl_news SET datetime = concat(date, " ", time);

