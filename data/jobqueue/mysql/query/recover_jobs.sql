UPDATE `{{.JobQueue}}` USE INDEX (PRIMARY)
SET status = 'claimed',
    grabber_id = NULL
WHERE status = 'grabbed' AND grabber_id != CONNECTION_ID() AND job_id IN
