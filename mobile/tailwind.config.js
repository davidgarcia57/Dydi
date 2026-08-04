/** @type {import('tailwindcss').Config} */
module.exports = {
  // NOTE: Update this to include the paths to all of your component files.
  content: [
    "./App.{js,jsx,ts,tsx}",
    "./app/**/*.{js,jsx,ts,tsx}",
    "./src/components/**/*.{js,jsx,ts,tsx}"
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
      // Escala tipográfica. Antes no existía ninguna: se usaba la default de
      // Tailwind más 77 tamaños ad-hoc en píxeles (de 8 a 15px), así que cada
      // pantalla acabó inventando la suya y el grueso del texto vivía entre 8 y
      // 14px — ilegible en un teléfono de verdad.
      //
      // El piso son 12px (`micro`): todos los 8/9/10/11px se subieron ahí. Y `xs`
      // y `sm` se redefinen hacia arriba a propósito, porque entre las dos cargan
      // ~190 usos: moverlas aquí levanta la legibilidad de toda la app sin tener
      // que editar cada pantalla.
      fontSize: {
        micro: ['12px', { lineHeight: '16px' }], // badges, pills, etiquetas
        xs: ['13px', { lineHeight: '18px' }], // era 12px
        sm: ['15px', { lineHeight: '21px' }], // era 14px
        base: ['16px', { lineHeight: '24px' }],
      },
    },
  },
  plugins: [],
}
