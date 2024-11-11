import { useMediaQuery } from '@monetr/interface/hooks/useMediaQuery';

export default function useIsMobile(): boolean {
  const matches = useMediaQuery('@media only screen and (max-width: 768px)');
  return matches;
}
