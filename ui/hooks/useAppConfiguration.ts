import { useQuery } from 'react-query';

export class AppConfiguration {
  iconsEnabled: boolean;

  constructor(data?: Partial<AppConfiguration>) {
    if (data) Object.assign(this, data);
  }
}

export function useAppConfiguration(): AppConfiguration {
  const { data } = useQuery<Partial<AppConfiguration>>('/api/config');
  return new AppConfiguration(data);
}
