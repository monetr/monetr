import { useCallback, useEffect, useRef } from 'react';

import { fetchChallenge, type PowPurpose, type PowSolution, solveChallenge } from '@monetr/interface/util/proofOfWork';

export interface UseProofOfWork {
  getSolution: () => Promise<PowSolution | null>;
  reset: () => void;
}

// Treat a challenge as stale a few seconds before its ttl to cover the request round trips. Absolute clock skew is
// irrelevant: each side only ever compares its own timestamps.
const STALE_MARGIN_MS = 5_000;

/**
 * useProofOfWork pre-solves a challenge on mount so the solution is ready by the time the user submits (getSolution
 * returns it, or null when disabled). It solves once, never in the background; if the user idles past the ttl,
 * getSolution fetches and solves a fresh one then. A challenge is single use and is consumed even on a failed submit,
 * so call reset() after a failure.
 *
 * @param {PowPurpose} purpose Which endpoint this challenge is for.
 * @param {boolean} enabled Whether proof of work is turned on for this server.
 * @returns {UseProofOfWork} getSolution and reset.
 */
export function useProofOfWork(purpose: PowPurpose, enabled: boolean): UseProofOfWork {
  const solutionRef = useRef<Promise<PowSolution> | null>(null);
  const abortRef = useRef<AbortController | null>(null);
  // When the current challenge goes stale (epoch ms); Infinity while a fetch is in flight so we never restart a
  // not-yet-solved one.
  const staleAtRef = useRef<number>(0);

  const start = useCallback(() => {
    if (!enabled) {
      return;
    }

    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    staleAtRef.current = Number.POSITIVE_INFINITY;

    const solution = fetchChallenge(purpose).then(challenge => {
      staleAtRef.current = Date.now() + challenge.ttl * 1000 - STALE_MARGIN_MS;
      return solveChallenge(challenge, controller.signal);
    });
    // Swallow so an unawaited rejection does not warn; getSolution surfaces it.
    solution.catch(() => {});
    solutionRef.current = solution;
  }, [enabled, purpose]);

  // Pre-fetch on mount, abort any in-flight work on unmount.
  useEffect(() => {
    start();
    return () => {
      abortRef.current?.abort();
    };
  }, [start]);

  const getSolution = useCallback((): Promise<PowSolution | null> => {
    if (!enabled) {
      return Promise.resolve(null);
    }

    // Fetch fresh if we have none or the pre-solved one went stale while idle. This and reset are the only places we
    // start solving (never in background).
    if (!solutionRef.current || Date.now() >= staleAtRef.current) {
      start();
    }

    return solutionRef.current ?? Promise.resolve(null);
  }, [enabled, start]);

  const reset = useCallback(() => {
    start();
  }, [start]);

  return {
    getSolution,
    reset,
  };
}
