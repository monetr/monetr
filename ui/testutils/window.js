let oldWindowLocation;

export function setupWindowLocationMock() {
  if (!oldWindowLocation) {
    oldWindowLocation = window.location;
  }
  delete window.location;

  window.location = Object.defineProperties(
    {},
    {
      ...Object.getOwnPropertyDescriptors(oldWindowLocation),
      assign: {
        configurable: true,
        value: jest.fn(),
      },
    },
  );
}

export function cleanupWindowLocationMock() {
  window.location = oldWindowLocation;
  oldWindowLocation = null;
}
