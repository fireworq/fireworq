SELECT job_id, category, url, payload, next_try, status, created_at, retry_count, retry_delay, fail_count, timeout
  FROM `{{.JobQueue}}`
WHERE status = ? AND job_id IN
