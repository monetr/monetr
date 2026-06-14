import { isValid, parseJSON } from 'date-fns';

// This is a typescript trick that makes it so that IF the input of the parse date function is nullable, then the
// function can return either null or a proper date. But IF the input is not nullable, such as an existing date object
// or a string. Then that input is handled. This does not promise that a date is returned, but it does throw an
// exception if there is a problem.

export default function parseDate(input: Date | string): Date;
export default function parseDate(input: null | undefined): null;
export default function parseDate(input: Date | string | null | undefined): Date | null;
export default function parseDate(input: Date | string | null | undefined): Date | null {
  if (input == null) {
    return null;
  }
  if (typeof input === 'string') {
    const result = parseJSON(input);
    if (!isValid(result)) {
      throw new Error(`invalid date provided: ${input}`);
    }
    return result;
  }
  return input;
}
