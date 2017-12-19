UPDATE `{{.JobQueue}}`
SET status = 'grabbed', grabber_id = CONNECTION_ID()
WHERE job_id IN
