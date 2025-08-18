package vcon

import (
	"strings"
	"testing"
	"time"
)

func TestGenerateKeyPair(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	if privateKey == nil {
		t.Error("Private key is nil")
		return
	}
	if publicKey == nil {
		t.Error("Public key is nil")
		return
	}

	// Verify key sizes
	if privateKey.Size() != 256 { // 2048 bits = 256 bytes
		t.Errorf("Expected private key size 256, got %d", privateKey.Size())
	}
}

func TestSignAndVerify(t *testing.T) {
	// Generate test keys
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create test vCon
	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})
	vcon.AddDialog(Dialog{
		Type:    "text",
		Start:   vcon.CreatedAt,
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Test message",
	})

	// Test signing
	err = vcon.Sign(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign vCon: %v", err)
	}

	// Verify signature was added
	if !vcon.IsSigned() {
		t.Error("vCon should be marked as signed")
	}

	if len(vcon.Signatures) == 0 {
		t.Error("No signatures found")
	}

	if vcon.Payload == nil {
		t.Error("Payload should not be nil after signing")
	}

	// Test verification with correct key
	valid, err := vcon.Verify(publicKey)
	if err != nil {
		t.Fatalf("Failed to verify vCon: %v", err)
	}
	if !valid {
		t.Error("Signature verification failed with correct key")
	}

	// Generate different key pair for negative test
	wrongPrivateKey, wrongPublicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate wrong key pair: %v", err)
	}

	// Test verification with wrong key
	valid, err = vcon.Verify(wrongPublicKey)
	if err != nil {
		t.Fatalf("Unexpected error during verification: %v", err)
	}
	if valid {
		t.Error("Signature verification should fail with wrong key")
	}

	// Test verification of unsigned vCon
	unsignedVCon := NewWithDefaults()
	_, err = unsignedVCon.Verify(publicKey)
	if err == nil {
		t.Error("Expected error when verifying unsigned vCon")
	}

	// Test signing already signed vCon (should add another signature)
	err = vcon.Sign(wrongPrivateKey)
	if err != nil {
		t.Fatalf("Failed to add second signature: %v", err)
	}

	if len(vcon.Signatures) != 2 {
		t.Errorf("Expected 2 signatures, got %d", len(vcon.Signatures))
	}
}

func TestKeyConversion(t *testing.T) {
	// Generate test keys
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Test private key to PEM conversion
	privatePEM, err := PrivateKeyToPEM(privateKey)
	if err != nil {
		t.Fatalf("Failed to convert private key to PEM: %v", err)
	}

	// Test private key from PEM conversion
	convertedPrivateKey, err := PrivateKeyFromPEM(privatePEM)
	if err != nil {
		t.Fatalf("Failed to convert private key from PEM: %v", err)
	}

	// Verify keys are equivalent by signing with both
	vcon1 := NewWithDefaults()
	vcon1.AddParty(Party{Name: StringPtr("Test")})

	vcon2 := NewWithDefaults()
	vcon2.AddParty(Party{Name: StringPtr("Test")})

	err = vcon1.Sign(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign with original key: %v", err)
	}

	err = vcon2.Sign(convertedPrivateKey)
	if err != nil {
		t.Fatalf("Failed to sign with converted key: %v", err)
	}

	// Test public key to PEM conversion
	publicPEM, err := PublicKeyToPEM(publicKey)
	if err != nil {
		t.Fatalf("Failed to convert public key to PEM: %v", err)
	}

	// Test public key from PEM conversion
	convertedPublicKey, err := PublicKeyFromPEM(publicPEM)
	if err != nil {
		t.Fatalf("Failed to convert public key from PEM: %v", err)
	}

	// Verify both keys can verify the signature
	valid1, err := vcon1.Verify(convertedPublicKey)
	if err != nil {
		t.Fatalf("Failed to verify with converted public key: %v", err)
	}
	if !valid1 {
		t.Error("Signature verification failed with converted public key")
	}

	valid2, err := vcon2.Verify(publicKey)
	if err != nil {
		t.Fatalf("Failed to verify with original public key: %v", err)
	}
	if !valid2 {
		t.Error("Signature verification failed with original public key")
	}
}

