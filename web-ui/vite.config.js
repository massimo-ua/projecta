import { execSync } from 'child_process';
import path from 'path';
import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import { intlayer } from 'vite-intlayer';

const getAppVersion = () => {
  if (process.env.VITE_APP_VERSION) {
    return process.env.VITE_APP_VERSION;
  }
  try {
    return execSync('git describe --tags --abbrev=0', { encoding: 'utf8' }).trim();
  } catch {
    try {
      return execSync('git describe --tags --always --dirty', { encoding: 'utf8' }).trim();
    } catch {
      return 'dev';
    }
  }
};

const appVersion = getAppVersion();

// https://vitejs.dev/config/
export default defineConfig({
  define: {
    'import.meta.env.VITE_APP_VERSION': JSON.stringify(appVersion),
  },
  plugins: [react(), intlayer()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    minify: 'esbuild',
    target: 'es2020',
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: [
            '@react-oauth/google',
            'date-fns',
            'dayjs',
            'jwt-decode',
            'lucide-react',
            'react',
            'react-dom',
            'react-router-dom',
          ],
        },
        chunkFileNames: 'assets/[hash].js',
        entryFileNames: 'assets/[hash].js',
        assetFileNames: 'assets/[hash].[ext]',
        compact: true,
      },
    },
    cssCodeSplit: true,
    sourcemap: false,
    chunkSizeWarningLimit: 500,
    assetsInlineLimit: 4096,
    minifyIdentifiers: true,
    minifySyntax: true,
    minifyWhitespace: true,
    treeShaking: true,
    dropLabels: ['DEBUG'],
  },
  server: {
    host: true,
    port: 5173,
    proxy: {
      '/api': {
        target: process.env.VITE_API_TARGET || 'http://127.0.0.1:8000',
        changeOrigin: true,
        rewrite: (urlPath) => urlPath.replace(/^\/api/, ''),
      },
    },
  },
});
