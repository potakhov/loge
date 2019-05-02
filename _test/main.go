package main

import (
    "fmt"
    "os"
    "os/signal"
    "syscall"

    "github.com/potakhov/loge"
)

type customTransport struct {
}

func (t *customTransport) WriteOutTransaction(tr *loge.Transaction) {
    fmt.Println("WriteOutTransaction")

    for _, be := range tr.Items {
        fmt.Printf("WriteOutTransaction:: [%v] %v\n", string(be.Timestring[:]), be.Message, )
    }
}
func (t *customTransport) FlushTransactions() {
    fmt.Println("FlushTransactions")
}

func main() {
    c := &customTransport{}

    loge.Init(
        loge.Configuration{
            Mode:          loge.OutputConsole | loge.OutputFile | loge.OutputFileRotate,
            Path:          ".",
            ConsoleOutput: os.Stdout,
            LogLevels:     loge.LogLevelDebug | loge.LogLevelInfo,
            Transports: func(list loge.TransactionList) []loge.Transport {
                transport := loge.WrapTransport(list, c)
                outputs := make([]loge.Transport, 0)
                outputs = append(outputs, transport)
                return outputs
            },
        })

    loge.Printf("Hello World\n")
    loge.Printf("Hello World\n")
    loge.With("uid", 32).With("nickname", "pap").Info("test")

    OsSignal := make(chan os.Signal, 1)
    LoopForever(OsSignal)
}

func LoopForever(OsSignal chan os.Signal) {
    loge.Info("Entering infinite loop\n")

    signal.Notify(OsSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
    <-OsSignal

    loge.Info("Exiting infinite loop received OsSignal\n")
}
