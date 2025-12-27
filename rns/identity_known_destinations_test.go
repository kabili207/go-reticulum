package rns

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"

	umsgpack "main/rns/vendor"
)

func pythonKnownDestinationsMsgpack(key16, packetHash, publicKey, appData []byte, seenAt float64) []byte {
	// msgpack encoding for: { <bin16 key>: [seenAt, packetHash, publicKey, appData] }
	//
	// fixmap(1)         0x81
	// bin8 + len        0xc4 0x10 + 16 bytes
	// fixarray(4)       0x94
	// float64           0xcb + 8
	// bin8 fields...
	buf := make([]byte, 0, 128)
	buf = append(buf, 0x81)
	buf = append(buf, 0xc4, byte(len(key16)))
	buf = append(buf, key16...)
	buf = append(buf, 0x94)
	buf = append(buf, 0xcb)
	var f [8]byte
	binary.BigEndian.PutUint64(f[:], math.Float64bits(seenAt))
	buf = append(buf, f[:]...)
	for _, b := range [][]byte{packetHash, publicKey, appData} {
		if len(b) < 256 {
			buf = append(buf, 0xc4, byte(len(b)))
		} else {
			buf = append(buf, 0xc5, byte(len(b)>>8), byte(len(b)))
		}
		buf = append(buf, b...)
	}
	return buf
}

func TestIdentityKnownDestinations_EncodeDecode_PythonCompatibleKeys(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte{0x01}, ReticulumTruncatedHashLength/8)
	pub := make([]byte, 0, 64)
	pub = append(pub, bytes.Repeat([]byte{0x02}, 32)...)
	pub = append(pub, bytes.Repeat([]byte{0x03}, 32)...)

	in := map[string]*knownDestinationEntry{
		string(key): {
			SeenAt:     123.0,
			PacketHash: []byte("ph"),
			PublicKey:  pub,
			AppData:    []byte("ad"),
		},
	}

	enc, err := encodeKnownDestinations(in)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	out, err := decodeKnownDestinations(enc)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	got := out[string(key)]
	if got == nil {
		t.Fatalf("missing decoded entry")
	}
	if got.SeenAt != 123.0 {
		t.Fatalf("unexpected SeenAt: %v", got.SeenAt)
	}
	if !bytes.Equal(got.PublicKey, pub) {
		t.Fatalf("unexpected PublicKey")
	}
	if !bytes.Equal(got.AppData, []byte("ad")) {
		t.Fatalf("unexpected AppData")
	}
}

func TestIdentityKnownDestinations_Decode_AcceptsStringKeys(t *testing.T) {
	t.Parallel()

	keyBytes := bytes.Repeat([]byte{0x01}, ReticulumTruncatedHashLength/8)
	keyStr := string(keyBytes)
	pub := make([]byte, 0, 64)
	pub = append(pub, bytes.Repeat([]byte{0x02}, 32)...)
	pub = append(pub, bytes.Repeat([]byte{0x03}, 32)...)

	legacy := map[any]any{
		keyStr: []any{1.0, []byte("ph"), pub, []byte("ad")},
	}
	raw, err := umsgpack.Packb(legacy)
	if err != nil {
		t.Fatalf("pack: %v", err)
	}

	out, err := decodeKnownDestinations(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out[keyStr] == nil {
		t.Fatalf("missing decoded entry")
	}
}

func TestIdentityKnownDestinations_Decode_PythonBytesKeys(t *testing.T) {
	t.Parallel()

	key := bytes.Repeat([]byte{0x01}, ReticulumTruncatedHashLength/8)
	pub := make([]byte, 0, 64)
	pub = append(pub, bytes.Repeat([]byte{0x02}, 32)...)
	pub = append(pub, bytes.Repeat([]byte{0x03}, 32)...)

	raw := pythonKnownDestinationsMsgpack(key, []byte("ph"), pub, []byte("ad"), 1.0)
	out, err := decodeKnownDestinations(raw)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out[string(key)] == nil {
		t.Fatalf("missing decoded entry")
	}
}
