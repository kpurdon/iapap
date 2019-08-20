package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func auth(next http.Handler, audience string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := validate(r, audience); err != nil {
			log.Printf("%+v", err)
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// TODO: cache the keys result
func getPublicKeys() (map[string]string, error) {
	resp, err := http.Get("https://www.gstatic.com/iap/verify/public_key")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var keys map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&keys); err != nil {
		return nil, err
	}

	return keys, nil
}

func tokenKeyFn(token *jwt.Token) (interface{}, error) {
	// the algorithm must me ES256
	if token.Header["alg"] != "ES256" {
		return nil, errors.New("invalid algorithm, must be ES256")
	}

	// only support ECDSA signing
	if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
		return nil, errors.New("invalid signing method, must be ECDSA")
	}

	// validate that the token contains a known public key from google
	publicKeys, err := getPublicKeys()
	if err != nil {
		return nil, err
	}
	publicKeyClaim, ok := token.Header["kid"].(string)
	if !ok {
		return nil, errors.New("missing token header \"kid\"")
	}
	publicKey, ok := publicKeys[publicKeyClaim]
	if !ok {
		return nil, errors.New("token public key claim does not match any google public key claims")
	}

	// validate that the token was signed by the google private key corresponding to the public key
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(publicKey))
	if err != nil {
		return nil, err
	}

	return key, nil
}

func validate(r *http.Request, audience string) error {
	tokenString := r.Header.Get("x-goog-iap-jwt-assertion")
	if len(tokenString) == 0 {
		return errors.New("missing header x-goog-iap-jwt-assertion")
	}

	var claims jwt.StandardClaims
	_, err := jwt.ParseWithClaims(tokenString, &claims, tokenKeyFn)
	if err != nil {
		return err
	}

	now := time.Now()

	// the token must not be expired
	if now.After(time.Unix(claims.ExpiresAt, 0)) {
		return errors.New("token must not be expired")
	}

	// the token must be issued before the current time
	if now.Before(time.Unix(claims.IssuedAt, 0)) {
		return errors.New("token must have been issued before the current time")
	}

	// the token must have an audience matching the provided value
	if claims.Audience != audience {
		return errors.New("token must have a matching audience")
	}

	// the token must be issued by google
	if claims.Issuer != "https://cloud.google.com/iap" {
		return errors.New("token must be issued by \"https://cloud.google.com/iap\"")
	}

	return nil
}
