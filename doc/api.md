API
===

All operations on Fireworq are done via HTTP API.  The following
operations are supported.

- [Queue Management][section-api-queue]
  - [`GET /queues`](#api-get-queues)
  - [`GET /queues/stats`](#api-get-queues-stats)
  - [<code>GET /queue/<var>{queue_name}</var></code>](#api-get-queue)
  - [<code>PUT /queue/<var>{queue_name}</var></code>](#api-put-queue)
  - [<code>DELETE /queue/<var>{queue_name}</var></code>](#api-delete-queue)
  - [<code>GET /queue/<var>{queue_name}</var>/node</code>](#api-get-queue-node)
  - [<code>GET /queue/<var>{queue_name}</var>/stats</code>](#api-get-queue-stats)
- [Routing Management][section-api-routing]
  - [`GET /routings`](#api-get-routings)
  - [<code>GET /routing/<var>{job_category}</var></code>](#api-get-routing)
  - [<code>PUT /routing/<var>{job_category}</var></code>](#api-put-routing)
  - [<code>DELETE /routing/<var>{job_category}</var></code>](#api-delete-routing)
- [Job Management][section-api-job]
  - [<code>GET /queue/<var>{queue_name}</var>/grabbed</code>](#api-get-queue-grabbed)
  - [<code>GET /queue/<var>{queue_name}</var>/waiting</code>](#api-get-queue-waiting)
  - [<code>GET /queue/<var>{queue_name}</var>/deferred</code>](#api-get-queue-deferred)
  - [<code>GET /queue/<var>{queue_name}</var>/job/<var>{id}</var></code>](#api-get-queue-job)
  - [<code>DELETE /queue/<var>{queue_name}</var>/job/<var>{id}</var></code>](#api-delete-queue-job)
  - [<code>GET /queue/<var>{queue_name}</var>/failed</code>](#api-get-queue-failed)
  - [<code>GET /queue/<var>{queue_name}</var>/failed/<var>{id}</var></code>](#api-get-queue-failed-job)
  - [<code>DELETE /queue/<var>{queue_name}</var>/failed/<var>{id}</var></code>](#api-delete-queue-failed-job)
  - [<code>POST /job/<var>{job_category}</var></code>](#api-post-job)

## <a name="api-queue">Queue Management</a>

### <a name="api-get-queues">`GET /queues`</a>

Returns defined queues.

```http
GET /queues HTTP/1.1
```

```http
HTTP/1.1 200 OK

[{
   "name": "test_queue1",
   "polling_interval": 100,
   "max_workers": 10
}, {
   "name": "test_queue2",
   "polling_interval": 200,
   "max_workers": 20
}]
```

### <a name="api-get-queues-stats">`GET /queues/stats`</a>

Returns stats of queues.

```http
GET /queues/stats HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "test_queue1": {
        "total_pushes": 10,
        "total_pops": 8,
        "total_successes": 3,
        "total_failures": 2,
        "total_permanent_failures": 1,
        "total_completes": 5,
        "total_elapsed": 718,
        "pushes_per_second": 2,
        "pops_per_second": 1,
        "outstanding_jobs": 0,
        "total_workers": 10,
        "idle_workers": 7,
        "active_nodes": 1
    },
    "test_queue2": {
        "total_pushes": 100,
        "total_pops": 100,
        "total_successes": 32,
        "total_failures": 0,
        "total_permanent_failures": 0,
        "total_completes": 32,
        "total_elapsed": 10944,
        "pushes_per_second": 10,
        "pops_per_second": 10,
        "outstanding_jobs": 48,
        "total_workers": 20,
        "idle_workers": 0,
        "active_nodes": 1
    },
    "test_queue3": {
        "total_pushes": 1,
        "total_pops": 0,
        "total_successes": 0,
        "total_failures": 0,
        "total_permanent_failures": 0,
        "total_completes": 0,
        "total_elapsed": 0,
        "pushes_per_second": 0,
        "pops_per_second": 0,
        "outstanding_jobs": 0,
        "total_workers": 30,
        "idle_workers": 29,
        "active_nodes": 1
    }
}
```

### <a name="api-get-queue"><code>GET /queue/<var>{queue_name}</var></code></a>

Returns the definition of a queue.

```http
GET /queue/test_queue1 HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
   "name": "test_queue1",
   "polling_interval": 100,
   "max_workers": 10
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`404 Not Found`          |The target queue is undefined.       |

### <a name="api-put-queue"><code>PUT /queue/<var>{queue_name}</var></code></a>

Creates a new queue or override the definition of an existing queue.

After putting a new queue, it may not be available immediately under [clustering multiple instances][section-backup].  In such case, a queue put to a host becomes available on another host after at most [`FIREWORQ_CONFIG_REFRESH_INTERVAL`][env-config-refresh-interval].

```http
PUT /queue/test_queue1 HTTP/1.1

{
   "polling_interval": 100,
   "max_workers": 10
}
```

```http
HTTP/1.1 200 OK

{
   "name": "test_queue1",
   "polling_interval": 100,
   "max_workers": 10
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`polling_interval`       |An interval, in milliseconds, at which Fireworq checks the arrival of new jobs in this queue.|optional, defaults to [`FIREWORQ_QUEUE_DEFAULT_POLLING_INTERVAL`][env-queue-default-polling-interval]|
|`max_workers`            |The maximum number of jobs that are processed simultaneously for this queue.|optional, defaults to [`FIREWORQ_QUEUE_DEFAULT_MAX_WORKERS`][env-queue-default-max-workers]|

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`400 Bad Request`        |A request parameter is invalid or missing.|

### <a name="api-delete-queue"><code>DELETE /queue/<var>{queue_name}</var></code></a>

Deletes a queue.

Deleting a queue will not delete routings related to the queue; you are responsible to [delete][api-delete-routing] or [modify][api-put-routing] them before deleting the queue.

Deleting a queue will not truncate the job list data store of the queue; it will be recovered if you recreate a queue of the same name.  Under [clustering multiple instances][section-backup], other instances may continue pushing jobs into the queue until the queue definition is synchronized.  This happens at most [`FIREWORQ_CONFIG_REFRESH_INTERVAL`][env-config-refresh-interval] after the queue has been deleted.

If you delete [the default queue][env-queue-default], the queue will be recreated when the queue definitions are reloaded.

```http
DELETE /queue/test_queue1 HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
   "name": "test_queue1",
   "polling_interval": 100,
   "max_workers": 10
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`404 Not Found`          |The target queue is undefined.       |

### <a name="api-get-queue-node"><code>GET /queue/<var>{queue_name}</var>/node</code></a>

Returns information of a node which is active on a queue.

```http
GET /queue/test_queue1/node HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "id": "104",
    "host": "172.17.0.1"
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`404 Not Found`          |The target queue is undefined or not working, or there is no node active on the queue.|

### <a name="api-get-queue-stats"><code>GET /queue/<var>{queue_name}</var>/stats</code></a>

Returns stats of a queue.

```http
GET /queue/test_queue1/stats HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "total_pushes": 10,
    "total_pops": 8,
    "total_successes": 3,
    "total_failures": 2,
    "total_permanent_failures": 1,
    "total_completes": 5,
    "total_elapsed": 718,
    "pushes_per_second": 2,
    "pops_per_second": 1,
    "total_workers": 10,
    "idle_workers": 7,
    "active_nodes": 1
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |

|Response code            |Meaning                                      |
|:------------------------|:--------------------------------------------|
|`404 Not Found`          |The target queue is undefined or not working.|

## <a name="api-routing">Routing Management</a>

### <a name="api-get-routings">`GET /routings`</a>

Returns defined routings.

```http
GET /routings HTTP/1.1
```

```http
HTTP/1.1 200 OK

[{
    "job_category": "test_job1",
    "queue_name": "test_queue1"
}, {
    "job_category": "test_job2",
    "queue_name": "test_queue2"
}, {
    "job_category": "test_job3",
    "queue_name": "test_queue2"
}]
```

### <a name="api-get-routing"><code>GET /routing/<var>{job_category}</var></code></a>

Returns the definition of a routing.

```http
GET /routing/test_job1 HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "queue_name": "test_queue1",
    "job_category": "test_job1"
}
```

|Field in the request|Meaning                              |Note               |
|:-------------------|:------------------------------------|:------------------|
|`job_category`      |The name of the target job category. |mandatory          |

|Response code            |Meaning                                 |
|:------------------------|:---------------------------------------|
|`404 Not Found`          |No routing of `job_category` is defined.|

### <a name="api-put-routing"><code>PUT /routing/<var>{job_category}</var></code></a>

Creates a new routing or override the definition of an existing routing.

After putting a new routing, it may not be available immediately under [clustering multiple instances][section-backup].  In such case, a routing put to a host becomes available on another host after at most [`FIREWORQ_CONFIG_REFRESH_INTERVAL`][env-config-refresh-interval].

```http
PUT /routing/test_job1 HTTP/1.1

{
    "queue_name": "test_queue1"
}
```

```http
HTTP/1.1 200 OK

{
    "queue_name": "test_queue1",
    "job_category": "test_job1"
}
```

|Field in the request|Meaning                              |Note               |
|:-------------------|:------------------------------------|:------------------|
|`job_category`      |A category of a job which will be delivered to a queue of `queue_name`.|mandatory|
|`queue_name`        |A name of a queue to which a job of `job_category` will be delivered.|mandatory|

|Response code            |Meaning                                   |
|:------------------------|:-----------------------------------------|
|`400 Bad Request`        |A request parameter is invalid or missing.|
|`404 Not Found`          |No queue of `queue_name` is defined.      |

### <a name="api-delete-routing"><code>DELETE /routing/<var>{job_category}</var></code></a>

Deletes the routing of a job category.

```http
DELETE /routing/test_job1 HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "queue_name": "test_queue1",
    "job_category": "test_job1"
}
```

|Field in the request|Meaning                              |Note               |
|:-------------------|:------------------------------------|:------------------|
|`job_category`      |The name of the target job category. |mandatory          |

|Response code            |Meaning                                 |
|:------------------------|:---------------------------------------|
|`404 Not Found`          |No routing of `job_category` is defined.|

## <a name="api-job">Job Management</a>

### <a name="api-get-queue-grabbed"><code>GET /queue/<var>{queue_name}</var>/grabbed</code></a>

Returns a list of grabbed jobs in a queue.  Grabbed jobs are running
or be prepared to run.

```http
GET /queue/test_queue1/grabbed?limit=10&cursor=MTQ5NzUxMDc4NiwxMw%3D%3D&order=desc HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "jobs": [{
        "id": 4,
        "category": "test",
        "url": "http://example.com/",
        "status": "grabbed",
        "created_at": "2017-06-26T00:51:54.537+09:00",
        "next_try": "2017-06-26T01:03:34.537+09:00",
        "timeout": 0,
        "fail_count": 0,
        "max_retries": 0,
        "retry_delay": 0
    }, {
        "id": 2,
        "category": "test",
        "url": "http://example.com/",
        "status": "grabbed",
        "created_at": "2017-06-26T00:51:26.33+09:00",
        "next_try": "2017-06-26T00:59:46.571+09:00",
        "timeout": 0,
        "fail_count": 1,
        "max_retries": 3,
        "retry_delay": 500
    }, {
        "id": 1,
        "category": "test",
        "url": "http://example.com/",
        "status": "grabbed",
        "created_at": "2017-06-26T00:51:14.724+09:00",
        "next_try": "2017-06-26T00:59:35.207+09:00",
        "timeout": 0,
        "fail_count": 1,
        "max_retries": 3,
        "retry_delay": 500
    }],
    "next_cursor": "MTQ5ODQwNjIwNDIxMSwz"
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`limit`                  |The maximum number of the jobs.      |default: `100`|
|`cursor`                 |A cursor to retrieve next items since the previous request.  Specify the value of `next_cursor` field in the previous response.|optional|
|`order`                  |Sort order of the jobs. `asc` or `desc` |default:`desc`|

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`404 Not Found`          |The target queue is undefined or not working.|
|`501 Not Implemented`    |Job inspection feature is not supported with this [driver][env-driver].|

### <a name="api-get-queue-waiting"><code>GET /queue/<var>{queue_name}</var>/waiting</code></a>

Returns a list of waiting jobs in a queue.  Waiting jobs are queued
but not grabbed yet just because there is not enough space or time to
grab them.

```http
GET /queue/test_queue1/waiting?limit=10&cursor=MTQ5NzUxMDc4NiwxMw%3D%3D&order=desc HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "jobs": [{
        "id": 4,
        "category": "test",
        "url": "http://example.com/",
        "status": "claimed",
        "created_at": "2017-06-26T00:51:54.537+09:00",
        "next_try": "2017-06-26T01:03:34.537+09:00",
        "timeout": 0,
        "fail_count": 0,
        "max_retries": 0,
        "retry_delay": 0
    }, {
        "id": 2,
        "category": "test",
        "url": "http://example.com/",
        "status": "claimed",
        "created_at": "2017-06-26T00:51:26.33+09:00",
        "next_try": "2017-06-26T00:59:46.571+09:00",
        "timeout": 0,
        "fail_count": 1,
        "max_retries": 3,
        "retry_delay": 500
    }, {
        "id": 1,
        "category": "test",
        "url": "http://example.com/",
        "status": "claimed",
        "created_at": "2017-06-26T00:51:14.724+09:00",
        "next_try": "2017-06-26T00:59:35.207+09:00",
        "timeout": 0,
        "fail_count": 1,
        "max_retries": 3,
        "retry_delay": 500
    }],
    "next_cursor": "MTQ5ODQwNjIwNDIxMSwz"
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`limit`                  |The maximum number of the jobs.      |default: `100`|
|`cursor`                 |A cursor to retrieve next items since the previous request.  Specify the value of `next_cursor` field in the previous response.|optional|
|`order`                  |Sort order of the jobs. `asc` or `desc` |default:`desc`|

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`404 Not Found`          |The target queue is undefined or not working.|
|`501 Not Implemented`    |Job inspection feature is not supported with this [driver][env-driver].|

### <a name="api-get-queue-deferred"><code>GET /queue/<var>{queue_name}</var>/deferred</code></a>

Returns a list of deferred jobs in a queue.  Deferred jobs are not
going to run for now because of specified delays.

```http
GET /queue/test_queue1/deferred?limit=10&cursor=MTQ5NzUxMDc4NiwxMw%3D%3D&order=desc HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "jobs": [{
        "id": 4,
        "category": "test",
        "url": "http://example.com/",
        "status": "claimed",
        "created_at": "2017-06-26T00:51:54.537+09:00",
        "next_try": "2017-06-26T01:03:34.537+09:00",
        "timeout": 0,
        "fail_count": 0,
        "max_retries": 0,
        "retry_delay": 0
    }, {
        "id": 2,
        "category": "test",
        "url": "http://example.com/",
        "status": "claimed",
        "created_at": "2017-06-26T00:51:26.33+09:00",
        "next_try": "2017-06-26T00:59:46.571+09:00",
        "timeout": 0,
        "fail_count": 1,
        "max_retries": 3,
        "retry_delay": 500
    }, {
        "id": 1,
        "category": "test",
        "url": "http://example.com/",
        "status": "claimed",
        "created_at": "2017-06-26T00:51:14.724+09:00",
        "next_try": "2017-06-26T00:59:35.207+09:00",
        "timeout": 0,
        "fail_count": 1,
        "max_retries": 3,
        "retry_delay": 500
    }],
    "next_cursor": "MTQ5ODQwNjIwNDIxMSwz"
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`limit`                  |The maximum number of the jobs.      |default: `100`|
|`cursor`                 |A cursor to retrieve next items since the previous request.  Specify the value of `next_cursor` field in the previous response.|optional|
|`order`                  |Sort order of the jobs. `asc` or `desc` |default:`desc`|

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`404 Not Found`          |The target queue is undefined or not working.|
|`501 Not Implemented`    |Job inspection feature is not supported with this [driver][env-driver].|

### <a name="api-get-queue-job"><code>GET /queue/<var>{queue_name}</var>/job/<var>{id}</var></code></a>

Returns a job in a queue.

```http
GET /queue/test_queue1/job/2
```

```http
HTTP/1.1 200 OK

{
    "id": 2,
    "category": "test",
    "url": "http://example.com/",
    "status": "grabbed",
    "created_at": "2017-06-26T00:51:26.33+09:00",
    "next_try": "2017-06-26T00:59:46.571+09:00",
    "timeout": 0,
    "fail_count": 1,
    "max_retries": 3,
    "retry_delay": 500
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`id`                     |The ID of the job.  This is the `id` field returned by [the job pushing API][api-post-job] or the job list APIs for [grabbed][api-get-queue-grabbed], [waiting][api-get-queue-waiting] or [deferred][api-get-queue-deferred] jobs.|mandatory     |

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`400 Bad Request`        |A request parameter is invalid or missing.|
|`404 Not Found`          |The target queue is undefined or not working, or the job is not found, possibly already has been completed and removed from the queue.|
|`501 Not Implemented`    |Job inspection feature is not supported with this [driver][env-driver].|

### <a name="api-delete-queue-job"><code>DELETE /queue/<var>{queue_name}</var>/job/<var>{id}</var></code></a>

Deletes a job in a queue.

```http
GET /queue/test_queue1/job/2
```

```http
HTTP/1.1 200 OK

{
    "id": 2,
    "category": "test",
    "url": "http://example.com/",
    "status": "grabbed",
    "created_at": "2017-06-26T00:51:26.33+09:00",
    "next_try": "2017-06-26T00:59:46.571+09:00",
    "timeout": 0,
    "fail_count": 1,
    "max_retries": 3,
    "retry_delay": 500
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`id`                     |The ID of the job.  This is the `id` field returned by [the job pushing API][api-post-job] or the job list APIs for [grabbed][api-get-queue-grabbed], [waiting][api-get-queue-waiting] or [deferred][api-get-queue-deferred] jobs.|mandatory     |

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`400 Bad Request`        |A request parameter is invalid or missing.|
|`404 Not Found`          |The target queue is undefined or not working, or the job is not found, possibly already has been completed and removed from the queue.|
|`501 Not Implemented`    |Job inspection feature is not supported with this [driver][env-driver].|

### <a name="api-get-queue-failed"><code>GET /queue/<var>{queue_name}</var>/failed</code></a>

Returns a list of failed jobs in a queue.

```http
GET /queue/test_queue1/failed?order=failed&limit=10&cursor=MTQ5NzUxMDc4NiwxMw%3D%3D HTTP/1.1
```

```http
HTTP/1.1 200 OK

{
    "failed_jobs": [{
        "id": 3,
        "job_id": 4,
        "category": "test",
        "url": "http://example.com/",
        "payload": {
            "tag": "test"
        },
        "result": {
            "status": "failure",
            "code": 200,
            "message": "Cannot parse body as JSON: invalid character '\u003c' looking for beginning of value"
        },
        "fail_count": 1,
        "failed_at": "2017-06-14T12:15:13.792+09:00",
        "created_at": "2017-06-14T12:15:12.635+09:00"
    }, {
        "id": 2,
        "job_id": 2,
        "category": "test",
        "url": "http://example.com/",
        "result": {
            "status": "failure",
            "code": 200,
            "message": "Cannot parse body as JSON: invalid character '\u003c' looking for beginning of value"
        },
        "fail_count": 1,
        "failed_at": "2017-06-14T12:03:53.636+09:00",
        "created_at": "2017-06-14T12:03:52.332+09:00"
    }, {
        "id": 1,
        "job_id": 1,
        "category": "test",
        "url": "http://example.com/",
        "result": {
            "status": "failure",
            "code": 200,
            "message": "Cannot parse body as JSON: invalid character '\u003c' looking for beginning of value"
        },
        "fail_count": 4,
        "failed_at": "2017-06-14T12:03:45.626+09:00",
        "created_at": "2017-06-14T12:03:01.258+09:00"
    }],
    "next_cursor": "MTQ5NzUxMDc4Niwz"
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`order`                  |The order of the jobs in the list.  If this value is `created`, then the most recently pushed job comes first.  Otherwise, the most recently failed job comes first.|default: `failed`|
|`limit`                  |The maximum number of the jobs.      |default: `100`|
|`cursor`                 |A cursor to retrieve next items since the previous request.  Specify the value of `next_cursor` field in the previous response.|optional|

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`404 Not Found`          |The target queue is undefined or not working.|
|`501 Not Implemented`    |Failure log feature is not supported with this [driver][env-driver].|

### <a name="api-get-queue-failed-job"><code>GET /queue/<var>{queue_name}</var>/failed/<var>{id}</var></code></a>

Returns a job in a failure log.

```http
GET /queue/test_queue1/failed/3
```

```http
HTTP/1.1 200 OK

{
    "id": 3,
    "job_id": 4,
    "category": "test",
    "url": "http://example.com/",
    "payload": {
        "tag": "test"
    },
    "result": {
        "status": "failure",
        "code": 200,
        "message": "Cannot parse body as JSON: invalid character '\u003c' looking for beginning of value"
    },
    "fail_count": 1,
    "failed_at": "2017-06-14T12:15:13.792+09:00",
    "created_at": "2017-06-14T12:15:12.635+09:00"
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`id`                     |The ID of the failure. This is the `id` field returned by [the failed job list API][api-get-queue-failed].|mandatory     |

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`400 Bad Request`        |A request parameter is invalid or missing.|
|`404 Not Found`          |The target queue is undefined or not working, or the job is not found.|
|`501 Not Implemented`    |Failure log feature is not supported with this [driver][env-driver].|

### <a name="api-delete-queue-failed-job"><code>DELETE /queue/<var>{queue_name}</var>/failed/<var>{id}</var></code></a>

Deletes a job in a failure log.

```http
GET /queue/test_queue1/failed/3
```

```http
HTTP/1.1 200 OK

{
    "id": 3,
    "job_id": 4,
    "category": "test",
    "url": "http://example.com/",
    "payload": {
        "tag": "test"
    },
    "result": {
        "status": "failure",
        "code": 200,
        "message": "Cannot parse body as JSON: invalid character '\u003c' looking for beginning of value"
    },
    "fail_count": 1,
    "failed_at": "2017-06-14T12:15:13.792+09:00",
    "created_at": "2017-06-14T12:15:12.635+09:00"
}
```

|Parameters in the request|Meaning                              |Note          |
|:------------------------|:------------------------------------|:-------------|
|`queue_name`             |The name of the target queue.        |mandatory     |
|`id`                     |The ID of the failure. This is the `id` field returned by [the failed job list API][api-get-queue-failed].|mandatory     |

|Response code            |Meaning                              |
|:------------------------|:------------------------------------|
|`400 Bad Request`        |A request parameter is invalid or missing.|
|`404 Not Found`          |The target queue is undefined or not working, or the job is not found.|
|`501 Not Implemented`    |Failure log feature is not supported with this [driver][env-driver].|

### <a name="api-post-job"><code>POST /job/<var>{job_category}</var></code></a>

Pushes a new job.

```http
POST /job/test_job1 HTTP/1.1

{
    "url": "http://example.com/process_job1",
    "payload": {
        "id": 1234,
        "value": "foo bar",
        "description": "The payload is just arbitrary data that will be passed to the target URL"
    },
    "run_after": 300,
    "max_retries": 3,
    "retry_delay": 60,
    "timeout": 30
}
```

```http
HTTP/1.1 200 OK

{
    "id": 5,
    "queue_name": "test_queue1",
    "category": "test_job1",
    "url": "http://example.com/process_job1",
    "payload": {
        "id": 1234,
        "value": "foo bar",
        "description": "The payload is just arbitrary data that will be passed to the target URL"
    },
    "run_after": 300,
    "max_retries": 3,
    "retry_delay": 60,
    "timeout": 30
}
```

|Field in the request|Meaning                              |Note               |
|:-------------------|:------------------------------------|:------------------|
|`job_category`      |The category of a job.  This name will be compared to `job_category` specified in the [routing API][api-put-routing] to decide to which queue to deliver the job.|mandatory|
|`url`               |An external destination to fire when the job is grabbed.|mandatory|
|`payload`           |A payload which will be `POST`ed to `url` on firing the job.  It can be any JSON value.  If it is a JSON string, then the raw string value not a JSON string will be a request body `POST`ed to `url`.|optional, defaults to nothing|
|`run_after`         |Seconds to wait before grabbing the job.|optional, defaults to `0`|
|`max_retries`       |The maximum number of retrying the job when the external destination returned a failure.|optional, defaults to `0`|
|`retry_delay`       |A delay in seconds to wait before grabbing the retrying job.|optional, defaults to `0`|
|`timeout`           |A timeout, in seconds, of the response from the external destination.  `0` means no timeout.|optional, defaults `0`|

|Response code            |Meaning                                   |
|:------------------------|:-----------------------------------------|
|`400 Bad Request`        |A request parameter is invalid or missing.|
|`405 Method Not Allowed` |Something other than `POST` is requested. |

[section-api-queue]: #api-queue
[section-api-routing]: #api-routing
[section-api-job]: #api-job
[section-backup]: ./production.md#backup

[api-put-routing]: #api-put-routing
[api-delete-routing]: #api-delete-routing
[api-post-job]: #api-post-job
[api-get-queue-grabbed]: #api-get-queue-grabbed
[api-get-queue-wating]: #api-get-queue-waiting
[api-get-queue-deferred]: #api-get-queue-deferred
[api-get-queue-failed]: #api-get-queue-failed

[env-config-refresh-interval]: ./config.md#env-config-refresh-interval
[env-driver]: ./config.md#env-driver
[env-queue-default]: ./config.md#env-queue-default
[env-queue-default-polling-interval]: ./config.md#env-queue-default-polling-interval
[env-queue-default-max-workers]: ./config.md#env-queue-default-max-workers
