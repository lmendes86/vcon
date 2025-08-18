package vcon

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// JWSHeader represents the JSON Web Signature header.
type JWSHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

// GenerateKeyPair generates a new RSA key pair for signing vCons.
func GenerateKeyPair() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate private key: %w", err)
	}

	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

// Sign signs the vCon using JWS (JSON Web Signature) with RS256 algorithm.
func (v *VCon) Sign(privateKey *rsa.PrivateKey) error {
	return v.SignWithHeaders(privateKey, nil)
}

// SignWithHeaders signs the vCon using JWS with additional header parameters.
func (v *VCon) SignWithHeaders(privateKey *rsa.PrivateKey, headers map[string]interface{}) error {
	// Create JWS header
	header := JWSHeader{
		Algorithm: "RS256",
		Type:      "JWS",
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return fmt.Errorf("failed to marshal header: %w", err)
	}

	// Base64URL encode header
	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)

	// Marshal the VCon payload (without signatures and payload fields)
	payload := v.createPayloadForSigning()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Base64URL encode payload
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Create signing input
	signingInput := encodedHeader + "." + encodedPayload

	// Hash the signing input
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	// Sign the hash
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed)
	if err != nil {
		return fmt.Errorf("failed to sign: %w", err)
	}

	// Base64URL encode signature
	encodedSignature := base64.RawURLEncoding.EncodeToString(signature)

	// Add signature to vCon
	if v.Signatures == nil {
		v.Signatures = []Signature{}
	}

	v.Signatures = append(v.Signatures, Signature{
		Protected: encodedHeader,
		Header:    headers, // JWS unprotected header parameters
		Signature: encodedSignature,
	})

	v.Payload = &encodedPayload
	v.UpdateTimestamp()

	return nil
}

// Verify verifies the JWS signature of the vCon using the provided public key.
func (v *VCon) Verify(publicKey *rsa.PublicKey) (bool, error) {
	if len(v.Signatures) == 0 || v.Payload == nil {
		return false, errors.New("vCon is not signed")
	}

	// Use the first signature
	sig := v.Signatures[0]

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(sig.Protected)
	if err != nil {
		return false, fmt.Errorf("failed to decode header: %w", err)
	}

	var header JWSHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return false, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	// Check algorithm
	if header.Algorithm != "RS256" {
		return false, fmt.Errorf("unsupported algorithm: %s", header.Algorithm)
	}

	// Create signing input
	signingInput := sig.Protected + "." + *v.Payload

	// Hash the signing input
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	// Decode signature
	signatureBytes, err := base64.RawURLEncoding.DecodeString(sig.Signature)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed, signatureBytes)
	// Signature verification failure should return false, not an error
	// This distinguishes between validation errors (wrong key) and structural errors
	return err == nil, nil
}

// createPayloadForSigning creates a copy of the vCon without signature-related fields.
func (v *VCon) createPayloadForSigning() map[string]interface{} {
	// Marshal and unmarshal to create a deep copy
	data, _ := json.Marshal(v)
	var payload map[string]interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		// This should not happen with valid vCon data
		return nil
	}

	// Remove signature-related fields
	delete(payload, "signatures")
	delete(payload, "payload")

	return payload
}

// Key utility functions

// PrivateKeyToPEM converts an RSA private key to PEM format.
func PrivateKeyToPEM(privateKey *rsa.PrivateKey) ([]byte, error) {
	privateKeyBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %w", err)
	}

	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: privateKeyBytes,
	})

	return privateKeyPEM, nil
}

// PublicKeyToPEM converts an RSA public key to PEM format.
func PublicKeyToPEM(publicKey *rsa.PublicKey) ([]byte, error) {
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %w", err)
	}

	publicKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	})

	return publicKeyPEM, nil
}

// PrivateKeyFromPEM parses an RSA private key from PEM format.
func PrivateKeyFromPEM(pemData []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("key is not an RSA private key")
	}

	return rsaKey, nil
}

// PublicKeyFromPEM parses an RSA public key from PEM format.
func PublicKeyFromPEM(pemData []byte) (*rsa.PublicKey, error) {
	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %w", err)
	}

	rsaKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("key is not an RSA public key")
	}

	return rsaKey, nil
}

// IsSigned returns true if the vCon has been digitally signed.
func (v *VCon) IsSigned() bool {
	return len(v.Signatures) > 0 && v.Payload != nil
}

