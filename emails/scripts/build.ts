import { createElement } from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import { buildSync } from 'esbuild';
import postcss from 'postcss';
import tailwindcss from 'tailwindcss';
import juice from 'juice';
import { convert as htmlToText } from 'html-to-text';
import { writeFileSync, mkdirSync, readdirSync, statSync, rmSync } from 'node:fs';
import { join, resolve, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createRequire } from 'node:module';

const __dirname = dirname(fileURLToPath(import.meta.url));
const emailsRoot = resolve(__dirname, '..');
const srcDir = join(emailsRoot, 'src', 'emails');
const tailwindConfigPath = join(emailsRoot, 'tailwind.config.ts');

// Parse --outDir argument
const outDirIndex = process.argv.indexOf('--outDir');
const outDir = outDirIndex !== -1 && process.argv[outDirIndex + 1]
  ? resolve(process.argv[outDirIndex + 1])
  : join(emailsRoot, 'dist');

const XHTML_DOCTYPE =
  '<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">';

/**
 * Discover all email template directories (each contains an index.tsx).
 */
function discoverTemplates(): { name: string; entryPath: string }[] {
  const entries = readdirSync(srcDir);
  const templates: { name: string; entryPath: string }[] = [];
  for (const entry of entries) {
    const entryPath = join(srcDir, entry, 'index.tsx');
    try {
      statSync(entryPath);
      templates.push({ name: entry, entryPath });
    } catch {
      // Not a template directory
    }
  }
  return templates;
}

/**
 * Bundle a template's TSX to a temporary CJS file using esbuild,
 * then require() it to get the component.
 */
function bundleAndLoadComponent(entryPath: string): React.ComponentType {
  const tmpDir = join(emailsRoot, '.build-tmp');
  mkdirSync(tmpDir, { recursive: true });

  const outfile = join(tmpDir, `template-${Date.now()}.cjs`);

  buildSync({
    entryPoints: [entryPath],
    bundle: true,
    platform: 'node',
    format: 'cjs',
    outfile,
    jsx: 'automatic',
    // Externalize react so we use the installed version
    external: ['react', 'react-dom'],
    logLevel: 'error',
    // Resolve .ts/.tsx imports with extensions
    resolveExtensions: ['.tsx', '.ts', '.jsx', '.js'],
  });

  const require = createRequire(import.meta.url);
  // Clear the require cache in case of multiple builds
  delete require.cache[outfile];
  const mod = require(outfile);
  const Component = mod.default || mod;

  // Clean up the temp file
  try { rmSync(outfile); } catch { /* ignore */ }

  return Component;
}

/**
 * Render a React component to an HTML string using default props
 * (which contain Go template placeholders).
 */
function renderTemplate(Component: React.ComponentType): string {
  return renderToStaticMarkup(createElement(Component));
}

/**
 * Generate Tailwind CSS for the given HTML content.
 */
async function generateTailwindCSS(html: string): Promise<string> {
  // Load the tailwind config by bundling it (it's a .ts file)
  const tmpDir = join(emailsRoot, '.build-tmp');
  mkdirSync(tmpDir, { recursive: true });
  const configOutfile = join(tmpDir, 'tailwind.config.cjs');

  buildSync({
    entryPoints: [tailwindConfigPath],
    bundle: true,
    platform: 'node',
    format: 'cjs',
    outfile: configOutfile,
    external: ['tailwindcss'],
    logLevel: 'error',
  });

  const require = createRequire(import.meta.url);
  delete require.cache[configOutfile];
  const tailwindConfig = require(configOutfile);
  const config = tailwindConfig.default || tailwindConfig;

  try { rmSync(configOutfile); } catch { /* ignore */ }

  // Override the content to use the raw HTML string instead of file globs
  const configWithContent = {
    ...config,
    content: [{ raw: html, extension: 'html' }],
  };

  // Only generate utilities and components — skip base/preflight to avoid
  // bloating every element with CSS variable resets that email clients ignore anyway.
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
  // Replace var(--tw-..., fallback) with the fallback value
  let result = html.replace(/var\(--tw-[\w-]+,\s*([^)]+)\)/g, '$1');

  // Simplify `rgb(R G B / 1)` to `rgb(R G B)` for cleaner output
  result = result.replace(/rgb\((\d+\s+\d+\s+\d+)\s*\/\s*1\)/g, 'rgb($1)');

  // Remove leftover --tw-* custom property declarations from inline styles
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
 * Main build pipeline.
 */
async function main() {
  console.log('Discovering email templates...');
  const templates = discoverTemplates();
  console.log(`Found ${templates.length} templates: ${templates.map(t => t.name).join(', ')}`);

  mkdirSync(outDir, { recursive: true });

  for (const template of templates) {
    console.log(`Building ${template.name}...`);

    // 1. Bundle and load the component
    const Component = bundleAndLoadComponent(template.entryPath);

    // 2. Render to HTML
    let html = renderTemplate(Component);

    // 3. Generate Tailwind CSS from the rendered HTML
    const css = await generateTailwindCSS(html);

    // 4. Inline CSS
    html = inlineCSS(html, css);

    // 5. Resolve CSS custom properties (email clients don't support var())
    html = resolveCSSVariables(html);

    // 6. Prepend doctype
    html = `${XHTML_DOCTYPE}\n${html}`;

    // 7. Generate plaintext
    const plaintext = toPlainText(html);

    // 8. Write output files
    const htmlPath = join(outDir, `${template.name}.html`);
    const txtPath = join(outDir, `${template.name}.txt`);
    writeFileSync(htmlPath, html);
    writeFileSync(txtPath, plaintext);

    console.log(`  -> ${htmlPath}`);
    console.log(`  -> ${txtPath}`);
  }

  // Clean up temp directory
  const tmpDir = join(emailsRoot, '.build-tmp');
  try { rmSync(tmpDir, { recursive: true }); } catch { /* ignore */ }

  console.log('Done!');
}

main().catch((err) => {
  console.error('Build failed:', err);
  process.exit(1);
});
