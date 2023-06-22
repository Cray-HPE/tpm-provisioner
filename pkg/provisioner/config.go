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
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"github.com/spf13/viper"
)

// Config contains the TPM Provisioner server configuration.
type Config struct {
	ManufactuerCAs *x509.CertPool
	ProviderCA     *x509.Certificate
	ProviderKey    *rsa.PrivateKey
	Port           int
	WhiteList      string
	SpireTokensURL string
}

// CFG stores the config in a global variable.
var CFG Config

// ParseConfig reads a TPM Server configuration file and sets the config.
func ParseConfig(path string, file string) error {
	viper.SetConfigName(file)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	manufacturerCAs, err := os.ReadFile(viper.GetString("manufacturerCAs"))
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(manufacturerCAs)

	pCAFile, err := os.ReadFile(viper.GetString("platformCA"))
	if err != nil {
		return err
	}

	dPCA, _ := pem.Decode(pCAFile)

	platformCA, err := x509.ParseCertificate(dPCA.Bytes)
	if err != nil {
		return err
	}

	pKeyFile, err := os.ReadFile(viper.GetString("platformKey"))
	if err != nil {
		return err
	}

	dPKey, _ := pem.Decode(pKeyFile)

	platformKey, err := x509.ParsePKCS8PrivateKey(dPKey.Bytes)
	if err != nil {
		return err
	}

	CFG = Config{
		ManufactuerCAs: certPool,
		ProviderCA:     platformCA,
		ProviderKey:    platformKey.(*rsa.PrivateKey),
		Port:           viper.GetInt("port"),
		WhiteList:      viper.GetString("whitelist"),
		SpireTokensURL: viper.GetString("spiretokensurl"),
	}

	return nil
}