func TestJWSFormat(t *testing.T) {
	// Generate test keys
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create test vCon
	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})

	// Sign the vCon
	err = vcon.Sign(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign vCon: %v", err)
	}

	// Get signed payload
	payload, err := vcon.GetSignedPayload()
	if err != nil {
		t.Fatalf("Failed to get signed payload: %v", err)
	}

	if payload == nil {
		t.Error("Signed payload should not be nil")
		return
	}

	// Test JWS token format
	sig := vcon.Signatures[0]
	jwsToken := sig.Protected + "." + *payload + "." + sig.Signature

	// Verify JWS token
	valid, err := VerifyJWS(jwsToken, publicKey)
	if err != nil {
		t.Fatalf("Failed to verify JWS token: %v", err)
	}
	if !valid {
		t.Error("JWS token verification failed")
	}

	// Test invalid JWS format
	_, err = VerifyJWS("invalid.format", publicKey)
	if err == nil {
		t.Error("Expected error for invalid JWS format")
	}
}

func TestSignatureUtilities(t *testing.T) {
	vcon := NewWithDefaults()

	// Test unsigned vCon
	if vcon.IsSigned() {
		t.Error("New vCon should not be signed")
	}

	_, err := vcon.GetSignedPayload()
	if err == nil {
		t.Error("Expected error getting payload from unsigned vCon")
	}

	// Sign the vCon
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	err = vcon.Sign(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign vCon: %v", err)
	}

	// Test signed vCon
	if !vcon.IsSigned() {
		t.Error("Signed vCon should be marked as signed")
	}

	payload, err := vcon.GetSignedPayload()
	if err != nil {
		t.Fatalf("Failed to get signed payload: %v", err)
	}
	if payload == nil {
		t.Error("Signed payload should not be nil")
	}
}

func TestCryptoEdgeCases(t *testing.T) {
	// Test signing with nil private key
	vcon := NewWithDefaults()

	// Check if it panics or returns error
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when signing with nil private key")
			}
		}()
		_ = vcon.Sign(nil)
	}()

	// Test verification with nil public key
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	err = vcon.Sign(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign vCon: %v", err)
	}

	// Check if it panics or returns error
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic when verifying with nil public key")
			}
		}()
		_, _ = vcon.Verify(nil)
	}()
}

func TestInvalidPEMData(t *testing.T) {
	// Test invalid private key PEM
	invalidPEM := []byte("-----BEGIN PRIVATE KEY-----\ninvalid data\n-----END PRIVATE KEY-----")
	_, err := PrivateKeyFromPEM(invalidPEM)
	if err == nil {
		t.Error("Expected error for invalid private key PEM")
	}

	// Test invalid public key PEM
	_, err = PublicKeyFromPEM(invalidPEM)
	if err == nil {
		t.Error("Expected error for invalid public key PEM")
	}

	// Test malformed PEM (no PEM block)
	malformedPEM := []byte("not a PEM file")
	_, err = PrivateKeyFromPEM(malformedPEM)
	if err == nil {
		t.Error("Expected error for malformed PEM")
	}

	_, err = PublicKeyFromPEM(malformedPEM)
	if err == nil {
		t.Error("Expected error for malformed PEM")
	}

	// Test wrong key type (try to load public key as private key)
	_, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	publicPEM, err := PublicKeyToPEM(publicKey)
	if err != nil {
		t.Fatalf("Failed to convert public key to PEM: %v", err)
	}

	_, err = PrivateKeyFromPEM(publicPEM)
	if err == nil {
		t.Error("Expected error when loading public key as private key")
	}

	// Test wrong key type (try to load EC key as RSA key)
	// Create a mock EC key PEM
	ecKeyPEM := []byte(`-----BEGIN PRIVATE KEY-----
MIGHAgEAMBMGByqGSM49AgEGCCqGSM49AwEHBG0wawIBAQQgg8XD7T5XVYEL/6PN
9pq6k8o6xh9a0d5qB9gq2Q/LGcKhRANCAAQ5I1bQj3cY6Zr1hEqZJ8HQj8K+R6dJ
7J9v1B7Qx8g6g7O+7H5E7D8R0e7+7K+6e7Z1g8m5g3o5k7m2v7z4r1w0
-----END PRIVATE KEY-----`)

	_, err = PrivateKeyFromPEM(ecKeyPEM)
	if err == nil {
		t.Error("Expected error when loading EC key as RSA key")
	}
}

