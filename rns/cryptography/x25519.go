package cryptography

import (
	"crypto/ecdh"
	"crypto/rand"
	"errors"
	"time"
)

const (
	MinExecTime = 2 * time.Millisecond
	MaxExecTime = 500 * time.Millisecond
	DelayWindow = 10 * time.Second
	keySize     = 32
)

var (
	curve  = ecdh.X25519()
	tClear time.Time
	tMax   time.Duration
)

// ---------- утилиты ----------

func clampSecret(b []byte) {
	// _fix_secret
	b[0] &^= 7     // n &= ~7
	b[31] &^= 0x80 // n &= ~(128 << 8 * 31)
	b[31] |= 0x40  // n |= 64 << 8 * 31
}

// ---------- публичный ключ ----------

type PublicKey struct {
	x []byte // 32 bytes, little-endian
}

func FromPublicBytes(data []byte) (*PublicKey, error) {
	if len(data) != keySize {
		return nil, errors.New("x25519: public key must be 32 bytes")
	}
	out := make([]byte, keySize)
	copy(out, data)
	return &PublicKey{x: out}, nil
}

func (p *PublicKey) PublicBytes() []byte {
	out := make([]byte, len(p.x))
	copy(out, p.x)
	return out
}

// ---------- приватный ключ ----------

type PrivateKey struct {
	a []byte // 32 bytes, clamped
}

func Generate() (*PrivateKey, error) {
	b := make([]byte, keySize)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	clampSecret(b)
	return &PrivateKey{a: b}, nil
}

func FromPrivateBytes(data []byte) (*PrivateKey, error) {
	if len(data) != keySize {
		return nil, errors.New("x25519: private key must be 32 bytes")
	}
	b := make([]byte, keySize)
	copy(b, data)
	clampSecret(b)
	return &PrivateKey{a: b}, nil
}

func (k *PrivateKey) PrivateBytes() []byte {
	out := make([]byte, len(k.a))
	copy(out, k.a)
	return out
}

func (k *PrivateKey) PublicKey() (*PublicKey, error) {
	priv, err := curve.NewPrivateKey(k.a)
	if err != nil {
		return nil, err
	}
	pub := priv.PublicKey()
	return &PublicKey{x: pub.Bytes()}, nil
}

// Exchange — аналог X25519PrivateKey.exchange(), с выравниванием времени
func (k *PrivateKey) Exchange(peer interface{}) ([]byte, error) {
	var peerPub *PublicKey
	switch v := peer.(type) {
	case []byte:
		p, err := FromPublicBytes(v)
		if err != nil {
			return nil, err
		}
		peerPub = p
	case *PublicKey:
		peerPub = v
	default:
		return nil, errors.New("x25519: unsupported peer key type")
	}

	// при желании — как в curve25519(): fixBasePoint(peerPub.x)
	// но обычно публичный ключ уже корректный X25519

	start := time.Now()

	priv, err := curve.NewPrivateKey(k.a)
	if err != nil {
		return nil, err
	}
	pub, err := curve.NewPublicKey(peerPub.x)
	if err != nil {
		return nil, err
	}
	shared, err := priv.ECDH(pub)
	if err != nil {
		return nil, err
	}

	end := time.Now()
	duration := end.Sub(start)

	// логика T_CLEAR / T_MAX как в Python
	if tClear.IsZero() {
		tClear = end.Add(DelayWindow)
	}

	if end.After(tClear) {
		tClear = end.Add(DelayWindow)
		tMax = 0
	}

	if duration < tMax || duration < MinExecTime {
		target := start.Add(tMax)

		if target.After(start.Add(MaxExecTime)) {
			target = start.Add(MaxExecTime)
		}
		if target.Before(start.Add(MinExecTime)) {
			target = start.Add(MinExecTime)
		}

		if sleep := time.Until(target); sleep > 0 {
			_ = sleep // errors игнорируем как в Python
			time.Sleep(sleep)
		}
	} else if duration > tMax {
		tMax = duration
	}

	return shared, nil
}
