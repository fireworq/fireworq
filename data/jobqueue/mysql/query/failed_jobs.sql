SELECT failure_id, job_id, category, url, payload, result, fail_count, failed_at, created_at FROM `{{.Failure}}`
WHERE created_at <= ? AND (created_at != ? OR failure_id <= ?)
ORDER BY created_at DESC, failure_id DESC LIMIT
