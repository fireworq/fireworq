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

3. Wait for [a new release published][releases] by [CI](https://github.com/fireworq/fireworq/blob/97dd254792ec36a648732c33c068d979772804e4/.travis.yml#L25).

4. Increment the patch version to state that the current source tree is not released yet.

   - `gobump patch -w`

[section-start]: ./README.md#start
[manual-setup-mysql]: ./doc/production.md#manual-setup-mysql

[releases]: https://github.com/fireworq/fireworq/releases

[Docker]: https://www.docker.com/
[Golang]: https://golang.org/
[MySQL]: https://www.mysql.com/
[golint]: https://github.com/golang/lint
[govet]: https://golang.org/cmd/vet/
[gofmt]: https://golang.org/cmd/gofmt/
[semver]: https://semver.org/
