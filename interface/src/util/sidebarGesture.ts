// This is true while the mobile sidebar is being dragged sideways, and is used to stop pull to refresh. It isn't
// context because pull to refresh is above the sidebar and reads this from a touch handler.
let active = false;

export function setSidebarGestureActive(next: boolean): void {
  active = next;
}

export function isSidebarGestureActive(): boolean {
  return active;
}
