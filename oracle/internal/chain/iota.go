package chain

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
)

var iotaObjectID = regexp.MustCompile(`^0x[0-9a-f]{64}$`)

var ErrRecordUnavailable = errors.New("on-chain record is unavailable")
var ErrTransactionPending = errors.New("transaction outcome is pending")
var ErrUnsupportedOperation = errors.New("operation is unsupported on this network")

type IOTAAdapter struct {
	name    string
	service string
	token   string
	client  *http.Client
	signer  string
	trusted map[string]struct{}
	index   *store.SQLiteIndex
	mu      sync.Mutex
}

type iotaPayload struct {
	SchemaVersion int    `json:"schemaVersion"`
	App           string `json:"app"`
	Version       string `json:"version"`
	ScanID        string `json:"scanId"`
	ArtifactHash  string `json:"artifactHash"`
	SBOMHash      string `json:"sbomHash"`
	ScanHash      string `json:"scanHash"`
}

type iotaRecord struct {
	iotaPayload
	store.IOTAReceipt
	Timestamp time.Time `json:"timestamp"`
}

type iotaServiceError struct {
	Status int
}

func (e *iotaServiceError) Error() string {
	return fmt.Sprintf("IOTA service returned HTTP %d", e.Status)
}

func NewIOTAAdapter(ctx context.Context, endpoint, token string, trustedSigners []string, index *store.SQLiteIndex) (*IOTAAdapter, error) {
	endpointURL, err := url.Parse(endpoint)
	if err != nil || endpointURL.Scheme != "http" || endpointURL.Hostname() != "127.0.0.1" ||
		endpointURL.User != nil || endpointURL.RawQuery != "" || endpointURL.Path != "" || endpointURL.Fragment != "" {
		return nil, errors.New("IOTA service URL must be an http://127.0.0.1 endpoint")
	}
	if len(token) < 32 || index == nil {
		return nil, errors.New("IOTA requires an index and a service token of at least 32 bytes")
	}
	a := &IOTAAdapter{
		service: endpoint,
		token:   token,
		index:   index,
		client:  &http.Client{Timeout: 20 * time.Second},
		trusted: make(map[string]struct{}),
	}
	var service struct {
		Name   string `json:"name"`
		Signer string `json:"signer"`
	}
	if err := a.call(ctx, http.MethodGet, "/network", nil, &service); err != nil {
		return nil, err
	}
	if (service.Name != "iota-testnet" && service.Name != "iota-localnet") || !iotaObjectID.MatchString(service.Signer) {
		return nil, errors.New("invalid IOTA service network or signer")
	}
	a.name, a.signer = service.Name, service.Signer
	for _, signer := range trustedSigners {
		signer = strings.ToLower(strings.TrimSpace(signer))
		if !iotaObjectID.MatchString(signer) {
			return nil, errors.New("IOTA trusted signers must be canonical addresses")
		}
		a.trusted[signer] = struct{}{}
	}
	if len(a.trusted) == 0 {
		return nil, errors.New("IOTA trusted signers are required")
	}
	return a, nil
}

func (a *IOTAAdapter) Name() string { return a.name }

func (a *IOTAAdapter) call(ctx context.Context, method, path string, input, output any) error {
	var data []byte
	var err error
	if input != nil {
		data, err = json.Marshal(input)
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, a.service+path, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.token)
	res, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return &iotaServiceError{Status: res.StatusCode}
	}
	return json.NewDecoder(res.Body).Decode(output)
}

func (a *IOTAAdapter) WriteRecord(ctx context.Context, record Record) (*Receipt, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, authorized := a.trusted[a.signer]; !authorized {
		return nil, errors.New("IOTA writer is not on the trusted signer list")
	}
	if record.Network != a.name {
		return nil, errors.New("IOTA record has a different network")
	}
	tx, err := a.index.IOTATransaction(ctx, a.name, record.ScanID)
	if err != nil {
		return nil, err
	}
	if tx == nil {
		payload := iotaPayload{
			SchemaVersion: 1, App: record.App, Version: record.Version, ScanID: record.ScanID,
			ArtifactHash: "0x" + hex.EncodeToString(record.ArtifactHash[:]),
			SBOMHash:     "0x" + hex.EncodeToString(record.SBOMHash[:]),
			ScanHash:     "0x" + hex.EncodeToString(record.ScanHash[:]),
		}
		var prepared store.IOTATransaction
		if err := a.call(ctx, http.MethodPost, "/prepare", payload, &prepared); err != nil {
			return nil, fmt.Errorf("prepare IOTA write: %w", err)
		}
		if prepared.Digest == "" || prepared.Bytes == "" || prepared.Signature == "" {
			return nil, errors.New("IOTA service returned an incomplete transaction")
		}
		if err := a.index.SaveIOTATransaction(ctx, a.name, record.ScanID, prepared); err != nil {
			return nil, err
		}
		tx, err = a.index.IOTATransaction(ctx, a.name, record.ScanID)
		if err != nil {
			return nil, err
		}
	}
	if err := a.index.UpdateStatus(ctx, a.name, record.ScanID, "pending", tx.Digest, "", nil); err != nil {
		return nil, errors.Join(ErrTransactionPending, err)
	}
	return a.submit(ctx, record.ScanID, *tx)
}

