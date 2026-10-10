// @ts-check
import { defineConfig } from 'astro/config';

// https://astro.build/config
export default defineConfig({
  site: 'https://famloop.my.id',
  output: 'static',
  build: {
    format: 'directory'
  }
});
