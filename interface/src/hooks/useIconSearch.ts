import { useQuery } from '@tanstack/react-query';

import { useAppConfiguration } from '@monetr/interface/hooks/useAppConfiguration';

export interface Icon {
  svg: string;
  colors: Array<string>;
}

export function useIconSearch(name: string): Icon | null {
  const { data: configuration } = useAppConfiguration();
  const { data } = useQuery<Icon>({
    queryKey: ['/icons/search', { name }],
    enabled: Boolean(configuration?.iconsEnabled) && Boolean(name) && name?.length > 0,
    staleTime: 60 * 60 * 1000, // 60 minutes in milliseconds
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
    refetchOnMount: false,
    retry: false,
  });

  return data;
}
