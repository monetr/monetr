import type { Config } from 'jest';

const config: Config = {
  testPathIgnorePatterns: [
    '.+/pkg/.+',
  ],
  modulePaths: [
    '<rootDir>',
    '<rootDir>/src',
    'node_modules',
  ],
  moduleNameMapper: {
    // eslint-disable-next-line max-len
    '^.+\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga|lottie.json|xlsx)$': '<rootDir>/src/testutils/mocks/fileMock.js',
  },
  resetMocks: false,
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.stories.{js,jsx,ts,tsx}',
  ],
  testEnvironment: 'jest-environment-jsdom',
  setupFilesAfterEnv: [
    '<rootDir>/src/setupTests.js',
  ],
  transform: {
    '^.+\\.(t|j)sx?$': '@swc/jest',
  },
  coverageReporters: [
    'lcov',
  ],
};

export default config;