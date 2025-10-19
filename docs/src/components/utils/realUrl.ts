export default function realUrl(path: string): string {
  const uri = new URL('https://monetr.app');
  uri.pathname = path;
  return uri.toString();
}
