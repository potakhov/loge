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

    logeShutdown := loge.Init(
        loge.WithDefault("ip", "127.0.0.1"),
        loge.WithDefault("process", "calc.exe"),
        loge.Path("."),
        loge.EnableOutputConsole(),
        loge.EnableOutputFile(),
        loge.EnableFileRotate(),
        loge.ConsoleOutput(os.Stdout),
        loge.EnableDebug(),
        loge.EnableError(),
        loge.EnableInfo(),
        loge.EnableWarning(),
        loge.Transports(func(list loge.TransactionList) []loge.Transport {
            transport := loge.WrapTransport(list, c)
            return []loge.Transport{transport}
        }),
    )

    defer logeShutdown()

    loge.Printf("Hello World\n")
    loge.Warn("Hello World Warn Level\n")
    loge.Info("Hello World Info Level\n")
    loge.Debug("Hello World Debug Level\n")
    loge.Trace("Hello World Trace Level\n")
    loge.Error("Hello World Error Level\n")
    loge.With("uid", 32).With("nickname", "pap").Info("Info Message Associated with user")

    OsSignal := make(chan os.Signal, 1)
    LoopForever(OsSignal)
}

func LoopForever(OsSignal chan os.Signal) {
    loge.Info("Entering infinite loop\n")

    signal.Notify(OsSignal, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1)
    <-OsSignal

    loge.Info("Exiting infinite loop received OsSignal\n")
}
