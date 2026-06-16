// MUST match the prefix the server uses in server/powchallenge/challenger.go.
const PROOF_PREFIX = 'monetr-pow-v1:';

const YIELD_EVERY = 256;

function leadingZeroBits(bytes: Uint8Array): number {
  let count = 0;
  for (const value of bytes) {
    if (value === 0) {
      count += 8;
      continue;
    }
    // clz32 counts leading zeros of a 32 bit number, our byte is in the low 8.
    count += Math.clz32(value) - 24;
    break;
  }
  return count;
}

/**
 * solve looks for the smallest nonce such that SHA-256 of the prefix, the
 * challenge and the nonce (as 8 big-endian bytes) has at least `difficulty`
 * leading zero bits. It has no DOM or Worker globals so the same function backs
 * both the worker and the inline fallback, they can never disagree.
 *
 * @param {string} challenge The opaque challenge token from the server.
 * @param {number} difficulty The number of leading zero bits the proof needs.
 * @param {AbortSignal} signal Optional signal to cancel the search.
 * @returns {Promise<number>} The nonce that solves the challenge.
 */
export async function solve(challenge: string, difficulty: number, signal?: AbortSignal): Promise<number> {
  // Any nonce satisfies difficulty 0, do not bother hashing.
  if (difficulty <= 0) {
    return 0;
  }

  const prefix = new TextEncoder().encode(`${PROOF_PREFIX}${challenge}:`);
  // prefix followed by 8 bytes for the nonce, reused and rewritten each iteration.
  const buffer = new Uint8Array(prefix.length + 8);
  buffer.set(prefix, 0);
  const view = new DataView(buffer.buffer);

  for (let nonce = 0; nonce < Number.MAX_SAFE_INTEGER; nonce++) {
    if (signal?.aborted) {
      throw new DOMException('aborted', 'AbortError');
    }

    view.setBigUint64(prefix.length, BigInt(nonce), false); // big-endian
    const digest = new Uint8Array(await crypto.subtle.digest('SHA-256', buffer));
    if (leadingZeroBits(digest) >= difficulty) {
      return nonce;
    }

    // Yield to the event loop so the inline fallback does not lock up the page.
    if (nonce % YIELD_EVERY === YIELD_EVERY - 1) {
      await new Promise<void>(resolve => {
        setTimeout(resolve, 0);
      });
    }
  }

  throw new Error('failed to find a proof of work solution');
}
