import { useMediaQuery } from '@monetr/interface/hooks/useMediaQuery';

export default function useIsMobile(): boolean {
  const matches = useMediaQuery('(max-width: 1024px)');
  return matches;
}
