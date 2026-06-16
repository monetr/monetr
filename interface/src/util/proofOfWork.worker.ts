import { solve } from './proofOfWorkSolver';

interface SolveRequest {
  challenge: string;
  difficulty: number;
}

// Cast to a tiny shape instead of pulling in the WebWorker lib (DOM types `self`
// as a Window).
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
