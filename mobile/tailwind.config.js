/** @type {import('tailwindcss').Config} */
module.exports = {
  // NOTE: Update this to include the paths to all of your component files.
  content: [
    "./App.{js,jsx,ts,tsx}",
    "./app/**/*.{js,jsx,ts,tsx}",
    "./components/**/*.{js,jsx,ts,tsx}"
  ],
  presets: [require("nativewind/preset")],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        cream: '#F4EEE3', // fondo principal de la app
        surface: '#FCF9F3', // cards, modales
        paper: '#FFFFFF',
        hairline: '#E7DECD', // bordes y divisores
        ink: {
          DEFAULT: '#2A251F', // texto principal
          soft: '#6F6557', // texto secundario
          faint: '#A89C89', // placeholders, deshabilitados
        },
        sage: {
          DEFAULT: '#A8C39A', // cumplió (fill)
          deep: '#7CA39D', // CTA primario / salvia profundo
          soft: '#E4EDDC', // wash de fondo
        },
        amber: {
          DEFAULT: '#E9C281', // pendiente (fill)
          deep: '#A57B33', // texto/íconos sobre claro
          soft: '#F5E8CD', // wash de fondo
        },
        coral: {
          DEFAULT: '#EDA48F', // falló (fill)
          deep: '#BC5C42', // texto/íconos sobre claro
          soft: '#F7E2DA', // wash de fondo
        },
        terracotta: '#C26F4D', // botón secundario / identidad Dydi
        'accent-deep': '#4C736C', // hover / texto-acento
        wash: '#DFEBE8', // fondos de acento suave
        'cream-2': '#EFE7D8', // fondo alterno / hendiduras
      },
      fontFamily: {
        serif: ['Newsreader', 'serif'],
        sans: ['HankenGrotesk', 'sans-serif'],
      },
    },
  },
  plugins: [],
}
