// Strips MDX/markdown artifacts from search index content. The built-in
// indexer uses remark (not MDX), so JSX tags and markdown syntax leak through.

import type { RspressPlugin } from '@rspress/shared';

// this is an absolute nightmare but it helps a ton. Ideally I should take the work that react-email or email.md has
// done for converting markdown to pure text and then use that to generate the search index entirely rather than however
// rspress is doing it here (which seems to be generating the index from the markdown content directly).
function cleanContent(text: string): string {
  return text
    .replace(/\{[^{}]*\}/g, '') // JSX expressions
    .replace(/!\[([^\]]*)\]\([^)]*\)/g, '$1') // images → alt text (before links)
    .replace(/\[([^\]]*)\]\([^)]*\)/g, '$1') // links → text
    .replace(/\*{2}/g, '') // ** markers around bold
    .replace(/^#{1,6}\s+/gm, '') // heading markers
    .replace(/^\s*\w+="[^"]*"\s*$/gm, '') // stray JSX attributes
    .replace(/\n{3,}/g, '\n\n') // collapse blank lines
    .trim();
}

export default function pluginSearchIndexCleanup(): RspressPlugin {
  return {
    name: 'plugin-search-index-cleanup',
    modifySearchIndexData(pages) {
      for (const page of pages) {
        if (page.content) {
          page.content = cleanContent(page.content);
        }
      }
    },
  };
}
