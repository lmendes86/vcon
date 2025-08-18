//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"time"

	vcon "github.com/lmendes86/vcon"
)

func main() {
	fmt.Println("=== Digital Signatures for vCon Demo ===")

	// 1. Basic signing and verification
	fmt.Println("1. Basic Signing and Verification")
	basicSigningDemo()

	// 2. Key management and PEM conversion
	fmt.Println("\n2. Key Management and PEM Conversion")
	keyManagementDemo()

	// 3. Multiple signatures
	fmt.Println("\n3. Multiple Signatures")
	multipleSignaturesDemo()

	// 4. JWS token validation
	fmt.Println("\n4. JWS Token Validation")
	jwsTokenDemo()

	// 5. Real-world workflow
	fmt.Println("\n5. Real-world Workflow: Secure vCon Transmission")
	realWorldWorkflowDemo()

	fmt.Println("\n=== Digital Signatures Demo Complete ===")
}

func basicSigningDemo() {
	// Create a sample vCon
	vconSample := createSampleVCon()
	fmt.Printf("  Created vCon with %d parties and %d dialogs\n",
		len(vconSample.Parties), len(vconSample.Dialog))

	// Generate key pair
	privateKey, publicKey, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}
	fmt.Println("  Generated RSA-2048 key pair")

	// Sign the vCon
	fmt.Println("  Signing vCon...")
	err = vconSample.Sign(privateKey)
	if err != nil {
		log.Fatalf("Failed to sign vCon: %v", err)
	}

	// Check if vCon is signed
	if vconSample.IsSigned() {
		fmt.Printf("  ✓ vCon is signed with %d signatures\n", len(vconSample.Signatures))
	}

	// Verify the signature
	fmt.Println("  Verifying signature...")
	valid, err := vconSample.Verify(publicKey)
	if err != nil {
		log.Fatalf("Failed to verify signature: %v", err)
	}

	if valid {
		fmt.Println("  ✓ Signature verification successful")
	} else {
		fmt.Println("  ✗ Signature verification failed")
	}

	// Try verification with wrong key
	fmt.Println("  Testing with wrong key...")
	_, wrongPublicKey, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate wrong key: %v", err)
	}

	valid, err = vconSample.Verify(wrongPublicKey)
	if err != nil {
		fmt.Printf("  Expected error with wrong key: %v\n", err)
	} else if !valid {
		fmt.Println("  ✓ Correctly rejected wrong key")
	} else {
		fmt.Println("  ✗ Incorrectly accepted wrong key")
	}
}

func keyManagementDemo() {
	// Generate keys
	privateKey, publicKey, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate keys: %v", err)
	}

	// Convert to PEM format
	privatePEM, err := vcon.PrivateKeyToPEM(privateKey)
	if err != nil {
		log.Fatalf("Failed to convert private key to PEM: %v", err)
	}

	publicPEM, err := vcon.PublicKeyToPEM(publicKey)
	if err != nil {
		log.Fatalf("Failed to convert public key to PEM: %v", err)
	}

	fmt.Printf("  Private key PEM length: %d bytes\n", len(privatePEM))
	fmt.Printf("  Public key PEM length: %d bytes\n", len(publicPEM))

	// Save keys to files
	privateKeyFile := "/tmp/private_key.pem"
	publicKeyFile := "/tmp/public_key.pem"

	err = os.WriteFile(privateKeyFile, privatePEM, 0600) // Restricted permissions
	if err != nil {
		log.Fatalf("Failed to save private key: %v", err)
	}

	err = os.WriteFile(publicKeyFile, publicPEM, 0644)
	if err != nil {
		log.Fatalf("Failed to save public key: %v", err)
	}

	fmt.Printf("  Saved private key to: %s (permissions: 0600)\n", privateKeyFile)
	fmt.Printf("  Saved public key to: %s (permissions: 0644)\n", publicKeyFile)

	// Load keys from PEM
	loadedPrivatePEM, err := os.ReadFile(privateKeyFile)
	if err != nil {
		log.Fatalf("Failed to read private key: %v", err)
	}

	loadedPublicPEM, err := os.ReadFile(publicKeyFile)
	if err != nil {
		log.Fatalf("Failed to read public key: %v", err)
	}

	loadedPrivateKey, err := vcon.PrivateKeyFromPEM(loadedPrivatePEM)
	if err != nil {
		log.Fatalf("Failed to parse private key from PEM: %v", err)
	}

	loadedPublicKey, err := vcon.PublicKeyFromPEM(loadedPublicPEM)
	if err != nil {
		log.Fatalf("Failed to parse public key from PEM: %v", err)
	}

	fmt.Println("  ✓ Successfully loaded keys from PEM files")

	// Test that loaded keys work
	testVCon := createSampleVCon()
	err = testVCon.Sign(loadedPrivateKey)
	if err != nil {
		log.Fatalf("Failed to sign with loaded key: %v", err)
	}

	valid, err := testVCon.Verify(loadedPublicKey)
	if err != nil {
		log.Fatalf("Failed to verify with loaded key: %v", err)
	}

	if valid {
		fmt.Println("  ✓ Loaded keys work correctly")
	}

	// Clean up
	os.Remove(privateKeyFile)
	os.Remove(publicKeyFile)
}

