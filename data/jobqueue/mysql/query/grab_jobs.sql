SELECT job_id FROM `{{.JobQueue}}`
WHERE status = 'claimed'
  AND next_try <= FLOOR(UNIX_TIMESTAMP(CURRENT_TIME(3)) * 1000)
ORDER BY next_try ASC
LIMIT
