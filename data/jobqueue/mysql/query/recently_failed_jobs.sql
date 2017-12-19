SELECT failure_id, job_id, category, url, payload, result, fail_count, failed_at, created_at FROM `{{.Failure}}`
WHERE ? = ? AND failure_id <= ?
ORDER BY failure_id DESC LIMIT
