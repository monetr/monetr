export default function getColor(input: string, saturation: number = 75, lightness: number = 75): string {
  let hash = 0;
  for (let i = 0; i < input.length; i++) {
    hash = input.charCodeAt(i) + ((hash << 5) - hash);
    hash = hash & hash;
  }
  return `hsl(${(hash % 360)}, ${saturation}%, ${lightness}%)`;
}
