package rns

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"sort"
	"testing"
	"time"
)

func mustHexT(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("invalid hex: %v", err)
	}
	return b
}

// Ported from `tests/identity.py` to ensure byte-for-byte compatibility with the Python implementation.
func TestIdentity_FromBytes_HashAndPrivateKey(t *testing.T) {
	maybeParallel(t)

	fixedKeys := [][2]string{
		{"f8953ffaf607627e615603ff1530c82c434cf87c07179dd7689ea776f30b964cfb7ba6164af00c5111a45e69e57d885e1285f8dbfe3a21e95ae17cf676b0f8b7", "650b5d76b6bec0390d1f8cfca5bd33f9"},
		{"d85d036245436a3c33d3228affae06721f8203bc364ee0ee7556368ac62add650ebf8f926abf628da9d92baaa12db89bd6516ee92ec29765f3afafcb8622d697", "1469e89450c361b253aefb0c606b6111"},
		{"8893e2bfd30fc08455997caf7abb7a6341716768dbbf9a91cc1455bd7eeaf74cdc10ec72a4d4179696040bac620ee97ebc861e2443e5270537ae766d91b58181", "e5fe93ee4acba095b3b9b6541515ed3e"},
		{"b82c7a4f047561d974de7e38538281d7f005d3663615f30d9663bad35a716063c931672cd452175d55bcdd70bb7aa35a9706872a97963dc52029938ea7341b39", "1333b911fa8ebb16726996adbe3c6262"},
		{"08bb35f92b06a0832991165a0d9b4fd91af7b7765ce4572aa6222070b11b767092b61b0fd18b3a59cae6deb9db6d4bfb1c7fcfe076cfd66eea7ddd5f877543b9", "d13712efc45ef87674fb5ac26c37c912"},
	}

	for _, entry := range fixedKeys {
		keyHex, idHashHex := entry[0], entry[1]
		id, err := IdentityFromBytes(mustHexT(t, keyHex))
		if err != nil {
			t.Fatalf("IdentityFromBytes failed: %v", err)
		}
		if got, want := hex.EncodeToString(id.Hash), idHashHex; got != want {
			t.Fatalf("hash mismatch: got %s want %s", got, want)
		}
		if got, want := hex.EncodeToString(id.GetPrivateKey()), keyHex; got != want {
			t.Fatalf("private key mismatch: got %s want %s", got, want)
		}
	}
}

func TestIdentity_Sign_KnownVector(t *testing.T) {
	maybeParallel(t)

	// From `tests/identity.py`
	privateKeyHex := "f8953ffaf607627e615603ff1530c82c434cf87c07179dd7689ea776f30b964cfb7ba6164af00c5111a45e69e57d885e1285f8dbfe3a21e95ae17cf676b0f8b7"
	message := []byte("e51a008b8b8ba855993d8892a40daad84a6fb69a7138e1b5f69b427fe03449826ab6ccb81f0d72b4725e8d55c814d3e8e151b495cf5b59702f197ec366d935ad04a98ca519d6964f96ea09910b020351d1cdff3befbad323a2a28a6ec7ced4d0d67f02c525f93b321d9b076d704408475bd2d123cd51916f7e49039246ac56add37ef87e32d7f9853ac44a7f77d26fedc83e4e67a45742b751c2599309f5eda6efa0dafd957f61af1f0e86c4d6c5052e0e5fa577db99846f2b7a0204c31cef4013ca51cb307506c9209fd18d0195a7c9ae628af1a1d9ee7a4cf30037ed190a9fdcaa4ce5bb7bea19803cb5b5cea8c21fdb98d8f73ff5aaad87f5f6c3b7bcfe8974e5b063cc1113d77b9e96bec1c9d10ed37b780c3f7349a34092bb3968daeced40eb0b5130c0d11595e30b9671896385d04289d067f671599386536eed8430a72e186fb95023d5ac5dd442443bfabfe13a84a38d060af73bf20f921f38a768672fdbcb1dfece7458166e2e15948d6b4fa81f42db48747d283c670f576a0b410b31a70d2594823d0e29135a488cb0408c9e5bc1e197ff99aef471924231ccc8e3eddc82dbcea4801f14c5fc7a389a26a52cc93cfe0770953ef595ff410b7033a6ed5c975dd922b3f48f9dffcfb412eeed5758f3aa51de7eb47cd2cb")
	wantSigHex := "3020ef58f861591826a61c3d2d4a25b949cdb3094085ba6b1177a6f2a05f3cdd24d1095d6fdd078f0b2826e80b261c93c1ff97fbfd4857f25706d57dd073590c"

	id, err := IdentityFromBytes(mustHexT(t, privateKeyHex))
	if err != nil {
		t.Fatalf("IdentityFromBytes failed: %v", err)
	}
	sig, err := id.Sign(message)
	if err != nil {
		t.Fatalf("Sign failed: %v", err)
	}
	if got, want := hex.EncodeToString(sig), wantSigHex; got != want {
		t.Fatalf("signature mismatch: got %s want %s", got, want)
	}
}

