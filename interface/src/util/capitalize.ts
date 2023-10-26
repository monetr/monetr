/**
 * Will capitalize the first letter of the provided string.
 */
export default function capitalize(input: string): string {
  const first = input.charAt(0).toUpperCase();
  const rest = input.slice(1);
  return first + rest;
}
