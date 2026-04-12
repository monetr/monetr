import { convert as htmlToText } from 'html-to-text';

export function toPlainText(html: string): string {
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
