// Copyright 2017 Martin Planer. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package felix

import (
	"context"
	"io"
	"testing"
)

// Make sure ScanFunc implements the Scanner interface.
var _ Scanner = ScanFunc(nil)

func TestScanFunc(t *testing.T) {
	success := false
	var s Scanner = ScanFunc(func(ctx context.Context, r io.Reader, e Emitter) error {
		success = true
		return nil
	})

	err := s.Scan(context.TODO(), nil, nil)

	if err != nil {
		t.Error("unexpected error:", err)
	}

	if !success {
		t.Errorf("ScanFunc not called successfully. expected %v, got %v", true, success)
	}
}
