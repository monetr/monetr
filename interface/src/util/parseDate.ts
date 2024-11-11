import { parseJSON } from 'date-fns';

export default function parseDate(input: Date | string | null | undefined): Date | null {
  if (typeof input === 'string') {
    return parseJSON(input);
  } else if (typeof input === 'object' && !!input) {
    return input;
  }

  return null;
}
