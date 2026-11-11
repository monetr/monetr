import path from 'node:path';
import { promisify } from 'node:util';
import type { RsbuildPlugin, Rspack } from '@rsbuild/core';
import sharp, { type Color } from 'sharp';

const transparent: Color = { r: 0, g: 0, b: 0, alpha: 0 };

type PluginPWAOptions = {
  logo: string;
  background: string;
  quality?: number;
};

const appleTouchIconName = 'apple-touch-icon.png';
const appleTouchIconPrecomposedName = 'apple-touch-icon-precomposed.png';

// mstileIcons provides all the information needed to generate the icons used by internet explorer. I still need to add
// support for microsoft edge though.
function mstileIcons(): Array<{ name: string; asset: string; width: number; height: number }> {
  const mstileSizes: Array<[number, number]> = [
    [70, 70],
    [128, 128],
    [150, 150],
    [270, 270],
    [310, 150],
    [310, 310],
    [558, 558],
  ];
  return mstileSizes.map(([width, height]) => ({
    // For mstile icons, if the width is greater than the height, then mark it as "wide".
    name: width > height ? `msapplication-wide${width}x${height}logo` : `msapplication-square${width}x${height}logo`,
    asset: `assets/resources/mstile-icon-${width}-${height}.png`,
    width,
    height,
  }));
}

interface AppleSplashScreenSpec {
  name: string;
  asset: string;
  width: number;
  height: number;
  ratio: number;
  orientation: 'portrait' | 'landscape';
}

function appleSplashScreens(): Array<AppleSplashScreenSpec> {
  // Taken from the "iOS, iPadOS device screen dimensions" table that used to be on Apple's layout guidelines page. Apple
  // removed it in 2026, the last archived copy is at
  // https://web.archive.org/web/20260702060104/https://developer.apple.com/tutorials/data/design/human-interface-guidelines/layout.json
  const spec: Array<Omit<AppleSplashScreenSpec, 'name' | 'asset'>> = [
    { width: 2064, height: 2752, ratio: 2, orientation: 'portrait' },
    { width: 2752, height: 2064, ratio: 2, orientation: 'landscape' },
    { width: 2048, height: 2732, ratio: 2, orientation: 'portrait' },
    { width: 2732, height: 2048, ratio: 2, orientation: 'landscape' },
    { width: 1668, height: 2420, ratio: 2, orientation: 'portrait' },
    { width: 2420, height: 1668, ratio: 2, orientation: 'landscape' },
    { width: 1668, height: 2388, ratio: 2, orientation: 'portrait' },
    { width: 2388, height: 1668, ratio: 2, orientation: 'landscape' },
    { width: 1536, height: 2048, ratio: 2, orientation: 'portrait' },
    { width: 2048, height: 1536, ratio: 2, orientation: 'landscape' },
    { width: 1488, height: 2266, ratio: 2, orientation: 'portrait' },
    { width: 2266, height: 1488, ratio: 2, orientation: 'landscape' },
    { width: 1640, height: 2360, ratio: 2, orientation: 'portrait' },
    { width: 2360, height: 1640, ratio: 2, orientation: 'landscape' },
    { width: 1668, height: 2224, ratio: 2, orientation: 'portrait' },
    { width: 2224, height: 1668, ratio: 2, orientation: 'landscape' },
    { width: 1620, height: 2160, ratio: 2, orientation: 'portrait' },
    { width: 2160, height: 1620, ratio: 2, orientation: 'landscape' },
    { width: 1320, height: 2868, ratio: 3, orientation: 'portrait' },
    { width: 2868, height: 1320, ratio: 3, orientation: 'landscape' },
    { width: 1260, height: 2736, ratio: 3, orientation: 'portrait' },
    { width: 2736, height: 1260, ratio: 3, orientation: 'landscape' },
    { width: 1206, height: 2622, ratio: 3, orientation: 'portrait' },
    { width: 2622, height: 1206, ratio: 3, orientation: 'landscape' },
    { width: 1290, height: 2796, ratio: 3, orientation: 'portrait' },
    { width: 2796, height: 1290, ratio: 3, orientation: 'landscape' },
    { width: 1179, height: 2556, ratio: 3, orientation: 'portrait' },
    { width: 2556, height: 1179, ratio: 3, orientation: 'landscape' },
    { width: 1284, height: 2778, ratio: 3, orientation: 'portrait' },
    { width: 2778, height: 1284, ratio: 3, orientation: 'landscape' },
    { width: 1170, height: 2532, ratio: 3, orientation: 'portrait' },
    { width: 2532, height: 1170, ratio: 3, orientation: 'landscape' },
    { width: 1080, height: 2340, ratio: 3, orientation: 'portrait' },
    { width: 2340, height: 1080, ratio: 3, orientation: 'landscape' },
    { width: 1125, height: 2436, ratio: 3, orientation: 'portrait' },
    { width: 2436, height: 1125, ratio: 3, orientation: 'landscape' },
    { width: 1242, height: 2688, ratio: 3, orientation: 'portrait' },
    { width: 2688, height: 1242, ratio: 3, orientation: 'landscape' },
    { width: 828, height: 1792, ratio: 2, orientation: 'portrait' },
    { width: 1792, height: 828, ratio: 2, orientation: 'landscape' },
    { width: 1242, height: 2208, ratio: 3, orientation: 'portrait' },
    { width: 2208, height: 1242, ratio: 3, orientation: 'landscape' },
    { width: 750, height: 1334, ratio: 2, orientation: 'portrait' },
    { width: 1334, height: 750, ratio: 2, orientation: 'landscape' },
    { width: 640, height: 1136, ratio: 2, orientation: 'portrait' },
    { width: 1136, height: 640, ratio: 2, orientation: 'landscape' },
  ];

  return spec.map(item => ({
    name: 'apple-touch-startup-image',
    asset: `assets/resources/apple-splash-${item.width}-${item.height}.png`,
    ...item,
  }));
}

