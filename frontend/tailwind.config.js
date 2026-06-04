/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        // Fondo & superficie
        cream:    '#F4EEE3',  // fondo principal de la app
        surface:  '#FCF9F3',  // cards, modales
        paper:    '#FFFFFF',
        hairline: '#E7DECD',  // bordes y divisores

        // Tinta
        ink: {
          DEFAULT: '#2A251F', // texto principal
          soft:    '#6F6557', // texto secundario
          faint:   '#A89C89', // placeholders, deshabilitados
        },

        // Estados semánticos
        sage: {
          DEFAULT: '#A8C39A', // cumplió
          deep:    '#7CA39D', // CTA primario / salvia profundo
        },
        amber:      '#E9C281', // pendiente
        coral:      '#EDA48F', // falló

        // Acentos de marca
        terracotta: '#C26F4D', // botón secundario / identidad Dydi
        wash:       '#DFEBE8', // fondos de acento suave
      },

      fontFamily: {
        serif: ['Newsreader', 'Georgia', 'serif'],
        sans:  ['Hanken Grotesk', 'system-ui', 'sans-serif'],
      },

      borderRadius: {
        card: '22px',  // cards elevadas y planas
        pill: '999px', // pills y tags de estado
      },

      boxShadow: {
        card: '0 4px 24px 0 rgba(42,37,31,0.08)', // card elevada
        flat: '0 1px 4px 0 rgba(42,37,31,0.06)',  // card plana
      },

      letterSpacing: {
        eyebrow: '0.1em',
      },
    },
  },
  plugins: [],
}