func multipleSignaturesDemo() {
	vconSample := createSampleVCon()

	// Generate multiple key pairs (simulating different signers)
	key1, pub1, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair 1: %v", err)
	}

	key2, pub2, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair 2: %v", err)
	}

	key3, pub3, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair 3: %v", err)
	}

	fmt.Println("  Generated 3 key pairs for different signers")

	// Sign with first key (e.g., call center system)
	err = vconSample.Sign(key1)
	if err != nil {
		log.Fatalf("Failed to sign with key 1: %v", err)
	}
	fmt.Printf("  Signed with key 1 - Total signatures: %d\n", len(vconSample.Signatures))

	// Sign with second key (e.g., supervisor)
	err = vconSample.Sign(key2)
	if err != nil {
		log.Fatalf("Failed to sign with key 2: %v", err)
	}
	fmt.Printf("  Signed with key 2 - Total signatures: %d\n", len(vconSample.Signatures))

	// Sign with third key (e.g., compliance officer)
	err = vconSample.Sign(key3)
	if err != nil {
		log.Fatalf("Failed to sign with key 3: %v", err)
	}
	fmt.Printf("  Signed with key 3 - Total signatures: %d\n", len(vconSample.Signatures))

	// Verify all signatures
	fmt.Println("  Verifying all signatures...")

	valid1, err := vconSample.Verify(pub1)
	if err != nil {
		log.Printf("  Error verifying signature 1: %v", err)
	} else {
		fmt.Printf("  Signature 1 valid: %t\n", valid1)
	}

	valid2, err := vconSample.Verify(pub2)
	if err != nil {
		log.Printf("  Error verifying signature 2: %v", err)
	} else {
		fmt.Printf("  Signature 2 valid: %t\n", valid2)
	}

	valid3, err := vconSample.Verify(pub3)
	if err != nil {
		log.Printf("  Error verifying signature 3: %v", err)
	} else {
		fmt.Printf("  Signature 3 valid: %t\n", valid3)
	}

	if valid1 && valid2 && valid3 {
		fmt.Println("  ✓ All signatures verified successfully")
	}
}

func jwsTokenDemo() {
	vconSample := createSampleVCon()

	// Generate key pair
	privateKey, publicKey, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate key pair: %v", err)
	}

	// Sign the vCon
	err = vconSample.Sign(privateKey)
	if err != nil {
		log.Fatalf("Failed to sign vCon: %v", err)
	}

	// Get signed payload
	payload, err := vconSample.GetSignedPayload()
	if err != nil {
		log.Fatalf("Failed to get signed payload: %v", err)
	}

	// Construct JWS token
	signature := vconSample.Signatures[0]
	jwsToken := signature.Protected + "." + *payload + "." + signature.Signature

	fmt.Printf("  JWS Token length: %d characters\n", len(jwsToken))
	fmt.Printf("  JWS Header: %s\n", signature.Protected)
	fmt.Printf("  Payload length: %d chars\n", len(*payload))
	fmt.Printf("  Signature length: %d chars\n", len(signature.Signature))

	// Verify JWS token directly
	valid, err := vcon.VerifyJWS(jwsToken, publicKey)
	if err != nil {
		log.Fatalf("Failed to verify JWS token: %v", err)
	}

	if valid {
		fmt.Println("  ✓ JWS token verification successful")
	} else {
		fmt.Println("  ✗ JWS token verification failed")
	}

	// Test invalid JWS format
	fmt.Println("  Testing invalid JWS format...")
	invalidToken := "invalid.jws.token"
	valid, err = vcon.VerifyJWS(invalidToken, publicKey)
	if err != nil {
		fmt.Printf("  Expected error for invalid format: %v\n", err)
	}
}

