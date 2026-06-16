import { solve } from './proofOfWorkSolver';

interface SolveRequest {
  challenge: string;
  difficulty: number;
}

// Cast to a tiny local shape rather than pulling in the WebWorker lib, the DOM
// lib types `self` as a Window whose postMessage wants a target origin.
const worker = self as unknown as {
  onmessage: ((event: MessageEvent<SolveRequest>) => void) | null;
  postMessage: (message: unknown) => void;
};

worker.onmessage = async (event: MessageEvent<SolveRequest>) => {
  try {
    const nonce = await solve(event.data.challenge, event.data.difficulty);
    worker.postMessage({ ok: true, nonce });
  } catch (err) {
    worker.postMessage({ ok: false, error: String(err) });
  }
};
