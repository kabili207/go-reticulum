package cryptography

import (
	"crypto/sha256"
	"hash"
)

// предвычисленные XOR-таблицы как в Python: trans_5C и trans_36
var (
	trans5C [256]byte
	trans36 [256]byte
)

func init() {
	for i := 0; i < 256; i++ {
		trans5C[i] = byte(i) ^ 0x5c
		trans36[i] = byte(i) ^ 0x36
	}
}

type HMAC struct {
	key        []byte // нормализованный и допадденный ключ
	blockSize  int    // эффективный block_size
	digestSize int    // размер дайджеста
	digestmod  func() hash.Hash
	data       []byte // все msg, полученные через Update
}

// New — аналог new(key, msg=None, digestmod=sha256)
func NewHMAC(key, msg []byte, digestmod func() hash.Hash) *HMAC {
	if digestmod == nil {
		digestmod = sha256.New
	}

	h0 := digestmod()
	blockSize := h0.BlockSize()
	if blockSize < 16 {
		blockSize = 64 // как в Python, fallback
	}

	// если ключ длиннее block_size — хешируем
	if len(key) > blockSize {
		h := digestmod()
		h.Write(key)
		key = h.Sum(nil)
	}

	// паддим нулями до block_size
	if len(key) < blockSize {
		k := make([]byte, blockSize)
		copy(k, key)
		key = k
	}

	h := &HMAC{
		key:        key,
		blockSize:  blockSize,
		digestSize: h0.Size(),
		digestmod:  digestmod,
		data:       nil,
	}

	if msg != nil {
		h.Update(msg)
	}
	return h
}

// Update — аналог update(msg)
func (h *HMAC) Update(msg []byte) {
	h.data = append(h.data, msg...)
}

// Copy — аналог copy(), полностью независимая копия
func (h *HMAC) Copy() *HMAC {
	kc := make([]byte, len(h.key))
	copy(kc, h.key)

	dc := make([]byte, len(h.data))
	copy(dc, h.data)

	return &HMAC{
		key:        kc,
		blockSize:  h.blockSize,
		digestSize: h.digestSize,
		digestmod:  h.digestmod,
		data:       dc,
	}
}

// compute() — внутренний расчёт HMAC, как digest() в Python-версии
func (h *HMAC) compute() []byte {
	// key ⊕ 0x36 и key ⊕ 0x5c через translate-таблицы
	kInner := make([]byte, len(h.key))
	kOuter := make([]byte, len(h.key))
	for i, b := range h.key {
		kInner[i] = trans36[b]
		kOuter[i] = trans5C[b]
	}

	inner := h.digestmod()
	inner.Write(kInner)
	inner.Write(h.data)
	innerSum := inner.Sum(nil)

	outer := h.digestmod()
	outer.Write(kOuter)
	outer.Write(innerSum)
	return outer.Sum(nil)
}

// Digest — аналог digest()
func (h *HMAC) Digest() []byte {
	sum := h.compute()
	out := make([]byte, len(sum))
	copy(out, sum)
	return out
}

// HexDigest — аналог hexdigest()
func (h *HMAC) HexDigest() string {
	sum := h.Digest()
	const hex = "0123456789abcdef"
	out := make([]byte, len(sum)*2)
	for i, b := range sum {
		out[i*2] = hex[b>>4]
		out[i*2+1] = hex[b&0x0f]
	}
	return string(out)
}

// New — точный аналог функции new(...) из Python
func New(key, msg []byte, digestmod func() hash.Hash) *HMAC {
	return NewHMAC(key, msg, digestmod)
}

// DigestFast — аналог функции digest(key, msg, digest)
func DigestFast(key, msg []byte, digestmod func() hash.Hash) []byte {
	h := NewHMAC(key, msg, digestmod)
	return h.Digest()
}