// GetSignedPayload returns the signed payload if the vCon is signed.
func (v *VCon) GetSignedPayload() (*string, error) {
	if !v.IsSigned() {
		return nil, errors.New("vCon is not signed")
	}
	return v.Payload, nil
}

// ValidateSignaturePresence validates that payload is present when signatures exist.
func (v *VCon) ValidateSignaturePresence() error {
	if len(v.Signatures) > 0 && v.Payload == nil {
		return errors.New("payload is required when signatures are present")
	}
	return nil
}

// SignedVCon represents a JWS container format for signed vCons per IETF specification.
// In this format, the payload contains the base64url-encoded unsigned vCon, and only
// signature-related fields are present at the top level.
type SignedVCon struct {
	Protected string                 `json:"protected"`        // Base64url-encoded JWS header
	Header    map[string]interface{} `json:"header,omitempty"` // Unprotected JWS header parameters
	Payload   string                 `json:"payload"`          // Base64url-encoded unsigned vCon JSON
	Signature string                 `json:"signature"`        // Base64url-encoded signature
}

// ToSignedFormat converts a signed VCon to the proper JWS container format.
// This separates the signature data from the vCon content as per IETF specification.
func (v *VCon) ToSignedFormat() (*SignedVCon, error) {
	if !v.IsSigned() {
		return nil, errors.New("vCon must be signed before converting to signed format")
	}

	if len(v.Signatures) == 0 {
		return nil, errors.New("no signatures found")
	}

	// Use the first signature for the signed format
	signature := v.Signatures[0]

	if v.Payload == nil {
		return nil, errors.New("payload is required for signed format")
	}

	return &SignedVCon{
		Protected: signature.Protected,
		Header:    signature.Header,
		Payload:   *v.Payload,
		Signature: signature.Signature,
	}, nil
}

// GetUnsignedVCon extracts and returns the unsigned vCon from a signed container format.
// This decodes the payload and reconstructs the original vCon without signatures.
func (s *SignedVCon) GetUnsignedVCon() (*VCon, error) {
	if s.Payload == "" {
		return nil, errors.New("signed vCon payload is empty")
	}

	// Decode the base64url-encoded payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(s.Payload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	// Unmarshal the unsigned vCon
	var vcon VCon
	if err := json.Unmarshal(payloadBytes, &vcon); err != nil {
		return nil, fmt.Errorf("failed to unmarshal unsigned vCon: %w", err)
	}

	// Ensure signature-related fields are not present in unsigned vCon
	vcon.Signatures = nil
	vcon.Payload = nil

	return &vcon, nil
}

// VerifySignedFormat verifies the signature of a JWS container format vCon.
func (s *SignedVCon) VerifySignedFormat(publicKey *rsa.PublicKey) error {
	if s.Protected == "" || s.Payload == "" || s.Signature == "" {
		return errors.New("invalid signed vCon format: missing required fields")
	}

	// Create signing input
	signingInput := s.Protected + "." + s.Payload

	// Decode the signature
	signatureBytes, err := base64.RawURLEncoding.DecodeString(s.Signature)
	if err != nil {
		return fmt.Errorf("failed to decode signature: %w", err)
	}

	// Hash the signing input
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	// Verify the signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed, signatureBytes)
	if err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
}

// VerifyJWS verifies a JWS token format string (header.payload.signature).
func VerifyJWS(jwsToken string, publicKey *rsa.PublicKey) (bool, error) {
	parts := strings.Split(jwsToken, ".")
	if len(parts) != 3 {
		return false, errors.New("invalid JWS format")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return false, fmt.Errorf("failed to decode header: %w", err)
	}

	var header JWSHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return false, fmt.Errorf("failed to unmarshal header: %w", err)
	}

	// Check algorithm
	if header.Algorithm != "RS256" {
		return false, fmt.Errorf("unsupported algorithm: %s", header.Algorithm)
	}

	// Create signing input
	signingInput := parts[0] + "." + parts[1]

	// Hash the signing input
	hasher := sha256.New()
	hasher.Write([]byte(signingInput))
	hashed := hasher.Sum(nil)

	// Decode signature
	signatureBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature
	err = rsa.VerifyPKCS1v15(publicKey, crypto.SHA256, hashed, signatureBytes)
	// Signature verification failure should return false, not an error
	// This distinguishes between validation errors (wrong key) and structural errors
	return err == nil, nil
}
