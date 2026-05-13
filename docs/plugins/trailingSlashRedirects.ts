// Rspress plugin that emits trailing-slash redirect stubs for GitHub Pages.
//
// Background: monetr's docs were Next.js with trailingSlash: true, so Google indexed URLs like /documentation/install/.
// The site is now Rspress with cleanUrls: true, which emits documentation/install.html. GitHub Pages serves that for
// /documentation/install but 404s on the slash form.
//
// For every clean-URL HTML page in the build output, this plugin writes a sibling <basename>/index.html redirect stub
// that bounces to the canonical clean URL. The stub also carries rel="canonical" and noindex so search engines
// consolidate ranking on the clean URL and drop the slash variant.

import { mkdir, readdir, stat, writeFile } from 'node:fs/promises';
import path from 'node:path';
import type { RspressPlugin } from '@rspress/core';

interface TrailingSlashRedirectsOptions {
  siteUrl: string;
}

export default function trailingSlashRedirects(options: TrailingSlashRedirectsOptions): RspressPlugin {
  if (!options?.siteUrl) {
    throw new Error('[trailing-slash-redirects] siteUrl option is required');
  }
  // Strip any trailing slash so concatenation always yields exactly one.
  const siteUrl = options.siteUrl.replace(/\/+$/, '');

  return {
    name: 'monetr-trailing-slash-redirects',
    async afterBuild(config, isProd) {
      if (!isProd) {
        return;
      }
      const outDir = path.resolve(config.outDir ?? 'doc_build');
      const count = await emitStubs(outDir, siteUrl);
      console.log(`[trailing-slash-redirects] wrote ${count} stubs`);
    },
  };
}

async function emitStubs(outDir: string, siteUrl: string): Promise<number> {
  const entries = await readdir(outDir, { recursive: true, withFileTypes: true });
  let count = 0;
  for (const entry of entries) {
    if (!entry.isFile() || !entry.name.endsWith('.html')) {
      continue;
    }
    if (entry.name === 'index.html' || entry.name === '404.html') {
      continue;
    }
    // Dirent.parentPath was added in Node 20.12; older 20.x exposes the same value under .path. Fall back so the plugin
    // survives either runtime.
    const parentDir =
      (entry as { parentPath?: string; path?: string }).parentPath ??
      (entry as { parentPath?: string; path?: string }).path ??
      outDir;
    const relDir = path.relative(outDir, parentDir);
    const base = entry.name.slice(0, -'.html'.length);
    const segments = relDir === '' ? [base] : [...relDir.split(path.sep), base];
    const routePath = segments.join('/');
    const stubDir = path.join(outDir, ...segments);
    const stubFile = path.join(stubDir, 'index.html');
    if (await fileExists(stubFile)) {
      // The route is already directory-style; leave it alone.
      continue;
    }
    const absoluteUrl = `${siteUrl}/${routePath}`;
    await mkdir(stubDir, { recursive: true });
    await writeFile(stubFile, renderStub(absoluteUrl), 'utf8');
    count += 1;
  }
  return count;
}

async function fileExists(p: string): Promise<boolean> {
  try {
    const s = await stat(p);
    return s.isFile();
  } catch {
    return false;
  }
}

// Route paths derive from rspress-generated .html filenames, which in this project are ASCII slugs joined with /. No
// HTML escaping needed.
function renderStub(absoluteUrl: string): string {
  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<title>Redirecting</title>
<link rel="canonical" href="${absoluteUrl}">
<meta name="robots" content="noindex">
<meta http-equiv="refresh" content="0; url=${absoluteUrl}">
</head>
<body>
<p>Redirecting to <a href="${absoluteUrl}">${absoluteUrl}</a>.</p>
</body>
</html>
`;
}
