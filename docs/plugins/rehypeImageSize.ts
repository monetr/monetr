// Rehype plugin that sets width and height on local images at build time. Without them the browser has no idea how big
// an image is until it loads, and the page shifts around as they come in. With width and height set the browser can
// reserve the space using the aspect ratio, and the max-width and height: auto from rspress's preflight css still scale
// it down to fit.
//
// Local images reach this plugin in one of two shapes. Absolute paths are left as a plain <img src> and are served from
// the public directory. Relative paths have already been rewritten by rspress into an import at the top of the file and
// an <img src={image0}> JSX element that uses it, that is what lets the bundler give them a hashed filename.

import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import type { Root, RootContent } from 'hast';
import { imageSize } from 'image-size';
import { visit } from 'unist-util-visit';

interface RehypeImageSizeOptions {
  // publicDir is where absolute image paths like /logo.svg are served from.
  publicDir: string;
}

// The hast types don't know about this node, so these are just the parts of it that we use.
export interface MdxjsEsm {
  type: 'mdxjsEsm';
  value: string;
  data?: {
    estree?: {
      type: string;
      sourceType: string;
      body: Array<{
        type: string;
        source?: { type: string; value: string; raw: string };
        specifiers?: Array<{ type: string; local: { type: string; name: string } }>;
      }>;
    };
  };
}

declare module 'hast' {
  interface RootContentMap {
    mdxjsEsm: MdxjsEsm;
  }
}

// The JSX nodes are already part of the hast types, rspress pulls in the package that adds them.
export type MdxJsxElement = Extract<RootContent, { type: 'mdxJsxFlowElement' | 'mdxJsxTextElement' }>;
export type MdxJsxAttribute = Extract<MdxJsxElement['attributes'][number], { type: 'mdxJsxAttribute' }>;

function isSizeAttribute(attribute: MdxJsxElement['attributes'][number]): boolean {
  return attribute.type === 'mdxJsxAttribute' && (attribute.name === 'width' || attribute.name === 'height');
}

export function isSrcAttribute(attribute: MdxJsxElement['attributes'][number]): boolean {
  return attribute.type === 'mdxJsxAttribute' && attribute.name === 'src';
}

function getImageSize(imagePath: string): { width: number; height: number } | null {
  // rspress's dead image check will already complain about this one
  if (!existsSync(imagePath)) {
    return null;
  }

  const { width, height } = imageSize(readFileSync(imagePath));
  if (!width || !height) {
    return null;
  }
  return { width, height };
}

// getImports maps the name of each default import in the file to the absolute path that it is importing.
export function getImports(tree: Root, filePath: string): Map<string, string> {
  const imports = new Map<string, string>();
  for (const node of tree.children) {
    if (node.type !== 'mdxjsEsm') {
      continue;
    }
    for (const statement of node.data?.estree?.body ?? []) {
      if (statement.type !== 'ImportDeclaration' || !statement.source) {
        continue;
      }
      const importPath = path.resolve(path.dirname(filePath), statement.source.value);
      for (const specifier of statement.specifiers ?? []) {
        if (specifier.type === 'ImportDefaultSpecifier') {
          imports.set(specifier.local.name, importPath);
        }
      }
    }
  }
  return imports;
}

export default function rehypeImageSize(options: RehypeImageSizeOptions) {
  return (tree: Root, file: { path: string }) => {
    visit(tree, 'element', node => {
      if (node.tagName !== 'img') {
        return;
      }
      // Don't override anything that was sized by hand
      if (node.properties.width || node.properties.height) {
        return;
      }
      const src = node.properties.src;
      // Only local images, external ones would need to be fetched during the build
      if (typeof src !== 'string' || !src.startsWith('/') || src.startsWith('//')) {
        return;
      }

      const size = getImageSize(path.join(options.publicDir, src.split(/[?#]/)[0]));
      if (!size) {
        return;
      }
      node.properties.width = size.width;
      node.properties.height = size.height;
    });

    const imports = getImports(tree, file.path);
    visit(tree, node => {
      if (node.type !== 'mdxJsxFlowElement' && node.type !== 'mdxJsxTextElement') {
        return;
      }
      if (node.name !== 'img') {
        return;
      }
      // Don't override anything that was sized by hand
      if (node.attributes.some(isSizeAttribute)) {
        return;
      }
      const src = node.attributes.find(isSrcAttribute)?.value;
      // Only an expression can point at an import, a string here would be an image we can't resolve
      if (!src || typeof src === 'string') {
        return;
      }
      const imagePath = imports.get(src.value);
      if (!imagePath) {
        return;
      }

      const size = getImageSize(imagePath);
      if (!size) {
        return;
      }
      node.attributes.push(
        { type: 'mdxJsxAttribute', name: 'width', value: String(size.width) },
        { type: 'mdxJsxAttribute', name: 'height', value: String(size.height) },
      );
    });
  };
}
