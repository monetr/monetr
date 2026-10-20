// Rehype plugin that sets width and height on local images at build time. rspress leaves images with absolute paths as
// a plain <img src> so the browser has no idea how big they are until they load, and the page shifts around as they
// come in. With width and height set the browser can reserve the space using the aspect ratio, and the max-width and
// height: auto from rspress's preflight css still scale it down to fit.

import { existsSync, readFileSync } from 'node:fs';
import path from 'node:path';
import type { Root } from 'hast';
import { imageSize } from 'image-size';
import { visit } from 'unist-util-visit';

interface RehypeImageSizeOptions {
  // publicDir is where absolute image paths like /v1/en/... are served from.
  publicDir: string;
}

export default function rehypeImageSize(options: RehypeImageSizeOptions) {
  return (tree: Root) => {
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

      const imagePath = path.join(options.publicDir, src.split(/[?#]/)[0]);
      // rspress's dead image check will already complain about this one
      if (!existsSync(imagePath)) {
        return;
      }

      const { width, height } = imageSize(readFileSync(imagePath));
      if (!width || !height) {
        return;
      }
      node.properties.width = width;
      node.properties.height = height;
    });
  };
}
