SELECT job_id FROM `{{.JobQueue}}`
WHERE status = 'grabbed' AND grabber_id != CONNECTION_ID()
LIMIT 1000