// renderIcon scales the logo to fit within `padding` of the target size, then centers it on a canvas of the full target
// size filled with the background color.
async function renderIcon(
  logo: Buffer,
  width: number,
  height: number,
  padding: number,
  background: Color,
  compressionLevel: number,
): Promise<Buffer> {
  const scaled = await sharp(logo)
    .resize(Math.round(width * padding), Math.round(height * padding), { fit: 'inside' })
    .toBuffer();
  return sharp({ create: { width, height, channels: 4, background } })
    .composite([{ input: scaled, gravity: 'center' }])
    .png({ compressionLevel })
    .toBuffer();
}

// encodeIco packs PNG images into an ICO file. Every entry stores PNG data directly, which every browser that matters
// supports.
function encodeIco(images: Array<{ width: number; height: number; png: Buffer }>): Buffer {
  const headerSize = 6;
  const entrySize = 16;
  const header = Buffer.alloc(headerSize + entrySize * images.length);
  header.writeUInt16LE(0, 0); // Reserved
  header.writeUInt16LE(1, 2); // Type, 1 is an icon
  header.writeUInt16LE(images.length, 4);

  let offset = header.length;
  images.forEach(({ width, height, png }, index) => {
    const entry = headerSize + entrySize * index;
    // Width and height are a single byte where 0 means 256. Anything 256 or larger gets written as 0, the real size is
    // read from the PNG itself.
    header.writeUInt8(width >= 256 ? 0 : width, entry);
    header.writeUInt8(height >= 256 ? 0 : height, entry + 1);
    header.writeUInt8(0, entry + 2); // Colors in the palette, 0 means no palette
    header.writeUInt8(0, entry + 3); // Reserved
    header.writeUInt16LE(1, entry + 4); // Color planes
    header.writeUInt16LE(32, entry + 6); // Bits per pixel
    header.writeUInt32LE(png.length, entry + 8);
    header.writeUInt32LE(offset, entry + 12);
    offset += png.length;
  });

  return Buffer.concat([header, ...images.map(({ png }) => png)]);
}

