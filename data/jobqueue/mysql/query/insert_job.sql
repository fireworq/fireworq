INSERT INTO `{{.JobQueue}}` (next_try, created_at, retry_count, retry_delay, fail_count, category, url, payload, timeout)
VALUES (FLOOR(UNIX_TIMESTAMP(CURRENT_TIME(3)) * 1000) + ?, FLOOR(UNIX_TIMESTAMP(CURRENT_TIME(3)) * 1000), ?, ?, ?, ?, ?, ?, ?)
