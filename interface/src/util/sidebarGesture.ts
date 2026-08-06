// Whether the mobile sidebar currently owns a horizontal drag. A module level flag rather than context because the
// only reader, PullToRefresh, is mounted above the sidebar's provider and reads it inside a touchmove handler.
//
// Only goes true once the gesture has picked an axis, which is after ~12px of movement.
let active = false;

export function setSidebarGestureActive(next: boolean): void {
  active = next;
}

export function isSidebarGestureActive(): boolean {
  return active;
}
