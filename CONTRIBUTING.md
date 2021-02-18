Contributing to Fireworq
========================

We appreciate your pull requests!

## Using docker

Run the following commands and you will get the whole system working
all at once.  Make sure you have [Docker][] and [Docker Compose][]
installed before running these commands.

```
$ git clone https://github.com/fireworq/fireworq
$ cd fireworq
$ script/docker/compose up
```

When Fireworq gets ready, it will listen on `localhost:8080` (on the
host machine).  Specify `FIREWORQ_PORT` environment variable if you
want Fireworq to listen on a different port.

Each time you run `script/docker/compose up`, a Fireworq instance with
the latest code will come up.  Note that sometimes you need
`script/docker/compose clean` for example when you add a new
dependency.

To run automated tests, the following command does it for you.

```
$ script/ci/test/docker-run
```

This command always clean up all the docker images used for the tests.
If you are going to run the tests many times in your iteration, you
should consider manual compilation and testing explained in the next
section.

## Manual compilation and testing

First make sure that you have

- [Golang][] environment,
- [MySQL][] installed and prepared [as described in the manual setup guide][manual-setup-mysql].

Now you can start developing with the following commands.

- `make build` to build a Fireworq binary.
- <code>FIREWORQ_MYSQL_DSN=<var>dsn</var> make test</code> to run tests.
- <code>FIREWORQ_MYSQL_DSN=<var>dsn</var> ./fireworq</code> to run a Fireworq instance locally.

## Coding conventions

- Use [`golint`][golint], [`go vet`][govet] and [`gofmt -s`][gofmt]
- `make clean lint` does these for you

## Release flow

1. Make sure that the current version in [`version.go`](./version.go)
   matches with what is expected to be released.

   - Follow [semver][] convention to decide the new version.
   - Do `gobump minor -w` or `gobump major -w` to increment.
   - Push the change in `version.go` and wait for @fireworq-bot to update `AUTHORS.md`.

2. Tag the new version on the master branch.

   - `git checkout master && git pull`
   - <code>git tag v<var>X</var>.<var>Y</var>.<var>Z</var></code>
   - `git push --tags`

3. [A new release][releases] and [a new Docker image](https://hub.docker.com/r/fireworq/fireworq) are published by [Release workflow](https://github.com/fireworq/fireworq/actions?query=workflow%3ARelease).

4. Increment the patch version to state that the current source tree is not released yet.

   - `gobump patch -w`

[section-start]: ./README.md#start
[manual-setup-mysql]: ./doc/production.md#manual-setup-mysql

[releases]: https://github.com/fireworq/fireworq/releases

[Docker]: https://www.docker.com/
[Docker Compose]: https://docs.docker.com/compose/
[Golang]: https://golang.org/
[MySQL]: https://www.mysql.com/
[golint]: https://github.com/golang/lint
[govet]: https://golang.org/cmd/vet/
[gofmt]: https://golang.org/cmd/gofmt/
[semver]: https://semver.org/
