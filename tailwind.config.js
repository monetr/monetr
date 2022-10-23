module.exports = {
  important: true,
  content: [
    './ui/**/*.tsx',
    './ui/**/*.jsx',
    './ui/**/*.js',
  ],
  safelist: [
    'text-green-500',
    'text-yellow-500',
    'text-red-500',
    'text-gray-500',
  ],
  theme: {
    screens: {
      'sm': '640px',
      // => @media (min-width: 640px) { ... }

      'md': '720px', // This is not the default.
      // => @media (min-width: 768px) { ... }

      'lg': '1024px',
      // => @media (min-width: 1024px) { ... }

      'xl': '1280px',
      // => @media (min-width: 1280px) { ... }

      '2xl': '1536px',
      // => @media (min-width: 1536px) { ... }
    }
  }
};
