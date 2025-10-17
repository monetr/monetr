import type { Config } from 'jest';
import path from 'node:path';

const config: Config = {
  rootDir: path.resolve(__dirname),
  roots: ['<rootDir>/src'],
  modulePaths: ['<rootDir>/src'],
  moduleNameMapper: {
    // eslint-disable-next-line max-len
    '^.+\\.(jpg|jpeg|png|gif|eot|otf|webp|svg|ttf|woff|woff2|mp4|webm|wav|mp3|m4a|aac|oga|lottie.json|xlsx)$':
      '<rootDir>/src/testutils/mocks/fileMock.js',
    '^.+\\.(css|scss|less)$': '<rootDir>/src/testutils/mocks/styleMock.js',
    '^@monetr/interface/(.*)$': '<rootDir>/src/$1',
  },
  testPathIgnorePatterns: ['node_modules'],
  resetMocks: false,
  collectCoverageFrom: ['src/**/*.{js,jsx,ts,tsx}', '!src/**/*.d.ts', '!src/**/*.stories.{js,jsx,ts,tsx}'],
  coveragePathIgnorePatterns: ['node_modules'],
  testEnvironment: 'jest-environment-jsdom', // '@happy-dom/jest-environment',
  setupFilesAfterEnv: ['<rootDir>/src/setupTests.ts'],
  transform: {
    '^.+\\.(t|j)sx?$': [
      '@swc/jest',
      {
        jsc: {
          transform: {
            react: {
              runtime: 'automatic',
            },
          },
        },
      },
    ],
  },
  coverageReporters: ['lcov'],
};

export default config;