func TestJWSEdgeCases(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Test invalid JWS format (wrong number of parts)
	invalidJWS := []string{
		"only.one.part.too.many",
		"only.one",
		"",
		"no.dots.here",
	}

	for _, jws := range invalidJWS {
		t.Run("invalid_"+jws, func(t *testing.T) {
			_, err := VerifyJWS(jws, publicKey)
			if err == nil {
				t.Error("Expected error for invalid JWS format")
			}
		})
	}

	// Test JWS with invalid base64 encoding
	invalidB64JWS := "invalid-base64.invalid-base64.invalid-base64"
	_, err = VerifyJWS(invalidB64JWS, publicKey)
	if err == nil {
		t.Error("Expected error for invalid base64 in JWS")
	}

	// Test JWS with invalid algorithm
	vcon := NewWithDefaults()
	err = vcon.Sign(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign vCon: %v", err)
	}

	// Create a JWS with wrong algorithm
	wrongAlgHeaderB64 := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXUyJ9" // Base64 of wrong header
	payload, _ := vcon.GetSignedPayload()
	sig := vcon.Signatures[0]
	wrongAlgJWS := wrongAlgHeaderB64 + "." + *payload + "." + sig.Signature

	_, err = VerifyJWS(wrongAlgJWS, publicKey)
	if err == nil {
		t.Error("Expected error for unsupported algorithm")
	}
}

func TestCorruptedSignatures(t *testing.T) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})

	err = vcon.Sign(privateKey)
	if err != nil {
		t.Fatalf("Failed to sign vCon: %v", err)
	}

	// Test with corrupted signature (but valid base64)
	originalSig := vcon.Signatures[0].Signature
	vcon.Signatures[0].Signature = "SGVsbG8gV29ybGQ" // "Hello World" in base64 - wrong signature

	valid, err := vcon.Verify(publicKey)
	if err != nil {
		t.Fatalf("Unexpected error during verification: %v", err)
	}
	if valid {
		t.Error("Signature should be invalid with corrupted signature")
	}

	// Restore original signature
	vcon.Signatures[0].Signature = originalSig

	// Test with corrupted payload
	originalPayload := *vcon.Payload
	corruptedPayload := "corrupted-payload"
	vcon.Payload = &corruptedPayload

	valid, err = vcon.Verify(publicKey)
	if err != nil {
		t.Fatalf("Unexpected error during verification: %v", err)
	}
	if valid {
		t.Error("Signature should be invalid with corrupted payload")
	}

	// Restore original payload
	vcon.Payload = &originalPayload

	// Test with corrupted protected header
	originalProtected := vcon.Signatures[0].Protected
	vcon.Signatures[0].Protected = "corrupted-header"

	_, err = vcon.Verify(publicKey)
	if err == nil {
		t.Error("Expected error for corrupted protected header")
	}

	// Restore original protected header
	vcon.Signatures[0].Protected = originalProtected

	// Verify it works again
	valid, err = vcon.Verify(publicKey)
	if err != nil {
		t.Fatalf("Unexpected error after restoration: %v", err)
	}
	if !valid {
		t.Error("Signature should be valid after restoration")
	}
}

