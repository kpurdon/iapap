// Package iapap provides methods for verifying IAP requests.
package iapap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go/v4"
)

// Verifier provides methods for verifying IAP requests.
type Verifier struct {
	audience string
}

// NewVerifier initializes a new Verifier for the given IAP audience.
func NewVerifier(audience string) *Verifier {
	return &Verifier{audience: audience}
}

func (v *Verifier) getPublicKey(token *jwt.Token) (interface{}, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		"https://www.gstatic.com/iap/verify/public_key", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating public key request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making public keys request: %w", err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Printf("error closing request body: %+v", err)
		}
	}()

	var keys map[string]string
	if err = json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, fmt.Errorf("error decoding public keys: %w", err)
	}

	publicKeyClaim, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("missing token header \"kid\"")
	}
	publicKey, ok := keys[publicKeyClaim]
	if !ok {
		return nil, errors.New("token public key claim does not match any google public key claims")
	}

	// verify the signature of the public key
	key, err := jwt.ParseECPublicKeyFromPEM([]byte(publicKey))
	if err != nil {
		return nil, fmt.Errorf("error verifying public signature key: %w", err)
	}

	return key, nil
}

// Verify attempts to verify the IAP request.
func (v *Verifier) Verify(r *http.Request) error {
	token := r.Header.Get("x-goog-iap-jwt-assertion")
	if len(token) == 0 {
		return errors.New("missing header x-goog-iap-jwt-assertion")
	}

	var claims jwt.StandardClaims
	parsedToken, err := jwt.ParseWithClaims(
		token, &claims, v.getPublicKey,
		jwt.WithLeeway(5*time.Minute),
		jwt.WithAudience(v.audience),
		jwt.WithIssuer("https://cloud.google.com/iap"),
	)
	if err != nil {
		return fmt.Errorf("error parsing token: %w", err)
	}

	if parsedToken.Header["alg"] != jwt.SigningMethodES256.Alg() {
		return errors.New("token is signed with an invalid algorithm")
	}
	if !parsedToken.Valid {
		return errors.New("token is invalid")
	}

	return nil
}

// Apply provides an http.Handler which calls Verify for each request.
func (v *Verifier) Apply(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := v.Verify(r); err != nil {
			log.Printf("error verifying IAP request: %+v", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
