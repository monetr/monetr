// This is a temporary fix to get the html.lang attribute set. Its not working quite right with next.js export atm.
// Basically when I add https://nextjs.org/docs/pages/building-your-application/routing/internationalization then the
// application won't build at all because it can't do that for static sites? But https://nextra.site/docs/guide/i18n
// also wasn't working properly? I need to fix this and its probably just user error. But this is an easy way to get it
// going for now. SEO is evil n all.

const fs = require('fs');
const path = require('path');
const cheerio = require('cheerio');

function processHtmlFile(filePath) {
  const html = fs.readFileSync(filePath, 'utf8');
  // eslint-disable-next-line id-length
  const $ = cheerio.load(html);
  $('html').attr('lang', 'en-US');
  fs.writeFileSync(filePath, $.html());
}

function walkDir(dir, callback) {
  console.log(`correcting html.lang attribute in ${dir}`);
  fs.readdirSync(dir).forEach(filePath => {
    const dirPath = path.join(dir, filePath);
    const isDirectory = fs.statSync(dirPath).isDirectory();
    isDirectory ? walkDir(dirPath, callback) : callback(path.join(dir, filePath));
  });
};

process.argv.slice(2).forEach(dirArg => {
  walkDir(dirArg, filePath => {
    if (path.extname(filePath) === '.html') {
      processHtmlFile(filePath);
    }
  });
});
