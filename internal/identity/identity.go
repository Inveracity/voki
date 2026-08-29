package identity

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// Identity is a self-signed certificate and its key, used by both peers of a link.
type Identity struct {
	Certificate tls.Certificate
	Fingerprint string
}

// Ensure loads the identity from dir, generating it on first use (trust on first use).
func Ensure(dir string) (*Identity, error) {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, err
	}

	keyPath := filepath.Join(dir, "key.pem")
	certPath := filepath.Join(dir, "cert.pem")

	if cert, err := tls.LoadX509KeyPair(certPath, keyPath); err == nil {
		return fromCert(cert)
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, err
	}

	hostname, _ := os.Hostname()
	template := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "voki-" + hostname},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, err
	}

	if err := writePrivate(keyPath, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})); err != nil {
		return nil, err
	}
	if err := os.WriteFile(certPath, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0644); err != nil {
		return nil, err
	}

	return fromCert(tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key})
}

func fromCert(cert tls.Certificate) (*Identity, error) {
	return &Identity{Certificate: cert, Fingerprint: Fingerprint(cert.Certificate[0])}, nil
}

// Fingerprint is the hex sha256 of a certificate in DER form.
func Fingerprint(der []byte) string {
	sum := sha256.Sum256(der)
	return hex.EncodeToString(sum[:])
}

func writePrivate(path string, data []byte) error {
	if err := os.WriteFile(path, data, 0600); err != nil {
		return err
	}
	return nil
}

// Pins maps "host:port" to an agent fingerprint, like an SSH known_hosts file.
type Pins struct {
	path string
	Host map[string]string `json:"hosts"`
}

func LoadPins(path string) (*Pins, error) {
	p := &Pins{path: path, Host: map[string]string{}}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return p, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, p); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	return p, nil
}

func (p *Pins) Put(host, fingerprint string) error {
	p.Host[host] = fingerprint
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p.path, data, 0600)
}

// Check returns an error when host has no pin yet, or when the pin changed.
func (p *Pins) Check(host, fingerprint string) error {
	pinned, ok := p.Host[host]
	if !ok {
		return fmt.Errorf("no fingerprint pinned for %s, run 'voki bond %s' first", host, host)
	}
	if pinned != fingerprint {
		return fmt.Errorf("fingerprint changed for %s (was %s, now %s), possible man-in-the-middle", host, short(pinned), short(fingerprint))
	}
	return nil
}

func short(fp string) string {
	if len(fp) > 16 {
		return fp[:16] + "…"
	}
	return fp
}
