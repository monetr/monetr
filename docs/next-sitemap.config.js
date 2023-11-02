/** @type {import('next-sitemap').IConfig} */
module.exports = {
  siteUrl: process.env.SITE_URL || 'https://monetr.app',
  generateRobotsTxt: true, // (optional)
  output: 'export',
  outDir: 'out',
  autoLastmod: false, // This isn't working quite right
}

