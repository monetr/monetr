
import { type TestingLibraryMatchers } from '@testing-library/jest-dom/matchers';

import { type expect } from 'bun:test';

export {};
declare module 'bun:test' {
  interface Matchers<T>
    extends TestingLibraryMatchers<
      ReturnType<typeof expect.stringContaining>,
      T
    > {}
}
