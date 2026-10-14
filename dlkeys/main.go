// dlkeys derives every DataLens key and token from one seed, then execs the service.
//
// Railway template variables can generate random strings but not RSA keypairs, and the
// DataLens services need matching keypairs (auth signs user JWTs with PS256, UI and the
// BI APIs sign US master tokens). Each service gets the same DATALENS_KEY_SEED and derives
// what it needs, so the keys are unique per deploy and identical across services.
//
// Usage: DL_KEYS="VAR=label:kind ..." dlkeys <command> [args...]
// kinds: priv / pub (RSA-2048 PKCS#8 / PKIX PEM), token (64 hex chars), fernet (Fernet key).
package main

import (
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

// stream is a SHA-256 counter-mode byte stream keyed by seed and label.
type stream struct {
	key [32]byte
	n   uint64
}

func newStream(seed, label string) *stream {
	return &stream{key: sha256.Sum256([]byte(seed + "\x00" + label))}
}

func (s *stream) read(n int) []byte {
	out := make([]byte, 0, n+32)
	for len(out) < n {
		var b [40]byte
		copy(b[:], s.key[:])
		binary.BigEndian.PutUint64(b[32:], s.n)
		s.n++
		h := sha256.Sum256(b[:])
		out = append(out, h[:]...)
	}
	return out[:n]
}

var e = big.NewInt(65537)

func prime(s *stream) *big.Int {
	one := big.NewInt(1)
	for {
		b := s.read(128)
		b[0] |= 0xC0 // top two bits set: p*q is a full 2048 bits
		b[127] |= 1
		p := new(big.Int).SetBytes(b)
		if p.ProbablyPrime(32) && new(big.Int).GCD(nil, nil, new(big.Int).Sub(p, one), e).Cmp(one) == 0 {
			return p
		}
	}
}

func rsaKey(seed, label string) *rsa.PrivateKey {
	s := newStream(seed, "rsa:"+label)
	p, q := prime(s), prime(s)
	for p.Cmp(q) == 0 {
		q = prime(s)
	}
	one := big.NewInt(1)
	phi := new(big.Int).Mul(new(big.Int).Sub(p, one), new(big.Int).Sub(q, one))
	k := &rsa.PrivateKey{
		PublicKey: rsa.PublicKey{N: new(big.Int).Mul(p, q), E: 65537},
		D:         new(big.Int).ModInverse(e, phi),
		Primes:    []*big.Int{p, q},
	}
	k.Precompute()
	if err := k.Validate(); err != nil {
		panic(err)
	}
	return k
}

func derive(seed, label, kind string) (string, error) {
	switch kind {
	case "priv":
		der, err := x509.MarshalPKCS8PrivateKey(rsaKey(seed, label))
		if err != nil {
			return "", err
		}
		return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der})), nil
	case "pub":
		der, err := x509.MarshalPKIXPublicKey(&rsaKey(seed, label).PublicKey)
		if err != nil {
			return "", err
		}
		return string(pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})), nil
	case "token":
		return hex.EncodeToString(newStream(seed, "token:"+label).read(32)), nil
	case "fernet":
		return base64.URLEncoding.EncodeToString(newStream(seed, "fernet:"+label).read(32)), nil
	}
	return "", fmt.Errorf("unknown kind %q", kind)
}

func die(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "dlkeys: "+f+"\n", a...)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		die("usage: DL_KEYS=\"VAR=label:kind ...\" dlkeys <command> [args...]")
	}
	seed := os.Getenv("DATALENS_KEY_SEED")
	if len(seed) < 32 {
		die("DATALENS_KEY_SEED must be set to at least 32 characters")
	}
	env := os.Environ()
	set := func(name, v string) {
		for i, kv := range env {
			if strings.HasPrefix(kv, name+"=") {
				env = append(env[:i], env[i+1:]...)
				break
			}
		}
		env = append(env, name+"="+v)
	}
	for _, spec := range strings.Fields(os.Getenv("DL_KEYS")) {
		name, src, ok := strings.Cut(spec, "=")
		label, kind, ok2 := strings.Cut(src, ":")
		if !ok || !ok2 {
			die("bad DL_KEYS entry %q", spec)
		}
		v, err := derive(seed, label, kind)
		if err != nil {
			die("%s: %v", name, err)
		}
		set(name, v)
	}
	path, err := exec.LookPath(os.Args[1])
	if err != nil {
		die("%v", err)
	}
	die("exec %s: %v", path, syscall.Exec(path, os.Args[1:], env))
}
