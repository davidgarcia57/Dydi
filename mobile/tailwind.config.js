/** @type {import('tailwindcss').Config} */
module.exports = {
  // NOTE: Update this to include the paths to all of your component files.
  content: [
    "./App.{js,jsx,ts,tsx}",
    "./app/**/*.{js,jsx,ts,tsx}",
    "./components/**/*.{js,jsx,ts,tsx}"
  ],
  presets: [require("nativewind/preset")],
  theme: {
    extend: {
      colors: {
        cream: '#F4EEE3',
        surface: '#FCF9F3',
        paper: '#FFFFFF',
        hairline: '#E7DECD',
        ink: {
          DEFAULT: '#2A251F',
          soft: '#6F6557',
          faint: '#A89C89',
        },
        sage: {
          DEFAULT: '#A8C39A',
          deep: '#7CA39D',
        },
        amber: '#E9C281',
        coral: '#EDA48F',
        terracotta: '#C26F4D',
        wash: '#DFEBE8',
      },
    },
  },
  plugins: [],
}
