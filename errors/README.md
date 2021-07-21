# Errors package

Package errors is a simple error handling facility for golang.

- [Rationale](#rationale)
- [Quick Start](#quick-start)
  - [Creating errors](#creating-errors)
  - [Wrap and WrapError](#wrap-and-wraperror)
  - [Hide and HideError](#hide-and-hideerror)
  - [Is and As](#is-and-as)
  - [Unwrap](#unwrap)
  - [UnwrapHidden](#unwraphidden)
  - [Creating custom error types](#creating-custom-error-types)
  - [Printing errors](#printing-errors)
  - [Collecting stack traces](#collecting-stack-traces)

## Rationale

There are various error handling packages out there extending the standard approach with things like 
wrapping, error chain searching, stack traces, and others.

Go 1.13 following the discussions around Go 2, already added some functionality like the ability
to wrap errors with Errorf `%w` verb, and the `Is` and `As` functions for searching into the
errors chain. But it lacks stack traces and error hiding.

Some of the available packages lack the features we need, some have more but different features than 
the ones we need.

Without further ado, here it is a list of the design goals and features that we want.
Detailed descriptions are provided by the sections that follow.

1. The errors package must be simple and thin as much as possible.
2. It must be compatible with Go 1.13 as much as possible.
3. It must support error wrapping creating an errors chain.
4. It must be possible to wrap an error with another user provided error, not just with implicitly
created wrapper based on a string.
5. It must support hiding internal and third party library errors from the chain search routines.
6. It must provide stack traces of all errors in the chain.
7. The stack traces should be available through a simple interface, not only as a formatted string.

### 1. The errors package must be simple and thin as much as possible.

Not much to say here. This must be a generic collection of routines, and must not try to solve
all kind of business problems.

### 2. It must be compatible with Go 1.13 as much as possible.

In particular we've made it compatible with `Is`, `As` and `Unwrap`, meaning that they will work 
with the error chain produced by our `Wrap`, `WrapError`, `Hide` and `HideError` functions.

We decided that we don't want to support Errorf `%w` though. 
We prefer explicit `Wrap` calls.

### 3. It must support error wrapping creating an errors chain.

This of course is one of the important features found in all the custom libraries out there.

Our chain though is a little different from the standard one, which will become clear on the next
point.

### 4. It must be possible to wrap an error with another user provided error, not just with implicitly created wrapper based on a string.

We think that this is a very important feature. Currently all the solutions known to us support 
adding an error to the chain only with internally created wrapper object that receives only a 
string as input.

```golang
newErr := fmt.Errorf("open file failed, %w", oldErr)
```

We want to be able to add sentinel and also custom error type instances to the chain.

And we want this to be easy for the user. We don't want the users to write wrapping logic into
their custom error types.

We achieved this by using internal wrapper type that has two pointers, `cause` that points to the
previous error in the chain, and `err` that points to the new error added to the chain.

The chain looks like this:
```
┌─────────┐       ┌─────────┐       ┌──────────┐
│         │ cause │         │ cause │          │
│ wrapper ├──────►│ wrapper ├──────►│ wrapper  ├──────► nil
│         │       │         │       │          │
└────┬────┘       └────┬────┘       └────┬─────┘
     │ err             │ err             │ err
┌────▼────┐       ┌────▼────┐       ┌────▼─────┐
│ Custom  │       │ Sentinel│       │ Custom   │
│ Error   │       │ Error   │       │ Error    │
└─────────┘       └─────────┘       └──────────┘
```

That way it is possible to have `Wrap` and `WrapError` functions, where the `Wrap` function
works as the standard one, while `WrapError` works by placing an user defined error in the chain.

This is useful for the user, who can use `Is` and `As` for searching those sentinel and user defined
error types.

```golang
// sentinel error
var ErrOpenFileFailed = errors.New("open file failed")

newErr := errors.WrapError(oldErr, ErrOpenFileFailed)

if errors.Is(newErr, ErrOpenFileFailed) {
}
```

### 5. It must support hiding internal and third party library errors from the chain search routines.

Errors returned by public API are part of the API.

If internal or third party library errors are part of the errors chain, inevitably some user
of the API will try to use `Is`, `As` on them, relying that they will always be there.

Furthermore if you change the internal implementation or update/replace the third party library,
even without changing your public API, it will break existing user code that relies on those errors.

The thing is that you in fact did change the API, because those errors were part of the API.

That's why we think that it must be possible to hide such errors.

Of course right now with the standard and other error libraries you can achieve this by simply 
returning a completely new error, thus starting a new chain of errors.

But this way you will loose the stack trace. We want to preserve the stack trace for logging,
but to prevent the hidden errors from being discoverable with `Is`, `As` and `Unwrap`.

For that reason we have `Hide` and `HideError` functions, that do almost the same as `Wrap` and
`WrapError`, but additionally set a flag in the internal wrapper object, so that it knows to return 
`nil` in the `Unwrap` method, thus preventing the traversal to go beyond a `hide barrier`.

In order to discover the errors for stack tracing, we added additional function `UnwrapHidden` 
that works as `Unwrap` but ignores the `hide barrier`.

**NOTE:** We saw the term `barrier` for the first time looking at 
[CockroachDB error package](https://github.com/cockroachdb/errors).


### 6. It must provide stack traces of all errors in the chain.

We decided to collect stack traces the same way package 
[xerrors](https://pkg.go.dev/golang.org/x/xerrors) does.

Meaning that we collect one frame per error, which can be used for extracting the function, file 
and line of the wrapping calls.

When printing with `fmt.Printf("%+v", err)` we use the same output format as xerrors package.

```
qux:
    github.com/sumup-oss/go-pkgs/errors_test.Example_printf
        /path/example_printf_test.go:15
  - baz:
    github.com/sumup-oss/go-pkgs/errors_test.Example_printf
        /path/example_printf_test.go:14
  - bar:
    github.com/sumup-oss/go-pkgs/errors_test.Example_printf
        /path/example_printf_test.go:13
  - foo:
    github.com/sumup-oss/go-pkgs/errors_test.Example_printf
        /path/example_printf_test.go:12
```

### 7. The stack traces should be available through a simple interface, not only as a formatted string.

It seems that every body makes the stack trace info private, and exposes it only with formatted string.

This is not very easy to work with when you need to format the data differently.
For example lets say that you want to put the stack trace into a structured log.

That's why the internal error wrapper type implements a `Locator` interface.

```golang
type Locator interface {
  Location() (function, file string, line int)
}
```

The user can traverse the chain using `UnwrapHidden`, check for `Locator` interface and collect
location data by calling the `Location` method.

```golang
err := someError

for err != nil {
  locator, ok := err.(interface {
    Location() (function, file string, line int)
  })

  if ok {
    fn, file, line := locator.Location()
    fmt.Printf("function=%s file=%s line=%s\n", fn, file, line)
  }

  err = UnwrapHidden(err)
}
```

## Quick Start

### Creating errors

You can create sentinel errors with `errors.New` as before.

```golang
var (
  ErrFoo = errors.New("foo error")
  ErrBar = errors.New("bar error")
)
```

You can create errors inside functions with `errors.New`.

```golang
func Foo() error {
  return errors.New("foo error %d", 100)
}
```

Normally when returning sentinel or custom error type instances you should use `Wrap`, `WrapError`,
`Hide` and `HideError`.

But sometimes the returned error already has enough context, and wrapping the error will just
add the same context twice.

```golang
func Foo() error {
  // creates a chain of two errors with the same description, which is redundant.
  return errors.Wrap(NewCustomError("foo failed"), "foo failed")
}
```

If you return the error without wrapping, it will not have trace information attached.

```golang
func Foo() error {
  // this will not add a stack trace information.
  return NewCustomError("foo failed")
}
```

In that case you can use `NewError`. It just adds stack trace information.

```golang
func Foo() error {
  // returns custom error wrapped with stack trace information.
  return errors.NewError(NewCustomError("foo failed"))
}
```

### Wrap and WrapError

You can wrap errors by using `Wrap` function.

```golang
newError := errors.Wrap(oldError, "cannot open file %s", filePath)
```

You can also wrap errors by using sentinel errors, or instances of your own error types by using
`WrapError`.

```golang
// sentinel error
var ErrFoo = errors.New("foo error")

newError := errors.WrapError(oldError, ErrFoo)

if errors.Is(newError, ErrFoo) {
  // ...
}
```

```golang
// custom error type
type PackageError struct {
  code int
  msg string
}

func NewPackageError(code int, msg string) *PackageError {
  return &PackageError{code: code, msg: msg}
}

func (e *PackageError) Error() string {
  return fmt.Sprintf("[%d] %s", e.code, e.msg)
}

func (e *PackageError) Code() int {
  return e.code
}

func (e *PackageError) Message() string {
  return e.msg
}

newError := errors.WrapError(oldError, NewPackageError(100, "foo error"))

var err *PackageError
if errors.As(newError, &err) {
  fmt.Printf("code=%s, msg=%s", err.Code(), err.Message())
}
```

### Hide and HideError

Check the `Wrap` and `WrapError` usage above.

`Hide` and `HideError` work almost the same as `Wrap` and `WrapError`, but `Is`, `As` will not 
discover any of the older errors on the chain. Also the `Unwrap` function will stop on the 
first error that was created with `Hide` and `HideError`.

The older errors are still discoverable by using `UnwrapHidden`.

**NOTE:** `UnwrapHidden` must be used only for generic processing of the errors.
For example to get the stack trace info for all errors in the chain.
It MUST NOT be used for searching of particular error or error type.

```golang
fooError := errors.New("foo")
barError := errors.Hide(fooError, "bar") // Unwrap will stop here, foo won't be discoverable
bazError := errors.Wrap(barError, "baz")
quxError := errors.Wrap(bazError, "qux")

err := quxError
for err != nil {
  fmt.Println(err.Error())

  err = errors.Unwrap(err)
}

// Output:
// qux
// baz
// bar
```

```golang
internalError := errors.New("internal")
apiError := errors.New("api error")

// Unwrap will stop here, internalError won't be discoverable
barError := errors.HideError(internalError, apiError)

bazError := errors.Wrap(barError, "baz")
quxError := errors.Wrap(bazError, "qux")

err := quxError
for err != nil {
  fmt.Println(err.Error())

  err = errors.Unwrap(err)
}

// Output:
// qux
// baz
// api error
```

### Is and As

As shown above `Is` and `As` are used for searching in the error chain. They are just dummy wrappers
of the standard ones, and exist only for ergonomics reasons, so that the users don't have to 
import the standard errors package.

```golang
fooError := NewAsWithHideCustomError("foo", 1)
barError := errors.Hide(fooError, "bar") // Since Unwrap will stop here, As cannot find foo.
bazError := errors.Wrap(barError, "baz")
quxError := errors.Wrap(bazError, "qux")

err := quxError

var customErr *AsWithHideCustomError
if errors.As(err, &customErr) {
  fmt.Printf("msg: %s, code: %d", customErr.Message(), customErr.Code())
} else {
  fmt.Println("no customErr found")
}

// Output:
// no customErr found
```

```golang
fooError := errors.New("foo")
barError := errors.Hide(fooError, "bar") // Since Unwrap will stop here, Is cannot find foo.
bazError := errors.Wrap(barError, "baz")
quxError := errors.Wrap(bazError, "qux")

err := quxError

if errors.Is(err, quxError) {
  fmt.Println("qux found")
}

if errors.Is(err, bazError) {
  fmt.Println("baz found")
}

if errors.Is(err, barError) {
  fmt.Println("bar found")
}

if !errors.Is(err, fooError) {
  fmt.Println("foo NOT found")
}

// Output:
// qux found
// baz found
// bar found
// foo NOT found
```

### Unwrap

Unwrap is used for traversing the errors chain.

It is a dummy wrapper of the standard `Unwrap` function and exists only for ergonomics reasons, 
so that the users don't have to import the standard errors package.

```golang
err := someError

for err != nil {
  fmt.Println(err.Error())

  err = Unwrap(err)
}
```

### UnwrapHidden

UnwrapHidden is used for traversing the errors chain. It works almost the same as `Unwrap`, but
it will traverse also the hidden errors.

**NOTE:** `UnwrapHidden` must be used only for generic processing of the errors.
For example to get the stack trace info for all errors in the chain.
It MUST NOT be used for searching of particular error or error type.

```golang
fooError := errors.New("foo")
barError := errors.Hide(fooError, "bar") // UnwrapHidden will NOT stop here, foo will be discoverable.
bazError := errors.Wrap(barError, "baz")
quxError := errors.Wrap(bazError, "qux")

err := quxError
for err != nil {
  fmt.Println(err.Error())

  err = errors.UnwrapHidden(err)
}

// Output:
// qux
// baz
// bar
// foo
```

### Creating custom error types

Creating a custom error type is nothing special. You do it the way you usually do.

All the wrap functions and `Is`, `As`, `Unwrap` will work with your custom types.

With a pure custom type if you want to collect stack trace you must return it with one of the
wrapper functions or using `NewError`.

```golang
func Foo() error {
  // ...
  return errors.Wrap(NewCustomError("custom error"), "something went wrong")
}
```

### Printing errors

Printing errors as string will print only the last error added to the chain.

```golang
fooError := errors.New("foo")
barError := errors.Hide(fooError, "bar")
bazError := errors.Hide(barError, "baz")
quxError := errors.Wrap(bazError, "qux")

fmt.Printf("%s\n", quxError)

// Output:
// qux
```

If you use `v` it will print all the errors in the chain separated with `: `.
Note that it will stop with the first hidden error.

```golang
fooError := errors.New("foo")
barError := errors.Hide(fooError, "bar")
bazError := errors.Hide(barError, "baz")
quxError := errors.Wrap(bazError, "qux")

fmt.Printf("%v\n", quxError)

// Output:
// qux: baz
```

If you use `+v` it will print all the errors including the hidden ones. It will also print
stack traces.

```golang
fooError := errors.New("foo")
barError := errors.Hide(fooError, "bar")
bazError := errors.Hide(barError, "baz")
quxError := errors.Wrap(bazError, "qux")

fmt.Printf("%+v\n", quxError)

// Output:
// qux:
//     github.com/sumup-oss/go-pkgs/errors_test.Example_printf
//         /path/example_printf_test.go:15
//   - baz:
//     github.com/sumup-oss/go-pkgs/errors_test.Example_printf
//         /path/example_printf_test.go:14
//   - bar:
//     github.com/sumup-oss/go-pkgs/errors_test.Example_printf
//         /path/example_printf_test.go:13
//   - foo:
//     github.com/sumup-oss/go-pkgs/errors_test.Example_printf
//         /path/example_printf_test.go:12
```

### Collecting stack traces

Collecting stack traces can be useful. For example if you want to put them in a structured log.

NOTE: If you just need a string representation `fmt.Printf("%+v", err)` will do the job.

```golang
err := someError

for err != nil {
  locator, ok := err.(interface {
    Location() (function, file string, line int)
  })

  if ok {
    // collect function, file and line
    function, file, line := locator.Location()
    fmt.Printf("function=%s file=%s line=%s\n", function, file, line)
  }

  err = UnwrapHidden(err)
}
```
