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
package provisioner

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// requestSpireWorkloads sends a request to the spire tokens service to create
// workloads for the newly joined via tpm spire client.
func requestSpireWorkloads(nodeType string, xname string, spireTokensURL string) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
	}

	log.Print("Creating spire workloads")
	log.Printf("type: %+v", nodeType)
	log.Printf("xname: %+v", xname)

	httpClient := http.Client{Transport: tr}

	data := url.Values{}
	data.Set("xname", xname)
	data.Set("type", nodeType)
	log.Printf("data: +%v", data)
	log.Printf("dataEncode: +%v", data.Encode())

	req, err := http.NewRequest(http.MethodPost, spireTokensURL, bytes.NewBufferString(data.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	log.Printf("req: %+v", req)

	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	d, _ := io.ReadAll(req.Body)
	log.Printf("body: %+v", d)

	if resp.StatusCode != 200 {
		data, e := io.ReadAll(resp.Body)
		if e != nil {
			return e
		}

		return fmt.Errorf("error requesting tpm workload creation: %+v: %v", resp.StatusCode, string(data))
	}

	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}
