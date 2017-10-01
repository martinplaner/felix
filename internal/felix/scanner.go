// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"context"
	"io"
)

// Scanner scans the contents of the reader and emits items, links or follow URLs.
type Scanner interface {
	Scan(context.Context, io.Reader, Emitter) error
}

// ScanFunc is an adapter to allow the use of ordinary functions as scanners.
// If f is a function with the appropriate signature, ScanFunc(f) is a Scanner that calls f.
type ScanFunc func(context.Context, io.Reader, Emitter) error

// Scan calls the underlying ScanFunc
func (f ScanFunc) Scan(ctx context.Context, r io.Reader, e Emitter) error {
	return f(ctx, r, e)
}
