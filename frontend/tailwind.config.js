/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js}'],
  theme: {
    extend: {
      colors: {
        // Fondo & superficie
        cream: '#F4EEE3', // fondo principal de la app
        surface: '#FCF9F3', // cards, modales
        paper: '#FFFFFF',
        hairline: '#E7DECD', // bordes y divisores

        // Tinta
        ink: {
          DEFAULT: '#2A251F', // texto principal
          soft: '#6F6557', // texto secundario
          faint: '#A89C89', // placeholders, deshabilitados
        },

        // Estados semánticos
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

        // Acentos de marca
        terracotta: '#C26F4D', // botón secundario / identidad Dydi
        'accent-deep': '#4C736C', // hover / texto-acento
        wash: '#DFEBE8', // fondos de acento suave
        'cream-2': '#EFE7D8', // fondo alterno / hendiduras
      },

      fontFamily: {
        serif: ['Newsreader', 'Georgia', 'serif'],
        sans: ['Hanken Grotesk', 'system-ui', 'sans-serif'],
      },

      borderRadius: {
        card: '22px', // cards elevadas y planas
        pill: '999px', // pills y tags de estado
      },

      boxShadow: {
        card: '0 4px 24px 0 rgba(42,37,31,0.08)', // card elevada
        flat: '0 1px 4px 0 rgba(42,37,31,0.06)', // card plana
      },

      letterSpacing: {
        eyebrow: '0.1em',
      },

      spacing: {
        22: '5.5rem', // anillo de éxito del check-in
      },

      keyframes: {
        'fade-up': {
          from: { opacity: '0', transform: 'translateY(20px)' },
          to: { opacity: '1', transform: 'translateY(0)' },
        },
        'fade-in': {
          from: { opacity: '0' },
          to: { opacity: '1' },
        },
      },
      animation: {
        'fade-up': 'fade-up 0.6s ease both',
        'fade-in': 'fade-in 0.6s ease both',
      },
    },
  },
  plugins: [],
}