func realWorldWorkflowDemo() {
	fmt.Println("  Scenario: Secure transmission of customer call recording")

	// Step 1: Create vCon with customer interaction
	vconSample := vcon.NewWithDefaults()
	subject := "Customer Support - Account Issue"
	vconSample.Subject = &subject

	// Add parties
	customer := vconSample.AddParty(vcon.Party{
		Name:   stringPtr("John Customer"),
		Mailto: stringPtr("john@customer.com"),
		Role:   stringPtr("customer"),
	})

	agent := vconSample.AddParty(vcon.Party{
		Name:   stringPtr("Sarah Agent"),
		Mailto: stringPtr("sarah@company.com"),
		Role:   stringPtr("agent"),
	})

	// Add call recording dialog
	callStart := time.Now().Add(-10 * time.Minute)
	duration := 300.0 // 5 minutes
	vconSample.AddDialog(vcon.Dialog{
		Type:        "recording",
		Start:       callStart,
		Duration:    &duration,
		Parties:     vcon.NewDialogPartiesArrayPtr([]int{customer, agent}),
		Originator:  &customer,
		Mediatype:   stringPtr("audio/wav"),
		URL:         stringPtr("https://secure.company.com/recordings/call-12345.wav"),
		ContentHash: vcon.NewContentHashSingle("sha-256:LCa0a2j_xo_5m0U8HTBBNBNCLXBkg7-g-YpeiGJm564"), // Required for external URLs
	})

	// Add tags for categorization
	vconSample.AddTag("category", "account_issue")
	vconSample.AddTag("priority", "medium")
	vconSample.AddTag("department", "customer_support")

	fmt.Printf("  Created vCon for call between %s and %s\n",
		*vconSample.Parties[customer].Name, *vconSample.Parties[agent].Name)

	// Step 2: Generate company signing key
	companyPrivateKey, companyPublicKey, err := vcon.GenerateKeyPair()
	if err != nil {
		log.Fatalf("Failed to generate company key: %v", err)
	}

	// Step 3: Sign vCon for authenticity
	err = vconSample.Sign(companyPrivateKey)
	if err != nil {
		log.Fatalf("Failed to sign vCon: %v", err)
	}
	fmt.Println("  ✓ vCon signed by company")

	// Step 4: Save to secure file
	secureFile := "/tmp/secure_call_recording.json"
	err = vconSample.SaveToFile(secureFile, vcon.PropertyHandlingDefault)
	if err != nil {
		log.Fatalf("Failed to save secure file: %v", err)
	}
	fmt.Printf("  ✓ Saved signed vCon to: %s\n", secureFile)

	// Step 5: Simulate transmission and verification
	fmt.Println("  Simulating secure transmission...")

	// Load the file (simulating receipt)
	receivedVCon, err := vcon.LoadFromFile(secureFile, vcon.PropertyHandlingDefault)
	if err != nil {
		log.Fatalf("Failed to load received vCon: %v", err)
	}

	// Verify signature (simulating verification by recipient)
	valid, err := receivedVCon.Verify(companyPublicKey)
	if err != nil {
		log.Fatalf("Failed to verify received vCon: %v", err)
	}

	if valid {
		fmt.Println("  ✓ Received vCon signature verified")
		fmt.Printf("  ✓ Call recording authenticated from %s\n",
			*receivedVCon.Parties[agent].Mailto)
	} else {
		fmt.Println("  ✗ Signature verification failed - potential tampering!")
	}

	// Step 6: Validate content integrity
	isValid, errors := receivedVCon.IsValid()
	if isValid {
		fmt.Println("  ✓ vCon structure validation passed")
	} else {
		fmt.Printf("  ✗ vCon validation failed: %v\n", errors)
	}

	// Step 7: Extract and display metadata
	fmt.Println("  Extracted metadata:")
	fmt.Printf("    UUID: %s\n", receivedVCon.UUID.String())
	fmt.Printf("    Subject: %s\n", *receivedVCon.Subject)
	fmt.Printf("    Created: %s\n", receivedVCon.CreatedAt.Format(time.RFC3339))

	category := receivedVCon.GetTag("category")
	priority := receivedVCon.GetTag("priority")
	if category != nil && priority != nil {
		fmt.Printf("    Category: %s, Priority: %s\n", *category, *priority)
	}

	fmt.Printf("    Call duration: %.0f seconds\n", *receivedVCon.Dialog[0].Duration)
	fmt.Printf("    Recording URL: %s\n", *receivedVCon.Dialog[0].URL)

	// Clean up
	os.Remove(secureFile)
	fmt.Println("  ✓ Workflow completed successfully")
}

func createSampleVCon() *vcon.VCon {
	v := vcon.NewWithDefaults()

	// Add subject
	subject := "Sample Conversation"
	v.Subject = &subject

	// Add parties
	caller := v.AddParty(vcon.Party{
		Name:   stringPtr("Alice Caller"),
		Tel:    stringPtr("+1234567890"),
		Mailto: stringPtr("alice@example.com"),
	})

	receiver := v.AddParty(vcon.Party{
		Name:   stringPtr("Bob Receiver"),
		Tel:    stringPtr("+0987654321"),
		Mailto: stringPtr("bob@example.com"),
	})

	// Add dialog
	now := time.Now()
	mediatype := "text/plain"
	encoding := "none"
	v.AddDialog(vcon.Dialog{
		Type:       "text",
		Start:      now,
		Parties:    vcon.NewDialogPartiesArrayPtr([]int{caller, receiver}),
		Originator: &caller,
		Body:       "Hello, this is a test conversation",
		Mediatype:  &mediatype,
		Encoding:   &encoding,
	})

	return v
}

func stringPtr(s string) *string {
	return &s
}
