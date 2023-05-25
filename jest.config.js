module.exports = {
  "testPathIgnorePatterns": [
    ".+/pkg/.+"
  ],
  "modulePaths": [
    "<rootDir>/ui"
  ],
  "moduleNameMapper": {
    "^.+\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga|lottie.json|xlsx)$": "<rootDir>/ui/testutils/mocks/fileMock.js"
  },
  "resetMocks": false,
  "collectCoverageFrom": [
    "ui/**/*.{js,jsx,ts,tsx}",
    "!ui/**/*.d.ts",
    "!ui/**/*.stories.{js,jsx,ts,tsx}"
  ],
  "testEnvironment": "jest-environment-jsdom",
  "setupFilesAfterEnv": [
    "<rootDir>/ui/setupTests.js"
  ],
  "transform": {
    "^.+\\.(t|j)sx?$": "@swc/jest"
  }
}
