Contributing to Fireworq
========================

We appreciate your pull requests!

## Using docker

The [Docker][] environment explained in [the first section][section-start] can also be used for developing.  Each time you run `script/docker/compose up`, a Fireworq instance with the latest code will come up.  Note that sometimes you need `script/docker/compose clean` for example when you add a new dependency.

To run automated tests, the following command does it for you.

```
$ script/ci/test/docker-run
```

This command always clean up all the docker images used for the tests.  If you are going to run the tests many times in your iteration, you should consider manual compilation and testing explained in the next section.

## Manual compilation and testing

First make sure that you have

- [Golang][] environment,
- [Glide][] installed, and,
- [MySQL][] installed and prepared [as described in the manual setup guide][manual-setup-mysql].

Now you can start developing with the following commands.

- `make build` to build a Fireworq binary.
- <code>FIREWORQ_MYSQL_DSN=<var>dsn</var> make test</code> to run tests.
- <code>FIREWORQ_MYSQL_DSN=<var>dsn</var> ./fireworq</code> to run a Fireworq instance locally.

## Coding conventions

- Use [`golint`][golint], [`go vet`][govet] and [`gofmt -s`][gofmt]
- `make clean lint` does these for you

[section-start]: ./README.md#start
[manual-setup-mysql]: ./doc/production.md#manual-setup-mysql

[Docker]: https://www.docker.com/
[Golang]: https://golang.org/
[Glide]: https://github.com/Masterminds/glide
[MySQL]: https://www.mysql.com/
[golint]: https://github.com/golang/lint
[govet]: https://golang.org/cmd/vet/
[gofmt]: https://golang.org/cmd/gofmt/
