// Returns a stub string for binary asset imports in test environments.
// Mirrors what the old Jest moduleNameMapper did with fileMock.js.
module.exports = function () {
  return 'module.exports = "test-file-stub";';
};
