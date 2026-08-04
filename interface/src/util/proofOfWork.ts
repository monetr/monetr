import request from '@monetr/interface/util/request';

import { solve } from './proofOfWorkSolver';

export type PowPurpose = 'register' | 'login' | 'forgot' | 'resend' | 'create_api_key' | 'delete_api_key';

export interface PowChallenge {
  challenge: string;
  difficulty: number;
  // How many seconds the challenge is valid for, used to spot a stale one.
  ttl: number;
}

export interface PowSolution {
  challenge: string;
  nonce: number;
}

/**
 * fetchChallenge gets a fresh challenge for the purpose (which binds it to that one endpoint).
 *
 * @param {PowPurpose} purpose Which endpoint the challenge is for.
 * @returns {Promise<PowChallenge>} The challenge, difficulty and ttl.
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
    ttl: response.data.ttl,
  };
}

/**
 * solveChallenge solves a challenge in a web worker so the form stays responsive, falling back to inline `solve` when a
 * worker is unavailable. Honors the signal.
 *
 * @param {PowChallenge} challenge The challenge to solve.
 * @param {AbortSignal} signal Optional cancel signal (e.g. on unmount).
 * @returns {Promise<PowSolution>} The challenge paired with its solving nonce.
 */
export async function solveChallenge(challenge: PowChallenge, signal?: AbortSignal): Promise<PowSolution> {
  const nonce = await solveNonce(challenge.challenge, challenge.difficulty, signal);
  return {
    challenge: challenge.challenge,
    nonce,
  };
}

function solveNonce(challenge: string, difficulty: number, signal?: AbortSignal): Promise<number> {
  // jsdom has no Worker, and construction can also throw (CSP, old browsers). Either way, solve inline.
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
