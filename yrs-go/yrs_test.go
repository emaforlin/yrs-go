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

// TestNewDoc tests creating a new YDoc instance
func TestNewDoc(t *testing.T) {
	doc := NewDoc()
	if doc == nil {
		t.Fatal("NewDoc() returned nil")
	}
	if doc.ptr == nil {
		t.Fatal("NewDoc() returned YDoc with nil ptr")
	}

	doc.Close()
}

// TestYDocClose tests closing a YDoc
func TestYDocClose(t *testing.T) {
	doc := NewDoc()
	if doc == nil {
		t.Fatal("NewDoc() returned nil")
	}

	// Close should not panic and should set ptr to nil
	doc.Close()
	if doc.ptr != nil {
		t.Error("After Close(), YDoc.ptr should be nil")
	}

	// Calling Close again should not panic
	doc.Close()
}

// TestYDocReadTransaction tests creating a read transaction
func TestYDocReadTransaction(t *testing.T) {
	doc := NewDoc()
	defer doc.Close()

	txn := doc.ReadTransaction()
	if txn == nil {
		t.Fatal("ReadTransaction() returned nil")
	}
	if txn.ptr == nil {
		t.Fatal("ReadTransaction() returned transaction with nil ptr")
	}

	// Commit the transaction
	txn.Commit()
	if txn.ptr != nil {
		t.Error("After Commit(), transaction ptr should be nil")
	}
}

// TestYDocWriteTransaction tests creating a write transaction
func TestYDocWriteTransaction(t *testing.T) {
	doc := NewDoc()
	defer doc.Close()

	// Test write transaction without origin
	txn := doc.WriteTransaction("")
	if txn == nil {
		t.Fatal("WriteTransaction('') returned nil")
	}
	if txn.ptr == nil {
		t.Fatal("WriteTransaction('') returned transaction with nil ptr")
	}
	txn.Commit()

	// Test write transaction with origin
	txn2 := doc.WriteTransaction("test-origin")
	if txn2 == nil {
		t.Fatal("WriteTransaction('test-origin') returned nil")
	}
	if txn2.ptr == nil {
		t.Fatal("WriteTransaction('test-origin') returned transaction with nil ptr")
	}
	txn2.Commit()
}

// TestYDocWriteTransactionWithClosedDoc tests transaction creation on a closed document
func TestYDocWriteTransactionWithClosedDoc(t *testing.T) {
	doc := NewDoc()
	doc.Close()

	// Should return nil when document is closed
	txn := doc.ReadTransaction()
	if txn != nil {
		t.Error("ReadTransaction() on closed doc should return nil")
	}

	txn2 := doc.WriteTransaction("test")
	if txn2 != nil {
		t.Error("WriteTransaction() on closed doc should return nil")
	}
}

// TestYTransactionCommit tests transaction commit behavior
func TestYTransactionCommit(t *testing.T) {
	doc := NewDoc()
	defer doc.Close()

	txn := doc.WriteTransaction("test")
	if txn == nil {
		t.Fatal("WriteTransaction() returned nil")
	}

	// First commit should work
	txn.Commit()
	if txn.ptr != nil {
		t.Error("After Commit(), transaction ptr should be nil")
	}

	// Second commit should not panic
	txn.Commit()
}

// TestYTransactionForceGC tests garbage collection
func TestYTransactionForceGC(t *testing.T) {
	doc := NewDoc()
	defer doc.Close()

	txn := doc.WriteTransaction("test")
	if txn == nil {
		t.Fatal("WriteTransaction() returned nil")
	}

	// ForceGC should not panic
	txn.ForceGC()

	// Should still be able to commit after GC
	txn.Commit()
}

// TestYTransactionNilBehavior tests behavior with nil transactions
func TestYTransactionNilBehavior(t *testing.T) {
	var txn *YTransaction

	// These should not panic with nil transaction
	txn.Commit()
	txn.ForceGC()
}

// TestYStringDestroy tests the string destruction utility
func TestYStringDestroy(t *testing.T) {
	// Test with nil pointer - should not panic
	YStringDestroy(nil)

	// We can't easily test with a real C string without calling
	// C functions that return strings, but we can at least verify
	// the function doesn't panic with nil input
}

// TestFreeYOutput tests the YOutput destruction utility
func TestFreeYOutput(t *testing.T) {
	// Test with nil pointer - should not panic
	FreeYOutput(nil)

	// Similar to YStringDestroy, we can't easily test with real
	// YOutput without calling C functions that return them
}
