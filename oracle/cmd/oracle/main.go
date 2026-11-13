// Command oracle runs the off-chain attestation API server.
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/api"
	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
	"github.com/brii26/eth-iota-attestation/oracle/internal/verify"
)

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	port := env("ORACLE_PORT", "8080")
	sqlitePath := env("SQLITE_PATH", "attestations.sqlite")
	verdaccioURL := env("VERDACCIO_URL", "http://127.0.0.1:4873")
	rustfsEndpoint := env("RUSTFS_ENDPOINT", "http://127.0.0.1:9000")
	rustfsAccessKey := env("RUSTFS_ACCESS_KEY", "rustfsadmin")
	rustfsSecretKey := env("RUSTFS_SECRET_KEY", "rustfsdevpassword123")

	maxAgeHours, _ := strconv.Atoi(env("MAX_GRYPE_DB_AGE_HOURS", "48"))
	maxAge := time.Duration(maxAgeHours) * time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Printf("Connecting to SQLite index: %s", sqlitePath)
	idx, err := store.NewSQLiteIndex(sqlitePath)
	if err != nil {
		log.Fatalf("failed to init SQLite index: %v", err)
	}
	defer idx.Close()

	log.Printf("Connecting to RustFS S3: %s", rustfsEndpoint)
	s3Store, err := store.NewS3Store(ctx, store.S3Config{
		Endpoint:  rustfsEndpoint,
		AccessKey: rustfsAccessKey,
		SecretKey: rustfsSecretKey,
	})
	if err != nil {
		log.Printf("warning: RustFS S3 store not available (%v); running with S3 disabled", err)
	}

	verdaccioClient := verify.NewVerdaccioClient(verdaccioURL)

	// ponytail: Use MockAdapter for standalone Issue #8. EthereumAdapter will be attached in Issue #9.
	chainAdapter := chain.NewMockAdapter()

	server := api.NewServer(api.Config{
		S3Store:         s3Store,
		Index:           idx,
		ChainAdapter:    chainAdapter,
		VerdaccioClient: verdaccioClient,
		MaxGrypeDBAge:   maxAge,
	})

	httpServer := &http.Server{
		Addr:         ":" + port,
		Handler:      server,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Oracle service listening on http://127.0.0.1:%s", port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	<-stopChan
	log.Println("Shutting down oracle service...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("Oracle service stopped cleanly")
}
