package main

import (
	networkservice "networkservice/app"
	"os"
)

func main() {

	networkserviceapp := networkservice.CreateServiceApp()
	networkserviceapp.StartApp()
	networkserviceapp.StartGRPC(os.Args)
}
