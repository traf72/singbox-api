package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/joho/godotenv"
	"github.com/traf72/singbox-api/internal/api/handlers"
	"github.com/traf72/singbox-api/internal/api/middleware"
	"github.com/traf72/singbox-api/internal/utils"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	router := http.NewServeMux()

	router.Handle("GET /health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("The service is alive"))
	}))
	router.Handle("PUT /dns-rules", handlers.AddDNSRuleHandler())
	router.Handle("DELETE /dns-rules", handlers.RemoveDNSRuleHandler())
	router.Handle("PUT /ip-rules", handlers.AddIPRuleHandler())
	router.Handle("DELETE /ip-rules", handlers.RemoveIPRuleHandler())
	router.Handle("GET /config", handlers.GetConfigHandler())

	// For Windows better to specify the full address (with IP instead of just ":8080") to avoid the Firewall issues
	// https://stackoverflow.com/questions/55201561/golang-run-on-windows-without-deal-with-the-firewall
	addr := utils.GetEnv("LISTEN_ADDR", "127.0.0.1:8080")
	server := http.Server{
		Addr:    addr,
		Handler: middleware.NewHandler(router).WithRequestLogging().Build(),
	}

	fmt.Println("Server is listening on", addr)
	server.ListenAndServe()
}
