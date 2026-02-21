import { defineConfig } from 'astro/config';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  site: 'https://yazhsab.github.io',
  base: '/qbitel-qsgw',
  output: 'static',
  vite: {
    plugins: [tailwindcss()],
  },
});
