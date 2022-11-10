
/** @type {import('tailwindcss').Config} */
module.exports = {
  important: true,
  darkMode: 'class',
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
    colors: {
      purple: {
        "50": "#EDE5FB",
        "100": "#D8C6F6",
        "200": "#B591ED",
        "300": "#8E58E4",
        "400": "#6823D7",
        "500": "#4E1AA0",
        "600": "#3E157F",
        "700": "#2F1060",
        "800": "#200B42",
        "900": "#0F051F"
      },
    },
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
