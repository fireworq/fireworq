package mysql

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/fireworq/fireworq/jobqueue"
)

type scanner interface {
	Scan(args ...interface{}) error
}

type inspector struct {
	db  *sql.DB
	sql *sqls
}

func (i *inspector) Delete(jobID uint64) error {
	_, err := i.db.Exec(i.sql.deleteJob, jobID)
	return err
}

func (i *inspector) Find(jobID uint64) (*jobqueue.InspectedJob, error) {
	j, err := i.scan(i.db.QueryRow(i.sql.inspectJob, jobID))
	if err != nil {
		return nil, err
	}
	return j, nil
}

func (i *inspector) FindAllGrabbed(limit uint, cursor string, order jobqueue.SortOrder) (*jobqueue.InspectedJobs, error) {
	var maxTime = time.Now().UnixNano() / int64(time.Millisecond)
	if order == jobqueue.Asc {
		return i.findAllAsc("grabbed", 0, maxTime, limit, cursor)
	}
	return i.findAllDesc("grabbed", 0, maxTime, limit, cursor)
}

func (i *inspector) FindAllWaiting(limit uint, cursor string, order jobqueue.SortOrder) (*jobqueue.InspectedJobs, error) {
	var maxTime = time.Now().UnixNano() / int64(time.Millisecond)
	if order == jobqueue.Asc {
		return i.findAllAsc("claimed", 0, maxTime, limit, cursor)
	}
	return i.findAllDesc("claimed", 0, maxTime, limit, cursor)
}

func (i *inspector) FindAllDeferred(limit uint, cursor string, order jobqueue.SortOrder) (*jobqueue.InspectedJobs, error) {
	var minTime = time.Now().UnixNano() / int64(time.Millisecond)
	if order == jobqueue.Asc {
		return i.findAllAsc("claimed", minTime, math.MaxInt64, limit, cursor)
	}
	return i.findAllDesc("claimed", minTime, 0, limit, cursor)
}

func decodeCursor(cursor string, time *int64, jobID *uint64) {
	if decoded, err := base64.StdEncoding.DecodeString(cursor); err == nil {
		if pair := strings.SplitN(string(decoded), ",", 2); len(pair) == 2 {
			t, err1 := strconv.Atoi(pair[0])
			j, err2 := strconv.Atoi(pair[1])
			if err1 == nil && err2 == nil {
				*time = int64(t)
				*jobID = uint64(j)
			}
		}
	}
}

func (i *inspector) findAllAsc(status string, minTime int64, maxTime int64, limit uint, cursor string) (*jobqueue.InspectedJobs, error) {
	if minTime >= math.MaxInt64 {
		minTime = 0
	}
	var minJobID uint64
	decodeCursor(cursor, &minTime, &minJobID)

	ids := make([]interface{}, 0, limit+1)
	placeholders := make([]string, 0, limit+1)
	results := make([]jobqueue.InspectedJob, 0, limit+1)

	if err := func() error {
		rows, err := i.db.Query(
			i.sql.inspectJobsAsc+strconv.FormatUint(uint64(limit)+1, 10),
			status,
			minTime,
			maxTime,
			minJobID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id uint64
			if err := rows.Scan(&id); err != nil {
				return err
			}
			placeholders = append(placeholders, "?")
			ids = append(ids, id)
		}
		return rows.Err()
	}(); err != nil {
		return nil, err
	}
	if len(ids) <= 0 { // no job to report
		return &jobqueue.InspectedJobs{Jobs: results}, nil
	}

	if err := func() error {
		rows, err := i.db.Query(
			i.sql.grabbed+"("+strings.Join(placeholders, ",")+")",
			append([]interface{}{status}, ids...)...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			j, err := i.scan(rows)
			if err != nil {
				return err
			}
			results = append(results, *j)
		}
		return rows.Err()
	}(); err != nil {
		return nil, err
	}

	// Emulate `ORDER BY next_try ASC, job_id ASC`, which causes
	// `using filesort` together with `SELECT ~ WHERE ~ IN`.
	sort.Slice(results, func(i, j int) bool {
		j1 := results[i]
		j2 := results[j]
		t1 := j1.NextTry.UnixNano()
		t2 := j2.NextTry.UnixNano()
		if t1 < t2 {
			return true
		}
		if t1 > t2 {
			return false
		}
		return j1.ID < j2.ID
	})

	nextCursor := ""
	if uint(len(results)) > limit {
		nextCursor = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(
			"%d,%d",
			results[limit].NextTry.UnixNano()/int64(time.Millisecond),
			results[limit].ID,
		)))
		results = results[:limit]
	}

	return &jobqueue.InspectedJobs{Jobs: results, NextCursor: nextCursor}, nil
}

