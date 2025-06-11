
import type { BaseRuntimePageInfo } from '@rspress/shared';
import { usePageData } from 'rspress/runtime';

export default function useGetPagesUnderRoute(path: string): Array<BaseRuntimePageInfo> {
  const data = usePageData();
  return data.siteData.pages.filter(page => page.routePath.startsWith(path));
}
