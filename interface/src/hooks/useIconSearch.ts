import { useQuery } from '@tanstack/react-query';

import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';

export interface Icon {
  svg: string;
  colors: Array<string>;
}

export function useIconSearch(name: string): Icon | null {
  const configuration = useAppConfiguration();
  const { data } = useQuery<Icon>(['/icons/search', { name }], {
    // Need to !! this otherwise it doesnt work right and evaluates to true when app config is loading.
    enabled: !!configuration?.iconsEnabled && !!name && name?.length > 0,
    staleTime: 60 * 60 * 1000, // 60 minutes in milliseconds
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: false,
    retry: false,
  });

  return data;
}
