package powchallenge_test

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/binary"
	"math/bits"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/benbjohnson/clock"
	"github.com/gomodule/redigo/redis"
	"github.com/monetr/monetr/server/cache"
	"github.com/monetr/monetr/server/internal/testutils"
	"github.com/monetr/monetr/server/powchallenge"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tokenEncoding mirrors the encoding the challenger uses for its tokens so the
// tamper tests can decode, poke at, and re-encode a real token.
var tokenEncoding = base32.StdEncoding.WithPadding(base32.NoPadding)

// These helpers are an independent implementation of the canonical proof of work
// algorithm. We deliberately do NOT reuse the production code here so that the
// tests actually cross-check the production [powchallenge.Challenger] against a
// from-scratch implementation of the spec. The frontend solver in
// interface/src/util/proofOfWorkSolver.ts is a third implementation of the same
// algorithm, and the shared vectors in TestGenerateFrontendVectors are what keep
// all three honest.

func computeTestDigest(challenge string, nonce uint64) []byte {
	h := sha256.New()
	h.Write([]byte("monetr-pow-v1:"))
	h.Write([]byte(challenge))
	h.Write([]byte(":"))
	var nonceBytes [8]byte
	binary.BigEndian.PutUint64(nonceBytes[:], nonce)
	h.Write(nonceBytes[:])
	return h.Sum(nil)
}

func testLeadingZeroBits(digest []byte) int {
	var count int
	for _, b := range digest {
		if b == 0 {
			count += 8
			continue
		}
		count += bits.LeadingZeros8(b)
		break
	}
	return count
}

// solveTestProof finds the smallest nonce that satisfies the difficulty. The
// frontend solver also returns the smallest nonce, so the values must agree.
func solveTestProof(t *testing.T, challenge string, difficulty int) uint64 {
	for nonce := uint64(0); nonce < 1<<32; nonce++ {
		if testLeadingZeroBits(computeTestDigest(challenge, nonce)) >= difficulty {
			return nonce
		}
	}
	t.Fatalf("failed to find a proof of work solution for difficulty %d", difficulty)
	return 0
}

func newTestCache(t *testing.T) cache.Cache {
	miniRedis := miniredis.NewMiniRedis()
	require.NoError(t, miniRedis.Start())
	pool := &redis.Pool{
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", miniRedis.Server().Addr().String())
		},
	}
	t.Cleanup(func() {
		require.NoError(t, pool.Close())
		miniRedis.Close()
	})
	return cache.NewCache(testutils.GetLog(t), pool)
}

func newTestClock() *clock.Mock {
	clk := clock.NewMock()
	clk.Set(time.Date(2023, 10, 9, 13, 32, 0, 0, time.UTC))
	return clk
}

// We use difficulty 4 almost everywhere in the tests. Anything higher just makes
// the tests slower without proving anything new, and high difficulties with
// their large variance can make solving flaky.
const testDifficulty = 4

func newTestChallenger(t *testing.T, difficulty int) (powchallenge.Challenger, *clock.Mock) {
	clk := newTestClock()
	secret := powchallenge.DeriveSecret([]byte("test-seed-for-proof-of-work"))
	challenger := powchallenge.NewChallenger(
		testutils.GetLog(t),
		newTestCache(t),
		clk,
		nil,
		secret,
		difficulty,
		5*time.Minute,
	)
	return challenger, clk
}

// compareAndSwapErrorCache wraps a real cache but makes CompareAndSwap always
// error. We use it to prove the challenger fails closed when redis is unhappy.
// SetTTL and everything else still go to the real cache so that Issue works.
type compareAndSwapErrorCache struct {
	cache.Cache
}

func (compareAndSwapErrorCache) CompareAndSwap(_ context.Context, _ string, _, _ []byte, _ time.Duration) (bool, error) {
	return false, errors.New("redis is down")
}

