package main

import (
	"io"
	"testing"

	"github.com/cray-hpe/tpm-provisioner/tests/simulateTPM"
	"github.com/google/go-tpm-tools/simulator"
)

func openTPM(tb testing.TB) io.ReadWriteCloser {
	tb.Helper()

	simulator, err := simulator.Get()
	if err != nil {
		tb.Fatalf("Simulator initialization failed: %v", err)
	}

	_, err = simulateTPM.CreateEK(simulator)
	if err != nil {
		tb.Fatalf("Unable to provision EK: %v", err)
	}

	return simulator
}

func TestGetEKPublicCertificate(t *testing.T) {
	rwc := openTPM(t)

	defer rwc.Close()

	_, err := getEKPublicCertificate(rwc)
	if err != nil {
		t.Fatalf("Error retrieving EK Public Certificate: %v", err)
	}
}
