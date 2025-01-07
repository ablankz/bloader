package slave

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ablankz/bloader/internal/runner"
	"github.com/ablankz/bloader/internal/slave/slcontainer"
)

// SlaveStore represents the store for the slave node
type SlaveStore struct {
	store                         *slcontainer.Store
	connectionID                  string
	receiveChanelRequestContainer *slcontainer.ReceiveChanelRequestContainer
	mapper                        *slcontainer.RequestConnectionMapper
}

// Store stores the data
func (s *SlaveStore) Store(ctx context.Context, data []runner.ValidStoreValueData, cb runner.StoreCallback) error {
	strData := make([]slcontainer.StoreData, len(data))
	for i, d := range data {
		valBytes, err := json.Marshal(d.Value)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %v", err)
		}
		if cb != nil {
			if err := cb(ctx, d, valBytes); err != nil {
				return fmt.Errorf("failed to store data: %v", err)
			}
		}
		strData[i] = slcontainer.StoreData{
			BucketID:   d.BucketID,
			StoreKey:   d.Key,
			Data:       valBytes,
			Encryption: slcontainer.Encryption(d.Encrypt),
		}
	}

	term, err := s.receiveChanelRequestContainer.SendStore(
		ctx,
		s.connectionID,
		s.mapper,
		slcontainer.StoreDataRequest{
			StoreData: strData,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send store data request: %v", err)
	}

	select {
	case <-ctx.Done():
		return nil
	case <-term:
	}

	return nil
}

// StoreWithExtractor stores the data with extractor
func (s *SlaveStore) StoreWithExtractor(ctx context.Context, res interface{}, data []runner.ValidExecRequestStoreData, cb runner.StoreWithExtractorCallback) error {
	strData := make([]slcontainer.StoreData, len(data))
	for i, d := range data {
		result, err := d.Extractor.Extract(res)
		if err != nil {
			return fmt.Errorf("failed to extract store data: %v", err)
		}
		valBytes, err := json.Marshal(result)
		if err != nil {
			return fmt.Errorf("failed to marshal store data: %v", err)
		}
		if cb != nil {
			if err := cb(ctx, d, valBytes); err != nil {
				return fmt.Errorf("failed to store data: %v", err)
			}
		}
		strData[i] = slcontainer.StoreData{
			BucketID:   d.BucketID,
			StoreKey:   d.StoreKey,
			Data:       valBytes,
			Encryption: slcontainer.Encryption(d.Encrypt),
		}
	}

	term, err := s.receiveChanelRequestContainer.SendStore(
		ctx,
		s.connectionID,
		s.mapper,
		slcontainer.StoreDataRequest{
			StoreData: strData,
		},
	)

	if err != nil {
		return fmt.Errorf("failed to send store data request: %v", err)
	}

	select {
	case <-ctx.Done():
		return nil
	case <-term:
	}

	return nil
}

// Import loads the data
func (s *SlaveStore) Import(ctx context.Context, data []runner.ValidStoreImportData, cb runner.ImportCallback) error {
	shortageData := make([]slcontainer.StoreRespectiveRequest, 0, len(data))
	for _, d := range data {
		ok := s.store.ExistData(d.BucketID, d.StoreKey)
		if !ok {
			shortageData = append(shortageData, slcontainer.StoreRespectiveRequest{
				BucketID:   d.BucketID,
				StoreKey:   d.StoreKey,
				Encryption: slcontainer.Encryption(d.Encrypt),
			})
		}
	}

	term, err := s.receiveChanelRequestContainer.SendStoreResourceRequests(
		ctx,
		s.connectionID,
		s.mapper,
		slcontainer.StoreResourceRequest{
			Requests: shortageData,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to send store resource request: %v", err)
	}

	select {
	case <-ctx.Done():
		return nil
	case <-term:
	}

	for _, d := range data {
		val, ok := s.store.GetData(d.BucketID, d.StoreKey)
		if !ok {
			return fmt.Errorf("failed to get data: %s", d.StoreKey)
		}
		if cb != nil {
			if err := cb(ctx, d, val, nil); err != nil {
				return fmt.Errorf("failed to import data: %v", err)
			}
		}
	}

	return nil
}

var _ runner.Store = &SlaveStore{}
