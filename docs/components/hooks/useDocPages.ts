import { usePages } from '@rspress/core/runtime';
import type { BaseRuntimePageInfo } from '@rspress/shared';

export interface Author {
  name: string;
  github?: string;
}

export type DocPageData = {
  authors: Array<Author>;
} & BaseRuntimePageInfo;

function buildAuthors(page: BaseRuntimePageInfo): Array<Author> {
  return Array.isArray(page?.frontmatter?.authors) ? (page?.frontmatter?.authors as Array<Author>) : [];
}

export default function useDocPages(): {
  pages: Array<DocPageData>;
} {
  const { pages } = usePages();
  return {
    pages: pages.map(page => ({
      authors: buildAuthors(page),
      ...page,
    })),
  };
}
