import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { mkdir, writeFile, rm } from 'node:fs/promises';
import { join } from 'node:path';
import { pathToFileURL } from 'node:url';
import postcss from 'postcss';
import tailwindcss from 'tailwindcss';
import juice from 'juice';
import { convert as htmlToText } from 'html-to-text';
import type { RsbuildPlugin } from '@rsbuild/core';

const SSG_BUNDLE_FOLDER = '__email_ssg__';
const SSG_BUNDLE_NAME = 'email-bundle.cjs';

const XHTML_DOCTYPE =
  '<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">';

export interface EmailPluginOptions {
  /** Directory to write the final .html and .txt files. */
  outDir: string;
  /** Already-loaded Tailwind config object. */
  tailwindConfig: any;
}

/**
 * Generate Tailwind CSS for the given HTML content.
 * Uses only components + utilities layers (no base/preflight) to avoid
 * bloating every element with CSS variable resets that email clients ignore.
 */
async function generateTailwindCSS(html: string, tailwindConfig: any): Promise<string> {
  const configWithContent = {
    ...tailwindConfig,
    content: [{ raw: html, extension: 'html' }],
  };

  const inputCSS = '@tailwind components;\n@tailwind utilities;';
  const result = await postcss([
    tailwindcss(configWithContent),
  ]).process(inputCSS, { from: undefined });

  return result.css;
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
 * Resolve Tailwind CSS custom properties that email clients don't support.
 * Replaces `var(--tw-*, fallback)` with the fallback value, and simplifies
 * resulting color functions like `rgb(78 26 160 / 1)` to `rgb(78 26 160)`.
 */
function resolveCSSVariables(html: string): string {
  let result = html.replace(/var\(--tw-[\w-]+,\s*([^)]+)\)/g, '$1');
  result = result.replace(/rgb\((\d+\s+\d+\s+\d+)\s*\/\s*1\)/g, 'rgb($1)');
  result = result.replace(/--tw-[\w-]+:\s*[^;]+;\s*/g, '');
  return result;
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
 * 2. Write the compiled bundle to disk
 * 3. Dynamically import it
 * 4. Render each exported component to HTML
 * 5. Post-process: Tailwind CSS → juice inline → CSS var resolution → plaintext
 * 6. Write .html and .txt files to the output directory
 */
export const rsbuildPluginEmail = ({
  outDir,
  tailwindConfig,
}: EmailPluginOptions): RsbuildPlugin => ({
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

            // 2. Generate Tailwind CSS from the rendered HTML
            const css = await generateTailwindCSS(html, tailwindConfig);

            // 3. Inline CSS
            html = inlineCSS(html, css);

            // 4. Resolve CSS custom properties
            html = resolveCSSVariables(html);

            // 5. Prepend XHTML doctype
            html = `${XHTML_DOCTYPE}\n${html}`;

            // 6. Generate plaintext
            const plaintext = toPlainText(html);

            // 7. Emit as rsbuild assets (they'll be written to the dist folder)
            // Also write directly to outDir for CMake integration
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