func TestChallenger_IssueAndVerify(t *testing.T) {
	purposes := []struct {
		name    string
		purpose powchallenge.Purpose
	}{
		{"register", powchallenge.PurposeRegister},
		{"login", powchallenge.PurposeLogin},
		{"forgot", powchallenge.PurposeForgot},
	}

	for _, p := range purposes {
		t.Run("happy path "+p.name, func(t *testing.T) {
			challenger, _ := newTestChallenger(t, testDifficulty)
			ctx := t.Context()

			challenge, err := challenger.Issue(ctx, p.purpose)
			require.NoError(t, err, "must be able to issue a challenge")
			require.NotNil(t, challenge, "issued challenge must not be nil")
			assert.NotEmpty(t, challenge.Token, "the challenge token must not be empty")
			assert.EqualValues(t, testDifficulty, challenge.Difficulty, "the difficulty should match the configured value")
			assert.EqualValues(t, 300, challenge.TTL, "the ttl should be the lifetime in seconds")

			nonce := solveTestProof(t, challenge.Token, challenge.Difficulty)
			err = challenger.Verify(ctx, p.purpose, challenge.Token, nonce)
			assert.NoError(t, err, "a freshly solved challenge should verify successfully")
		})
	}

	t.Run("cannot use a register challenge for login", func(t *testing.T) {
		// A challenge is bound to a single purpose. Even though this is a perfectly
		// valid register challenge, it should not be accepted on the login
		// endpoint.
		challenger, _ := newTestChallenger(t, testDifficulty)
		ctx := t.Context()

		challenge, err := challenger.Issue(ctx, powchallenge.PurposeRegister)
		require.NoError(t, err, "must be able to issue a register challenge")

		nonce := solveTestProof(t, challenge.Token, challenge.Difficulty)
		err = challenger.Verify(ctx, powchallenge.PurposeLogin, challenge.Token, nonce)
		assert.ErrorIs(t, err, powchallenge.ErrInvalidChallenge, "a register challenge must not be usable for login")
	})

	t.Run("an expired challenge is rejected", func(t *testing.T) {
		challenger, clk := newTestChallenger(t, testDifficulty)
		ctx := t.Context()

		challenge, err := challenger.Issue(ctx, powchallenge.PurposeLogin)
		require.NoError(t, err, "must be able to issue a challenge")
		nonce := solveTestProof(t, challenge.Token, challenge.Difficulty)

		// Jump past the 5 minute lifetime. The mock clock is what the challenger
		// reads, so this makes the challenge look old.
		clk.Add(6 * time.Minute)

		err = challenger.Verify(ctx, powchallenge.PurposeLogin, challenge.Token, nonce)
		assert.ErrorIs(t, err, powchallenge.ErrChallengeExpired, "an expired challenge must be rejected")
	})

	t.Run("a challenge cannot be used twice", func(t *testing.T) {
		challenger, _ := newTestChallenger(t, testDifficulty)
		ctx := t.Context()

		challenge, err := challenger.Issue(ctx, powchallenge.PurposeForgot)
		require.NoError(t, err, "must be able to issue a challenge")
		nonce := solveTestProof(t, challenge.Token, challenge.Difficulty)

		err = challenger.Verify(ctx, powchallenge.PurposeForgot, challenge.Token, nonce)
		assert.NoError(t, err, "the first use of a challenge should succeed")

		err = challenger.Verify(ctx, powchallenge.PurposeForgot, challenge.Token, nonce)
		assert.ErrorIs(t, err, powchallenge.ErrChallengeReplay, "a second use of the same challenge must be rejected as a replay")
	})

	t.Run("a tampered signature is rejected", func(t *testing.T) {
		challenger, _ := newTestChallenger(t, testDifficulty)
		ctx := t.Context()

		challenge, err := challenger.Issue(ctx, powchallenge.PurposeRegister)
		require.NoError(t, err, "must be able to issue a challenge")

		raw, err := tokenEncoding.DecodeString(challenge.Token)
		require.NoError(t, err, "must be able to decode the issued token")
		// Flip every bit of the very last byte, which is part of the HMAC.
		raw[len(raw)-1] ^= 0xFF
		tampered := tokenEncoding.EncodeToString(raw)

		// The HMAC is checked before the proof, so the nonce here does not matter.
		err = challenger.Verify(ctx, powchallenge.PurposeRegister, tampered, 0)
		assert.ErrorIs(t, err, powchallenge.ErrInvalidChallenge, "a token with a broken HMAC must be rejected")
	})

	t.Run("a tampered purpose is rejected", func(t *testing.T) {
		// Changing the purpose byte without re-signing should break the HMAC.
		challenger, _ := newTestChallenger(t, testDifficulty)
		ctx := t.Context()

		challenge, err := challenger.Issue(ctx, powchallenge.PurposeRegister)
		require.NoError(t, err, "must be able to issue a challenge")

		raw, err := tokenEncoding.DecodeString(challenge.Token)
		require.NoError(t, err, "must be able to decode the issued token")
		// The purpose byte lives right after the version byte.
		raw[1] = byte(powchallenge.PurposeLogin)
		tampered := tokenEncoding.EncodeToString(raw)

		err = challenger.Verify(ctx, powchallenge.PurposeLogin, tampered, 0)
		assert.ErrorIs(t, err, powchallenge.ErrInvalidChallenge, "flipping the purpose must break the HMAC and be rejected")
	})

	t.Run("a stale low difficulty challenge is rejected after a policy bump", func(t *testing.T) {
		// Simulate raising the difficulty: a token issued at difficulty 4 should be
		// rejected once the policy is 8, even though the token is otherwise valid.
		// The two challengers share the same secret and cache so the only thing
		// that differs is the configured difficulty.
		clk := newTestClock()
		secret := powchallenge.DeriveSecret([]byte("shared-secret-for-difficulty-test"))
		sharedCache := newTestCache(t)
		ctx := t.Context()

		issuer := powchallenge.NewChallenger(testutils.GetLog(t), sharedCache, clk, nil, secret, 4, 5*time.Minute)
		verifier := powchallenge.NewChallenger(testutils.GetLog(t), sharedCache, clk, nil, secret, 8, 5*time.Minute)

		challenge, err := issuer.Issue(ctx, powchallenge.PurposeRegister)
		require.NoError(t, err, "must be able to issue a low difficulty challenge")

		// The difficulty is checked before the proof, so the nonce does not matter.
		err = verifier.Verify(ctx, powchallenge.PurposeRegister, challenge.Token, 0)
		assert.ErrorIs(t, err, powchallenge.ErrInvalidChallenge, "a difficulty 4 token must be rejected when the policy is 8")
	})

	t.Run("a solution with too few zero bits is rejected", func(t *testing.T) {
		challenger, _ := newTestChallenger(t, 12)
		ctx := t.Context()

		challenge, err := challenger.Issue(ctx, powchallenge.PurposeLogin)
		require.NoError(t, err, "must be able to issue a challenge")

		// Find a nonce that we KNOW does not satisfy the difficulty so the test is
		// deterministic rather than relying on nonce 0 happening to be bad.
		var badNonce uint64
		for n := uint64(0); ; n++ {
			if testLeadingZeroBits(computeTestDigest(challenge.Token, n)) < challenge.Difficulty {
				badNonce = n
				break
			}
		}

		err = challenger.Verify(ctx, powchallenge.PurposeLogin, challenge.Token, badNonce)
		assert.ErrorIs(t, err, powchallenge.ErrInvalidProof, "a solution that does not meet the difficulty must be rejected")
	})

	t.Run("a corrupted token is rejected", func(t *testing.T) {
		challenger, _ := newTestChallenger(t, testDifficulty)

		// The spaces and exclamation points are not in the base32 alphabet, so this
		// fails to even decode.
		err := challenger.Verify(t.Context(), powchallenge.PurposeRegister, "this is not a valid token!!!", 0)
		assert.ErrorIs(t, err, powchallenge.ErrInvalidChallenge, "a token that cannot be decoded must be rejected")
	})

	t.Run("a token of the wrong length is rejected", func(t *testing.T) {
		challenger, _ := newTestChallenger(t, testDifficulty)

		// Valid base32, but nowhere near the right number of bytes.
		short := tokenEncoding.EncodeToString([]byte("too short"))
		err := challenger.Verify(t.Context(), powchallenge.PurposeRegister, short, 0)
		assert.ErrorIs(t, err, powchallenge.ErrInvalidChallenge, "a token of the wrong length must be rejected")
	})

	t.Run("a redis error fails closed", func(t *testing.T) {
		clk := newTestClock()
		secret := powchallenge.DeriveSecret([]byte("test-seed-for-proof-of-work"))
		ctx := t.Context()

		// Issue with a real cache so the marker is written, but verify against a
		// cache that errors on CompareAndSwap.
		challenger := powchallenge.NewChallenger(
			testutils.GetLog(t),
			compareAndSwapErrorCache{Cache: newTestCache(t)},
			clk,
			nil,
			secret,
			testDifficulty,
			5*time.Minute,
		)

		challenge, err := challenger.Issue(ctx, powchallenge.PurposeRegister)
		require.NoError(t, err, "must be able to issue a challenge even with the wrapped cache")
		nonce := solveTestProof(t, challenge.Token, challenge.Difficulty)

		err = challenger.Verify(ctx, powchallenge.PurposeRegister, challenge.Token, nonce)
		assert.Error(t, err, "a redis error during verification must fail closed")
		assert.NotErrorIs(t, err, powchallenge.ErrChallengeReplay, "the error should be the underlying redis error, not a replay")
	})

	t.Run("an unknown purpose cannot be issued", func(t *testing.T) {
		challenger, _ := newTestChallenger(t, testDifficulty)

		_, err := challenger.Issue(t.Context(), powchallenge.Purpose(99))
		assert.Error(t, err, "issuing a challenge for an unknown purpose must error")
	})
}

