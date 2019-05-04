package loge

import (
    "fmt"
    "io"
    "log"
    "os"
    "sync"
    "time"
)

// Various log output modes
const (
    OutputConsole             uint32 = 1  // OutputConsole outputs to the stderr
    OutputFile                uint32 = 2  // OutputFile adds a file output
    OutputFileRotate          uint32 = 4  // OutputFileRotate adds an automatic file rotation based on current date
    OutputIncludeLine         uint32 = 8  // Include file and line into the output
    OutputConsoleInJSONFormat uint32 = 16 // Switch console output to JSON serialized format
    OutputConsoleOptionalData uint32 = 32 // Print optional data values to the console too
)

// Various selectable log levels
const (
    LogLevelInfo    uint32 = 1
    LogLevelDebug   uint32 = 2
    LogLevelTrace   uint32 = 4
    LogLevelWarning uint32 = 8
    LogLevelError   uint32 = 16
)

// TransportCreator is an interface to create new optional transports when the log is initialized
type TransportCreator func(TransactionList) []Transport

// Configuration defines the logger startup configuration
type configuration struct {
    Mode                     uint32                 // work mode
    Path                     string                 // output path for the file mode
    Filename                 string                 // log file name (ignored if rotation is enabled)
    TransactionSize          int                    // transaction size limit in bytes (default 10KB)
    TransactionTimeout       time.Duration          // transaction length limit (default 3 seconds)
    ConsoleOutput            io.Writer              // output writer for console (default os.Stderr)
    BacklogExpirationTimeout time.Duration          // transaction backlog expiration timeout (default is time.Hour)
    LogLevels                uint32                 // selectable log levels
    defaultData              map[string]interface{} // default Data added to each Element
    Transports               func(list TransactionList) []Transport
}



var std *logger

func init() {
    std = newLogger(
        configuration{
            Mode:          OutputConsole,
            ConsoleOutput: os.Stderr,
            defaultData:   make(map[string]interface{}),
        })
}

const (
    defaultTransactionSize   = 10 * 1024
    defaultTransactionLength = time.Second * 3
    defaultBacklogTimeout    = time.Minute * 15
)

type logger struct {
    configuration        configuration
    writeTimestampBuffer []byte
    buffer               *buffer

    customTimestampBuffer []byte
    customTimestampLock   sync.Mutex
}

// Init initializes the library and returns the shutdown handler to defer, must defer call the shutdown handler to ensure log messages are flushed.
func Init(decorators ...func(*configuration) *configuration) func() {
    c := &configuration{defaultData: make(map[string]interface{})}

    for _, decorator := range decorators {
        c = decorator(c)
    }

    std = newLogger(*c)
    return std.shutdown
}

// Path returns a function to set the log file path.
func Path(p string) func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Path = p
        return l
    }
}

// // Mode returns a function to set the Mode of config. (Deprecated in favour of calls like loge.EnableOutputConsole etc.
// func Mode(v uint32) func(*configuration) *configuration {
//     return func(l *configuration) *configuration {
//         l.Mode = v
//         return l
//     }
// }

// EnableOutputConsole returns a function to enable the output console.
func EnableOutputConsole() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Mode |= OutputConsole
        return l
    }
}

// EnableOutputFile returns a function to enable the output file.
func EnableOutputFile() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Mode |= OutputFile
        return l
    }
}

// EnableFileRotate returns a function to enable the output file rotation.
func EnableFileRotate() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Mode |= OutputFileRotate
        return l
    }
}

// EnableOutputIncludeLine returns a function to enable the Include file and line into the output.
func EnableOutputIncludeLine() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Mode |= OutputIncludeLine
        return l
    }
}

// EnableOutputConsoleInJSONFormat returns a function to enable the console output to JSON serialized format.
func EnableOutputConsoleInJSONFormat() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Mode |= OutputConsoleInJSONFormat
        return l
    }
}

// EnableOutputConsoleOptionalData returns a function to enable optional With() fields to the console output if turned on.  By default optional fields are only serialized into JSON format.
func EnableOutputConsoleOptionalData() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Mode |= OutputConsoleOptionalData
        return l
    }
}


// Filename returns a function to set the log file name (ignored if rotation is enabled).
func Filename(p string) func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Filename = p
        return l
    }
}

