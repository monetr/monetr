type ClassNameMap = { [key: string]: boolean | undefined | null | 0 | string };
type ClassNameValue = string | undefined | null | false | 0 | ClassNameValue[];

// mergeClasses joins class name arguments into a single space-separated string.
// It accepts plain strings, arrays (flattened recursively), and conditional
// maps (`{ className: boolean }`) where keys with a truthy value are included.
//
// This replaces the previous `mergeClasses` helper, which wrapped
// `tailwind-merge`'s `twMerge` to de-duplicate conflicting Tailwind utilities.
// Now that the UI is styled entirely with SCSS Modules there are no Tailwind
// utilities to reconcile, so a plain join is sufficient.
export default function mergeClasses(...args: (ClassNameValue | ClassNameMap)[]): string {
  const result: string[] = [];

  for (const arg of args) {
    if (!arg) continue;

    if (typeof arg === 'string') {
      result.push(arg);
    } else if (Array.isArray(arg)) {
      const nested = mergeClasses(...arg);
      if (nested) result.push(nested);
    } else if (typeof arg === 'object') {
      for (const key of Object.keys(arg)) {
        if (arg[key]) result.push(key);
      }
    }
  }

  return result.join(' ');
}
