package log_test

import (
	"fmt"

	"ngrd.no/log"
)

func Example_basic() {
	logger, err := log.New(log.WithDisabledTimestamp())
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	logger.Infof("Hello")
	logger.Warnf("This is a warning")
	logger.Debugf("This is not logged by default")

	// Output:
	// 	ngrd.no/log_test	INFO	Hello
	// 	ngrd.no/log_test	WARN	This is a warning
}
