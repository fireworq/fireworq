SELECT failure_id, job_id, category, url, payload, result, fail_count, failed_at, created_at FROM `{{.Failure}}`
WHERE failure_id = ?
