package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"net"
	"testing"
)

const seed = "0123456789abcdef0123456789abcdef"

func TestDerive(t *testing.T) {
	priv, _ := derive(seed, "auth", "priv")
	again, _ := derive(seed, "auth", "priv")
	pub, _ := derive(seed, "auth", "pub")
	other, _ := derive(seed+"x", "auth", "priv")
	if priv != again || priv == other {
		t.Fatal("priv must be deterministic per seed and differ across seeds")
	}
	b, _ := pem.Decode([]byte(priv))
	k, err := x509.ParsePKCS8PrivateKey(b.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	pb, _ := pem.Decode([]byte(pub))
	pk, err := x509.ParsePKIXPublicKey(pb.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if k.(*rsa.PrivateKey).N.BitLen() != 2048 {
		t.Fatal("want 2048-bit modulus")
	}
	h := sha256.Sum256([]byte("x"))
	sig, err := rsa.SignPSS(rand.Reader, k.(*rsa.PrivateKey), crypto.SHA256, h[:], nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := rsa.VerifyPSS(pk.(*rsa.PublicKey), crypto.SHA256, h[:], sig, nil); err != nil {
		t.Fatal("pub does not match priv:", err)
	}
	f, _ := derive(seed, "crypto", "fernet")
	if raw, err := base64.URLEncoding.DecodeString(f); err != nil || len(raw) != 32 {
		t.Fatal("fernet key must be 32 url-safe base64 bytes")
	}
	if tok, _ := derive(seed, "us", "token"); len(tok) != 64 {
		t.Fatal("token must be 64 hex chars")
	}
}

func TestWaitFor(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()
	waitFor("postgres://u:p@" + l.Addr().String() + "/db,postgres://other:1/db") // returns at once
}
