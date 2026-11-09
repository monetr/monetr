// Rehype plugin that swaps large local images for a smaller copy, so the page isn't downloading a 2560 pixel wide
// screenshot to show it in a column that is a fraction of that. The original stays around as data-zoom-src, which
// medium-zoom loads when someone clicks on the image.
//
// This needs to run before rehypeImageSize. That plugin then measures the smaller copy, which has the same aspect ratio
// as the original, so the browser still reserves the right amount of space.
//
// Only images that rspress has turned into imports are scaled, those are the ones that the bundler gives a hashed
// filename. The scaled copies are cached on disk, named by a hash of the original, and then imported like any other
// image so they get a hashed filename too.

import { getImports, isSrcAttribute, type MdxJsxAttribute, type MdxJsxElement, type MdxjsEsm } from './rehypeImageSize';

import { createHash } from 'node:crypto';
import { existsSync, readFileSync } from 'node:fs';
import { mkdir, rename, writeFile } from 'node:fs/promises';
import path from 'node:path';
import type { Root } from 'hast';
import { imageSize } from 'image-size';
import sharp from 'sharp';
import { visit } from 'unist-util-visit';

// Screenshots render around 824 pixels wide on a desktop sized screen, anything wider than this gets scaled down to it.
const SCALED_WIDTH = 900;
const CACHE_DIR = path.resolve(__dirname, '../node_modules/.cache/monetr-image-scale');

// Pages get compiled for both the web and the server builds at the same time, this keeps them from doing the work twice.
const pending = new Map<string, Promise<string>>();

async function writeScaledImage(original: Buffer, scaledPath: string): Promise<string> {
  if (existsSync(scaledPath)) {
    return scaledPath;
  }

  await mkdir(path.dirname(scaledPath), { recursive: true });
  const scaled = await sharp(original).resize({ width: SCALED_WIDTH }).png({ compressionLevel: 9 }).toBuffer();
  // Rename into place so a build that gets killed halfway through can't leave a broken image in the cache.
  const temporaryPath = `${scaledPath}.tmp`;
  await writeFile(temporaryPath, scaled);
  await rename(temporaryPath, scaledPath);
  return scaledPath;
}

function scaleImage(imagePath: string, original: Buffer): Promise<string> {
  const hash = createHash('sha256').update(String(SCALED_WIDTH)).update(original).digest('hex').slice(0, 16);
  const scaledPath = path.join(CACHE_DIR, hash, `${path.parse(imagePath).name}.scaled.png`);

  const existing = pending.get(scaledPath);
  if (existing) {
    return existing;
  }
  const result = writeScaledImage(original, scaledPath);
  pending.set(scaledPath, result);
  return result;
}

function getImportNode(name: string, from: string): MdxjsEsm {
  return {
    type: 'mdxjsEsm',
    value: `import ${name} from ${JSON.stringify(from)}`,
    data: {
      estree: {
        type: 'Program',
        sourceType: 'module',
        body: [
          {
            type: 'ImportDeclaration',
            specifiers: [{ type: 'ImportDefaultSpecifier', local: { type: 'Identifier', name } }],
            source: { type: 'Literal', value: from, raw: JSON.stringify(from) },
          },
        ],
      },
    },
  };
}

function getExpressionAttribute(name: string, identifier: string): MdxJsxAttribute {
  return {
    type: 'mdxJsxAttribute',
    name,
    value: {
      type: 'mdxJsxAttributeValueExpression',
      value: identifier,
      data: {
        estree: {
          type: 'Program',
          sourceType: 'module',
          body: [{ type: 'ExpressionStatement', expression: { type: 'Identifier', name: identifier } }],
        },
      },
    },
  };
}

export default function rehypeImageScale() {
  return async (tree: Root, file: { path: string }) => {
    const imports = getImports(tree, file.path);

    const images: Array<{ element: MdxJsxElement; originalName: string; imagePath: string }> = [];
    visit(tree, node => {
      if (node.type !== 'mdxJsxFlowElement' && node.type !== 'mdxJsxTextElement') {
        return;
      }
      if (node.name !== 'img') {
        return;
      }
      const src = node.attributes.find(isSrcAttribute)?.value;
      // Only an expression can point at an import
      if (!src || typeof src === 'string') {
        return;
      }
      const imagePath = imports.get(src.value);
      // rspress's dead image check will already complain about one that doesn't exist
      if (imagePath && existsSync(imagePath)) {
        images.push({ element: node, originalName: src.value, imagePath });
      }
    });

    await Promise.all(
      images.map(async ({ element, originalName, imagePath }) => {
        const original = readFileSync(imagePath);
        const { width } = imageSize(original);
        if (!width || width <= SCALED_WIDTH) {
          return;
        }

        const scaledPath = await scaleImage(imagePath, original);
        const scaledName = `${originalName}Scaled`;
        if (!imports.has(scaledName)) {
          imports.set(scaledName, scaledPath);
          tree.children.unshift(getImportNode(scaledName, scaledPath));
        }

        element.attributes = [
          ...element.attributes.filter(attribute => !isSrcAttribute(attribute)),
          getExpressionAttribute('src', scaledName),
          getExpressionAttribute('data-zoom-src', originalName),
        ];
      }),
    );
  };
}