func TestIdentity_Decrypt_KnownToken(t *testing.T) {
	maybeParallel(t)

	// From `tests/identity.py`
	privateKeyHex := "f8953ffaf607627e615603ff1530c82c434cf87c07179dd7689ea776f30b964cfb7ba6164af00c5111a45e69e57d885e1285f8dbfe3a21e95ae17cf676b0f8b7"
	tokenHex := "e37705f9b432d3711acf028678b0b9d37fdf7e00a3b47c95251aad61447df2620b5b9978783c3d9f2fb762e68c8b57c554928fb70dd79c1033ce5865f91761aad3e992790f63456092cb69b7b045f539147f7ba10d480e300f193576ae2d75a7884809b76bd17e05a735383305c0aa5621395bbf51e8cc66c1c536f339f2bea600f08f8f9a76564b2522cd904b6c2b6e553ec3d4df718ae70434c734297b313539338d184d2c64a9c4ddbc9b9a4947d0b45f5a274f65ae9f6bb203562fd5cede6abd3c615b699156e08fa33b841647a0"
	wantPlaintextHex := "71884a271ead43558fcf1e331c5aebcd43498f16da16f8056b0893ce6b15d521eaa4f31639cd34da1b57995944076c4f14f300f2d2612111d21a3429a9966ac1da68545c00c7887d8b26f6c1ab9defa020b9519849ca41b7904199882802b6542771df85144a79890289d3c02daef6c26652c5ce9de231a2"

	id, err := IdentityFromBytes(mustHexT(t, privateKeyHex))
	if err != nil {
		t.Fatalf("IdentityFromBytes failed: %v", err)
	}
	plaintext, err := id.Decrypt(mustHexT(t, tokenHex), nil, false)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}
	if got, want := hex.EncodeToString(plaintext), wantPlaintextHex; got != want {
		t.Fatalf("plaintext mismatch: got %s want %s", got, want)
	}
}

func TestIdentity_SignAndValidate_RandomMessages(t *testing.T) {
	maybeParallel(t)

	rounds := 200
	if testing.Short() {
		rounds = 50
	}

	for i := 0; i < rounds; i++ {
		msg := make([]byte, 512)
		if _, err := rand.Read(msg); err != nil {
			t.Fatalf("rand.Read: %v", err)
		}

		id1, err := NewIdentity()
		if err != nil {
			t.Fatalf("NewIdentity: %v", err)
		}
		id2 := &Identity{}
		if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
			t.Fatalf("LoadPublicKey: %v", err)
		}

		sig, err := id1.Sign(msg)
		if err != nil {
			t.Fatalf("Sign: %v", err)
		}
		if ok := id2.Validate(sig, msg); !ok {
			t.Fatalf("Validate failed on round %d", i)
		}
	}
}

func TestIdentity_SignAndValidate_AllZeroAndAllFF(t *testing.T) {
	maybeParallel(t)

	rounds := 200
	if testing.Short() {
		rounds = 50
	}

	cases := []struct {
		name string
		b    byte
	}{
		{name: "all_zero", b: 0x00},
		{name: "all_ff", b: 0xFF},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg := make([]byte, 512)
			for i := range msg {
				msg[i] = tc.b
			}

			for i := 0; i < rounds; i++ {
				id1, err := NewIdentity()
				if err != nil {
					t.Fatalf("NewIdentity: %v", err)
				}
				id2 := &Identity{}
				if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
					t.Fatalf("LoadPublicKey: %v", err)
				}

				sig, err := id1.Sign(msg)
				if err != nil {
					t.Fatalf("Sign: %v", err)
				}
				if ok := id2.Validate(sig, msg); !ok {
					t.Fatalf("Validate failed on round %d", i)
				}
			}
		})
	}
}

