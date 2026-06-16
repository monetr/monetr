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
 * solve finds the smallest nonce whose SHA-256 of prefix + challenge + nonce (8
 * big-endian bytes) has at least `difficulty` leading zero bits. No DOM or Worker
 * globals, so the worker and the inline fallback share it and cannot disagree.
 *
 * @param {string} challenge The opaque challenge token from the server.
 * @param {number} difficulty Leading zero bits the proof needs.
 * @param {AbortSignal} signal Optional cancel signal.
 * @returns {Promise<number>} The solving nonce.
 */
export async function solve(challenge: string, difficulty: number, signal?: AbortSignal): Promise<number> {
  // Any nonce satisfies difficulty 0.
  if (difficulty <= 0) {
    return 0;
  }

  // SubtleCrypto is only exposed in a secure context (HTTPS, or a localhost
  // address). On a plain-http non-localhost origin crypto.subtle is undefined,
  // so fail with a message that points at the cause rather than a bare TypeError.
  if (!globalThis.crypto?.subtle) {
    throw new Error('proof of work requires a secure context (HTTPS or localhost)');
  }

  const prefix = new TextEncoder().encode(`${PROOF_PREFIX}${challenge}:`);
  // prefix + 8 nonce bytes, rewritten each iteration.
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