// TransactionSize returns a function to set the transaction size limit in bytes (default 10KB).
func TransactionSize(p int) func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.TransactionSize = p
        return l
    }
}

// TransactionTimeout returns a function to set the transaction length limit (default 3 seconds).
func TransactionTimeout(p time.Duration) func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.TransactionTimeout = p
        return l
    }
}

// ConsoleOutput returns a function to set the output writer for console (default os.Stderr).
func ConsoleOutput(p io.Writer) func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.ConsoleOutput = p
        return l
    }
}

// BacklogExpirationTimeout returns a function to set the transaction backlog expiration timeout (default is time.Hour).
func BacklogExpirationTimeout(p time.Duration) func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.BacklogExpirationTimeout = p
        return l
    }
}

// // LogLevels returns a function to set the selectable log levels. (Deprecated in favour of calls like loge.EnableDebug, loge.EnableTrace
// func LogLevels(p uint32) func(*configuration) *configuration {
//     return func(l *configuration) *configuration {
//         l.LogLevels = p
//         return l
//     }
// }


// EnableDebug returns a function to enable the logging of Debug level messages.
func EnableDebug() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.LogLevels |= LogLevelDebug
        return l
    }
}

// EnableInfo returns a function to enable the logging of Info level messages.
func EnableInfo() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.LogLevels |= LogLevelInfo
        return l
    }
}

// EnableTrace returns a function to enable the logging of Trace level messages.
func EnableTrace() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.LogLevels |= LogLevelTrace
        return l
    }
}

// EnableWarning returns a function to enable the logging of Warning level messages.
func EnableWarning() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.LogLevels |= LogLevelWarning
        return l
    }
}

// EnableError returns a function to enable the logging of Error level messages.
func EnableError() func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.LogLevels |= LogLevelError
        return l
    }
}


// Transports returns a function to set the Optional transports creator.
func Transports(s func(list TransactionList) []Transport) func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.Transports = s
        return l
    }
}


// WithDefault returns a function to sets default parameters that will be included with each entry. Such as ip, processName etc.
func WithDefault(key string, value interface{}) func(*configuration) *configuration {
    return func(l *configuration) *configuration {
        l.defaultData[key] = value
        return l
    }
}

func newLogger(c configuration) *logger {
    l := &logger{
        configuration: c,
    }

    flag := 0
    if (c.Mode & OutputIncludeLine) != 0 {
        flag |= log.Lshortfile
    }

    if (c.Mode & OutputFile) != 0 {
        validPath := false

        if fileInfo, err := os.Stat(c.Path); !os.IsNotExist(err) {
            if fileInfo.IsDir() {
                validPath = true
            }
        }

        if !validPath {
            l.configuration.Mode = l.configuration.Mode & (^OutputFile)
            os.Stderr.Write([]byte("Log path is invalid.  Log file output is disabled.\n"))
        }
    }

    if l.configuration.TransactionSize == 0 {
        l.configuration.TransactionSize = defaultTransactionSize
    }

    if l.configuration.TransactionTimeout == 0 {
        l.configuration.TransactionTimeout = defaultTransactionLength
    }

    if l.configuration.ConsoleOutput == nil {
        l.configuration.ConsoleOutput = os.Stderr
    }

    if l.configuration.BacklogExpirationTimeout == 0 {
        l.configuration.BacklogExpirationTimeout = defaultBacklogTimeout
    }

    if ((l.configuration.Mode & OutputFile) != 0) || (l.configuration.Transports != nil) {
        buffer := newBuffer(l)

        var outputs []Transport

        if (l.configuration.Mode & OutputFile) != 0 {
            outputs = make([]Transport, 1)
            outputs[0] = newFileTransport(buffer, c.Path, c.Filename, (c.Mode & OutputFileRotate) != 0, (c.Mode & OutputConsoleInJSONFormat) != 0)
        } else {
            outputs = make([]Transport, 0)
        }

        if l.configuration.Transports != nil {
            outputs = append(outputs, l.configuration.Transports(buffer)...)
        }

        if len(outputs) > 0 {
            l.buffer = buffer
            l.buffer.start(outputs)
        }
    }

    log.SetFlags(flag)
    log.SetOutput(l)

    return l
}

func (l *logger) shutdown() {
    if l.buffer != nil {
        l.buffer.shutdown()
    }
}

