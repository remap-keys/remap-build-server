package web

import (
	"cloud.google.com/go/firestore"
	"context"
	"fmt"
)

type FirestoreCertCache struct {
	client *firestore.Client
}

func NewFirestoreCertCache(client *firestore.Client) *FirestoreCertCache {
	return &FirestoreCertCache{
		client: client,
	}
}
func (fc *FirestoreCertCache) Get(ctx context.Context, key string) ([]byte, error) {
	doc, err := fc.client.Collection("certifications").Doc(key).Get(ctx)
	if err != nil {
		return nil, err
	}

	data, ok := doc.Data()["data"].([]byte)
	if !ok {
		return nil, fmt.Errorf("data not found or is not a byte slice")
	}

	return data, nil
}

func (fc *FirestoreCertCache) Put(ctx context.Context, key string, data []byte) error {
	_, err := fc.client.Collection("certifications").Doc(key).Set(ctx, map[string]interface{}{
		"data": data,
	})
	return err
}

func (fc *FirestoreCertCache) Delete(ctx context.Context, key string) error {
	_, err := fc.client.Collection("certifications").Doc(key).Delete(ctx)
	return err
}
