export function getLocale(): string {
  let locale: string = Intl.DateTimeFormat().resolvedOptions().locale;
  if (!locale && navigator.languages !== undefined) {
    locale = navigator.languages[0];
  }
  if (!locale) {
    locale = navigator.language;
  }

  // Transform en-US to en_US
  return locale.replace('-', '_');
}

export function getTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}
