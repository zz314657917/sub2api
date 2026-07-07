/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,js,ts,jsx,tsx}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        primary: {
          50: '#fbf4ef',
          100: '#f6e4d8',
          200: '#edc7b4',
          300: '#df9e7f',
          400: '#d48667',
          500: '#cc785c',
          600: '#a9583e',
          700: '#88432e',
          800: '#6f3829',
          900: '#5b3025',
          950: '#321610'
        },
        accent: {
          50: '#faf6ec',
          100: '#f0e5ca',
          200: '#dfca9c',
          300: '#caa86b',
          400: '#b4894b',
          500: '#8b6f47',
          600: '#75573b',
          700: '#5e432f',
          800: '#463225',
          900: '#33251b',
          950: '#1d130d'
        },
        dark: {
          50: '#f8fafc',
          100: '#f1f5f9',
          200: '#e2e8f0',
          300: '#cbd5e1',
          400: '#94a3b8',
          500: '#64748b',
          600: '#475569',
          700: '#334155',
          800: '#1e293b',
          900: '#0f172a',
          950: '#020617'
        }
      },
      fontFamily: {
        sans: [
          'system-ui',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'PingFang SC',
          'Hiragino Sans GB',
          'Microsoft YaHei',
          'sans-serif'
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace']
      },
      boxShadow: {
        glass: '0 6px 0 rgba(70, 50, 35, 0.22), 0 16px 28px rgba(20, 31, 18, 0.12)',
        'glass-sm': '0 3px 0 rgba(70, 50, 35, 0.2), 0 10px 18px rgba(20, 31, 18, 0.1)',
        glow: '0 0 0 2px rgba(204, 120, 92, 0.24), 0 0 22px rgba(204, 120, 92, 0.18)',
        'glow-lg':
          '0 0 0 2px rgba(204, 120, 92, 0.32), 0 0 44px rgba(204, 120, 92, 0.24)',
        card: '0 4px 0 rgba(70, 50, 35, 0.18), 0 12px 24px rgba(20, 31, 18, 0.08)',
        'card-hover':
          '0 6px 0 rgba(70, 50, 35, 0.2), 0 18px 34px rgba(20, 31, 18, 0.12)',
        'inner-glow': 'inset 0 1px 0 rgba(255, 255, 255, 0.1)'
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
        'gradient-primary': 'linear-gradient(135deg, #d48667 0%, #a9583e 100%)',
        'gradient-dark': 'linear-gradient(135deg, #1e293b 0%, #020617 100%)',
        'gradient-glass':
          'linear-gradient(135deg, rgba(255,255,255,0.1) 0%, rgba(255,255,255,0.05) 100%)',
        'mesh-gradient':
          'linear-gradient(45deg, rgba(204, 120, 92, 0.08) 25%, transparent 25%), linear-gradient(-45deg, rgba(139, 111, 71, 0.08) 25%, transparent 25%), linear-gradient(rgba(91, 78, 60, 0.08) 1px, transparent 1px), linear-gradient(90deg, rgba(91, 78, 60, 0.08) 1px, transparent 1px)'
      },
      animation: {
        'fade-in': 'fadeIn 0.3s ease-out',
        'slide-up': 'slideUp 0.3s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'slide-in-right': 'slideInRight 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
        'pulse-slow': 'pulse 3s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        shimmer: 'shimmer 2s linear infinite',
        glow: 'glow 2s ease-in-out infinite alternate'
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' }
        },
        slideUp: {
          '0%': { opacity: '0', transform: 'translateY(10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideDown: {
          '0%': { opacity: '0', transform: 'translateY(-10px)' },
          '100%': { opacity: '1', transform: 'translateY(0)' }
        },
        slideInRight: {
          '0%': { opacity: '0', transform: 'translateX(20px)' },
          '100%': { opacity: '1', transform: 'translateX(0)' }
        },
        scaleIn: {
          '0%': { opacity: '0', transform: 'scale(0.95)' },
          '100%': { opacity: '1', transform: 'scale(1)' }
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' }
        },
        glow: {
          '0%': {
            boxShadow:
              '0 0 0 2px rgba(204, 120, 92, 0.22), 0 0 18px rgba(204, 120, 92, 0.18)'
          },
          '100%': {
            boxShadow:
              '0 0 0 2px rgba(139, 111, 71, 0.28), 0 0 28px rgba(139, 111, 71, 0.24)'
          }
        }
      },
      backdropBlur: {
        xs: '2px'
      },
      borderRadius: {
        '4xl': '2rem'
      }
    }
  },
  plugins: []
}
