import { describe, expect, it } from '@rstest/core';

import { solveChallenge } from '@monetr/interface/util/proofOfWork';
import { solve } from '@monetr/interface/util/proofOfWorkSolver';

// These vectors are generated and locked by TestGenerateFrontendVectors in
// server/powchallenge/challenger_test.go. They MUST match the values over there
// exactly, that is the whole point of them, it proves the frontend solver and
// the backend verifier agree on the wire format right down to the byte. If you
// change the hashing, run that Go test, and paste the new nonces into both
// places.
const vectors = [
  { challenge: 'monetr-pow-test-vector-1', difficulty: 4, nonce: 21 },
  { challenge: 'monetr-pow-test-vector-1', difficulty: 8, nonce: 104 },
  { challenge: 'monetr-pow-test-vector-2', difficulty: 12, nonce: 1654 },
];

describe('proof of work', () => {
  for (const vector of vectors) {
    // The direct solver and the public solveChallenge wrapper (which would use a
    // worker in a real browser but falls back to the exact same solve here in
    // jsdom) must produce identical nonces.
    it(`will solve ${vector.challenge} at difficulty ${vector.difficulty} directly`, async () => {
      const nonce = await solve(vector.challenge, vector.difficulty);
      expect(nonce).toBe(vector.nonce);
    });

    it(`will solve ${vector.challenge} at difficulty ${vector.difficulty} via solveChallenge`, async () => {
      const solution = await solveChallenge({
        challenge: vector.challenge,
        difficulty: vector.difficulty,
      });
      expect(solution.challenge).toBe(vector.challenge);
      expect(solution.nonce).toBe(vector.nonce);
    });
  }

  it('will return a zero nonce immediately for difficulty zero', async () => {
    const nonce = await solve('anything', 0);
    expect(nonce).toBe(0);
  });

  it('will reject with an AbortError when the signal is already aborted', async () => {
    const controller = new AbortController();
    controller.abort();
    // Even at a difficulty that would normally take a while, an already-aborted
    // signal should make it bail right away.
    await expect(solve('monetr-pow-test-vector-1', 24, controller.signal)).rejects.toMatchObject({
      name: 'AbortError',
    });
  });
});