func (a *IOTAAdapter) submit(ctx context.Context, scanID string, tx store.IOTATransaction) (*Receipt, error) {
	if _, authorized := a.trusted[a.signer]; !authorized {
		return nil, ErrTransactionPending
	}
	var result store.IOTAReceipt
	if err := a.call(ctx, http.MethodPost, "/submit", tx, &result); err != nil {
		var serviceErr *iotaServiceError
		if errors.As(err, &serviceErr) && serviceErr.Status == http.StatusUnprocessableEntity {
			return nil, err
		}
		return nil, errors.Join(ErrTransactionPending, err)
	}
	if result.TxRef != tx.Digest || !iotaObjectID.MatchString(result.RecordID) || result.Signer != a.signer || result.IncludedAt == nil {
		return nil, fmt.Errorf("%w: incomplete IOTA receipt", ErrTransactionPending)
	}
	if err := a.index.UpdateIOTAReceipt(ctx, a.name, scanID, result); err != nil {
		return nil, errors.Join(ErrTransactionPending, err)
	}
	if err := a.index.UpdateStatus(ctx, a.name, scanID, "included", result.TxRef, result.RecordID, result.IncludedAt); err != nil {
		return nil, errors.Join(ErrTransactionPending, err)
	}
	return &Receipt{TxRef: result.TxRef, RecordID: result.RecordID, TxSize: result.TxSize, IncludedAt: *result.IncludedAt}, nil
}

func (a *IOTAAdapter) read(ctx context.Context, id string) (*iotaRecord, error) {
	var record iotaRecord
	err := a.call(ctx, http.MethodGet, "/record/"+url.PathEscape(id), nil, &record)
	var serviceErr *iotaServiceError
	if errors.As(err, &serviceErr) && serviceErr.Status == http.StatusNotFound {
		return nil, ErrRecordUnavailable
	}
	return &record, err
}

func (a *IOTAAdapter) ReadRecord(ctx context.Context, query Query) (*ChainRecord, error) {
	id := query.ExternalRecordID
	if id == "" {
		var err error
		id, err = a.index.IOTARecordID(ctx, a.name, query.App, query.Version)
		if err != nil {
			return nil, err
		}
		if id == "" {
			return nil, ErrRecordNotFound
		}
	}
	if !iotaObjectID.MatchString(id) {
		return nil, ErrRecordNotFound
	}
	record, err := a.read(ctx, id)
	if err != nil {
		return nil, err
	}
	result := &ChainRecord{
		RecordID: id, App: record.App, Version: record.Version, ScanID: record.ScanID,
		Signer: record.Signer, Timestamp: record.Timestamp,
	}
	for _, hash := range []struct {
		source string
		target *[32]byte
	}{{record.ArtifactHash, &result.ArtifactHash}, {record.SBOMHash, &result.SBOMHash}, {record.ScanHash, &result.ScanHash}} {
		if !iotaObjectID.MatchString(hash.source) {
			return nil, errors.New("invalid IOTA record hash")
		}
		_, _ = hex.Decode(hash.target[:], []byte(hash.source[2:]))
	}
	return result, nil
}

func (a *IOTAAdapter) IsSignerAuthorized(_ context.Context, signer string) (bool, error) {
	_, authorized := a.trusted[strings.ToLower(signer)]
	return authorized, nil
}

func (a *IOTAAdapter) FulfillVerification(context.Context, *big.Int, uint8, bool) (*Receipt, error) {
	return nil, ErrUnsupportedOperation
}

func (a *IOTAAdapter) FilterVerificationRequests(context.Context, *uint64, *uint64) ([]VerificationEvent, error) {
	return nil, ErrUnsupportedOperation
}

func (a *IOTAAdapter) CurrentBlockNumber(context.Context) (uint64, error) {
	return 0, ErrUnsupportedOperation
}

func (a *IOTAAdapter) Reconcile(ctx context.Context) error {
	scanIDs, err := a.index.PendingIOTATransactions(ctx, a.name)
	if err != nil {
		return err
	}
	for _, scanID := range scanIDs {
		a.mu.Lock()
		tx, err := a.index.IOTATransaction(ctx, a.name, scanID)
		if err == nil && tx != nil {
			_, err = a.submit(ctx, scanID, *tx)
			if err != nil && !errors.Is(err, ErrTransactionPending) {
				err = errors.Join(err, a.index.UpdateStatus(ctx, a.name, scanID, "failed", tx.Digest, "", nil))
			}
		}
		a.mu.Unlock()
		if err != nil {
			log.Printf("IOTA transaction %q: %v", scanID, err)
		}
	}
	for owner := range a.trusted {
		if err := a.reconcileOwner(ctx, owner); err != nil {
			return err
		}
	}
	return nil
}

func (a *IOTAAdapter) reconcileOwner(ctx context.Context, owner string) error {
	cursor := ""
	for {
		query := url.Values{"owner": {owner}}
		if cursor != "" {
			query.Set("cursor", cursor)
		}
		var page struct {
			IDs    []string `json:"ids"`
			Cursor string   `json:"cursor"`
		}
		if err := a.call(ctx, http.MethodGet, "/records?"+query.Encode(), nil, &page); err != nil {
			return err
		}
		for _, id := range page.IDs {
			record, err := a.read(ctx, id)
			if errors.Is(err, ErrRecordUnavailable) {
				continue
			}
			if err != nil {
				return err
			}
			if record.Signer != owner || record.IncludedAt == nil {
				continue
			}
			err = a.index.SaveIOTAReceipt(ctx, store.AttestationRecord{
				Network: a.name, App: record.App, Version: record.Version, ScanID: record.ScanID,
				ArtifactHash: strings.TrimPrefix(record.ArtifactHash, "0x"),
				SBOMHash:     strings.TrimPrefix(record.SBOMHash, "0x"), ScanHash: strings.TrimPrefix(record.ScanHash, "0x"),
				SubmittedAt: record.Timestamp,
			}, record.IOTAReceipt)
			if err != nil {
				log.Printf("IOTA index record %s: %v", id, err)
			}
		}
		if page.Cursor == "" || page.Cursor == cursor {
			return nil
		}
		cursor = page.Cursor
	}
}
