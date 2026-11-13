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
	"strings"
	"syscall"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/api"
	"github.com/brii26/eth-iota-attestation/oracle/internal/chain"
	"github.com/brii26/eth-iota-attestation/oracle/internal/listener"
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

	// Build network adapter router
	router := chain.NewRouter()

	// 1. Ethereum Private (Geth --dev)
	privRPC := env("ETHEREUM_PRIVATE_RPC_URL", "http://127.0.0.1:8545")
	privContract := env("ETHEREUM_PRIVATE_CONTRACT_ADDRESS", "")
	oracleKey := env("ETHEREUM_ORACLE_PRIVATE_KEY", "")

	if privContract != "" && oracleKey != "" {
		privAdapter, err := chain.NewEthereumAdapter(ctx, chain.EthereumConfig{
			Name:            "ethereum-private",
			Endpoints:       []string{privRPC},
			ContractAddress: privContract,
			PrivateKeyHex:   oracleKey,
		})
		if err != nil {
			log.Printf("warning: failed to init ethereum-private adapter: %v", err)
		} else {
			router.Register(privAdapter)
			log.Printf("Registered chain adapter: ethereum-private (contract: %s, rpc: %s)", privContract, privRPC)
		}
	} else {
		// ponytail: If Geth private chain keys are not supplied, register MockAdapter for ethereum-private
		mockPriv := chain.NewMockAdapter()
		router.Register(mockPriv)
		log.Printf("Notice: ETHEREUM_ORACLE_PRIVATE_KEY or contract address not set; using mock adapter for local testing")
	}

	// 2. Ethereum Sepolia
	sepoliaRPC := env("ETHEREUM_SEPOLIA_RPC_URL", "")
	sepoliaBackupRPC := env("ETHEREUM_SEPOLIA_BACKUP_RPC_URL", "")
	sepoliaContract := env("ETHEREUM_SEPOLIA_CONTRACT_ADDRESS", "")

	if sepoliaContract != "" && sepoliaRPC != "" && oracleKey != "" {
		endpoints := []string{sepoliaRPC}
		if sepoliaBackupRPC != "" {
			endpoints = append(endpoints, sepoliaBackupRPC)
		}
		sepoliaAdapter, err := chain.NewEthereumAdapter(ctx, chain.EthereumConfig{
			Name:            "ethereum-sepolia",
			Endpoints:       endpoints,
			ContractAddress: sepoliaContract,
			PrivateKeyHex:   oracleKey,
		})
		if err != nil {
			log.Printf("warning: failed to init ethereum-sepolia adapter: %v", err)
		} else {
			router.Register(sepoliaAdapter)
			log.Printf("Registered chain adapter: ethereum-sepolia (endpoints: %s)", strings.Join(endpoints, ", "))
		}
	}

	server := api.NewServer(api.Config{
		S3Store:         s3Store,
		Index:           idx,
		Router:          router,
		VerdaccioClient: verdaccioClient,
		MaxGrypeDBAge:   maxAge,
	})

	// Start background on-chain verification listeners for Ethereum networks
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()

	verifyEngine := verify.NewEngine(verify.EngineConfig{
		BlobStore:       s3Store,
		ArtifactFetcher: verdaccioClient,
		MaxGrypeDBAge:   maxAge,
	})

	for _, netName := range router.SupportedNetworks() {
		if strings.HasPrefix(netName, "ethereum-") {
			ad, err := router.Get(netName)
			if err == nil {
				w := listener.NewWorker(listener.Config{
					Adapter:      ad,
					Engine:       verifyEngine,
					Index:        idx,
					Network:      netName,
					PollInterval: 3 * time.Second,
				})
				go w.Start(workerCtx)
				log.Printf("Started on-chain verification listener for network: %s", netName)
			}
		}
	}

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
		log.Printf("Supported networks: %v", router.SupportedNetworks())
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
