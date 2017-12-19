UPDATE `{{.JobQueue}}`
SET grabber_id = NULL, status = 'claimed',
	next_try = FLOOR(UNIX_TIMESTAMP(CURRENT_TIME(3)) * 1000) + ?, retry_count = ?, fail_count = ?
WHERE job_id = ?
