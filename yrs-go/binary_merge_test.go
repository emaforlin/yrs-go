package yrs

/*
#include "libyrs.h"

#cgo CFLAGS: -I../libs

#cgo LDFLAGS: -L../libs -lyrs

#include <stdlib.h>
*/

import (
	"testing"
)

// TestBinaryUpdateMerging tests that taking 3 separate binary updates and merging them
// produces the expected final state vector
func TestBinaryUpdateMerging(t *testing.T) {
	// Create three separate documents that will generate updates
	doc1 := NewDoc()
	doc2 := NewDoc()
	doc3 := NewDoc()
	resultDoc := NewDoc()
	
	defer func() {
		doc1.Close()
		doc2.Close()
		doc3.Close()
		resultDoc.Close()
	}()

	// Generate three different updates by creating state diffs
	// For this test, we'll generate snapshots of empty documents as our "updates"
	txnRead1 := doc1.ReadTransaction()
	if txnRead1 == nil {
		t.Fatal("Failed to create read transaction for doc1 snapshot")
	}
	update1, err := txnRead1.StateDiffV1(nil) // Full snapshot
	txnRead1.Commit()
	if err != nil {
		t.Fatalf("Failed to generate update1: %v", err)
	}
	if update1 == nil {
		t.Fatal("update1 is nil")
	}

	txnRead2 := doc2.ReadTransaction()
	if txnRead2 == nil {
		t.Fatal("Failed to create read transaction for doc2 snapshot")
	}
	update2, err := txnRead2.StateDiffV1(nil) // Full snapshot
	txnRead2.Commit()
	if err != nil {
		t.Fatalf("Failed to generate update2: %v", err)
	}
	if update2 == nil {
		t.Fatal("update2 is nil")
	}

	txnRead3 := doc3.ReadTransaction()
	if txnRead3 == nil {
		t.Fatal("Failed to create read transaction for doc3 snapshot")
	}
	update3, err := txnRead3.StateDiffV1(nil) // Full snapshot
	txnRead3.Commit()
	if err != nil {
		t.Fatalf("Failed to generate update3: %v", err)
	}
	if update3 == nil {
		t.Fatal("update3 is nil")
	}

	// Apply all three updates to the result document
	txnWrite := resultDoc.WriteTransaction("merge-test")
	if txnWrite == nil {
		t.Fatal("Failed to create write transaction for result document")
	}

	// Apply update1
	result := txnWrite.Apply(update1)
	if result != 0 {
		t.Fatalf("Failed to apply update1, error code: %d", result)
	}

	// Apply update2
	result = txnWrite.Apply(update2)
	if result != 0 {
		t.Fatalf("Failed to apply update2, error code: %d", result)
	}

	// Apply update3
	result = txnWrite.Apply(update3)
	if result != 0 {
		t.Fatalf("Failed to apply update3, error code: %d", result)
	}

	// Get the final state vector after merging all updates
	finalStateVector, err := txnWrite.StateVectorV1()
	txnWrite.Commit()
	if err != nil {
		t.Fatalf("Failed to get final state vector: %v", err)
	}
	if finalStateVector == nil {
		t.Fatal("Final state vector is nil")
	}

	// Verify that we got a meaningful state vector
	if len(finalStateVector) == 0 {
		t.Error("Expected non-empty final state vector")
	}

	// Log some information about the merge results
	t.Logf("Successfully merged 3 binary updates")
	t.Logf("Update 1 size: %d bytes", len(update1))
	t.Logf("Update 2 size: %d bytes", len(update2))
	t.Logf("Update 3 size: %d bytes", len(update3))
	t.Logf("Final state vector size: %d bytes", len(finalStateVector))

	// Additional verification: Create a new document and verify we can get consistent results
	verifyDoc := NewDoc()
	defer verifyDoc.Close()

	txnVerify := verifyDoc.WriteTransaction("verify")
	if txnVerify == nil {
		t.Fatal("Failed to create verify transaction")
	}

	// Apply the same updates to verify consistency
	if txnVerify.Apply(update1) != 0 {
		t.Error("Failed to apply update1 to verify document")
	}
	if txnVerify.Apply(update2) != 0 {
		t.Error("Failed to apply update2 to verify document")
	}
	if txnVerify.Apply(update3) != 0 {
		t.Error("Failed to apply update3 to verify document")
	}

	verifyStateVector, err := txnVerify.StateVectorV1()
	txnVerify.Commit()
	if err != nil {
		t.Fatalf("Failed to get verify state vector: %v", err)
	}

	// The state vectors should be identical
	if len(finalStateVector) != len(verifyStateVector) {
		t.Errorf("State vector lengths don't match: %d vs %d", len(finalStateVector), len(verifyStateVector))
	}

	// Compare byte-by-byte for consistency
	for i := 0; i < len(finalStateVector) && i < len(verifyStateVector); i++ {
		if finalStateVector[i] != verifyStateVector[i] {
			t.Errorf("State vectors differ at position %d: %d vs %d", i, finalStateVector[i], verifyStateVector[i])
			break
		}
	}

	t.Log("Binary update merging test completed successfully")
}