func TestDeriveSecret(t *testing.T) {
	t.Run("is deterministic for the same seed", func(t *testing.T) {
		seed := []byte("the same seed every single time")
		first := powchallenge.DeriveSecret(seed)
		second := powchallenge.DeriveSecret(seed)
		assert.Equal(t, first, second, "the same seed must always derive the same secret")
		assert.Len(t, first, 32, "the derived secret should be 32 bytes")
	})

	t.Run("different seeds derive different secrets", func(t *testing.T) {
		first := powchallenge.DeriveSecret([]byte("seed one"))
		second := powchallenge.DeriveSecret([]byte("seed two"))
		assert.NotEqual(t, first, second, "different seeds must derive different secrets")
	})
}

// TestGenerateFrontendVectors locks the wire format. The exact same triples are
// hardcoded in interface/src/util/proofOfWork.spec.ts. If you ever change the
// hashing then these nonces will change, in which case run this test (the values
// are logged) and paste the new nonces into both this test and the frontend
// spec. The minimal-nonce contract matters: both this solver and the frontend
// solver start at 0 and return the first nonce that works.
func TestGenerateFrontendVectors(t *testing.T) {
	vectors := []struct {
		challenge  string
		difficulty int
		nonce      uint64
	}{
		{"monetr-pow-test-vector-1", 4, 21},
		{"monetr-pow-test-vector-1", 8, 104},
		{"monetr-pow-test-vector-2", 12, 1654},
	}

	for _, vector := range vectors {
		nonce := solveTestProof(t, vector.challenge, vector.difficulty)
		t.Logf("challenge=%q difficulty=%d nonce=%d", vector.challenge, vector.difficulty, nonce)

		// Whatever we computed must actually satisfy the difficulty.
		assert.GreaterOrEqualf(
			t,
			testLeadingZeroBits(computeTestDigest(vector.challenge, nonce)),
			vector.difficulty,
			"the generated nonce must satisfy the difficulty for %q", vector.challenge,
		)

		// Lock the value down so it cannot silently change.
		assert.Equalf(
			t,
			vector.nonce,
			nonce,
			"wire format vector for %q at difficulty %d must be stable",
			vector.challenge,
			vector.difficulty,
		)
	}
}
