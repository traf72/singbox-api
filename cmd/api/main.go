package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/traf72/singbox-api/internal/api/handlers"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()

	router.Handle("PUT /dns-rules", handlers.AddDNSRuleHandler())

	// For Windows better to specify the full address (with IP instead of just ":8080") to avoid the Firewall issues
	// https://stackoverflow.com/questions/55201561/golang-run-on-windows-without-deal-with-the-firewall
	addr := "127.0.0.1:8080"
	server := http.Server{
		Addr:    addr,
		Handler: handlers.NewHandler(router).WithRequestLogging().Build(),
	}

	fmt.Println("Server is listening on", addr)
	server.ListenAndServe()
}
