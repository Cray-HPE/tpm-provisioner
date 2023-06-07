/*
 *
 *  MIT License
 *
 *  (C) Copyright 2023 Hewlett Packard Enterprise Development LP
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a
 *  copy of this software and associated documentation files (the "Software"),
 *  to deal in the Software without restriction, including without limitation
 *  the rights to use, copy, modify, merge, publish, distribute, sublicense,
 *  and/or sell copies of the Software, and to permit persons to whom the
 *  Software is furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included
 *  in all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL
 *  THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR
 *  OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE,
 *  ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
 *  OTHER DEALINGS IN THE SOFTWARE.
 *
 */
package blob_test

import (
	"io"
	"testing"

	"github.com/cray-hpe/tpm-provisioner/pkg/blob"
	"github.com/cray-hpe/tpm-provisioner/pkg/test"
	"github.com/google/go-tpm-tools/simulator"
)

// openTPM opens a simulated tpm and populate's its EK Certificate.
func openTPM(tb testing.TB) io.ReadWriteCloser {
	tb.Helper()

	simulator, err := simulator.Get()
	if err != nil {
		tb.Fatalf("Simulator initialization failed: %v", err)
	}

	_, err = test.CreateEK(simulator)
	if err != nil {
		tb.Fatalf("Unable to provision EK: %v", err)
	}

	return simulator
}

// TestBlob validates that a blob can be saved and then loaded from a TPM.
func TestBlob(t *testing.T) {
	rwc := openTPM(t)
	addr := 0x1800000
	data := []byte(`TESTBYTES`)

	err := blob.SaveData(rwc, addr, data)
	if err != nil {
		t.Logf("Failed To Save Data to TPM: %v", err)
		t.Fail()
	}

	actual, err := blob.LoadData(rwc, addr)
	if err != nil {
		t.Logf("Failed To Load Data from TPM: %v", err)
		t.Fail()
	}

	if string(actual) != string(data) {
		t.Logf("Data does not match.\nExpected: %v\nActual: %v", data, actual)
		t.Fail()
	}
}
