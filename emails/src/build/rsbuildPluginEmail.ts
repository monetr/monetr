import { createElement } from 'react';

import { mkdir, rm, writeFile } from 'node:fs/promises';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
import type { RsbuildPlugin } from '@rsbuild/core';
import { convert as htmlToText } from 'html-to-text';
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

/**
 * Inline CSS into HTML using juice.
 */
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

/**
 * Convert HTML to plaintext.
 */
function toPlainText(html: string): string {
  return htmlToText(html, {
    selectors: [
      { selector: 'img', format: 'skip' },
      { selector: '[data-skip-in-text]', format: 'skip' },
      {
        selector: 'a',
        options: {
          hideLinkHrefIfSameAsText: true,
        },
      },
    ],
  });
}

/**
 * rsbuild plugin that renders email templates from the node environment bundle.
 *
 * Follows the same pattern as rspress's rsbuildPluginSSG:
 * 1. Hook into processAssets on the node environment
 * 2. Write the compiled JS bundle to disk, capture the compiled CSS
 * 3. Dynamically import the JS bundle
 * 4. Render each exported component to HTML
 * 5. Post-process: juice CSS inlining → plaintext generation
 * 6. Write .html and .txt files to the output directory
 */
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

          // Collect compiled CSS from the build output (SCSS modules → extracted CSS)
          let css = '';
          for (const [assetName, assetSource] of Object.entries(assets)) {
            if (assetName.endsWith('.css')) {
              css += assetSource.source().toString() + '\n';
              compilation.deleteAsset(assetName);
            }
          }

          // Write the compiled node bundle(s) to disk so we can import them
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

          // Import the bundle to get the template registry.
          // Node.js wraps CJS module.exports as the `default` export when using import().
          const rawModule = await import(pathToFileURL(bundleAbsolutePath).href);
          const bundleExports = rawModule.default || rawModule;
          const templates: Record<string, React.ComponentType> = bundleExports.templates;

          // Create output directory
          await mkdir(outDir, { recursive: true });

          const templateNames = Object.keys(templates);
          console.log(`[email] Found ${templateNames.length} templates: ${templateNames.join(', ')}`);

          for (const name of templateNames) {
            const Component = templates[name];
            console.log(`[email] Building ${name}...`);

            // 1. Render to HTML (uses default props which contain Go template placeholders)
            let html = renderToStaticMarkup(createElement(Component));

            // 2. Inline the compiled SCSS module CSS
            html = inlineCSS(html, css);

            // 3. Prepend XHTML doctype
            html = `${XHTML_DOCTYPE}\n${html}`;

            // 4. Generate plaintext
            const plaintext = toPlainText(html);

            // 5. Write output files
            const htmlPath = join(outDir, `${name}.html`);
            const txtPath = join(outDir, `${name}.txt`);
            await writeFile(htmlPath, html);
            await writeFile(txtPath, plaintext);

            console.log(`[email]   -> ${htmlPath}`);
            console.log(`[email]   -> ${txtPath}`);
          }

          // Clean up the SSG bundle folder
          await rm(ssgFolderPath, { recursive: true }).catch(() => {});

          console.log('[email] Done!');
        },
      );
    });
  },
});
