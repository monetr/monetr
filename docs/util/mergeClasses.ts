type ClassValue = string | number | boolean | undefined | null | ClassValue[];
type ClassNameMap = { [key: string]: boolean | undefined | null | 0 | string };

export default function mergeClasses(...args: (ClassValue | ClassNameMap)[]): string {
  const parts: string[] = [];
  const walk = (v: unknown): void => {
    if (!v) {
      return;
    }
    if (typeof v === 'string') {
      parts.push(v);
      return;
    }
    if (typeof v === 'number') {
      parts.push(String(v));
      return;
    }
    if (Array.isArray(v)) {
      v.forEach(walk);
      return;
    }
    if (typeof v === 'object') {
      for (const [k, val] of Object.entries(v)) {
        if (val) {
          parts.push(k);
        }
      }
    }
  };
  args.forEach(walk);
  return parts.filter(Boolean).join(' ');
}
