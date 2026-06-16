import request from '@monetr/interface/util/request';

import { solve } from './proofOfWorkSolver';

export type PowPurpose = 'register' | 'login' | 'forgot';

export interface PowChallenge {
  challenge: string;
  difficulty: number;
}

export interface PowSolution {
  challenge: string;
  nonce: number;
}

/**
 * fetchChallenge asks the API for a fresh challenge for the given purpose. The
 * purpose binds it to one endpoint, so a `login` challenge cannot register.
 *
 * @param {PowPurpose} purpose Which endpoint the challenge is for.
 * @returns {Promise<PowChallenge>} The challenge token and its difficulty.
 */
export async function fetchChallenge(purpose: PowPurpose): Promise<PowChallenge> {
  const response = await request<PowChallenge>({
    method: 'POST',
    url: '/api/authentication/challenge',
    data: { purpose },
  });
  return {
    challenge: response.data.challenge,
    difficulty: response.data.difficulty,
  };
}

/**
 * solveChallenge solves a challenge in a dedicated web worker so the form stays
 * responsive. If the worker cannot be constructed (jsdom in tests, a strict CSP,
 * an old browser) it falls back to solving inline with the same `solve`. Both
 * paths honor the abort signal.
 *
 * @param {PowChallenge} challenge The challenge to solve.
 * @param {AbortSignal} signal Optional signal to cancel the work (e.g. on unmount).
 * @returns {Promise<PowSolution>} The challenge paired with the nonce that solves it.
 */
export async function solveChallenge(challenge: PowChallenge, signal?: AbortSignal): Promise<PowSolution> {
  const nonce = await solveNonce(challenge.challenge, challenge.difficulty, signal);
  return {
    challenge: challenge.challenge,
    nonce,
  };
}

function solveNonce(challenge: string, difficulty: number, signal?: AbortSignal): Promise<number> {
  // Check before constructing so we never hit the worker URL machinery where it
  // cannot work (jsdom), and fall back inline if construction throws (CSP, old
  // browsers).
  if (typeof Worker === 'undefined') {
    return solve(challenge, difficulty, signal);
  }

  let worker: Worker;
  try {
    worker = new Worker(new URL('./proofOfWork.worker.ts', import.meta.url), {
      type: 'module',
    });
  } catch {
    return solve(challenge, difficulty, signal);
  }

  return new Promise<number>((resolve, reject) => {
    function cleanup() {
      worker.terminate();
      signal?.removeEventListener('abort', onAbort);
    }

    function onAbort() {
      cleanup();
      reject(new DOMException('aborted', 'AbortError'));
    }

    if (signal?.aborted) {
      cleanup();
      reject(new DOMException('aborted', 'AbortError'));
      return;
    }
    signal?.addEventListener('abort', onAbort);

    worker.onmessage = (event: MessageEvent<{ ok: boolean; nonce?: number; error?: string }>) => {
      cleanup();
      if (event.data.ok && typeof event.data.nonce === 'number') {
        resolve(event.data.nonce);
        return;
      }
      reject(new Error(event.data.error || 'proof of work worker failed'));
    };

    worker.onerror = (event: ErrorEvent) => {
      cleanup();
      reject(new Error(event.message || 'proof of work worker errored'));
    };

    worker.postMessage({ challenge, difficulty });
  });
}
