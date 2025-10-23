//go:build dev
// +build dev

package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

func main() {
	// pastikan folder certs ada
	certDir := "certs"
	_ = os.MkdirAll(certDir, 0755)

	certFile := filepath.Join(certDir, "server.crt")
	keyFile := filepath.Join(certDir, "server.key")

	// generate jika belum ada
	if _, err := os.Stat(certFile); os.IsNotExist(err) {
		log.Println("Sertifikat belum ditemukan, membuat self-signed TLS...")
		if err := generateSelfSignedCert(certFile, keyFile); err != nil {
			log.Fatalf("Gagal membuat sertifikat: %v", err)
		}
	}

	r := chi.NewRouter()

	// Rate limit: 100 request per menit per IP
	r.Use(httprate.LimitByIP(100, time.Minute))

	// Simple route
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Server berjalan dengan HTTPS dan Rate Limit aktif"))
	})

	srv := &http.Server{
		Addr:         ":8443",
		Handler:      r,
		TLSConfig:    &tls.Config{MinVersion: tls.VersionTLS12},
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Server HTTPS berjalan di https://localhost:8443")
	log.Fatal(srv.ListenAndServeTLS(certFile, keyFile))
}

// generateSelfSignedCert membuat sertifikat TLS lokal untuk pengujian
func generateSelfSignedCert(certFile, keyFile string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(time.Now().Unix()),
		Subject: pkix.Name{
			Organization: []string{"Local Dev"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Sertifikat valid untuk localhost
	template.IPAddresses = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")}
	template.DNSNames = []string{"localhost"}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	certOut, err := os.Create(certFile)
	if err != nil {
		return err
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return err
	}

	keyOut, err := os.Create(keyFile)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(priv)}); err != nil {
		return err
	}

	return nil
}
