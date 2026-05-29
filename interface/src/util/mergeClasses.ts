type ClassNameMap = { [key: string]: boolean | undefined | null | 0 | string };
type ClassNameValue = string | undefined | null | false | 0 | ClassNameValue[];

// mergeClasses joins class name arguments into a single space-separated string.
// It accepts plain strings, arrays (flattened recursively), and conditional
// maps (`{ className: boolean }`) where keys with a truthy value are included.
// Use it for conditional styling or when a component takes a `className`; prefer
// standalone classes or @extend for unconditional merges within a module.
export default function mergeClasses(...args: (ClassNameValue | ClassNameMap)[]): string {
  const result: string[] = [];

  for (const arg of args) {
    if (!arg) {
      continue;
    }

    if (typeof arg === 'string') {
      result.push(arg);
    } else if (Array.isArray(arg)) {
      const nested = mergeClasses(...arg);
      if (nested) {
        result.push(nested);
      }
    } else if (typeof arg === 'object') {
      for (const key of Object.keys(arg)) {
        if (arg[key]) {
          result.push(key);
        }
      }
    }
  }

  return result.join(' ');
}
