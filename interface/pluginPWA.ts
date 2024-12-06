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

const wasmLocation = require.resolve('@imagemagick/magick-wasm/magick.wasm') ;
const wasmBytes = readFileSync(wasmLocation);

type PluginPWAOptions = {
  logo: string;
  background: string;
  quality?: number;
}

const appleTouchIconName = 'apple-touch-icon.png';
const appleTouchIconPrecomposedName = 'apple-touch-icon-precomposed.png';

// mstileIcons provides all the information needed to generate the icons used by internet explorer. I still need to add
// support for microsoft edge though.
function mstileIcons(): Array<{ name: string; asset: string, width: number, height: number }> {
  const mstileSizes: Array<[number, number]> = [
    [70,  70],
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

export const pluginPWA = (options: PluginPWAOptions): RsbuildPlugin => ({
  name: 'pwa',

  setup(api) {
    api.processAssets(
      { stage: 'additional' },
      async ({ compilation, sources }) => {
        await initializeImageMagick(wasmBytes);
        const backgroundColor = new MagickColor(options.background);

        if (!compilation.inputFileSystem) {
          throw new Error(
            '[pluginPWA] \'compilation.inputFileSystem\' is not available.',
          );
        }

        const quality = options.quality ?? 90;
        const source = await promisify(
          compilation.inputFileSystem.readFile,
        )(options.logo);
        if (!source) {
          throw new Error(
            `[pluginPWA] Failed to read the PWA logo file, please check if the '${options.logo}' file exists'.`,
          );
        }
        const sourceBytes = Uint8Array.from(source);
        // Taken from https://developer.apple.com/design/human-interface-guidelines/layout/#Specifications
        const appleSplashSizes: Array<[number, number]> = [
          [2048, 2732],
          [2732, 2048],
          [1668, 2388],
          [2388, 1668],
          [1536, 2048],
          [2048, 1536],
          [1488, 2266],
          [2266, 1488],
          [1640, 2360],
          [2360, 1640],
          [1668, 2224],
          [2224, 1668],
          [1620, 2160],
          [2160, 1620],
          [1320, 2868],
          [2868, 1320],
          [1206, 2622],
          [2622, 1206],
          [1290, 2796],
          [2796, 1290],
          [1179, 2556],
          [2556, 1179],
          [1284, 2778],
          [2778, 1284],
          [1170, 2532],
          [2532, 1170],
          [1125, 2436],
          [2436, 1125],
          [1242, 2688],
          [2688, 1242],
          [828, 1792],
          [1792, 828],
          [1242, 2208],
          [2208, 1242],
          [750, 1334],
          [1334, 750],
          [640, 1136],
          [1136, 640],
        ];
        appleSplashSizes.forEach(([width, height]) => compilation.emitAsset(
          `assets/resources/apple-splash-${width}-${height}.png`,
          new sources.RawSource(Buffer.from(ImageMagick.read(sourceBytes, image => {
            const padding = 0.3;
            const logoWidth = +((width * padding).toFixed(0));
            const logoHeight = +((height * padding).toFixed(0));
            image.resize(logoWidth, logoHeight);
            image.extent(new MagickGeometry(width, height), Gravity.Center, backgroundColor);
            image.quality = quality;
            // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
            // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
            // See https://github.com/dlemstra/magick-wasm/issues/185
            return image.write(MagickFormat.Png, data => [...data]);
          }))),
        ));

        // Apple touch icon
        compilation.emitAsset(
          appleTouchIconName,
          new sources.RawSource(Buffer.from(ImageMagick.read(sourceBytes, image => {
            const width = 180;
            const height = 180;
            const padding = 0.70;
            const logoWidth = +((width * padding).toFixed(0));
            const logoHeight = +((height * padding).toFixed(0));
            image.resize(logoWidth, logoHeight);
            image.extent(new MagickGeometry(width, height), Gravity.Center, backgroundColor);
            image.quality = quality;
            // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
            // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
            // See https://github.com/dlemstra/magick-wasm/issues/185
            return image.write(MagickFormat.Png, data => [...data]);
          }))),
        );
        compilation.emitAsset(
          appleTouchIconPrecomposedName,
          new sources.RawSource(Buffer.from(ImageMagick.read(sourceBytes, image => {
            const width = 180;
            const height = 180;
            const padding = 0.70;
            const logoWidth = +((width * padding).toFixed(0));
            const logoHeight = +((height * padding).toFixed(0));
            image.resize(logoWidth, logoHeight);
            image.extent(new MagickGeometry(width, height), Gravity.Center, backgroundColor);
            image.quality = quality;
            // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
            // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
            // See https://github.com/dlemstra/magick-wasm/issues/185
            return image.write(MagickFormat.Png, data => [...data]);
          }))),
        );

        compilation.emitAsset(
          'assets/resources/transparent-128.png',
          new sources.RawSource(Buffer.from(ImageMagick.read(sourceBytes, image => {
            const width = 128;
            const height = 128;
            const padding = 1.0;
            const logoWidth = +((width * padding).toFixed(0));
            const logoHeight = +((height * padding).toFixed(0));
            image.resize(logoWidth, logoHeight);
            image.extent(new MagickGeometry(width, height), Gravity.Center, MagickColors.Transparent);
            image.quality = quality;
            // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
            // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
            // See https://github.com/dlemstra/magick-wasm/issues/185
            return image.write(MagickFormat.Png, data => [...data]);
          }))),
        );

        mstileIcons().forEach(({ asset, width, height }) => compilation.emitAsset(
          asset,
          new sources.RawSource(Buffer.from(ImageMagick.read(sourceBytes, image => {
            const padding = 0.70;
            const logoWidth = +((width * padding).toFixed(0));
            const logoHeight = +((height * padding).toFixed(0));
            image.resize(logoWidth, logoHeight);
            image.extent(new MagickGeometry(width, height), Gravity.Center, backgroundColor);
            image.quality = quality;
            // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
            // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
            // See https://github.com/dlemstra/magick-wasm/issues/185
            return image.write(MagickFormat.Png, data => [...data]);
          }))),
        ));

        const androidChromeSizes: Array<[number, number, boolean]> = [
          [192, 192, false],
          [192, 192, true],
          [512, 512, false],
          [512, 512, true],
        ];
        androidChromeSizes.forEach(([width, height, maskable]) => compilation.emitAsset(
          `assets/resources/android-chrome-${width}-${height}${maskable ? '_maskable' : ''}.png`,
          new sources.RawSource(Buffer.from(ImageMagick.read(sourceBytes, image => {
            const padding = 0.70;
            const logoWidth = +((width * padding).toFixed(0));
            const logoHeight = +((height * padding).toFixed(0));
            image.resize(logoWidth, logoHeight);
            const background = maskable ? backgroundColor : MagickColors.Transparent;
            image.extent(new MagickGeometry(width, height), Gravity.Center, background);
            image.quality = quality;
            // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
            // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
            // See https://github.com/dlemstra/magick-wasm/issues/185
            return image.write(MagickFormat.Png, data => [...data]);
          }))),
        ));

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
          new sources.RawSource(Buffer.from((() => {
            const favicon = MagickImageCollection.create();
            // Add each favicon size to our collection
            favicon.push(...faviconSizes.map(([width, height]) => ImageMagick.read(sourceBytes, image => {
              image.resize(width, height);
              image.extent(new MagickGeometry(width, height), Gravity.Center, MagickColors.Transparent);
              image.quality = quality;
              // The `image` variable is "disposed" after this callback. So we basically need to copy the image entirely
              // when we return it so that it can be used by the collection. There might be a better way to do this but
              // I haven't figured it out yet.
              return MagickImage.create(Uint8Array.from(image.write(MagickFormat.Png, data => [...data])));
            })));
            // Need to do `[...data]`. The data array is in managemend memory and may be freed after this function is
            // complete. Unpacking it into another array copies the memory so we don't get corrupt files.
            // See https://github.com/dlemstra/magick-wasm/issues/185
            return favicon.write(MagickFormat.Ico, data => [...data]);
          })()))
        );

        return;
      }
    );

    api.modifyHTMLTags({
      order: 'post',
      handler: html => {
        mstileIcons().forEach(({ name, asset: assetPath }) => html.headTags.push({
          tag: 'meta',
          attrs: {
            name: name,
            content: path.join('/', assetPath),
          },
        }));

        { // Apple touch icons
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
            rel: 'icon',
            href: '/favicon.ico',
          },
        });

        return html;
      },
    });
  },
});