func TestIdentity_EncryptDecrypt_RandomSmallChunks(t *testing.T) {
	maybeParallel(t)

	rounds := 200
	if testing.Short() {
		rounds = 50
	}

	// Mirrors `tests/identity.py` intent: varying sizes around MTU/2..MTU.
	for i := 1; i <= rounds; i++ {
		mlen := (i % (MTU / 2)) + (MTU / 2)
		msg := make([]byte, mlen)
		if _, err := rand.Read(msg); err != nil {
			t.Fatalf("rand.Read: %v", err)
		}

		id1, err := NewIdentity()
		if err != nil {
			t.Fatalf("NewIdentity: %v", err)
		}
		id2, err := NewIdentity()
		if err != nil {
			t.Fatalf("NewIdentity: %v", err)
		}
		// id2 should encrypt to id1
		if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
			t.Fatalf("LoadPublicKey: %v", err)
		}

		token, err := id2.Encrypt(msg, nil)
		if err != nil {
			t.Fatalf("Encrypt: %v", err)
		}
		plain, err := id1.Decrypt(token, nil, false)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}
		if hex.EncodeToString(plain) != hex.EncodeToString(msg) {
			t.Fatalf("decrypt mismatch on round %d", i)
		}
	}
}

func TestIdentity_EncryptDecrypt_LargeChunk(t *testing.T) {
	maybeParallel(t)

	mlen := 1024 * 1024 // keep test runtime reasonable
	if !testing.Short() {
		mlen = 2 * 1024 * 1024
	}

	msg := make([]byte, mlen)
	if _, err := rand.Read(msg); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	id1, err := NewIdentity()
	if err != nil {
		t.Fatalf("NewIdentity: %v", err)
	}
	id2 := &Identity{}
	if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
		t.Fatalf("LoadPublicKey: %v", err)
	}

	token, err := id2.Encrypt(msg, nil)
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	plain, err := id1.Decrypt(token, nil, false)
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if hex.EncodeToString(plain) != hex.EncodeToString(msg) {
		t.Fatalf("decrypt mismatch for large chunk")
	}
}

type timingStats struct {
	minMs  float64
	maxMs  float64
	meanMs float64
	medMs  float64
	mdevMs float64
	mpct   float64
}

func computeTimingStats(samples []time.Duration) timingStats {
	if len(samples) == 0 {
		return timingStats{}
	}
	ms := make([]float64, 0, len(samples))
	var sum float64
	for _, d := range samples {
		v := float64(d) / float64(time.Millisecond)
		ms = append(ms, v)
		sum += v
	}
	sort.Float64s(ms)

	minV := ms[0]
	maxV := ms[len(ms)-1]
	meanV := sum / float64(len(ms))
	medV := ms[len(ms)/2]
	if len(ms)%2 == 0 {
		medV = (ms[len(ms)/2-1] + ms[len(ms)/2]) / 2
	}
	mdevV := maxV - minV
	mpct := 0.0
	if medV > 0 {
		mpct = (maxV / medV) * 100
	}

	return timingStats{
		minMs:  minV,
		maxMs:  maxV,
		meanMs: meanV,
		medMs:  medV,
		mdevMs: mdevV,
		mpct:   mpct,
	}
}

func sizeStr(num float64, suffix string) string {
	units := []string{"", "K", "M", "G", "T", "P", "E", "Z"}
	lastUnit := "Y"
	if suffix == "b" {
		num *= 8
	}
	for _, unit := range units {
		if math.Abs(num) < 1000.0 {
			if unit == "" {
				return fmt.Sprintf("%.0f %s%s", num, unit, suffix)
			}
			return fmt.Sprintf("%.2f %s%s", num, unit, suffix)
		}
		num /= 1000.0
	}
	return fmt.Sprintf("%.2f%s%s", num, lastUnit, suffix)
}

