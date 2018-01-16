# Conventions

## Multiformats

Like IPFS, self-describing standards are preferred.

[Multiformats](http://multiformats.io) defines a set of self-describing
protocols:

* multihash - self-describing hashes
* multiaddr - self-describing network addresses
* multibase - self-describing base encodings
* multicodec - self-describing serialization
* multistream - self-describing stream network protocols
* multigram (WIP) - self-describing packet network protocols

Alice uses these when possible.

The string representation of the formats should be used in documentation and
human readable files such as configuration files.
For instance, to refer to a network address, use `/ip4/127.0.0.1/tcp/5990`.

## API

The API is an exception to the last section and uses gRPC to send Protobuf
messages over the wire. This is a pragmatic choice to allow applications
developed in a wide range of programming languages to connect to a host.

## Dependencies

The [dep](https://github.com/golang/dep) dependency manager is preferred and
should be used when possible.

Libp2p and Multiaddr packages are exceptions and are usually distributed using
the [gx](https://github.com/whyrusleeping/gx) package manager, which uses IPFS
to fetch the packages.

## Git

Squash multiple commits into commits that make sense before doing a pull
request. The final commits should all pass tests.

Use commit messages of the form `package: short description`. For instance:

* `cmd: add debug flag`
* `*: update license`
* `cli+api: add peer listing`

## Logging

For consistency with libp2p, Alice uses
[go-log](https://github.com/ipfs/go-log) as the logger.

It is an event based logger (use events, not the deprecated Debug, Info,
Warning, and Error methods). Add metadata to events.

## Go Code

Use the [Go Code Review guidelines](https://github.com/golang/go/wiki/CodeReviewComments).

Watch [Go Proverbs](https://go-proverbs.github.io).

Also read [Idiomatic Go](https://dmitri.shuralyov.com/idiomatic-go).

Use a context whenever it makes sense.

Keep comment lines shorter than 80 characters. Code can be over that limit.

Do not use underscores in Go source file names except for test files such as
`api_test.go`.

If a package has `init` functions, put them at the top of the main source file
so it's easy to see them.

Keep the dependency graph as simple as possible. Remember you don't have to
explicitly implement interfaces in Go. If you are developing a package that is
meant to use types from others packages, a lot of time you don't have to
import them, especially if you only need a couple of methods from those types.
You can just define a local type with the needed methods:

```go
package a

type S struct {}

func (S) MethodA() {}

func (S) MethodB() {}
```

```go
package b

type MyType interface {
    MethodA()
}

func MyFunc(t MyType) {
}
```

In addition to simplifying the depedency graph, it also makes it possible to
use the package with any type that satisfies the interface. A general rule of
thumb is that **interfaces should be defined by the consumer**, even if it
involves a little repetition.

You can view the internal dependency graph of this package using:

```bash
$ go get -u github.com/davecheney/graphpkg
$ graphpkg -match github.com/stratumn/alice github.com/stratumn/alice
```

Use the [github.com/pkg/errors](http://github.com/pkg/errors) package to add
stack frames to all errors returned by functions from third-party packages.
For instance:

```go
if err, _ := ExternalFunc(); err != nil {
    return errors.WithStack(err)
}

// Or if it REALLY makes sense to wrap the error message:
if err, _ := ExternalFunc(); err != nil {
    return errors.Wrap(err, "third-party")
}

// Don't add a stack to an internal error that already has one.
if err, _ := InternalFunc(); err != nil {
    return err
}
```

This allows you to print detailed error messages using `%+v` in formatting
strings when debugging.

Error messages should neither begin with an upper-case letter nor end with
punctuation.

Use the standard Go test packages with assertions from
[testify](https://github.com/stretchr/testify).
Error messages should only begin with an upper-case letter if they begin with
an identifier, and should not end with punctuation.
Use the error message to specify the entity tested:

```go
assert.Equal(t, 500, status, "response http status code")
assert.True(t, ok, "type assertion should succeed")
```

When you use multiple asserts in the same test, it's worth creating an instance
to avoid having to pass `t` repeatedly:

```go
assert := assert.New(t)
assert.Equal(500, status, "response http status code")
```

Note that assertions failures are not fatal, so if you need to stop test
execution you need to use `require`:

```go
import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestSomething(t *testing.T) {
    res, err := DoSomething()
    require.NoError(t, err)

    // Assert things on res
}
```

Use table-driven tests when possible.
Use short test names that underline what is tested.

```go
type myTableTest struct {
    name            string
    // other variables specific to the test case
}

func TestCore(t *testing.T) {
    testcases := []myTableTest{{
        "with something turned on"
    }, {
        "with something turned off"
    }}

    for _, tt := range testcases {
        t.Run(tt.name, func(t *testing.T) {
            // Shared testing logic
        })
    }
}
```

When relevant, use test fixtures.
Test fixtures should provide a good description for each test case.

```go
func TestMyComponent(t *testing.T) {
    assert := assert.New(t)
    // Shared initialization of my component
    // ...

    t.Run("Does X when Y", func(t *testing.T) {
        // ...
    })

    t.Run("Returns XX when YY", func(t *testing.T) {
        // ...
    })
}
```

Use [mock](https://github.com/golang/mock) to mock interfaces for unit tests.

Check for data races using the `-race` flag when running tests.

Create packages for domains, not types.

Use singular package names, for instance `util` instead of `utils`, unless the
noun is typically plural (ex: `metrics`).

Comment your code, including unexposed types and functions. Don't describe
implementation details in your comments, as they are likely to change and
become out of sync with the code itself.

## Protobuf

Use the [Protobuf style guide](https://developers.google.com/protocol-buffers/docs/style).

## Configuration Files

Use TOML for configuration files. Use underscored lower-case names.
