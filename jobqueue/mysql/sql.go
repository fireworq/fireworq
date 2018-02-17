//go:generate go-assets-builder -p mysql -o assets.go ../../data/jobqueue/mysql

package mysql

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"text/template"

	"github.com/fireworq/fireworq/model"
)

func newTableName(definition *model.Queue) *tableName {
	re := invalidTablenameChars
	name := string(re.ReplaceAll([]byte(definition.Name), []byte{'_'}))
	return &tableName{
		JobQueue: strings.Join([]string{"fireworq_jq(", name, ")"}, ""),
		Failure:  strings.Join([]string{"fireworq_jq_fail(", name, ")"}, ""),
	}
}

type tableName struct {
	JobQueue string
	Payload  string
	Failure  string
}

func (tn *tableName) makeQueries() *sqls {
	return &sqls{
		createJobqueue:     tn.makeQuery(tmplCreateJobqueue),
		createFailure:      tn.makeQuery(tmplCreateFailure),
		grab:               tn.makeQuery(tmplGrabJobs),
		grabbed:            tn.makeQuery(tmplGrabbedJobs),
		launch:             tn.makeQuery(tmplLaunchJobs),
		insertJob:          tn.makeQuery(tmplInsertJob),
		insertFailedJob:    tn.makeQuery(tmplInsertFailedJob),
		deleteFailedJob:    tn.makeQuery(tmplDeleteFailedJob),
		deleteJob:          tn.makeQuery(tmplDeleteJob),
		updateJob:          tn.makeQuery(tmplUpdateJob),
		orphan:             tn.makeQuery(tmplOrphanJobs),
		recover:            tn.makeQuery(tmplRecoverJobs),
		inspectJob:         tn.makeQuery(tmplInspectJob),
		inspectJobs:        tn.makeQuery(tmplInspectJobs),
		failedJob:          tn.makeQuery(tmplFailedJob),
		failedJobs:         tn.makeQuery(tmplFailedJobs),
		recentlyFailedJobs: tn.makeQuery(tmplRecentlyFailedJobs),
	}
}

func (tn *tableName) makeQuery(tmpl *template.Template) string {
	buffer := new(bytes.Buffer)
	_ = tmpl.Execute(buffer, tn) // ignore error
	return buffer.String()
}

type sqls struct {
	createJobqueue     string
	createFailure      string
	grab               string
	grabbed            string
	launch             string
	insertJob          string
	insertFailedJob    string
	deleteFailedJob    string
	deleteJob          string
	updateJob          string
	orphan             string
	recover            string
	inspectJob         string
	inspectJobs        string
	failedJob          string
	failedJobs         string
	recentlyFailedJobs string
}

var (
	invalidTablenameChars  *regexp.Regexp
	tmplCreateJobqueue     *template.Template
	tmplCreateFailure      *template.Template
	tmplGrabJobs           *template.Template
	tmplGrabbedJobs        *template.Template
	tmplLaunchJobs         *template.Template
	tmplInsertJob          *template.Template
	tmplInsertFailedJob    *template.Template
	tmplDeleteFailedJob    *template.Template
	tmplDeleteJob          *template.Template
	tmplUpdateJob          *template.Template
	tmplOrphanJobs         *template.Template
	tmplRecoverJobs        *template.Template
	tmplInspectJob         *template.Template
	tmplInspectJobs        *template.Template
	tmplFailedJob          *template.Template
	tmplFailedJobs         *template.Template
	tmplRecentlyFailedJobs *template.Template
)

func mustLoadTemplate(name string) *template.Template {
	f, err := Assets.Open(fmt.Sprintf("/data/jobqueue/mysql/%s.sql", name))
	if err != nil {
		panic("Cannot load template (" + name + "): " + err.Error())
	}

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		panic("Cannot load template (" + name + "): " + err.Error())
	}

	tmpl, err := template.New(name).Parse(string(buf))
	if err != nil {
		panic("Cannot load template (" + name + "): " + err.Error())
	}
	return tmpl
}

func init() {
	invalidTablenameChars = regexp.MustCompile("[^0-9a-z_]")
	tmplCreateJobqueue = mustLoadTemplate("schema/job_queue")
	tmplCreateFailure = mustLoadTemplate("schema/job_failure")
	tmplGrabJobs = mustLoadTemplate("query/grab_jobs")
	tmplGrabbedJobs = mustLoadTemplate("query/grabbed_jobs")
	tmplLaunchJobs = mustLoadTemplate("query/launch_jobs")
	tmplInsertJob = mustLoadTemplate("query/insert_job")
	tmplInsertFailedJob = mustLoadTemplate("query/insert_failed_job")
	tmplDeleteFailedJob = mustLoadTemplate("query/delete_failed_job")
	tmplDeleteJob = mustLoadTemplate("query/delete_job")
	tmplUpdateJob = mustLoadTemplate("query/update_job")
	tmplOrphanJobs = mustLoadTemplate("query/orphan_jobs")
	tmplRecoverJobs = mustLoadTemplate("query/recover_jobs")
	tmplInspectJob = mustLoadTemplate("query/inspect_job")
	tmplInspectJobs = mustLoadTemplate("query/inspect_jobs")
	tmplFailedJob = mustLoadTemplate("query/failed_job")
	tmplFailedJobs = mustLoadTemplate("query/failed_jobs")
	tmplRecentlyFailedJobs = mustLoadTemplate("query/recently_failed_jobs")
}
