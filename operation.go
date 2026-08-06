package mycel

import (
	"crypto/rand"
	"fmt"
)

// NewOperationID returns a new UUID v4 string suitable for
// BeginTransactionRequest.operation_id.
//
// Operation IDs are client-side correlation metadata only. They are not
// idempotency keys, authorization credentials, replay protection, or commit
// ordering guarantees.
func NewOperationID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(fmt.Errorf("mycel: generate operation ID: %w", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