func TestSigningLargePayload(t *testing.T) {
	// Test signing a vCon with large amounts of data
	vcon := NewWithDefaults()

	// Add many parties
	for i := 0; i < 100; i++ {
		party := Party{
			Name: StringPtr("User " + string(rune(48+i%10))), // Use digits 0-9
			Tel:  StringPtr("+123456789" + string(rune(48+i%10))),
			Meta: map[string]any{
				"index":       i,
				"description": "This is a test party with a longer description to increase data size",
			},
		}
		vcon.AddParty(party)
	}

	// Add many dialogs
	for i := 0; i < 500; i++ {
		dialog := Dialog{
			Type:    "text",
			Start:   time.Now().Add(time.Duration(i) * time.Second),
			Parties: NewDialogPartiesArrayPtr([]int{i % 100}),
			Body:    "This is a test message with some content to make it larger. Message number: " + string(rune(48+i%10)),
		}
		vcon.AddDialog(dialog)
	}

	// Generate keys and sign
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	start := time.Now()
	err = vcon.Sign(privateKey)
	signingTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to sign large vCon: %v", err)
	}

	t.Logf("Signing time for large vCon: %v", signingTime)

	// Verify signature
	start = time.Now()
	valid, err := vcon.Verify(publicKey)
	verificationTime := time.Since(start)

	if err != nil {
		t.Fatalf("Failed to verify large vCon: %v", err)
	}

	if !valid {
		t.Error("Large vCon signature should be valid")
	}

	t.Logf("Verification time for large vCon: %v", verificationTime)

	// Performance thresholds (generous for large payloads)
	if signingTime > 10*time.Second {
		t.Errorf("Signing took too long: %v", signingTime)
	}

	if verificationTime > 10*time.Second {
		t.Errorf("Verification took too long: %v", verificationTime)
	}
}

