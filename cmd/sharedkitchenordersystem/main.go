package main

import (
	"flag"
	system "sharedkitchenordersystem/internal/app/sharedkitchenordersystem"

	"go.uber.org/zap"
)

func main() {

	// Create a zap logger with appropriate configuration.
	logger := createZapLogger()

	// Replace the global logger with created logger. After this
	// any package in the process can use zap.L() or zap.S()
	zap.ReplaceGlobals(logger)
	zap.S().Info("Starting main method")

	var noOfOrdersToRead int
	flag.IntVar(&noOfOrdersToRead, "noOfOrdersToRead", 2, "Orders receive rate")
	flag.Parse()

	// Start application
	system.Start(noOfOrdersToRead)
}

func createZapLogger() *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		zap.S().Fatal(err)
	}
	return logger
}
