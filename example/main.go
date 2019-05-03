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
	fmt.Println(">>> New transaction")

	for _, be := range tr.Items {
		record, err := be.Marshal()
		if err == nil {
			fmt.Println(string(record))
		}
	}

	fmt.Println("<<< Transaction ends")
}
func (t *customTransport) FlushTransactions() {
	fmt.Println(">>> Flush transactions <<<")
}

func main() {
	defer loge.Init(
		loge.Configuration{
			Mode:          loge.OutputConsole | loge.OutputConsoleOptionalData,
			ConsoleOutput: os.Stdout,
			LogLevels:     loge.LogLevelDebug | loge.LogLevelInfo,
			Transports: func(list loge.TransactionList) []loge.Transport {
				return []loge.Transport{loge.WrapTransport(list, &customTransport{})}
			},
		})()

	log.Printf("Plain record via standard logger")
	loge.Printf("Plain record via extended logger")
	loge.With("uid", 42).Info("Info message with additional data")
	loge.With("uid", 33).With("token", "cafebabe").Debug("Debug message with additional data")
	loge.Trace("Trace message, not enabled in Init()")
}