func (i *inspector) findAllDesc(status string, minTime int64, maxTime int64, limit uint, cursor string) (*jobqueue.InspectedJobs, error) {
	if maxTime <= 0 {
		maxTime = math.MaxInt64
	}

	var maxJobID uint64 = math.MaxUint64
	decodeCursor(cursor, &maxTime, &maxJobID)

	ids := make([]interface{}, 0, limit+1)
	placeholders := make([]string, 0, limit+1)
	results := make([]jobqueue.InspectedJob, 0, limit+1)

	if err := func() error {
		rows, err := i.db.Query(
			i.sql.inspectJobs+strconv.FormatUint(uint64(limit)+1, 10),
			status,
			minTime,
			maxTime,
			maxJobID,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var id uint64
			if err := rows.Scan(&id); err != nil {
				return err
			}
			placeholders = append(placeholders, "?")
			ids = append(ids, id)
		}
		return rows.Err()
	}(); err != nil {
		return nil, err
	}
	if len(ids) <= 0 { // no job to report
		return &jobqueue.InspectedJobs{Jobs: results}, nil
	}

	if err := func() error {
		rows, err := i.db.Query(
			i.sql.grabbed+"("+strings.Join(placeholders, ",")+")",
			append([]interface{}{status}, ids...)...,
		)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			j, err := i.scan(rows)
			if err != nil {
				return err
			}
			results = append(results, *j)
		}
		return rows.Err()
	}(); err != nil {
		return nil, err
	}

	// Emulate `ORDER BY next_try DESC, job_id DESC`, which causes
	// `using filesort` together with `SELECT ~ WHERE ~ IN`.
	sort.Slice(results, func(i, j int) bool {
		j1 := results[i]
		j2 := results[j]
		t1 := j1.NextTry.UnixNano()
		t2 := j2.NextTry.UnixNano()
		if t1 > t2 {
			return true
		}
		if t1 < t2 {
			return false
		}
		return j1.ID > j2.ID
	})

	nextCursor := ""
	if uint(len(results)) > limit {
		nextCursor = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(
			"%d,%d",
			results[limit].NextTry.UnixNano()/int64(time.Millisecond),
			results[limit].ID,
		)))
		results = results[:limit]
	}

	return &jobqueue.InspectedJobs{Jobs: results, NextCursor: nextCursor}, nil
}

func (i *inspector) scan(s scanner) (*jobqueue.InspectedJob, error) {
	var j jobqueue.InspectedJob
	var createdAt uint64
	var nextTry uint64
	var retryCount uint

	if err := s.Scan(&(j.ID), &(j.Category), &(j.URL), &(j.Payload), &nextTry, &(j.Status), &createdAt, &retryCount, &(j.RetryDelay), &(j.FailCount), &(j.Timeout)); err != nil {
		return nil, err
	}
	if _, err := json.Marshal(j.Payload); err != nil {
		payload, _ := json.Marshal(string(j.Payload))
		j.Payload = json.RawMessage(payload)
	}

	secInMillisec := int64(time.Second / time.Millisecond)
	j.NextTry = time.Unix(int64(nextTry)/secInMillisec, int64(nextTry)%secInMillisec*int64(time.Millisecond))
	j.CreatedAt = time.Unix(int64(createdAt)/secInMillisec, int64(createdAt)%secInMillisec*int64(time.Millisecond))
	j.MaxRetries = j.FailCount + retryCount

	return &j, nil
}
