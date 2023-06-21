// Package test provides functions for use when testing the TPM Provisioner.
package simulateTPM

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/pem"
	"io"
	"math/big"
	"time"

	"github.com/google/go-tpm-tools/client"
	"github.com/google/go-tpm/legacy/tpm2"
	"github.com/google/go-tpm/tpmutil"
)

// GenerateCA creates a TPM Provisioner test CA.
func GenerateCA(cname string) (*x509.Certificate, *rsa.PrivateKey, *bytes.Buffer, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{"TPM-Provisioner"},
			Country:       []string{"TPM"},
			Province:      []string{"TPM"},
			Locality:      []string{"TPM-Provisioner"},
			StreetAddress: []string{"TPM-Provisioner"},
			CommonName:    cname,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		UnknownExtKeyUsage:    []asn1.ObjectIdentifier{{2, 23, 133, 8, 1}},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return &x509.Certificate{}, nil, nil, err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return &x509.Certificate{}, nil, nil, err
	}

	caPEM := new(bytes.Buffer)

	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return &x509.Certificate{}, nil, nil, err
	}

	caPrivKeyPEM := new(bytes.Buffer)

	err = pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	if err != nil {
		return &x509.Certificate{}, nil, nil, err
	}

	return ca, caPrivKey, caPEM, nil
}

// GenerateEK creates an EK Certificate for use with the TPM simulator.
func GenerateEK(ca *x509.Certificate, caKey *rsa.PrivateKey, ekPubKey *rsa.PublicKey) ([]byte, error) {
	// This long string creates a parsable DirName for use with the EK's
	// SubjectAltName
	tpmDirName, err := hex.DecodeString("3040313E301406056781050201130B69643A344535343433303030100605678105020213074E504354373578301406056781050203130B69643A3030303730303032")
	if err != nil {
		return nil, err
	}

	rawDirValues := []asn1.RawValue{{Tag: 4, Class: asn1.ClassContextSpecific, Bytes: tpmDirName}}

	dirName, err := asn1.Marshal(rawDirValues)
	if err != nil {
		return nil, err
	}

	extSubjectAltName := pkix.Extension{}
	extSubjectAltName.Id = asn1.ObjectIdentifier{2, 5, 29, 17}
	extSubjectAltName.Critical = true
	extSubjectAltName.Value = dirName

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: "TPM-Provisioner Test EK",
		},
		IsCA:                  false,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		ExtraExtensions:       []pkix.Extension{extSubjectAltName},
		KeyUsage:              x509.KeyUsageKeyEncipherment,
		UnknownExtKeyUsage:    []asn1.ObjectIdentifier{{2, 23, 133, 8, 1}},
		BasicConstraintsValid: true,
		PolicyIdentifiers:     []asn1.ObjectIdentifier{{2, 5, 29, 32, 0}},
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, ekPubKey, caKey)
	if err != nil {
		return nil, err
	}

	certPEM := new(bytes.Buffer)

	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, err
	}

	return certBytes, nil
}

// CreateEK Creates an EK Certificate and loads it into the simulated TPM.
func CreateEK(rwc io.ReadWriter) ([]byte, error) {
	// Get EK Public Key
	ekPubKey, err := GetPublicEK(rwc)
	if err != nil {
		return nil, err
	}

	// Create CA Certificate
	ca, caKey, caPEM, err := GenerateCA("TPM Manufacturer Test")
	if err != nil {
		return nil, err
	}

	// Create TPM EK Certificate
	ekCRT, err := GenerateEK(ca, caKey, ekPubKey)
	if err != nil {
		return nil, err
	}

	// Load EK Certificate into NVRam
	err = LoadEK(rwc, ekCRT)
	if err != nil {
		return nil, err
	}

	return caPEM.Bytes(), nil
}

// LoadEK loads the EK into the simulated TPM.
func LoadEK(rwc io.ReadWriter, cert []byte) error {
	attr := tpm2.NVAttr(0x42072001)
	authHandle := tpmutil.Handle(0x4000000C)
	idx := tpmutil.Handle(0x1c00002)

	err := tpm2.NVDefineSpace(rwc, authHandle, idx, "", "", nil, attr, uint16(len(cert)))
	if err != nil {
		return err
	}

	for i := 0; i <= len(cert); i += 1024 {
		var maxLen int
		if len(cert)-i >= 1024 {
			maxLen += i + 1024
		} else {
			maxLen = len(cert)
		}

		err = tpm2.NVWrite(rwc, authHandle, idx, "", cert[i:maxLen], uint16(i))
		if err != nil {
			return err
		}
	}

	return nil
}

// GetPublicEK returns the EK Certificate stores in the TPM.
func GetPublicEK(rwc io.ReadWriter) (*rsa.PublicKey, error) {
	// Obtain the Public EK
	ek, err := client.EndorsementKeyRSA(rwc)
	if err != nil {
		return nil, err
	}

	ekPublicKey := ek.PublicKey().(*rsa.PublicKey)

	return ekPublicKey, nil
}
