import { useCallback, useEffect, useRef } from 'react';

import {
  fetchChallenge,
  type PowPurpose,
  type PowSolution,
  solveChallenge,
} from '@monetr/interface/util/proofOfWork';

export interface UseProofOfWork {
  getSolution: () => Promise<PowSolution | null>;
  reset: () => void;
}

/**
 * useProofOfWork pre-fetches and starts solving a challenge as soon as the form
 * mounts, so by the time the user submits the solution is already waiting and
 * they never wait on it. getSolution hands back that solution (or null when
 * proof of work is disabled). A challenge is single use and is consumed server
 * side even on a failed submit, so callers must call reset() after a failure to
 * line up a fresh one for the retry.
 *
 * @param {PowPurpose} purpose Which endpoint this challenge is for.
 * @param {boolean} enabled Whether proof of work is turned on for this server.
 * @returns {UseProofOfWork} getSolution and reset.
 */
export function useProofOfWork(purpose: PowPurpose, enabled: boolean): UseProofOfWork {
  const solutionRef = useRef<Promise<PowSolution> | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const start = useCallback(() => {
    if (!enabled) {
      return;
    }

    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    const solution = fetchChallenge(purpose).then(challenge => solveChallenge(challenge, controller.signal));
    // Swallow rejections here so we do not get an unhandled rejection warning,
    // getSolution surfaces the real result.
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

    if (!solutionRef.current) {
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
