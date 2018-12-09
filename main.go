package main

import (
	"fmt"
	app "pdu-server/app"
)

func main() {

	mainApp := app.NewApp()

	go grpcMain(mainApp)
	go httpMain(mainApp)

	fmt.Scanln()
}
