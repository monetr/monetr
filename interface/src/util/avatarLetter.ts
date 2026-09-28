/**
 * Returns the first letter or number (in any script) of the provided string, uppercased, for use as an avatar
 * fallback. Leading punctuation, symbols and whitespace are skipped. Returns '?' if there is no letter or number.
 */
export default function avatarLetter(input?: string | null): string {
  const match = input?.match(/[\p{L}\p{N}]/u);
  return match?.[0]?.toUpperCase() || '?';
}
