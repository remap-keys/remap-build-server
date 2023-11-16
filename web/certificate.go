package web

import (
	"cloud.google.com/go/firestore"
	"context"
	"golang.org/x/crypto/acme/autocert"
	"log"
	"remap-keys.app/remap-build-server/database"
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
	certificate, err := database.FetchCertificate(ctx, fc.client, key)
	if err != nil {
		log.Printf("Failed to fetch the certificate from the Firestore: %v", err)
		return nil, err
	}
	if certificate == nil {
		return nil, autocert.ErrCacheMiss
	}
	return certificate.Data, nil
}

func (fc *FirestoreCertCache) Put(ctx context.Context, key string, data []byte) error {
	err := database.SaveCertificate(ctx, fc.client, key, data)
	if err != nil {
		log.Printf("Failed to save the certificate to the Firestore: %v", err)
		return err
	}
	return nil
}

func (fc *FirestoreCertCache) Delete(ctx context.Context, key string) error {
	err := database.DeleteCertificate(ctx, fc.client, key)
	if err != nil {
		log.Printf("Failed to delete the certificate from the Firestore: %v", err)
		return err
	}
	return nil
}
