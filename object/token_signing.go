// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package object

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
)

// getJwtSigningKey parses the private key for a supported public-key JWT signing method.
func getJwtSigningKey(method jwt.SigningMethod, cert *Cert) (interface{}, error) {

	privateKey := cert.PrivateKey
	switch method.(type) {
	case *jwt.SigningMethodECDSA:
		return jwt.ParseECPrivateKeyFromPEM([]byte(privateKey))
	case *jwt.SigningMethodEd25519:
		return jwt.ParseEdPrivateKeyFromPEM([]byte(privateKey))
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		return jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKey))
	}
	return nil, fmt.Errorf("unexpected signing method: %s", method.Alg())
}

// getJwtVerifyKey parses the certificate of a cert for the signing method of a token being verified
func getJwtVerifyKey(token *jwt.Token, certificate string) (interface{}, error) {
	switch token.Method.(type) {
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		return jwt.ParseRSAPublicKeyFromPEM([]byte(certificate))
	case *jwt.SigningMethodECDSA:
		return jwt.ParseECPublicKeyFromPEM([]byte(certificate))
	case *jwt.SigningMethodEd25519:
		return parseEdPublicKeyFromPem(certificate)
	}
	return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
}

// parseEdPublicKeyFromPem takes the Ed25519 public key from a certificate, or from a PEM public key
func parseEdPublicKeyFromPem(certificate string) (interface{}, error) {
	block, _ := pem.Decode([]byte(certificate))
	if block == nil || block.Type != "CERTIFICATE" {
		return jwt.ParseEdPublicKeyFromPEM([]byte(certificate))
	}

	parsedCertificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, err
	}
	publicKey, ok := parsedCertificate.PublicKey.(ed25519.PublicKey)
	if !ok {
		return nil, fmt.Errorf("the certificate has no Ed25519 public key")
	}
	return publicKey, nil
}
