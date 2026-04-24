import { usePages } from '@rspress/core/runtime';
import type { BaseRuntimePageInfo } from '@rspress/shared';

export interface Author {
  name: string;
  github?: string;
}

export type DocPageData = {
  authors: Array<Author>;
} & BaseRuntimePageInfo;

function isAuthor(value: unknown): value is Author {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    return false;
  }

  const author = value as Record<string, unknown>;
  return (
    typeof author.name === 'string' &&
    author.name.trim().length > 0 &&
    (author.github === undefined || typeof author.github === 'string')
  );
}

export function buildAuthors(page: Pick<BaseRuntimePageInfo, 'frontmatter'>): Array<Author> {
  return Array.isArray(page?.frontmatter?.authors) ? page.frontmatter.authors.filter(isAuthor) : [];
}

export default function useDocPages(): {
  pages: Array<DocPageData>;
} {
  const { pages } = usePages();
  return {
    pages: pages.map(page => ({
      ...page,
      authors: buildAuthors(page),
    })),
  };
}
