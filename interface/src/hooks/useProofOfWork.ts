import { useCallback, useEffect, useRef } from 'react';

import { fetchChallenge, type PowPurpose, type PowSolution, solveChallenge } from '@monetr/interface/util/proofOfWork';

export interface UseProofOfWork {
  getSolution: () => Promise<PowSolution | null>;
  reset: () => void;
  warmup: () => void;
}

// Treat a challenge as stale a few seconds before its ttl to cover the request round trips. Absolute clock skew is
// irrelevant: each side only ever compares its own timestamps.
const STALE_MARGIN_MS = 5_000;

/**
 * useProofOfWork fetches and solves a proof of work challenge on demand, it never solves on mount. The first call to
 * warmup (which the page ties to the user actually typing) or getSolution kicks off the fetch and solve, and the result
 * is reused while it is still fresh; if the user idles past the ttl getSolution grabs a fresh one then. We deliberately
 * do NOT pre-solve when the hook mounts. The login page can remount for a few milliseconds during the post-login
 * navigation (the router briefly bounces back to /login before the new auth state settles), and pre-solving on that
 * remount would burn a challenge we never send. warmup is the latency hider and getSolution is the guarantee, so a user
 * who pastes or autofills without typing still gets a solution, just solved at submit time instead. A challenge is
 * single use and is consumed even on a failed submit, so call reset() after a failure to line up a fresh one.
 *
 * @param {PowPurpose} purpose Which endpoint this challenge is for.
 * @param {boolean} enabled Whether proof of work is turned on for this server.
 * @returns {UseProofOfWork} getSolution, reset and warmup.
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

  // warmup kicks off a fetch and solve if we do not already have a fresh one in hand. It is idempotent and cheap, so a
  // page can call it on every keystroke. This is the only thing that pre-solves now, and it is tied to the user
  // actually typing rather than to mount so a transient remount never triggers it.
  const warmup = useCallback(() => {
    if (!enabled) {
      return;
    }

    if (!solutionRef.current || Date.now() >= staleAtRef.current) {
      start();
    }
  }, [enabled, start]);

  // We do not solve on mount anymore, all this does is abort an in-flight solve when the page goes away. See warmup for
  // why mount is the wrong trigger.
  useEffect(() => {
    return () => {
      abortRef.current?.abort();
    };
  }, []);

  const getSolution = useCallback((): Promise<PowSolution | null> => {
    if (!enabled) {
      return Promise.resolve(null);
    }

    // warmup may never have run (the user pasted, autofilled, or never typed) so make sure something is in flight, then
    // hand back whatever we have. This and reset are the only places we are guaranteed to line up a solution.
    warmup();

    return solutionRef.current ?? Promise.resolve(null);
  }, [enabled, warmup]);

  const reset = useCallback(() => {
    start();
  }, [start]);

  return {
    getSolution,
    reset,
    warmup,
  };
}
