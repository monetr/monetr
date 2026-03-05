// Returns an empty CSS module object for test environments.
// Mirrors what the old Jest moduleNameMapper did with styleMock.js.
module.exports = function () {
  return 'module.exports = {};';
};
