package main

import (
    "fmt"
    "log"
    "os"

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

    log.Printf("Plain record via standard logger")
    loge.Printf("Plain record via extended logger")
    loge.With("uid", 42).Info("Info message with additional data")
    loge.With("uid", 33).With("token", "cafebabe").Debug("Debug message with additional data")
    loge.Trace("Trace message")

    loge.Trace("Trace Level, not enabled in Init()\n")
    loge.Debug("Debug Level\n")
    loge.Info("Info Level\n")
    loge.Warn("Warn Level \n")
    loge.Error("Error Level\n")

}