// Timing/jitter tests from `tests/identity.py` can't be made identical across runtimes.
// We mirror their purpose: collect distributions and report min/avg/med/max/mdev + max deviation from median.
// Optionally enforce a very loose limit via `RNS_TIMING_ASSERT=1`.
func TestIdentity_TimingReport(t *testing.T) {
	if testing.Short() {
		t.Skip("timing report skipped in -short")
	}

	rounds := 2000
	switch os.Getenv("RNS_TIMING_ROUNDS") {
	case "500":
		rounds = 500
	case "20000":
		rounds = 20000
	}

	run := func(name string, msgFn func() []byte, op func(id *Identity, msg []byte) error) timingStats {
		t.Helper()
		id, err := NewIdentity()
		if err != nil {
			t.Fatalf("NewIdentity: %v", err)
		}
		samples := make([]time.Duration, 0, rounds)
		for i := 0; i < rounds; i++ {
			msg := msgFn()
			start := time.Now()
			if err := op(id, msg); err != nil {
				t.Fatalf("%s op failed: %v", name, err)
			}
			samples = append(samples, time.Since(start))
		}
		return computeTimingStats(samples)
	}

	randomMsg := func() []byte {
		msg := make([]byte, 512)
		_, _ = rand.Read(msg)
		return msg
	}
	allFF := func() []byte {
		msg := make([]byte, 512)
		for i := range msg {
			msg[i] = 0xFF
		}
		return msg
	}
	all00 := func() []byte {
		return make([]byte, 512)
	}

	signOp := func(id *Identity, msg []byte) error {
		_, err := id.Sign(msg)
		return err
	}

	t.Run("SignRandomMessages", func(t *testing.T) {
		s := run("sign_random", randomMsg, signOp)
		fmt.Println("")
		fmt.Println("Random messages:")
		fmt.Printf("  Signature timing min/avg/med/max/mdev: %s/%s/%s/%s/%s\n",
			fmt.Sprintf("%.3f", s.minMs),
			fmt.Sprintf("%.3f", s.meanMs),
			fmt.Sprintf("%.3f", s.medMs),
			fmt.Sprintf("%.3f", s.maxMs),
			fmt.Sprintf("%.3f", s.mdevMs),
		)
		fmt.Printf("  Max deviation from median: %s%%\n", fmt.Sprintf("%.1f", s.mpct))
		fmt.Println("")
		if os.Getenv("RNS_TIMING_ASSERT") == "1" && !(s.mpct < 10000 || math.IsInf(s.mpct, 0) || math.IsNaN(s.mpct)) {
			t.Fatalf("timing jitter too high: %.1f%%", s.mpct)
		}
	})

	t.Run("SignAllFF", func(t *testing.T) {
		s := run("sign_all_ff", allFF, signOp)
		fmt.Println("All 0xff messages:")
		fmt.Printf("  Signature timing min/avg/med/max/mdev: %s/%s/%s/%s/%s\n",
			fmt.Sprintf("%.3f", s.minMs),
			fmt.Sprintf("%.3f", s.meanMs),
			fmt.Sprintf("%.3f", s.medMs),
			fmt.Sprintf("%.3f", s.maxMs),
			fmt.Sprintf("%.3f", s.mdevMs),
		)
		fmt.Printf("  Max deviation from median: %s%%\n", fmt.Sprintf("%.1f", s.mpct))
		fmt.Println("")
	})

	t.Run("SignAll00", func(t *testing.T) {
		s := run("sign_all_00", all00, signOp)
		fmt.Println("All 0x00 messages:")
		fmt.Printf("  Signature timing min/avg/med/max/mdev: %s/%s/%s/%s/%s\n",
			fmt.Sprintf("%.3f", s.minMs),
			fmt.Sprintf("%.3f", s.meanMs),
			fmt.Sprintf("%.3f", s.medMs),
			fmt.Sprintf("%.3f", s.maxMs),
			fmt.Sprintf("%.3f", s.mdevMs),
		)
		fmt.Printf("  Max deviation from median: %s%%\n", fmt.Sprintf("%.1f", s.mpct))
		fmt.Println("")
	})

	// Throughput reports (mirroring python output). These measure sign+validate together.
	t.Run("SignValidateThroughput", func(t *testing.T) {
		rounds := 500
		if testing.Short() {
			rounds = 100
		}

		var totalBytes float64
		var totalTime float64

		for i := 1; i < rounds; i++ {
			mlen := (i % (MTU / 2)) + (MTU / 2)
			msg := make([]byte, mlen)
			if _, err := rand.Read(msg); err != nil {
				t.Fatalf("rand.Read: %v", err)
			}
			totalBytes += float64(mlen)

			id1, err := NewIdentity()
			if err != nil {
				t.Fatalf("NewIdentity: %v", err)
			}
			id2 := &Identity{}
			if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
				t.Fatalf("LoadPublicKey: %v", err)
			}

			start := time.Now()
			sig, err := id1.Sign(msg)
			if err != nil {
				t.Fatalf("Sign: %v", err)
			}
			if ok := id2.Validate(sig, msg); !ok {
				t.Fatalf("Validate failed")
			}
			totalTime += time.Since(start).Seconds()
		}

		fmt.Printf("Sign/validate chunks < MTU: %sps\n", sizeStr(totalBytes/totalTime, "b"))

		totalBytes = 0
		totalTime = 0
		mlen := 16 * 1024
		for i := 1; i < rounds; i++ {
			msg := make([]byte, mlen)
			if _, err := rand.Read(msg); err != nil {
				t.Fatalf("rand.Read: %v", err)
			}
			totalBytes += float64(mlen)

			id1, err := NewIdentity()
			if err != nil {
				t.Fatalf("NewIdentity: %v", err)
			}
			id2 := &Identity{}
			if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
				t.Fatalf("LoadPublicKey: %v", err)
			}

			start := time.Now()
			sig, err := id1.Sign(msg)
			if err != nil {
				t.Fatalf("Sign: %v", err)
			}
			if ok := id2.Validate(sig, msg); !ok {
				t.Fatalf("Validate failed")
			}
			totalTime += time.Since(start).Seconds()
		}

		fmt.Printf("Sign/validate 16KB chunks: %sps\n", sizeStr(totalBytes/totalTime, "b"))
	})

	// Encrypt/decrypt throughput + timing stats (mirroring python output shape).
	t.Run("EncryptDecryptReport", func(t *testing.T) {
		rounds := 500
		if testing.Short() {
			rounds = 100
		}

		fmt.Println("")
		fmt.Println("Testing random small chunk encrypt/decrypt")

		var totalBytes float64
		var eTotal, dTotal float64
		eTimes := make([]time.Duration, 0, rounds)
		dTimes := make([]time.Duration, 0, rounds)

		for i := 1; i < rounds; i++ {
			mlen := (i % (MTU / 2)) + (MTU / 2)
			msg := make([]byte, mlen)
			if _, err := rand.Read(msg); err != nil {
				t.Fatalf("rand.Read: %v", err)
			}
			totalBytes += float64(mlen)

			id1, err := NewIdentity()
			if err != nil {
				t.Fatalf("NewIdentity: %v", err)
			}
			id2 := &Identity{}
			if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
				t.Fatalf("LoadPublicKey: %v", err)
			}

			eStart := time.Now()
			token, err := id2.Encrypt(msg, nil)
			if err != nil {
				t.Fatalf("Encrypt: %v", err)
			}
			eDur := time.Since(eStart)
			eTotal += eDur.Seconds()
			eTimes = append(eTimes, eDur)

			dStart := time.Now()
			plain, err := id1.Decrypt(token, nil, false)
			if err != nil {
				t.Fatalf("Decrypt: %v", err)
			}
			dDur := time.Since(dStart)
			dTotal += dDur.Seconds()
			dTimes = append(dTimes, dDur)

			if hex.EncodeToString(plain) != hex.EncodeToString(msg) {
				t.Fatalf("decrypt mismatch")
			}
		}

		es := computeTimingStats(eTimes)
		ds := computeTimingStats(dTimes)

		fmt.Printf("  Encrypt chunks < MTU: %sps\n", sizeStr(totalBytes/eTotal, "b"))
		fmt.Printf("    Encryption timing min/avg/med/max/mdev: %s/%s/%s/%s/%s\n",
			fmt.Sprintf("%.3f", es.minMs), fmt.Sprintf("%.3f", es.meanMs), fmt.Sprintf("%.3f", es.medMs), fmt.Sprintf("%.3f", es.maxMs), fmt.Sprintf("%.3f", es.mdevMs))
		fmt.Printf("    Max deviation from median: %s%%\n", fmt.Sprintf("%.1f", es.mpct))
		fmt.Println("")

		fmt.Printf("  Decrypt chunks < MTU: %sps\n", sizeStr(totalBytes/dTotal, "b"))
		fmt.Printf("    Decryption timing min/avg/med/max/mdev: %s/%s/%s/%s/%s\n",
			fmt.Sprintf("%.3f", ds.minMs), fmt.Sprintf("%.3f", ds.meanMs), fmt.Sprintf("%.3f", ds.medMs), fmt.Sprintf("%.3f", ds.maxMs), fmt.Sprintf("%.3f", ds.mdevMs))
		fmt.Printf("    Max deviation from median: %s%%\n", fmt.Sprintf("%.1f", ds.mpct))
		fmt.Println("")

		fmt.Println("Testing large chunk encrypt/decrypt")
		mlen := 8 * 1000 * 1000
		if testing.Short() {
			mlen = 1 * 1000 * 1000
		}

		msg := make([]byte, mlen)
		if _, err := rand.Read(msg); err != nil {
			t.Fatalf("rand.Read: %v", err)
		}

		id1, err := NewIdentity()
		if err != nil {
			t.Fatalf("NewIdentity: %v", err)
		}
		id2 := &Identity{}
		if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
			t.Fatalf("LoadPublicKey: %v", err)
		}

		eStart := time.Now()
		token, err := id2.Encrypt(msg, nil)
		if err != nil {
			t.Fatalf("Encrypt: %v", err)
		}
		eDur := time.Since(eStart)

		dStart := time.Now()
		plain, err := id1.Decrypt(token, nil, false)
		if err != nil {
			t.Fatalf("Decrypt: %v", err)
		}
		dDur := time.Since(dStart)
		if hex.EncodeToString(plain) != hex.EncodeToString(msg) {
			t.Fatalf("decrypt mismatch")
		}

		fmt.Printf("  Encrypt %s chunks: %sps\n", sizeStr(float64(mlen), "B"), sizeStr(float64(mlen)/eDur.Seconds(), "b"))
		fmt.Printf("  Decrypt %s chunks: %sps\n", sizeStr(float64(mlen), "B"), sizeStr(float64(mlen)/dDur.Seconds(), "b"))
	})
}

