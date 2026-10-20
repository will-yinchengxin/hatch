import { defineConfig } from 'vite';

export default defineConfig({
  server: {
    open: true,
  },
  build: {
    // The Go backend (../src) embeds this bundle via //go:embed all:web/dist.
    outDir: '../src/web/dist',
    emptyOutDir: true,
  },
});
