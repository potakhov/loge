# LOGE - an extension for the standard log library in Go

This package is meant to be used as an addition to the standard log package extending its' functionality adding
an ability to write log files (with optional rotation) and serializing the output into JSON format.  File output
transport is transactional and abstracted to leave the possibility of adding optional transports (for example networking).

## Usage

`loge.Init()` function returns callable object that finalizes the log output. In order to use it simply add following
to the beginning of your `main()` function:

```go
    defer loge.Init(
        loge.Configuration{
            Mode:          loge.OutputConsole | loge.OutputFile | loge.OutputFileRotate,
            Path:          ".",
            ConsoleOutput: os.Stdout,
        })()
```

To use the log simply use the default `log.Println()` and `log.Printf()` functions.  Additionally `loge` package adds two more
output log levels `LogLevelInfo` and `LogLevelDebug` with corresponding `Infof()`/`Infoln()` and `Debugf()`/`Debugln()` functions.

## Configuration

Parameter|Type|Description
---------|----|-----------
Mode|uint32|Work mode options
Path|string|Output path for file output (ignored if file output is disabled)
Filename|string|Log file name (ignored if rotation is enabled)
TransactionSize|int|Transaction size limit in bytes (default `10KB`)
TransactionTimeout|time.Duration|Transaction flush timeout (default `3 seconds`)
ConsoleOutput|io.Writer|Output writer for console output (default os.Stderr, ignored if console output is disabled)
BacklogExpirationTimeout|time.Duration|Transaction backlog expiration timeout (default is `15 minutes`)
LogLevels|uint32|Enabled log levels (defaults to `0` meaning that only default `log.Print()` messages go to the output)
Transports|TransportCreator|Optional transports creator

## Work mode options

Mode|Description
----|-----------
OutputConsole|Enable the console output
OutputFile|Enable the file output
OutputFileRotate|Add an automatic file rotation based on current date
OutputIncludeLine|Include file and line into the output
OutputConsoleInJSONFormat|Switch console output to JSON serialized format

## Optional transports

In order to create additional logging transports the library should be initialized with a `TransportCreator` - a function returning an array of external transports conforming the `Transport` interface.

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
