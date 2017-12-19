package mysql

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/fireworq/fireworq/jobqueue"
)

type failureLog struct {
	db  *sql.DB
	sql *sqls
}

func (l *failureLog) Add(failed jobqueue.Job, result *jobqueue.Result) error {
	log := log.With().Str("method", "failureLog.Add").Logger()

	j, ok := failed.(*job)
	if !ok {
		return fmt.Errorf("Invalid job structure: %v", failed)
	}

	res, err := json.Marshal(result)
	if err != nil {
		return err
	}

	if _, err := l.db.Exec(
		l.sql.insertFailedJob,
		j.id,
		j.Category(),
		failed.URL(),
		failed.Payload(),
		res,
		failed.FailCount()+1,
		time.Now().UnixNano()/int64(time.Millisecond),
		j.CreatedAt(),
	); err != nil {
		log.Debug().Msgf("Failed to Insert a job: %s", err)
	}

	return err
}

func (l *failureLog) Delete(failureID uint64) error {
	_, err := l.db.Exec(l.sql.deleteFailedJob, failureID)
	return err
}

func (l *failureLog) Find(failureID uint64) (*jobqueue.FailedJob, error) {
	j, err := l.scan(l.db.QueryRow(l.sql.failedJob, failureID))
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (l *failureLog) FindAll(limit uint, cursor string) (*jobqueue.FailedJobs, error) {
	return l.findAllByQuery(l.sql.failedJobs, limit, cursor)
}

func (l *failureLog) FindAllRecentFailures(limit uint, cursor string) (*jobqueue.FailedJobs, error) {
	return l.findAllByQuery(l.sql.recentlyFailedJobs, limit, cursor)
}

func (l *failureLog) findAllByQuery(query string, limit uint, cursor string) (*jobqueue.FailedJobs, error) {
	var maxTime int64 = math.MaxInt64
	var maxID uint64 = math.MaxUint64
	if decoded, err := base64.StdEncoding.DecodeString(cursor); err == nil {
		if pair := strings.SplitN(string(decoded), ",", 2); len(pair) == 2 {
			t, err1 := strconv.Atoi(pair[0])
			j, err2 := strconv.Atoi(pair[1])
			if err1 == nil && err2 == nil {
				maxTime = int64(t)
				maxID = uint64(j)
			}
		}
	}

	rows, err := l.db.Query(
		query+strconv.FormatUint(uint64(limit)+1, 10),
		maxTime,
		maxTime,
		maxID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make([]jobqueue.FailedJob, 0, limit)
	ids := make([]uint64, 0, limit)
	for rows.Next() {
		j, err := l.scan(rows)
		if err != nil {
			return nil, err
		}

		results = append(results, *j)
		ids = append(ids, j.ID)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	nextCursor := ""
	if uint(len(results)) > limit {
		nextCursor = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(
			"%d,%d",
			results[limit].CreatedAt.UnixNano()/int64(time.Millisecond),
			ids[limit],
		)))
		results = results[:limit]
	}

	return &jobqueue.FailedJobs{FailedJobs: results, NextCursor: nextCursor}, nil
}

func (l *failureLog) scan(s scanner) (*jobqueue.FailedJob, error) {
	var j jobqueue.FailedJob
	var result []byte
	var failedAt uint64
	var createdAt uint64

	if err := s.Scan(&(j.ID), &(j.JobID), &(j.Category), &(j.URL), &(j.Payload), &result, &(j.FailCount), &failedAt, &createdAt); err != nil {
		return nil, err
	}
	if _, err := json.Marshal(j.Payload); err != nil {
		payload, _ := json.Marshal(string(j.Payload))
		j.Payload = json.RawMessage(payload)
	}

	if err := json.Unmarshal(result, &(j.Result)); err != nil {
		return nil, err
	}

	secInMillisec := int64(time.Second / time.Millisecond)
	j.FailedAt = time.Unix(int64(failedAt)/secInMillisec, int64(failedAt)%secInMillisec*int64(time.Millisecond))
	j.CreatedAt = time.Unix(int64(createdAt)/secInMillisec, int64(createdAt)%secInMillisec*int64(time.Millisecond))

	return &j, nil
}
