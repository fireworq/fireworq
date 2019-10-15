package config

type configItem struct {
	category     string
	label        string
	defaultValue string
	description  string
}

var logLevelDescription = `The level is either a name or a numeric value.  The following table describes the meaning of the value.

|Value|Name   |
|-----|-------|
|` + "`" + `0` + "`" + `    |` + "`" + `debug` + "`" + `  |
|` + "`" + `1` + "`" + `    |` + "`" + `info` + "`" + `   |
|` + "`" + `2` + "`" + `    |` + "`" + `warn` + "`" + `   |
|` + "`" + `3` + "`" + `    |` + "`" + `error` + "`" + `  |
|` + "`" + `4` + "`" + `    |` + "`" + `fatal` + "`" + `  |
`

var defaultConf = map[string]*configItem{
	"bind": {
		category:     "manual",
		defaultValue: "127.0.0.1:8080",
		label:        "<address>:<port>",
		description: `
Specifies the address and the port number of a daemon in a form <code><var>address</var>:<var>port</var></code>.
`,
	},
	"pid": {
		category:     "common",
		defaultValue: "",
		label:        "<file>",
		description: `
Specifies a file where PID is written to.
`,
	},
	"access_log": {
		category:     "common",
		defaultValue: "",
		label:        "<file>",
		description: `
Specifies a file where API access log is written to.  It defaults to standard output.

Each line in the file is a JSON string corresponds to a single log item.
`,
	},
	"access_log_tag": {
		category:     "common",
		defaultValue: "fireworq.access",
		label:        "<tag>",
		description: `
Specifies the value of ` + "`" + `tag` + "`" + ` field in a access log item.
`,
	},
	"error_log": {
		category:     "common",
		defaultValue: "",
		label:        "<file>",
		description: `
Specifies a file where error logs are written to.  It defaults to standard error output.

If this value is specified, each line in the file is a JSON string corresponds to a single log item.  Otherwise, each line of the output is a prettified log item.
`,
	},
	"error_log_level": {
		category:     "common",
		defaultValue: "",
		label:        "<level>",
		description: `
Specifies a log level of the access log.  ` + logLevelDescription + `
If none of these values is specified, the level is determined by ` + "`" + `DEBUG` + "`" + ` environment variable.  If ` + "`" + `DEBUG` + "`" + ` has a non-empty value, then the level is ` + "`" + `debug` + "`" + `.  Otherwise, the level is ` + "`" + `info` + "`" + `.
`,
	},
	"shutdown_timeout": {
		category:     "manual",
		defaultValue: "30",
		label:        "<seconds>",
		description: `
Specifies a timeout, in seconds, which the daemon waits on [gracefully shutting down or restarting][section-graceful-restart].
`,
	},
	"keep_alive": {
		category:     "common",
		defaultValue: "false",
		label:        "true|false",
		description: `
Specifies whether connections should be reused.
`,
	},
	"config_refresh_interval": {
		category:     "manual",
		defaultValue: "1000",
		label:        "<milliseconds>",
		description: `
Specifies an interval, in milliseconds, at which a Fireworq daemon checks if configurations (such as queue definitions or routings) are changed by other daemons.
`,
	},
	"driver": {
		category:     "manual",
		defaultValue: "mysql",
		label:        "<driver>",
		description: `
Specifies a driver for job queues and repositories.  The available values are ` + "`" + `mysql` + "`" + ` and ` + "`in-memory`" + `.

Note that ` + "`in-memory`" + ` driver is not for production use.  It is intended to be used for just playing with Fireworq without a storage middleware or to show the upper bound of performance in a benchmark.
`,
	},
	"mysql_dsn": {
		category:     "manual",
		defaultValue: "tcp(localhost:3306)/fireworq",
		label:        "<DSN>",
		description: `
Specifies a data source name for the job queue and the repository database in a form <code><var>user</var>:<var>password</var>@tcp(<var>mysql_host</var>:<var>mysql_port</var>)/<var>database</var>?<var>options</var></code>.  This is in effect only when [the driver](#env-driver) is ` + "`" + `mysql` + "`" + ` and is mandatory for that case.
`,
	},
	"repository_mysql_dsn": {
		category:     "manual",
		defaultValue: "",
		label:        "<DSN>",
		description: `
Specifies a data source name for the repository database in a form <code><var>user</var>:<var>password</var>@tcp(<var>mysql_host</var>:<var>mysql_port</var>)/<var>database</var>?<var>options</var></code>.  This is in effect only when the [driver](#env-driver) is ` + "`" + `mysql` + "`" + ` and overrides [the default DSN](#env-mysql-dsn).  This should be used when you want to specify a DSN differs from [the queue DSN](#env-queue-mysql-dsn).
`,
	},
	"queue_default": {
		category:     "common",
		defaultValue: "",
		label:        "<name>",
		description: `
Specifies the name of a default queue.  A job whose ` + "`" + `category` + "`" + ` is not defined via the [routing API][api-put-routing] will be delivered to this queue.  If no default queue name is specified, pushing a job with an unknown category will fail for a [manual setup][section-manual-setup].  A docker-composed instance uses ` + "`" + `default` + "`" + ` as a default value.

If you already have a queue with the specified name in the job queue database, that one is used.  Or otherwise a new queue is created automatically.
`,
	},
	"queue_default_polling_interval": {
		category:     "common",
		defaultValue: "200",
		label:        "<milliseconds>",
		description: `
Specifies the default interval, in milliseconds, at which Fireworq checks the arrival of new jobs, used when ` + "`" + `polling_interval` + "`" + ` in the [queue API][api-put-queue] is omitted.
`,
	},
	"queue_default_max_workers": {
		category:     "common",
		defaultValue: "20",
		label:        "<number>",
		description: `
Specifies the default maximum number of jobs that are processed simultaneously in a queue, used when ` + "`" + `max_workers` + "`" + ` in the [queue API][api-put-queue] is omitted.
`,
	},
	"queue_log": {
		category:     "common",
		defaultValue: "",
		label:        "<file>",
		description: `
Specifies a file where the job queue logs are written to.  It defaults to standard output. No other logs than the job queue logs are written to this file.

Each line in the file is a JSON string corresponds to a single log item.
`,
	},
	"queue_log_tag": {
		category:     "common",
		defaultValue: "fireworq.queue",
		label:        "<tag>",
		description: `
Specifies the value of ` + "`" + `tag` + "`" + ` field in a job queue log item JSON.
`,
	},
	"queue_log_level": {
		category:     "common",
		defaultValue: "",
		label:        "<level>",
		description: `
Specifies a log level of the job queue logs.  ` + logLevelDescription + `
If none of these values is specified, the level is determined by ` + "`" + `DEBUG` + "`" + ` environment variable.  If ` + "`" + `DEBUG` + "`" + ` has a non-empty value, then the level is ` + "`" + `debug` + "`" + `.  Otherwise, the level is ` + "`" + `info` + "`" + `.
`,
	},
	"queue_mysql_dsn": {
		category:     "manual",
		defaultValue: "",
		label:        "<DSN>",
		description: `
Specifies a data source name for the job queue database in a form <code><var>user</var>:<var>password</var>@tcp(<var>mysql_host</var>:<var>mysql_port</var>)/<var>database</var>?<var>options</var></code>.  This is in effect only when the [driver](#env-driver) is ` + "`" + `mysql` + "`" + ` and overrides [the default DSN](#env-mysql-dsn).  This should be used when you want to specify a DSN differs from [the repository DSN](#env-repository-mysql-dsn).
`,
	},
	"dispatch_user_agent": {
		category:     "common",
		defaultValue: "",
		label:        "<agent>",
		description: `
Specifies the value of ` + "`" + `User-Agent` + "`" + ` header field used for an HTTP request to a worker.  The default value is <code>Fireworq/<var>version</var></code>.
`,
	},
	"dispatch_keep_alive": {
		category: "common",
		label:    "true|false",
		description: `
Specifies whether a connection to a worker should be reused.  This overrides [the default keep-alive setting](#env-keep-alive).
`,
	},
	"dispatch_max_conns_per_host": {
		category:     "common",
		defaultValue: "10",
		label:        "<number>",
		description: `
Specifies maximum idle connections to keep per-host. This value work only when [connections of the dispatcher are reused](#env-dispatch-keep-alive).
`,
	},
	"dispatch_idle_conn_timeout": {
		category:     "common",
		defaultValue: "0",
		label:        "<seconds>",
		description: `
Specifies the maximum amount of time of an idle (keep-alive) connection will remain idle before closing itself. If zero, an idle connections will not be closed. 
`,
	},
}
