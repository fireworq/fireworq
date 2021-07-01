![Fireworq][logo]
=================

[![Build Status](https://github.com/fireworq/fireworq/workflows/CI/badge.svg)](https://github.com/fireworq/fireworq/actions)
[![Coverage Status](https://coveralls.io/repos/github/fireworq/fireworq/badge.svg?branch=master)](https://coveralls.io/github/fireworq/fireworq?branch=master)
[![Docker Hub](https://img.shields.io/badge/docker%20build-ready-blue)](https://hub.docker.com/r/fireworq/fireworq)

Fireworq is a lightweight, high-performance job queue system with the
following abilities.

- **Portability** - It is available from ANY programming language
  which can talk HTTP.  It works with a single binary without external
  dependencies.

- **Reliability** - It is built on top of RDBMS (MySQL), so that jobs
  won't be lost even if the job queue process dies.  You can apply an
  ordinary replication scheme to the underlying DB for the reliability
  of the DB itself.

- **Availability** - It supports primary/backup nodes.  Only one node
  becomes primary simultaneously and the others become backup nodes.
  A backup node will automatically be active when the primary node
  dies.

- **Scalability** - It always works with a single dispatcher per queue
  which can concurrently dispatch jobs to workers via HTTP.
  Scalability of workers themselves should be maintained by a load
  balancer in the ordinary way.  This means that adding a worker will
  never harm performance of grabbing jobs from a queue.

- **Flexibility** - It supports the following features.

  - **Multiple queues** - You can define multiple queues and use them
    in different ways: for example, one for a low priority queue for a
    limited number of high latency workers and another one for a high
    priority queue for a large number of low latency workers.
  - **Delayed jobs** - You can specify a delay for each job, which
    will make the job dispatched after the delay.
  - **Job retrying** - You can specify the maximum number of retries
    for each job.

- **Maintainability** - It can be managed on [a Web UI][Fireworqonsole].  It also [provides metrics suitable for monitoring][section-monitoring].

----

- [Getting Started][section-start]
- [Using the API][section-api]
- [Inspecting Running Queues][section-inspecting]
- [Configuration][section-configuration]
- [Other Topics][section-other]
  - [Full List of API Endpoints][page-api]
  - [Full List of Configurations][page-configuration]
  - [Make It Production-Ready][page-production-ready]
  - [Developer Guide][page-developing]
- [License][section-license]

## <a name="start">Getting Started</a>

Run the following commands and you will get the whole system working
all at once.  Make sure you have [Docker][] installed before running
these commands.

```
$ docker run -p 8080:8080 fireworq/fireworq --driver=in-memory --queue-default=default
```

Pressing `Ctrl+C` will gracefully shut it down.

Note that `in-memory` driver is not for production use.  It is
intended to be used for just playing with Fireworq without a storage
middleware.

## <a name="api">Using the API</a>

### Preparing a Worker

First of all, you need a Web server which does an actual work for a
job.  We call it a 'worker'.

A worker must accept a `POST` request with a body, which is typically
a JSON value, and respond a JSON result.  For example, if you have a
worker at `localhost:3000`, it must handle a request like the
following.

```http
POST /work HTTP/1.1
Host: localhost:3000

{"id":12345}
```

```http
HTTP/1.1 200 OK

{"status":"success","message":"It's working!"}
```

The response JSON must have `status` field, which describes whether
the job  succeeded.  It must be one of the following values.

|Value                |Meaning                                 |
|:--------------------|:---------------------------------------|
|`"success"`          |The job succeeded.                      |
|`"failure"`          |The job failed and it can be retried.   |
|`"permanent-failure"`|The job failed and it cannot be retried.|

Any other values are regarded as `"failure"`.  The HTTP status code is
always ignored.

### Enqueuing a Job to Fireworq

Let's make the job asynchronous using Fireworq.  All you have to do is
to make a `POST` request to Fireworq with a worker URL and a job
payload.  If you have [a Fireworq docker instance][section-start] and
your docker host IP (from the container's point of view) is
`172.17.0.1`, then requesting something like the following will
enqueue exactly the same job in the previous example.

```
$ curl -XPOST -d '{"url":"http://172.17.0.1:3000/work","payload":{"id":12345}}' http://localhost:8080/job/foo
```

When Fireworq gets ready to grab this job, it will `POST` the
`payload` to the `url`.  When the job is completed on the worker, the
log output of Fireworq should say something like this.

```
fireworq_1  | {"level":"info","time":1507128673123,"tag":"fireworq.dev","action":"complete","queue":"default","category":"foo","id":2,"status":"completed","created_at":1507128673025,"elapsed":98,"url":"http://172.17.0.1:3000/work","payload":"{\"id\":12345}","next_try":1507128673025,"retry_count":0,"retry_delay":0,"fail_count":0,"timeout":0,"message":"It's working!"}
```

### Further Reading

See [the full list of API endpoints][page-api] for the details of the
API.

## <a name="inspecting">Inspecting Running Queues</a>

There is only a set of API endpoints provided by Fireworq itself to
inspect running queues.  They are useful for machine monitoring but
not intended for human use.

Instead, use [Fireworqonsole][], a powerful Web UI which enables
monitoring stats of queues, inspecting running or failed jobs and
defining queues and routings.

> ![Web UI](https://github.com/fireworq/fireworqonsole/raw/master/doc/images/console.png "Web UI")

## <a name="config">Configuration</a>

You can configure Fireworq by providing environment variables on
starting a daemon.  There are many of them but we only describe
important ones here.  See [the full list][page-configuration] for the
other variables.

- `FIREWORQ_MYSQL_DSN`

  Specifies a data source name for the job queue and the repository
  database in a form
  <code><var>user</var>:<var>password</var>@tcp(<var>mysql_host</var>:<var>mysql_port</var>)/<var>database</var>?<var>options</var></code>.
  This is for a manual setup and is mandatory for it.

- `FIREWORQ_QUEUE_DEFAULT`

  Specifies the name of a default queue.  A job whose `category` is
  not defined via the [routing API][api-put-routing] will be delivered
  to this queue.  If no default queue name is specified, pushing a job
  with an unknown category will fail.

  If you already have a queue with the specified name in the job queue
  database, that one is used.  Or otherwise a new queue is created
  automatically.

- `FIREWORQ_QUEUE_DEFAULT_POLLING_INTERVAL`

  Specifies the default interval, in milliseconds, at which Fireworq
  checks the arrival of new jobs, used when `polling_interval` in the
  [queue API][api-put-queue] is omitted.  The default value is `200`.

- `FIREWORQ_QUEUE_DEFAULT_MAX_WORKERS`

  Specifies the default maximum number of jobs that are processed
  simultaneously in a queue, used when `max_workers` in the
  [queue API][api-put-queue] is omitted.  The default value is `20`.

## <a name="other">Other Topics</a>

- [Full List of API Endpoints][page-api]
  - [Queue Management][section-api-queue]
  - [Routing Management][section-api-routing]
  - [Job Management][section-api-job]
- [Full List of Configurations][page-configuration]
- [Make It Production-Ready][page-production-ready]
  - [Using a Release Build (Manual setup)][section-manual-setup]
  - [Preparing a Backup Instance][section-backup]
  - [Graceful Shutdown/Restart][section-graceful-restart]
  - [Logging][section-logging]
  - [Monitoring][section-monitoring]
- [Developer Guide][page-developing]

## <a name="license">License</a>

- Copyright (c) 2017 The [Fireworq Authors][authors]. All rights reserved.
- Fireworq is licensed under the Apache License, Version 2.0. See
  [LICENSE][license] for the full license text.

[section-start]: #start
[section-configuration]: #config
[section-api]: #api
[section-inspecting]: #inspecting
[section-other]: #other
[section-license]: #license

[page-configuration]: ./doc/config.md
[page-api]: ./doc/api.md
[section-api-queue]: ./doc/api.md#api-queue
[section-api-routing]: ./doc/api.md#api-routing
[section-api-job]: ./doc/api.md#api-job
[page-production-ready]: ./doc/production.md
[section-manual-setup]: ./doc/production.md#manual-setup
[section-backup]: ./doc/production.md#backup
[section-graceful-restart]: ./doc/production.md#graceful-restart
[section-logging]: ./doc/production.md#logging
[section-monitoring]: ./doc/production.md#monitoring
[page-developing]: ./CONTRIBUTING.md

[api-put-queue]: ./doc/api.md#api-put-queue
[api-put-routing]: ./doc/api.md#api-put-routing

[logo]: ./doc/images/logo.png "Fireworq"
[license]: ./LICENSE
[authors]: ./AUTHORS.md

[Docker]: https://www.docker.com/
[Fireworqonsole]: https://github.com/fireworq/fireworqonsole