export const pluginPWA = (options: PluginPWAOptions): RsbuildPlugin => ({
  name: 'pwa',

  setup(api) {
    // processAssets runs on every compilation, so in dev every hot reload would regenerate every image which is really
    // slow. The logo isn't going to change while the dev server is running, so generate everything once and then just
    // emit the same sources again on every compilation after that.
    const generated = new Map<string, Rspack.sources.Source>();

    api.processAssets({ stage: 'additional' }, async ({ compilation, sources }) => {
      if (generated.size > 0) {
        generated.forEach((source, name) => void compilation.emitAsset(name, source));
        return;
      }

      if (!compilation.inputFileSystem) {
        throw new Error("[pluginPWA] 'compilation.inputFileSystem' is not available.");
      }

      // ImageMagick treated quality on a PNG as the zlib compression level in its tens digit, so keep that meaning.
      // Passing quality to sharp's PNG encoder would instead quantize the image down to a palette.
      const compressionLevel = Math.min(9, Math.floor((options.quality ?? 90) / 10));
      const source = await promisify(compilation.inputFileSystem.readFile)(options.logo);
      if (!source) {
        throw new Error(
          `[pluginPWA] Failed to read the PWA logo file, please check if the '${options.logo}' file exists'.`,
        );
      }
      const logo = Buffer.from(source);
      const background = options.background;

      const pending = new Map<string, Promise<Buffer>>();

      appleSplashScreens().forEach(
        ({ asset, width, height }) =>
          void pending.set(asset, renderIcon(logo, width, height, 0.3, background, compressionLevel)),
      );

      // Apple touch icon
      pending.set(appleTouchIconName, renderIcon(logo, 180, 180, 0.7, background, compressionLevel));
      pending.set(appleTouchIconPrecomposedName, renderIcon(logo, 180, 180, 0.7, background, compressionLevel));

      pending.set(
        'assets/resources/transparent-128.png',
        renderIcon(logo, 128, 128, 1.0, transparent, compressionLevel),
      );

      mstileIcons().forEach(
        ({ asset, width, height }) =>
          void pending.set(asset, renderIcon(logo, width, height, 0.7, background, compressionLevel)),
      );

      const androidChromeSizes: Array<[number, number, boolean]> = [
        [192, 192, false],
        [192, 192, true],
        [512, 512, false],
        [512, 512, true],
      ];
      androidChromeSizes.forEach(
        ([width, height, maskable]) =>
          void pending.set(
            `assets/resources/android-chrome-${width}-${height}${maskable ? '_maskable' : ''}.png`,
            renderIcon(logo, width, height, 0.7, maskable ? background : transparent, compressionLevel),
          ),
      );

      const faviconSizes: Array<[number, number]> = [
        [16, 16],
        [32, 32],
        [48, 48],
        [96, 96],
        [180, 180],
        [192, 192],
        [256, 256],
        [512, 512],
      ];
      pending.set(
        'favicon.ico',
        Promise.all(
          faviconSizes.map(async ([width, height]) => ({
            width,
            height,
            png: await renderIcon(logo, width, height, 1.0, transparent, compressionLevel),
          })),
        ).then(encodeIco),
      );

      const rendered = await Promise.all(Array.from(pending, async ([name, image]) => [name, await image] as const));
      rendered.forEach(([name, image]) => void generated.set(name, new sources.RawSource(image)));

      generated.forEach((source, name) => void compilation.emitAsset(name, source));
    });

    api.modifyHTMLTags({
      order: 'post',
      handler: html => {
        mstileIcons().forEach(
          ({ name, asset }) =>
            void html.headTags.push({
              tag: 'meta',
              attrs: {
                name: name,
                content: path.join('/', asset),
              },
            }),
        );

        appleSplashScreens().forEach(
          ({ name, asset, width, height, ratio, orientation }) =>
            void html.headTags.push({
              tag: 'link',
              attrs: {
                rel: name,
                href: path.join('/', asset),
                media: `(device-width: ${(orientation === 'portrait' ? width : height) / ratio}px) and (device-height: ${(orientation === 'portrait' ? height : width) / ratio}px) and (-webkit-device-pixel-ratio: ${ratio}) and (orientation: ${orientation})`,
              },
            }),
        );
        // Apple touch icons
        html.headTags.push({
          tag: 'link',
          attrs: {
            rel: 'apple-touch-icon',
            href: path.join('/', appleTouchIconName),
          },
        });
        html.headTags.push({
          tag: 'link',
          attrs: {
            rel: 'apple-touch-icon-precomposed',
            href: path.join('/', appleTouchIconPrecomposedName),
          },
        });

        html.headTags.push({
          tag: 'link',
          attrs: {
            rel: 'shortcut icon',
            href: '/assets/resources/transparent-128.png',
          },
        });

        html.headTags.push({
          tag: 'link',
          attrs: {
            rel: 'icon',
            href: '/favicon.ico',
          },
        });

        return html;
      },
    });
  },
});
