import type { Config } from 'jest';

const config: Config = {
  modulePaths: [
    '<rootDir>/src',
  ],
  moduleNameMapper: {
    // eslint-disable-next-line max-len
    '^.+\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga|lottie.json|xlsx)$': '<rootDir>/src/testutils/mocks/fileMock.js',
    '^@monetr/interface/(.*)$': '<rootDir>/src/$1',
  },
  resetMocks: false,
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/*.stories.{js,jsx,ts,tsx}',
  ],
  testEnvironment: 'jest-environment-jsdom',
  setupFilesAfterEnv: [
    '<rootDir>/src/setupTests.ts',
    '@testing-library/jest-dom/extend-expect',
  ],
  transform: {
    '^.+\\.(t|j)sx?$': '@swc/jest',
  },
  coverageReporters: [
    'lcov',
  ],
};

export default config;
