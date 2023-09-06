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
// Package blob stores and retrieves data from a TPM's NVRam.
package blob

import (
	"fmt"
	"io"

	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// Clear undefines the space used for storing the TPM Provisioner blobs,
// removing the data and allowing it to be reused.
func Clear(rwc io.ReadWriter, cfg Config) error {
	err := ClearAddr(rwc, cfg.DevCertAddr)
	if err != nil {
		return err
	}

	err = ClearAddr(rwc, cfg.DevPubBlobAddr)
	if err != nil {
		return err
	}

	err = ClearAddr(rwc, cfg.DevPrivBlobAddr)
	if err != nil {
		return err
	}

	return nil
}

// ClearAddr undefines the NVRam address space, allowing it to be used in the futrue.
func ClearAddr(rwc io.ReadWriter, addr int) error {
	fmt.Printf("Clearing %x\n", addr)

	err := tpm2.NVUndefineSpace(rwc, "", tpm2.HandleOwner, tpmutil.Handle(addr))
	if err != nil {
		return err
	}

	return nil
}
