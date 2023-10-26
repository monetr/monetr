
export function queryText(query: string): string|null {
  return document.querySelector(query).textContent;
}
