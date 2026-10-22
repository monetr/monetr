// jsdom doesn't have PointerEvent, so testing-library falls back to a plain Event and drops clientX and everything
// else. MouseEvent already has the coordinates, this just adds the pointer fields on top.
export class TestPointerEvent extends MouseEvent {
  public readonly pointerId: number;
  public readonly isPrimary: boolean;
  public readonly pointerType: string;

  constructor(type: string, init: PointerEventInit & { timeStamp?: number } = {}) {
    super(type, init);
    this.pointerId = init.pointerId ?? 0;
    this.isPrimary = init.isPrimary ?? false;
    this.pointerType = init.pointerType ?? '';
    // The real timestamps are less than a millisecond apart in tests, which makes the fling velocity random.
    if (init.timeStamp !== undefined) {
      Object.defineProperty(this, 'timeStamp', {
        configurable: true,
        value: init.timeStamp,
      });
    }
  }
}

export function installPointerEvents(): () => void {
  const original = Object.getOwnPropertyDescriptor(window, 'PointerEvent');
  Object.defineProperty(window, 'PointerEvent', {
    configurable: true,
    writable: true,
    value: TestPointerEvent,
  });
  return () => {
    if (original) {
      Object.defineProperty(window, 'PointerEvent', original);
      return;
    }
    Reflect.deleteProperty(window, 'PointerEvent');
  };
}
