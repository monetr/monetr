import { createElement } from 'react';

import { toPlainText } from '../toPlainText';

import { mkdir, rm, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
import type { RsbuildPlugin } from '@rsbuild/core';
import { load as loadHtml } from 'cheerio';
import juice from 'juice';
import { renderToStaticMarkup } from 'react-dom/server';

const SSG_BUNDLE_FOLDER = '__email_ssg__';
const SSG_BUNDLE_NAME = 'email-bundle.cjs';

const XHTML_DOCTYPE =
  '<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">';

export interface EmailPluginOptions {
  /** Directory to write the final .html and .txt files. */
  outDir: string;
}

function inlineCSS(html: string, css: string): string {
  return juice.inlineContent(html, css, {
    preserveMediaQueries: true,
    preserveFontFaces: true,
    preserveKeyFrames: true,
    applyStyleTags: true,
    removeStyleTags: false,
    preserveImportant: true,
  });
}

// After juice inlines styles, CSS-module class names serve no purpose in the
// final email. Strip them so clients see leaner markup.
function stripClassAttributes(html: string): string {
  const $ = loadHtml(html);
  $('[class]').removeAttr('class');
  return $.html();
}

// Renders email templates to static HTML+text at build time, following the
// same processAssets pattern as rspress's rsbuildPluginSSG.
export const rsbuildPluginEmail = ({ outDir }: EmailPluginOptions): RsbuildPlugin => ({
  name: 'monetr-rsbuild-plugin-email',
  async setup(api) {
    api.onBeforeBuild(() => {
      api.processAssets(
        { stage: 'optimize-transfer', environments: ['node'] },
        async ({ assets, compilation, environment }) => {
          if (compilation.errors.length > 0) {
            return;
          }

          const distPath = environment.distPath;
          const ssgFolderPath = join(distPath, SSG_BUNDLE_FOLDER);
          const bundleAbsolutePath = join(ssgFolderPath, SSG_BUNDLE_NAME);

          // Collect compiled CSS from the build output
          let css = '';
          for (const [assetName, assetSource] of Object.entries(assets)) {
            if (assetName.endsWith('.css')) {
              css += `${assetSource.source().toString()}\n`;
              compilation.deleteAsset(assetName);
            }
          }

          await mkdir(ssgFolderPath, { recursive: true });
          await Promise.all(
            Object.entries(assets).map(async ([assetName, assetSource]) => {
              if (assetName.startsWith(`${SSG_BUNDLE_FOLDER}/`)) {
                const fileAbsolutePath = join(distPath, assetName);
                await writeFile(fileAbsolutePath, assetSource.source().toString());
                compilation.deleteAsset(assetName);
              }
            }),
          );

          // Node.js wraps CJS module.exports as `default` when using import()
          const rawModule = await import(pathToFileURL(bundleAbsolutePath).href);
          const bundleExports = rawModule.default || rawModule;
          const templates: Record<string, React.ComponentType> = bundleExports.templates;

          await mkdir(outDir, { recursive: true });

          const templateNames = Object.keys(templates);
          console.log(`[email] Found ${templateNames.length} templates: ${templateNames.join(', ')}`);

          for (const name of templateNames) {
            const Component = templates[name];
            console.log(`[email] Building ${name}...`);

            // Default props contain Go template placeholders for production
            let html = renderToStaticMarkup(createElement(Component));
            html = inlineCSS(html, css);
            html = stripClassAttributes(html);
            html = `${XHTML_DOCTYPE}\n${html}`;
            const plaintext = toPlainText(html);
            const htmlPath = join(outDir, `${name}.html`);
            const txtPath = join(outDir, `${name}.txt`);
            await writeFile(htmlPath, html);
            await writeFile(txtPath, plaintext);

            console.log(`[email]   -> ${htmlPath}`);
            console.log(`[email]   -> ${txtPath}`);
          }

          await rm(ssgFolderPath, { recursive: true }).catch(() => {});

          console.log('[email] Done!');
        },
      );
    });
  },
});
