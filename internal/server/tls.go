package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
)

// TLSConfig holds the TLS certificate and key paths
type TLSConfig struct {
	CertPath string
	KeyPath  string
	Enabled  bool
}

// CertificateManager handles certificate generation and validation
type CertificateManager struct {
	certsDir string
	certFile string
	keyFile  string
}

// NewCertificateManager creates a new certificate manager
func NewCertificateManager(claudeDir string) *CertificateManager {
	certsDir := filepath.Join(claudeDir, "analytics", "certs")
	return &CertificateManager{
		certsDir: certsDir,
		certFile: filepath.Join(certsDir, "server.crt"),
		keyFile:  filepath.Join(certsDir, "server.key"),
	}
}

// EnsureCertificates checks if certificates exist and are valid, generates new ones if needed
func (cm *CertificateManager) EnsureCertificates() (*TLSConfig, error) {
	// Create certs directory if it doesn't exist
	if err := os.MkdirAll(cm.certsDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create certs directory: %w", err)
	}

	// Check if certificates exist
	certExists := fileExists(cm.certFile)
	keyExists := fileExists(cm.keyFile)

	if certExists && keyExists {
		// Validate existing certificate
		if valid, daysUntilExpiry := cm.validateCertificate(); valid {
			if daysUntilExpiry < 30 {
				pterm.Warning.Printf("TLS certificate expires in %d days. Consider regenerating.\n", daysUntilExpiry)
			}
			return &TLSConfig{
				CertPath: cm.certFile,
				KeyPath:  cm.keyFile,
				Enabled:  true,
			}, nil
		}
		pterm.Info.Println("Existing certificate is invalid or expired, generating new one...")
	}

	// Generate new certificates
	pterm.Info.Println("Generating self-signed TLS certificate...")
	if err := cm.generateSelfSignedCert(); err != nil {
		return nil, fmt.Errorf("failed to generate certificate: %w", err)
	}

	pterm.Success.Println("TLS certificate generated successfully")
	return &TLSConfig{
		CertPath: cm.certFile,
		KeyPath:  cm.keyFile,
		Enabled:  true,
	}, nil
}

// generateSelfSignedCert creates a new self-signed certificate
func (cm *CertificateManager) generateSelfSignedCert() error {
	// Generate private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("failed to generate private key: %w", err)
	}

	// Create certificate template
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valid for 1 year

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Claude Control Terminal"},
			CommonName:   "localhost",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
	}

	// Create certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %w", err)
	}

	// Write certificate to file
	certOut, err := os.OpenFile(cm.certFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open cert file for writing: %w", err)
	}
	defer certOut.Close()

	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write certificate: %w", err)
	}

	// Write private key to file
	keyOut, err := os.OpenFile(cm.keyFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open key file for writing: %w", err)
	}
	defer keyOut.Close()

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return fmt.Errorf("failed to marshal private key: %w", err)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("failed to write private key: %w", err)
	}

	return nil
}

// validateCertificate checks if the existing certificate is valid
func (cm *CertificateManager) validateCertificate() (bool, int) {
	certPEM, err := os.ReadFile(cm.certFile)
	if err != nil {
		return false, 0
	}

	block, _ := pem.Decode(certPEM)
	if block == nil {
		return false, 0
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return false, 0
	}

	// Check if certificate is expired or not yet valid
	now := time.Now()
	if now.Before(cert.NotBefore) || now.After(cert.NotAfter) {
		return false, 0
	}

	// Calculate days until expiry
	daysUntilExpiry := int(cert.NotAfter.Sub(now).Hours() / 24)

	return true, daysUntilExpiry
}

// GetTLSConfig returns the TLS configuration
func (cm *CertificateManager) GetTLSConfig() *TLSConfig {
	return &TLSConfig{
		CertPath: cm.certFile,
		KeyPath:  cm.keyFile,
		Enabled:  true,
	}
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
