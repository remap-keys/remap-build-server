package web

import (
	"cloud.google.com/go/firestore"
	"context"
	"golang.org/x/crypto/acme/autocert"
)

type LetsEncryptCert struct {
	Data []byte `firestore:"data"`
}

type FirestoreCertCache struct {
	client *firestore.Client
}

func NewFirestoreCertCache(client *firestore.Client) *FirestoreCertCache {
	return &FirestoreCertCache{
		client: client,
	}
}
func (c *FirestoreCertCache) Get(ctx context.Context, key string) ([]byte, error) {
	ref, err := c.client.Collection("certifications").Doc(key).Get(ctx)
	if err != nil {
		return nil, err
	}
	if !ref.Exists() {
		return nil, autocert.ErrCacheMiss
	}
	var cert LetsEncryptCert
	ref.DataTo(&cert)
	return cert.Data, nil
}

func (c *FirestoreCertCache) Put(ctx context.Context, key string, data []byte) error {
	_, err := c.client.Collection("certifications").Doc(key).Set(ctx, LetsEncryptCert{Data: data})
	return err
}

func (c *FirestoreCertCache) Delete(ctx context.Context, key string) error {
	_, err := c.client.Collection("certifications").Doc(key).Delete(ctx)
	return err
}
