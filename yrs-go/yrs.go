package yrs

/*
#include "libyrs.h"

#cgo CFLAGS: -I../libs

#cgo LDFLAGS: -L../libs -lyrs

#include <stdlib.h>
*/
import "C"
import (
	"unsafe"
)

type YDoc struct {
	ptr unsafe.Pointer
}

// NewDoc inicializa un nuevo documento Yrs.
func NewDoc() *YDoc {
	// Llama al constructor de C.
	cPtr := C.ydoc_new()

	// Envuelve el puntero C en la estructura Go.
	return &YDoc{ptr: unsafe.Pointer(cPtr)}
}

// Close libera la memoria del documento Yrs en el lado de Rust.
func (d *YDoc) Close() {
	if d.ptr != nil {
		// Castea de vuelta a YDoc* para el destructor de C.
		cPtr := (*C.YDoc)(d.ptr)

		C.ydoc_destroy(cPtr)
		d.ptr = nil
	}
}

// FreeRustMemory libera punteros crudos devueltos por Rust (strings, etc.).
// Es crucial usar esta función cuando Rust nos entrega un puntero que debemos liberar.
func YStringDestroy(ptr unsafe.Pointer) {
	if ptr != nil {
		// ystring_destroy expects a `char *` (C string). Convert the
		// unsafe.Pointer we received to *C.char before calling.
		C.ystring_destroy((*C.char)(ptr))
	}
}

// FreeYOutput libera la estructura YOutput devuelta por algunas funciones.
func FreeYOutput(output *C.YOutput) {
	if output != nil {
		C.youtput_destroy(output)
	}
}

// YTransaction is a Go wrapper around the C YTransaction type.
// Use `ReadTransaction` and `WriteTransaction` on `YDoc` to obtain one.
type YTransaction struct {
	ptr *C.YTransaction
}

// ReadTransaction starts a read-only transaction on the document.
// Returns nil if a transaction couldn't be created.
func (d *YDoc) ReadTransaction() *YTransaction {
	if d.ptr == nil {
		return nil
	}
	cDoc := (*C.YDoc)(d.ptr)
	cTxn := C.ydoc_read_transaction(cDoc)
	if cTxn == nil {
		return nil
	}
	return &YTransaction{ptr: cTxn}
}

// WriteTransaction starts a read-write transaction on the document.
// If origin is empty, no origin is set.
func (d *YDoc) WriteTransaction(origin string) *YTransaction {
	if d.ptr == nil {
		return nil
	}
	cDoc := (*C.YDoc)(d.ptr)
	if origin == "" {
		cTxn := C.ydoc_write_transaction(cDoc, 0, nil)
		if cTxn == nil {
			return nil
		}
		return &YTransaction{ptr: cTxn}
	}

	cOrigin := C.CString(origin)
	defer C.free(unsafe.Pointer(cOrigin))
	cTxn := C.ydoc_write_transaction(cDoc, C.uint32_t(len(origin)), cOrigin)
	if cTxn == nil {
		return nil
	}
	return &YTransaction{ptr: cTxn}
}

// Commit commits and disposes a read-write transaction.
// After calling Commit, the transaction pointer is nil'ed.
func (t *YTransaction) Commit() {
	if t == nil || t.ptr == nil {
		return
	}
	C.ytransaction_commit(t.ptr)
	t.ptr = nil
}

// ForceGC triggers a garbage collection of deleted blocks for the document
// within the scope of this transaction.
func (t *YTransaction) ForceGC() {
	if t == nil || t.ptr == nil {
		return
	}
	C.ytransaction_force_gc(t.ptr)
}

// StateVectorV1 returns the state vector of the transaction's document
// encoded as a binary payload using lib0 version 1 encoding.
// Returns the binary data and its length.
func (t *YTransaction) StateVectorV1() ([]byte, error) {
	if t == nil || t.ptr == nil {
		return nil, nil
	}
	var length C.uint32_t
	cResult := C.ytransaction_state_vector_v1(t.ptr, &length)
	if cResult == nil {
		return nil, nil
	}

	// Convert C data to Go slice
	data := C.GoBytes(unsafe.Pointer(cResult), C.int(length))

	// Free the C memory
	C.ybinary_destroy(cResult, length)

	return data, nil
}

// StateDiffV1 returns a delta difference between current state and a state vector
// encoded as binary payload. If sv is nil, returns a full snapshot.
func (t *YTransaction) StateDiffV1(sv []byte) ([]byte, error) {
	if t == nil || t.ptr == nil {
		return nil, nil
	}

	var length C.uint32_t
	var cResult *C.char

	if sv == nil {
		// Generate full snapshot
		cResult = C.ytransaction_state_diff_v1(t.ptr, nil, 0, &length)
	} else {
		// Generate diff from state vector
		cResult = C.ytransaction_state_diff_v1(t.ptr, (*C.char)(unsafe.Pointer(&sv[0])), C.uint32_t(len(sv)), &length)
	}

	if cResult == nil {
		return nil, nil
	}

	// Convert C data to Go slice
	data := C.GoBytes(unsafe.Pointer(cResult), C.int(length))

	// Free the C memory
	C.ybinary_destroy(cResult, length)

	return data, nil
}

// Apply applies a binary update to the transaction's document.
// Returns 0 on success, or an error code on failure.
func (t *YTransaction) Apply(update []byte) uint8 {
	if t == nil || t.ptr == nil || len(update) == 0 {
		return 1 // Error
	}

	return uint8(C.ytransaction_apply(t.ptr, (*C.char)(unsafe.Pointer(&update[0])), C.uint32_t(len(update))))
}

// YBinaryDestroy frees binary data returned by Yrs functions
func YBinaryDestroy(ptr *C.char, length uint32) {
	if ptr != nil {
		C.ybinary_destroy(ptr, C.uint32_t(length))
	}
}
