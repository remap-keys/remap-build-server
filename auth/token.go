package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"regexp"
	"strings"
)

func CheckAuthenticationToken(r *http.Request) error {
	authenticationToken, err := parseAuthenticationToken(r)
	if err != nil {
		return err
	}
	token, err := jwt.Parse(authenticationToken, func(token *jwt.Token) (interface{}, error) {
		alg := token.Header["alg"]
		if alg != "RS256" {
			return nil, fmt.Errorf("alg is not RS256")
		}
		kid := token.Header["kid"]
		if kid == "" {
			return nil, fmt.Errorf("kid is empty")
		}
		publicKey, err := FetchPublicKey(kid.(string))
		if err != nil {
			return nil, err
		}
		return ConvertPublickeyToPEM(publicKey)
	})
	if err != nil {
		return err
	}
	if !token.Valid {
		return fmt.Errorf("token is invalid")
	}
	// Checks the iss, email and aud claims.
	claims := token.Claims.(jwt.MapClaims)
	iss := claims["iss"]
	if iss.(string) != "https://accounts.google.com" {
		return fmt.Errorf("iss is invalid")
	}
	email := claims["email"]
	if email.(string) != "remap-build-server-task-auth@remap-b2d08.iam.gserviceaccount.com" {
		return fmt.Errorf("email is invalid")
	}
	aud := claims["aud"]
	if !strings.HasPrefix(aud.(string), "https://build.remap-keys.app/build?") {
		return fmt.Errorf("aud is invalid")
	}
	return nil
}

func parseAuthenticationToken(r *http.Request) (string, error) {
	authorizationHeader := r.Header.Get("Authorization")
	if authorizationHeader == "" {
		return "", fmt.Errorf("authorization header is empty")
	}
	re := regexp.MustCompile(`Bearer (.+)`)
	matches := re.FindStringSubmatch(authorizationHeader)
	if len(matches) != 2 {
		return "", errors.New("invalid authorization header format")
	}
	return matches[1], nil
}