// Benchmark tests for crypto operations.
func BenchmarkGenerateKeyPair(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, err := GenerateKeyPair()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSignVCon(b *testing.B) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})
	vcon.AddDialog(Dialog{
		Type:    "text",
		Start:   time.Now(),
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Test message",
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Create a copy for each iteration to avoid state pollution
		testVCon := NewWithDefaults()
		testVCon.AddParty(Party{Name: StringPtr("Test User")})
		testVCon.AddDialog(Dialog{
			Type:    "text",
			Start:   time.Now(),
			Parties: NewDialogPartiesArrayPtr([]int{0}),
			Body:    "Test message",
		})

		err := testVCon.Sign(privateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyVCon(b *testing.B) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})
	vcon.AddDialog(Dialog{
		Type:    "text",
		Start:   time.Now(),
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Test message",
	})

	err = vcon.Sign(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vcon.Verify(publicKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrivateKeyToPEM(b *testing.B) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := PrivateKeyToPEM(privateKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPrivateKeyFromPEM(b *testing.B) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	pem, err := PrivateKeyToPEM(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := PrivateKeyFromPEM(pem)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestSignWithHeaders(t *testing.T) {
	// Generate test keys
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create test vCon
	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})
	vcon.AddDialog(Dialog{
		Type:    "text",
		Start:   vcon.CreatedAt,
		Parties: NewDialogPartiesArrayPtr([]int{0}),
		Body:    "Test message",
	})

	// Test signing with custom headers
	customHeaders := map[string]interface{}{
		"x5c": []string{"certificate1", "certificate2"}, // X.509 certificate chain
		"kid": "key-id-123",                             // Key ID
		"cty": "application/json",                       // Content type
	}

	err = vcon.SignWithHeaders(privateKey, customHeaders)
	if err != nil {
		t.Fatalf("Failed to sign vCon with headers: %v", err)
	}

	// Verify signature was added with headers
	if !vcon.IsSigned() {
		t.Error("vCon should be marked as signed")
	}

	if len(vcon.Signatures) == 0 {
		t.Error("No signatures found")
	}

	signature := vcon.Signatures[0]
	if signature.Header == nil {
		t.Error("Signature should have header parameters")
	}

	// Check header parameters
	if signature.Header["kid"] != "key-id-123" {
		t.Errorf("Expected kid 'key-id-123', got %v", signature.Header["kid"])
	}

	if signature.Header["cty"] != "application/json" {
		t.Errorf("Expected cty 'application/json', got %v", signature.Header["cty"])
	}

	// Verify the signature works
	valid, err := vcon.Verify(publicKey)
	if err != nil {
		t.Fatalf("Failed to verify vCon: %v", err)
	}
	if !valid {
		t.Error("Signature verification failed")
	}

	// Test signing with nil headers (should work like normal Sign)
	vcon2 := NewWithDefaults()
	vcon2.AddParty(Party{Name: StringPtr("Test User")})

	err = vcon2.SignWithHeaders(privateKey, nil)
	if err != nil {
		t.Fatalf("Failed to sign vCon with nil headers: %v", err)
	}

	if vcon2.Signatures[0].Header != nil {
		t.Error("Header should be nil when no headers provided")
	}
}

func TestSignatureValidation(t *testing.T) {
	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})

	// Test unsigned vCon - should pass validation
	err := vcon.ValidateSignaturePresence()
	if err != nil {
		t.Errorf("Unsigned vCon should pass signature validation: %v", err)
	}

	// Add signature without payload - should fail
	vcon.Signatures = []Signature{
		{
			Protected: "test-header",
			Signature: "test-signature",
		},
	}

	err = vcon.ValidateSignaturePresence()
	if err == nil {
		t.Error("vCon with signature but no payload should fail validation")
	}
	if err.Error() != "payload is required when signatures are present" {
		t.Errorf("Expected specific error message, got: %v", err)
	}

	// Add payload - should pass
	payload := "test-payload"
	vcon.Payload = &payload

	err = vcon.ValidateSignaturePresence()
	if err != nil {
		t.Errorf("vCon with signature and payload should pass validation: %v", err)
	}

	// Test validation through ValidateAdvanced
	err = vcon.ValidateAdvanced()
	if err != nil {
		// This might fail for other reasons (like invalid parties), but shouldn't fail on signature validation
		if strings.Contains(err.Error(), "payload is required when signatures are present") {
			t.Errorf("ValidateAdvanced should not fail on signature validation when payload is present: %v", err)
		}
	}
}

func TestX509CertificateChainHandling(t *testing.T) {
	// This test simulates handling X.509 certificate chains in JWS headers
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("Failed to generate key pair: %v", err)
	}

	// Create mock certificate chain
	cert1 := "MIIC...mock-cert-1"
	cert2 := "MIIC...mock-cert-2"
	rootCert := "MIIC...mock-root-cert"

	headers := map[string]interface{}{
		"x5c": []string{cert1, cert2, rootCert}, // X.509 certificate chain
		"x5t": "thumbprint123",                  // X.509 certificate thumbprint
		"kid": "cert-key-id",                    // Key identifier
	}

	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})

	err = vcon.SignWithHeaders(privateKey, headers)
	if err != nil {
		t.Fatalf("Failed to sign with X.509 headers: %v", err)
	}

	// Verify certificate chain is preserved in signature
	signature := vcon.Signatures[0]
	x5c, ok := signature.Header["x5c"].([]string)
	if !ok {
		t.Error("x5c header should be []string")
	} else {
		if len(x5c) != 3 {
			t.Errorf("Expected 3 certificates in chain, got %d", len(x5c))
		}
		if x5c[0] != cert1 {
			t.Errorf("First certificate mismatch: expected %s, got %s", cert1, x5c[0])
		}
	}

	// Verify signature still works
	valid, err := vcon.Verify(publicKey)
	if err != nil {
		t.Fatalf("Failed to verify signed vCon: %v", err)
	}
	if !valid {
		t.Error("Signature should be valid")
	}
}

func BenchmarkVerifyJWS(b *testing.B) {
	privateKey, publicKey, err := GenerateKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	vcon := NewWithDefaults()
	vcon.AddParty(Party{Name: StringPtr("Test User")})

	err = vcon.Sign(privateKey)
	if err != nil {
		b.Fatal(err)
	}

	payload, err := vcon.GetSignedPayload()
	if err != nil {
		b.Fatal(err)
	}

	sig := vcon.Signatures[0]
	jwsToken := sig.Protected + "." + *payload + "." + sig.Signature

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := VerifyJWS(jwsToken, publicKey)
		if err != nil {
			b.Fatal(err)
		}
	}
}
