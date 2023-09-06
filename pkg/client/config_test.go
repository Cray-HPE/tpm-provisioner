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
package client_test

import (
	"testing"

	"github.com/cray-hpe/tpm-provisioner/pkg/client"
)

// TestParseConfig validates that we're able to parse a test configuration file
// correctly.
func TestParseConfig(t *testing.T) {
	cfg, err := client.ParseConfig("../../tests", "client.conf")
	if err != nil {
		t.Fatalf("Unable to parse config: %v", err)
	}

	expected := "/tmp/output"
	actual := cfg.OutputDir
	matchType := "OutputDir"

	if actual != expected {
		t.Fatalf("Invalid %s:\nExpected: %v\nActual: %v", matchType, expected, actual)
	}

	expected = "https://127.0.0.1:8080"
	actual = cfg.URL
	matchType = "URL"

	if actual != expected {
		t.Fatalf("Invalid %s:\nExpected: %v\nActual: %v", matchType, expected, actual)
	}

	expected = "/var/lib/spire/agent.sock"
	actual = cfg.SocketPath
	matchType = "socketPath"

	if actual != expected {
		t.Fatalf("Invalid %s:\nExpected: %v\nActual: %v", matchType, expected, actual)
	}
}
