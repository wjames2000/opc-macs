export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      colors: {
        primary: {'50':'#EFF6FF','100':'#DBEAFE','200':'#BFDBFE','400':'#60A5FA','500':'#3B82F6','700':'#1E3A5F','800':'#1E3A5F','900':'#1E3A5F'},
        secondary: '#0891B2',
        accent: '#F59E0B',
        surface: '#F8FAFC',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      }
    },
  },
  plugins: [],
};
