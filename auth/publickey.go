package auth

import (
	"bytes"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"io"
	"log"
	"math/big"
	"net/http"
)

type openIDConfiguration struct {
	JWKsURI string `json:"jwks_uri"`
}

type jwk struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
	Alg string `json:"alg"`
}

type jwks struct {
	Keys []jwk `json:"keys"`
}

const openIDConfigurationURL = "https://accounts.google.com/.well-known/openid-configuration"

func ConvertPublickeyToPEM(publicKey *rsa.PublicKey) (*rsa.PublicKey, error) {
	derRsaPublicKey, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = pem.Encode(&buf, &pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: derRsaPublicKey,
	})
	if err != nil {
		return nil, err
	}
	return jwt.ParseRSAPublicKeyFromPEM(buf.Bytes())
}

func FetchPublicKey(kid string) (*rsa.PublicKey, error) {
	jwksURL, err := fetchJWksURL()
	if err != nil {
		return nil, err
	}
	keys, err := fetchJWKs(jwksURL)
	if err != nil {
		return nil, err
	}
	for _, key := range keys.Keys {
		if key.Kid == kid {
			pubkey := rsa.PublicKey{}
			number, _ := base64.RawURLEncoding.DecodeString(key.N)
			pubkey.N = new(big.Int).SetBytes(number)
			pubkey.E = 65537
			return &pubkey, nil
		}
	}
	return nil, fmt.Errorf("failed to find the public key")
}

func fetchJWKs(jwksURL string) (*jwks, error) {
	response, err := http.Get(jwksURL)
	if err != nil {
		log.Println("Failed to fetch jwks_uri from the hosting server.")
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Println("Failed to fetch jwks_uri from the hosting server.")
		return nil, fmt.Errorf("failed to fetch jwks_uri from the hosting server")
	}
	body, _ := io.ReadAll(response.Body)
	var keys jwks
	if err := json.Unmarshal(body, &keys); err != nil {
		log.Println("Failed to unmarshal jwks_uri.")
		return nil, err
	}
	return &keys, nil
}

func fetchJWksURL() (string, error) {
	// Fetch the openid-configuration from the hosting server.
	response, err := http.Get(openIDConfigurationURL)
	if err != nil {
		log.Println("Failed to fetch openid-configuration from the hosting server.")
		return "", err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		log.Println("Failed to fetch openid-configuration from the hosting server.")
		return "", fmt.Errorf("failed to fetch openid-configuration from the hosting server")
	}
	body, _ := io.ReadAll(response.Body)
	var config openIDConfiguration
	if err := json.Unmarshal(body, &config); err != nil {
		log.Println("Failed to unmarshal openid-configuration.")
		return "", err
	}
	return config.JWKsURI, nil
}
