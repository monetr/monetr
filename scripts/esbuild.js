require('esbuild').buildSync({
  entryPoints: ['./src/index.js'],
  bundle: true,
  minify: true,
  sourcemap: true,
  target: ['chrome58', 'firefox57', 'safari11', 'edge16'],
  define: {
    'process.env.NODE_ENV': '"production"',
  },
})
