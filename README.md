# LOGE - an extension for the standard log library in Go.

## Usage

```go
    defer loge.Init(
        loge.Configuration{
            Mode:          loge.OutputConsole | loge.OutputFile | loge.OutputFileRotate,
            Path:          ".",
            ConsoleOutput: os.Stdout,
        })()
```