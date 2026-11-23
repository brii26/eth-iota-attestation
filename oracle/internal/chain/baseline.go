package chain

import (
	"context"
	"errors"
	"net/http"
	"net/url"

	"github.com/brii26/eth-iota-attestation/oracle/internal/store"
)

type BaselineHashes struct {
	ArtifactHash string `json:"artifactHash"`
	SBOMHash     string `json:"sbomHash"`
	ScanHash     string `json:"scanHash"`
}

func (a *IOTAAdapter) WriteBaseline(ctx context.Context, hashes BaselineHashes) (*store.IOTAReceipt, error) {
	if _, allowed := a.trusted[a.signer]; !allowed {
		return nil, errors.New("IOTA writer is not trusted")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	tx, err := a.prepare(ctx, "/baseline/prepare", hashes)
	if err != nil {
		return nil, err
	}
	var receipt store.IOTAReceipt
	if err := a.call(ctx, http.MethodPost, "/submit", tx, &receipt); err != nil {
		return &store.IOTAReceipt{TxRef: tx.Digest}, errors.Join(ErrTransactionPending, err)
	}
	return &receipt, nil
}

func (a *IOTAAdapter) VerifyBaseline(ctx context.Context, objectID string, expected BaselineHashes) (bool, error) {
	if !iotaObjectID.MatchString(objectID) {
		return false, errors.New("invalid IOTA baseline object ID")
	}
	var record struct {
		BaselineHashes
		Signer string `json:"signer"`
	}
	err := a.call(ctx, http.MethodGet, "/baseline/record/"+url.PathEscape(objectID), nil, &record)
	var serviceError *iotaServiceError
	if errors.As(err, &serviceError) && serviceError.Status == http.StatusNotFound {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	trusted, err := a.IsSignerAuthorized(ctx, record.Signer)
	return trusted && record.BaselineHashes == expected, err
}
