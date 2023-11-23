import { twMerge } from 'tailwind-merge';
// @ts-ignore
import type { ClassNameValue } from 'tailwind-merge/dist/lib/tw-join';

type ClassNameMap = {[key: string]: boolean | undefined | null | 0 | string};

export default function mergeTailwind(...args: (ClassNameValue | ClassNameMap)[]): string {
  const flattened = args.map(arg => {
    if (typeof arg === 'object' && !Array.isArray(arg) && arg !== null) {
      return Object.keys(arg).filter(key => !!arg[key]);
    }

    return arg;
  }) as ClassNameValue;

  return twMerge(flattened);
}
