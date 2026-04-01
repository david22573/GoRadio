/** @type {import('tailwindcss').Config} */
export default {
  content: ['./src/**/*.{html,js,svelte,ts}'],
  darkMode: 'class',
  theme: {
    extend: {
      colors: {
        "on-background": "#ffffff",
        "on-primary-container": "#00253f",
        "tertiary-container": "#b091ff",
        "inverse-surface": "#fcf9f8",
        "secondary-dim": "#d4d3dd",
        "surface-bright": "#2c2c2c",
        "on-primary": "#003151",
        "surface": "#0e0e0e",
        "on-secondary-fixed": "#3e3f47",
        "on-tertiary-fixed": "#1e0053",
        "on-surface-variant": "#adaaaa",
        "surface-container": "#1a1a1a",
        "secondary-fixed": "#e3e1ec",
        "on-error-container": "#ffa8a3",
        "outline": "#767575",
        "surface-container-highest": "#262626",
        "outline-variant": "#484847",
        "on-tertiary-fixed-variant": "#43208b",
        "on-error": "#490006",
        "error": "#ff716c",
        "error-dim": "#d7383b",
        "surface-container-lowest": "#000000",
        "tertiary-dim": "#ab8bf9",
        "on-tertiary": "#3a1482",
        "on-secondary-container": "#d0cfd9",
        "primary": "#5eb4ff",
        "tertiary": "#bca2ff",
        "primary-fixed": "#2aa7ff",
        "primary-container": "#2aa7ff",
        "on-tertiary-container": "#2e0075",
        "primary-fixed-dim": "#0099f0",
        "on-secondary": "#515159",
        "secondary": "#e3e1ec",
        "surface-tint": "#5eb4ff",
        "tertiary-fixed": "#bca2ff",
        "surface-container-low": "#131313",
        "on-surface": "#ffffff",
        "inverse-on-surface": "#565555",
        "surface-variant": "#262626",
        "background": "#0e0e0e",
        "secondary-fixed-dim": "#d4d3dd",
        "tertiary-fixed-dim": "#b091ff",
        "surface-container-high": "#20201f",
        "inverse-primary": "#00639e",
        "error-container": "#9f0519",
        "on-primary-fixed": "#000000",
        "primary-dim": "#15a4ff",
        "on-primary-fixed-variant": "#002e4d",
        "on-secondary-fixed-variant": "#5a5b63",
        "surface-dim": "#0e0e0e",
        "secondary-container": "#46464e"
      },
      fontFamily: {
        "headline": ["Manrope", "sans-serif"],
        "body": ["Inter", "sans-serif"],
        "label": ["Inter", "sans-serif"]
      },
      borderRadius: {
        "DEFAULT": "1rem",
        "lg": "2rem",
        "xl": "3rem",
        "full": "9999px"
      }
    }
  },
  plugins: []
};