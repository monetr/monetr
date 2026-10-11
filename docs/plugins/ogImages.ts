// Rspress plugin that generates an og:image for every page that doesn't already have one in its frontmatter.
//
// The image is rendered from the page's title and description with satori, then turned into a PNG with resvg. Pages get
// an `ogImage` field on their page data (not the frontmatter, the runtime reads frontmatter from the MDX module so
// changing it here does nothing) and PageMetadata falls back to that when there is no frontmatter ogImage.
//
// Images are only generated for production builds. They get emitted as assets of the web build, the same way rspress
// emits its own search index.

import type { ReactNode } from 'react';

import { readFileSync } from 'node:fs';
import path from 'node:path';
import { Resvg } from '@resvg/resvg-js';
import type { RsbuildPlugin } from '@rsbuild/core';
import type { RspressPlugin } from '@rspress/core';
import satori from 'satori';

const WIDTH = 1200;
const HEIGHT = 630;

interface OgImage {
  path: string;
  title: string;
  description: string;
}

export default function ogImages(): RspressPlugin {
  // Keyed by the image path so if page data gets extended more than once we just overwrite.
  const images = new Map<string, OgImage>();

  return {
    name: 'monetr-og-images',
    async extendPageData(pageData, isProd) {
      if (!isProd) {
        return;
      }
      if (pageData.frontmatter?.ogImage || pageData.frontmatter?.pageType === '404') {
        return;
      }

      // Custom pages put "- monetr" in their frontmatter title, the image already says monetr.app so drop it.
      const title = ((pageData.frontmatter?.title as string | undefined) ?? pageData.title)?.replace(/ - monetr$/, '');
      if (!title) {
        return;
      }

      // Routes are like /documentation/configure/cors, the homepage is just / so give it a name.
      const imagePath = `/og${pageData.routePath === '/' ? '/index' : pageData.routePath}.png`;
      images.set(imagePath, {
        path: imagePath,
        title,
        description: pageData.description ?? '',
      });
      Object.assign(pageData, { ogImage: imagePath });
    },
    builderConfig: {
      plugins: [
        {
          name: 'monetr-og-images-emit',
          setup(api) {
            // Page data is extended before the compile starts, so by the time assets are processed the map is full.
            // Only the web environment though, otherwise the images would end up in the SSR bundle too.
            api.processAssets({ stage: 'additional', targets: ['web'] }, async ({ compilation, sources }) => {
              if (images.size === 0) {
                return;
              }
              const fonts = loadFonts();
              const logo = readFileSync(path.resolve(__dirname, '../src/public/logo192.png')).toString('base64');
              await Promise.all(
                Array.from(images.values()).map(async image => {
                  const png = await renderImage(image, fonts, logo);
                  // Asset names are relative to the output directory, no leading slash.
                  compilation.emitAsset(image.path.slice(1), new sources.RawSource(png));
                }),
              );
            });
          },
        } satisfies RsbuildPlugin,
      ],
    },
  };
}

// Satori can't read woff2, so this uses the static woff files from @fontsource/inter instead of the variable font the
// site itself uses.
function loadFonts(): Array<{ name: string; data: Buffer; weight: 400 | 600 | 700; style: 'normal' }> {
  const fontsDir = path.resolve(__dirname, '../node_modules/@fontsource/inter/files');
  return ([400, 600, 700] as const).map(weight => ({
    name: 'Inter',
    data: readFileSync(path.join(fontsDir, `inter-latin-${weight}-normal.woff`)),
    weight,
    style: 'normal',
  }));
}

async function renderImage(image: OgImage, fonts: ReturnType<typeof loadFonts>, logoBase64: string): Promise<Buffer> {
  const svg = await satori(
    // Satori takes plain objects shaped like react elements just fine, its types just only allow actual react nodes.
    {
      type: 'div',
      props: {
        style: {
          width: '100%',
          height: '100%',
          display: 'flex',
          flexDirection: 'column',
          justifyContent: 'space-between',
          padding: '96px',
          background: '#19161f',
          color: '#fafafa',
          fontFamily: 'Inter',
        },
        children: [
          {
            type: 'div',
            props: {
              style: {
                display: 'flex',
                alignItems: 'center',
                gap: '24px',
              },
              children: [
                {
                  type: 'div',
                  props: {
                    style: {
                      width: 96,
                      height: 96,
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      overflow: 'hidden',
                    },
                    children: [
                      {
                        type: 'img',
                        props: {
                          src: `data:image/png;base64,${logoBase64}`,
                          style: {
                            width: '100%',
                            height: '100%',
                            display: 'block',
                            objectFit: 'contain',
                            objectPosition: 'center',
                          },
                        },
                      },
                    ],
                  },
                },
                {
                  type: 'span',
                  props: {
                    style: {
                      fontSize: 40,
                      fontWeight: 600,
                    },
                    children: 'monetr.app',
                  },
                },
              ],
            },
          },
          {
            type: 'div',
            props: {
              style: {
                display: 'flex',
                flexDirection: 'column',
              },
              children: [
                {
                  type: 'div',
                  props: {
                    style: {
                      display: 'flex',
                      fontSize: 64,
                      fontWeight: 700,
                      lineHeight: 1.1,
                      marginBottom: '24px',
                    },
                    children: image.title,
                  },
                },
                {
                  type: 'div',
                  props: {
                    style: {
                      display: 'flex',
                      fontSize: 32,
                      opacity: 0.85,
                    },
                    children: image.description,
                  },
                },
              ],
            },
          },
        ],
      },
    } as unknown as ReactNode,
    {
      width: WIDTH,
      height: HEIGHT,
      fonts,
    },
  );

  const resvg = new Resvg(svg, {
    fitTo: {
      mode: 'width',
      value: WIDTH,
    },
  });

  return resvg.render().asPng();
}
