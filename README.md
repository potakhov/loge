# LOGE - an extension for the standard log library in Go

This package is meant to be used as an addition to the standard log package extending its' functionality adding
an ability to write log files (with optional rotation) and serializing the output into JSON format.

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

## Work mode options

Mode|Description
----|-----------
OutputConsole|Enable the console output
OutputFile|Enable the file output
OutputFileRotate|Add an automatic file rotation based on current date
OutputIncludeLine|Include file and line into the output
OutputConsoleInJSONFormat|Switch console output to JSON serialized format