func BenchmarkIdentity_Sign_512B(b *testing.B) {
	id, err := NewIdentity()
	if err != nil {
		b.Fatalf("NewIdentity: %v", err)
	}
	msg := make([]byte, 512)
	if _, err := rand.Read(msg); err != nil {
		b.Fatalf("rand.Read: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := id.Sign(msg); err != nil {
			b.Fatalf("Sign: %v", err)
		}
	}
}

func BenchmarkIdentity_EncryptDecrypt_ChunkAroundMTU(b *testing.B) {
	id1, err := NewIdentity()
	if err != nil {
		b.Fatalf("NewIdentity: %v", err)
	}
	id2 := &Identity{}
	if err := id2.LoadPublicKey(id1.GetPublicKey()); err != nil {
		b.Fatalf("LoadPublicKey: %v", err)
	}

	mlen := MTU
	msg := make([]byte, mlen)
	if _, err := rand.Read(msg); err != nil {
		b.Fatalf("rand.Read: %v", err)
	}

	b.SetBytes(int64(mlen))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		token, err := id2.Encrypt(msg, nil)
		if err != nil {
			b.Fatalf("Encrypt: %v", err)
		}
		plain, err := id1.Decrypt(token, nil, false)
		if err != nil {
			b.Fatalf("Decrypt: %v", err)
		}
		if len(plain) != len(msg) {
			b.Fatalf("length mismatch")
		}
	}
}
