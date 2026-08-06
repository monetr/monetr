// jsdom has no PointerEvent, so testing-library falls back to a bare Event and quietly drops every init property
// including clientX. MouseEvent already carries the coordinates, this only adds the pointer specific fields.
export class TestPointerEvent extends MouseEvent {
  public readonly pointerId: number;
  public readonly isPrimary: boolean;
  public readonly pointerType: string;

  constructor(type: string, init: PointerEventInit & { timeStamp?: number } = {}) {
    super(type, init);
    this.pointerId = init.pointerId ?? 0;
    this.isPrimary = init.isPrimary ?? false;
    this.pointerType = init.pointerType ?? '';
    // Velocity gets measured over event timestamps. On the real clock those are sub millisecond apart, which makes
    // fling estimates either zero or enormous depending on scheduling.
    if (init.timeStamp !== undefined) {
      Object.defineProperty(this, 'timeStamp', { configurable: true, value: init.timeStamp });
    }
  }
}

export function installPointerEvents(): () => void {
  const original = Object.getOwnPropertyDescriptor(window, 'PointerEvent');
  Object.defineProperty(window, 'PointerEvent', { configurable: true, writable: true, value: TestPointerEvent });
  return () => {
    if (original) {
      Object.defineProperty(window, 'PointerEvent', original);
      return;
    }
    Reflect.deleteProperty(window, 'PointerEvent');
  };
}
