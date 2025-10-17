/* eslint-disable max-len */
import {
  Gravity,
  ImageMagick,
  initializeImageMagick,
  MagickColor,
  MagickColors,
  MagickFormat,
  MagickGeometry,
  MagickImage,
  MagickImageCollection,
} from '@imagemagick/magick-wasm';
import { RsbuildPlugin } from '@rsbuild/core';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { promisify } from 'node:util';

const wasmLocation = require.resolve('@imagemagick/magick-wasm/magick.wasm');
const wasmBytes = readFileSync(wasmLocation);

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
  // Taken from https://developer.apple.com/design/human-interface-guidelines/layout/#Specifications
  const spec: Array<Omit<AppleSplashScreenSpec, 'name' | 'asset'>> = [
    { width: 2048, height: 2732, ratio: 2, orientation: 'portrait' },
    { width: 2732, height: 2048, ratio: 2, orientation: 'landscape' },
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

export const pluginPWA = (options: PluginPWAOptions): RsbuildPlugin => ({
  name: 'pwa',

  setup(api) {
    api.processAssets({ stage: 'additional' }, async ({ compilation, sources }) => {
      await initializeImageMagick(wasmBytes);
      const backgroundColor = new MagickColor(options.background);

      if (!compilation.inputFileSystem) {
        throw new Error("[pluginPWA] 'compilation.inputFileSystem' is not available.");
      }

      const quality = options.quality ?? 90;
      const source = await promisify(compilation.inputFileSystem.readFile)(options.logo);
      if (!source) {
        throw new Error(
          `[pluginPWA] Failed to read the PWA logo file, please check if the '${options.logo}' file exists'.`,
        );
      }
      const sourceBytes = Uint8Array.from(source);

      appleSplashScreens().forEach(({ width, height }) =>
        compilation.emitAsset(
          `assets/resources/apple-splash-${width}-${height}.png`,
          new sources.RawSource(
            Buffer.from(
              ImageMagick.read(sourceBytes, image => {
                const padding = 0.3;
                const logoWidth = +(width * padding).toFixed(0);
                const logoHeight = +(height * padding).toFixed(0);
                image.resize(logoWidth, logoHeight);
                image.extent(new MagickGeometry(width, height), Gravity.Center, backgroundColor);
                image.quality = quality;
                // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
                // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
                // See https://github.com/dlemstra/magick-wasm/issues/185
                return image.write(MagickFormat.Png, data => [...data]);
              }),
            ),
          ),
        ),
      );

      // Apple touch icon
      compilation.emitAsset(
        appleTouchIconName,
        new sources.RawSource(
          Buffer.from(
            ImageMagick.read(sourceBytes, image => {
              const width = 180;
              const height = 180;
              const padding = 0.7;
              const logoWidth = +(width * padding).toFixed(0);
              const logoHeight = +(height * padding).toFixed(0);
              image.resize(logoWidth, logoHeight);
              image.extent(new MagickGeometry(width, height), Gravity.Center, backgroundColor);
              image.quality = quality;
              // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
              // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
              // See https://github.com/dlemstra/magick-wasm/issues/185
              return image.write(MagickFormat.Png, data => [...data]);
            }),
          ),
        ),
      );
      compilation.emitAsset(
        appleTouchIconPrecomposedName,
        new sources.RawSource(
          Buffer.from(
            ImageMagick.read(sourceBytes, image => {
              const width = 180;
              const height = 180;
              const padding = 0.7;
              const logoWidth = +(width * padding).toFixed(0);
              const logoHeight = +(height * padding).toFixed(0);
              image.resize(logoWidth, logoHeight);
              image.extent(new MagickGeometry(width, height), Gravity.Center, backgroundColor);
              image.quality = quality;
              // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
              // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
              // See https://github.com/dlemstra/magick-wasm/issues/185
              return image.write(MagickFormat.Png, data => [...data]);
            }),
          ),
        ),
      );

      compilation.emitAsset(
        'assets/resources/transparent-128.png',
        new sources.RawSource(
          Buffer.from(
            ImageMagick.read(sourceBytes, image => {
              const width = 128;
              const height = 128;
              const padding = 1.0;
              const logoWidth = +(width * padding).toFixed(0);
              const logoHeight = +(height * padding).toFixed(0);
              image.resize(logoWidth, logoHeight);
              image.extent(new MagickGeometry(width, height), Gravity.Center, MagickColors.Transparent);
              image.quality = quality;
              // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
              // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
              // See https://github.com/dlemstra/magick-wasm/issues/185
              return image.write(MagickFormat.Png, data => [...data]);
            }),
          ),
        ),
      );

      mstileIcons().forEach(({ asset, width, height }) =>
        compilation.emitAsset(
          asset,
          new sources.RawSource(
            Buffer.from(
              ImageMagick.read(sourceBytes, image => {
                const padding = 0.7;
                const logoWidth = +(width * padding).toFixed(0);
                const logoHeight = +(height * padding).toFixed(0);
                image.resize(logoWidth, logoHeight);
                image.extent(new MagickGeometry(width, height), Gravity.Center, backgroundColor);
                image.quality = quality;
                // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
                // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
                // See https://github.com/dlemstra/magick-wasm/issues/185
                return image.write(MagickFormat.Png, data => [...data]);
              }),
            ),
          ),
        ),
      );

      const androidChromeSizes: Array<[number, number, boolean]> = [
        [192, 192, false],
        [192, 192, true],
        [512, 512, false],
        [512, 512, true],
      ];
      androidChromeSizes.forEach(([width, height, maskable]) =>
        compilation.emitAsset(
          `assets/resources/android-chrome-${width}-${height}${maskable ? '_maskable' : ''}.png`,
          new sources.RawSource(
            Buffer.from(
              ImageMagick.read(sourceBytes, image => {
                const padding = 0.7;
                const logoWidth = +(width * padding).toFixed(0);
                const logoHeight = +(height * padding).toFixed(0);
                image.resize(logoWidth, logoHeight);
                const background = maskable ? backgroundColor : MagickColors.Transparent;
                image.extent(new MagickGeometry(width, height), Gravity.Center, background);
                image.quality = quality;
                // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
                // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
                // See https://github.com/dlemstra/magick-wasm/issues/185
                return image.write(MagickFormat.Png, data => [...data]);
              }),
            ),
          ),
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
      compilation.emitAsset(
        'favicon.ico',
        new sources.RawSource(
          Buffer.from(
            (() => {
              const favicon = MagickImageCollection.create();
              // Add each favicon size to our collection
              favicon.push(
                ...faviconSizes.map(([width, height]) =>
                  ImageMagick.read(sourceBytes, image => {
                    image.resize(width, height);
                    image.extent(new MagickGeometry(width, height), Gravity.Center, MagickColors.Transparent);
                    image.quality = quality;
                    // The `image` variable is "disposed" after this callback. So we basically need to copy the image entirely
                    // when we return it so that it can be used by the collection. There might be a better way to do this but
                    // I haven't figured it out yet.
                    return MagickImage.create(Uint8Array.from(image.write(MagickFormat.Png, data => [...data])));
                  }),
                ),
              );
              // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
              // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
              // See https://github.com/dlemstra/magick-wasm/issues/185
              return favicon.write(MagickFormat.Ico, data => [...data]);
            })(),
          ),
        ),
      );

      return;
    });

    api.modifyHTMLTags({
      order: 'post',
      handler: html => {
        mstileIcons().forEach(({ name, asset }) =>
          html.headTags.push({
            tag: 'meta',
            attrs: {
              name: name,
              content: path.join('/', asset),
            },
          }),
        );

        appleSplashScreens().forEach(({ name, asset, width, height, ratio, orientation }) =>
          html.headTags.push({
            tag: 'link',
            attrs: {
              rel: name,
              href: path.join('/', asset),
              media: `(device-width: ${(orientation === 'portrait' ? width : height) / ratio}px) and (device-height: ${(orientation === 'portrait' ? height : width) / ratio}px) and (-webkit-device-pixel-ratio: ${ratio}) and (orientation: ${orientation})`,
            },
          }),
        );

        {
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
        }

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
