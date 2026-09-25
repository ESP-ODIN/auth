package testutil

import (
	"auth/internal/utils"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
)

// Signer creates an ephemeral RSA key in a test-owned temporary directory.
func Signer(t testing.TB) (*utils.JWTSigner, *rsa.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "private.pem")
	data := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	signer, err := utils.LoadJWTSigner(path, "test-key", "https://auth.odin.test", "odin-api")
	if err != nil {
		t.Fatal(err)
	}
	return signer, &key.PublicKey
}
