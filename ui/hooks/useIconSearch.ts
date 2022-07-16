import { useQuery } from 'react-query';

import { useAppConfiguration } from 'hooks/useAppConfiguration';

export interface Icon {
  svg: string;
  colors: Array<string>;
}

export function useIconSearch(name: string): Icon | null {
  const configuration = useAppConfiguration();
  const { data } = useQuery<Icon>(`/api/icons/search?name=${ name }`, {
    // Need to !! this otherwise it doesnt work right and evaluates to true when app config is loading.
    enabled: !!configuration.iconsEnabled,
    staleTime: 60 * 60 * 1000, // 60 minutes in milliseconds
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: false,
  });

  return data;
}