func (l *logger) Write(d []byte) (int, error) {
    if (l.buffer != nil) || ((l.configuration.Mode & OutputConsole) != 0) {
        t := time.Now()
        dumpTimeToBuffer(&l.writeTimestampBuffer, t) // don't have to lock this buf here because Write events are serialized
        l.write(
            NewBufferElement(t, l.writeTimestampBuffer, d, 0),
        )
    }

    return len(d), nil
}

func levelToString(level uint32) string {
    switch level {
    case LogLevelInfo:
        return "info"
    case LogLevelDebug:
        return "debug"
    case LogLevelTrace:
        return "trace"
    case LogLevelWarning:
        return "warning"
    case LogLevelError:
        return "error"
    default:
        return ""
    }
}

func (l *logger) write(be *BufferElement) {
    if (l.configuration.Mode & OutputConsole) != 0 {
        if (l.configuration.Mode & OutputConsoleInJSONFormat) != 0 {
            json, err := be.Marshal()
            if err == nil {
                l.configuration.ConsoleOutput.Write(json)
                l.configuration.ConsoleOutput.Write([]byte("\n"))
            }
        } else {
            l.configuration.ConsoleOutput.Write(be.Timestring[:])
            if ((l.configuration.Mode & OutputConsoleOptionalData) != 0) && (be.Data != nil) {
                l.configuration.ConsoleOutput.Write([]byte(be.serializeData()))
            }
            l.configuration.ConsoleOutput.Write([]byte(be.Message))
            l.configuration.ConsoleOutput.Write([]byte("\n"))
        }
    }

    if l.buffer != nil {
        l.buffer.write(
            be,
        )
    }
}

func (l *logger) writeLevel(level uint32, message string) {
    if (l.buffer != nil) || ((l.configuration.Mode & OutputConsole) != 0) {
        l.customTimestampLock.Lock()
        defer l.customTimestampLock.Unlock()
        t := time.Now()
        dumpTimeToBuffer(&l.customTimestampBuffer, t)
        be := NewBufferElement(t, l.customTimestampBuffer, []byte(message), level)
        l.write(be)
    }
}

// Printf creates creates a new log entry
func Printf(format string, v ...interface{}) {
    std.writeLevel(0, fmt.Sprintf(format, v...))
}

// Println creates creates a new log entry
func Println(v ...interface{}) {
    std.writeLevel(0, fmt.Sprintln(v...))
}

// Info creates creates a new "info" log entry
func Info(format string, v ...interface{}) {
    if (std.configuration.LogLevels & LogLevelInfo) != 0 {
        std.writeLevel(LogLevelInfo, fmt.Sprintf(format, v...))
    }
}

// Debug creates creates a new "debug" log entry
func Debug(format string, v ...interface{}) {
    if (std.configuration.LogLevels & LogLevelDebug) != 0 {
        std.writeLevel(LogLevelDebug, fmt.Sprintf(format, v...))
    }
}

// Trace creates creates a new "trace" log entry
func Trace(format string, v ...interface{}) {
    if (std.configuration.LogLevels & LogLevelTrace) != 0 {
        std.writeLevel(LogLevelTrace, fmt.Sprintf(format, v...))
    }
}

// Warn creates creates a new "warning" log entry
func Warn(format string, v ...interface{}) {
    if (std.configuration.LogLevels & LogLevelWarning) != 0 {
        std.writeLevel(LogLevelWarning, fmt.Sprintf(format, v...))
    }
}

// Error creates creates a new "error" log entry
func Error(format string, v ...interface{}) {
    if (std.configuration.LogLevels & LogLevelError) != 0 {
        std.writeLevel(LogLevelError, fmt.Sprintf(format, v...))
    }
}

// With creates a new log entry with optional parameters
func With(key string, value interface{}) *BufferElement {
    be := inPlaceBufferElement(std)
    be.Data[key] = value
    return be
}

func (l *logger) submit(be *BufferElement, message string, level uint32) {
    if (l.buffer != nil) || ((l.configuration.Mode & OutputConsole) != 0) {
        l.customTimestampLock.Lock()
        defer l.customTimestampLock.Unlock()
        t := time.Now()
        dumpTimeToBuffer(&l.customTimestampBuffer, t)
        be.fill(t, l.customTimestampBuffer, []byte(message), level)
        l.write(be)
    }
}
