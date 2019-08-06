# LOGE - an extension for the standard log library in Go

This package is meant to be used as an addition to the standard log package extending its' functionality adding
an ability to write log files (with optional rotation) and serializing the output into JSON format.  File output
transport is transactional and abstracted to leave the possibility of adding optional transports (for example networking).

## Usage

`loge.Init()` function returns callable object that finalizes the log output. In order to use it simply add following
to the beginning of your `main()`. The defer will invoke a finalizer shutdown function returned by Init. This will ensure
all pending messages are flushed out and handled.

Usage:

```go
    defer loge.Init(
        loge.WithDefault("ip", "127.0.0.1"),
        loge.EnableOutputConsole(true),
        loge.ConsoleOutput(os.Stdout),
        loge.EnableOutputFile(true),
        loge.EnableFileRotate(true),
        loge.Path("."),
        loge.LogLevels(loge.LogLevelDebug|loge.LogLevelInfo),
    )()

```

To use the log simply use the default `log.Println()` and `log.Printf()` functions or alternative versions from `loge` package.
Additionally `loge` package adds five more output log levels with corresponding`Info()`, `Debug()`, `Trace()`, `Warn()`, and
`Error()` functions.

## Optional key-value parameters

If required it is possible to attach an optional key-value parameter (parameters) to any given log entry using a helper function
`loge.With(key, val).Printf()`.  Those calls can be stacked in order to add multiple parameters to the record as
`loge.With(key, val).With(key, val).Info()`.

```go
loge.With("uid", 32).With("nickname", "pap").Info("Info Message Associated with user")
```

## Configuration

Configuration is handled by passing an arbitrary config functions to the Init function.

Function|Input|Description
--------|-----|-----------
loge.EnableOutputConsole|bool|Enable the output console.
loge.EnableOutputFile|bool|Enable the output file.
loge.EnableFileRotate|bool|Enable the output file rotation.
loge.Path|string|Output path for file output (ignored if file output is disabled).
loge.Filename|string|Log file name (ignored if rotation is enabled).
loge.TransactionSize|int|Transaction size limit in bytes (default `10KB`).
loge.TransactionTimeout|time.Duration|Transaction flush timeout (default `3 seconds`).
loge.ConsoleOutput|io.Writer|Output writer for console output (default os.Stderr, ignored if console output is disabled).
loge.BacklogExpirationTimeout|time.Duration|Transaction backlog expiration timeout (default is `15 minutes`).
loge.Transports|TransportCreator|Optional transports creator.
loge.WithDefault|key string, value interface{}|WithDefault returns a function to sets default parameters that will be included with each entry. Such as ip, processName etc.
loge.LogLevels|uint32|Set the log level as a bitmask value.

## Optional log levels

Level|Description
-----|-----------
loge.EnableDebug|Enable the logging of LogLevelDebug level messages.
loge.EnableInfo|Enable the logging of LogLevelInfo level messages.
loge.EnableTrace|Enable the logging of LogLevelTrace level messages.
loge.EnableWarning|Enable the logging of LogLevelWarning level messages.
loge.EnableError|Enable the logging of LogLevelError level messages.

## Work mode options

Mode|Description
----|-----------
loge.EnableOutputConsole|Enable the output console.
loge.EnableOutputFile|Enable the output file.
loge.EnableFileRotate|Enable the output file rotation.
loge.EnableOutputIncludeLine|Include file and line into the output.
loge.EnableOutputConsoleInJSONFormat|Switch console output to JSON serialized format.
loge.EnableOutputConsoleOptionalData|Display optional With() fields to the console output if turned on.  By default optional fields are only serialized into JSON format.

## Optional transports

In order to create additional logging transports the library should be initialized with a `TransportCreator` - a function returning an array of external transports conforming to the `Transport` interface.

```go
type TransportCreator func(TransactionList) []Transport
```

`TransportCreator` receives `TransactionList` interface as a parameter. `TransactionList` provides an unified way for all transports to read the transaction log and expire the records that were delivered to the destination.

## Transport interface

```go
type Transport interface {
	NewTransaction(uint64)
	Stop()
}
```

`NewTransaction` is getting called every time the logger has a new transaction available to push down the transport. Function receives the transaction ID as a parameter and should never block the execution but rather signal the transport about the new transaction availability and return immediately.

`Stop` is getting called at the program exit. The transport should flush all the outputs and could potentially block the execution not returning until the flush is complete.

## TransactionList interface

```go
type TransactionList interface {
	Get(id uint64, autofree bool) (*Transaction, bool)
	Free(id uint64)
}
```

`Get` returns the transaction contents if transaction is still available (not purged). `autofree` parameter signals the transaction list handler that the transport is not going to request the transaction contents in the future anymore and the operation is complete.

`Free` purges the transaction when it's not needed by the transport anymore. Transactions are also automatically purged after configured period of time even if not accessed.

## Transaction interface

```go
type Transaction struct {
	ID         uint64
	Items      []*BufferElement
}

type BufferElement struct {
	Timestamp  time.Time
	Message    string
	Level      string
}
```

Transaction interface describes a single logger transaction containing multiple log records in `BufferElement` format. Each `BufferElement` carries a `Timestamp` in UTC format, a `Message` and an optional `Level` for `debug` or `info` log message types.

Each `BufferElement` also provides an additional helper functions: `Marshal`, allowing to marshal the element into JSON format.
