/// <reference types="vitest" />
import { defineConfig, loadEnv, type Plugin } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'path'
import fs from 'fs'

/**
 * Plugin to copy sql.js WASM file to public/wasm directory.
 * This ensures the WASM file is available at /wasm/sql-wasm.wasm in both dev and build.
 */
function copySqlJsWasm(): Plugin {
  return {
    name: 'copy-sqljs-wasm',
    buildStart() {
      const wasmDir = path.resolve(__dirname, 'public/wasm')
      const source = path.resolve(__dirname, 'node_modules/sql.js/dist/sql-wasm.wasm')
      const dest = path.resolve(wasmDir, 'sql-wasm.wasm')

      // Ensure directory exists
      if (!fs.existsSync(wasmDir)) {
        fs.mkdirSync(wasmDir, { recursive: true })
      }

      // Copy WASM file if it doesn't exist or is outdated
      if (!fs.existsSync(dest) || fs.statSync(source).mtime > fs.statSync(dest).mtime) {
        fs.copyFileSync(source, dest)
        console.log('[copySqlJsWasm] Copied sql-wasm.wasm to public/wasm/')
      }
    },
  }
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const buildMode = env.VITE_BUILD_MODE || 'server'
  const isWasmMode = buildMode === 'wasm'
  const isDemoMode = env.VITE_DEMO_MODE === 'true'

  return {
    plugins: [
      react(),
      tailwindcss(),
      // Include sql.js WASM copy plugin in wasm mode or demo mode
      ...(isWasmMode || isDemoMode ? [copySqlJsWasm()] : []),
    ],
    resolve: {
      alias: {
        '@': path.resolve(__dirname, './src'),
      },
    },
    define: {
      __BUILD_MODE__: JSON.stringify(buildMode),
      __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
    },
    server: {
      port: 5173,
      // Disable proxy in WASM mode (no backend)
      proxy: isWasmMode
        ? undefined
        : {
            '/api': {
              target: 'http://localhost:8081',
              changeOrigin: true,
            },
            '/badges': {
              target: 'http://localhost:8081',
              changeOrigin: true,
            },
          },
    },
    build: {
      // Different output directories for each mode
      outDir: isDemoMode ? 'dist-demo' : isWasmMode ? 'dist-wasm' : 'dist',
      sourcemap: true,
      // Use terser in production for better console removal control
      minify: mode === 'production' ? 'terser' : 'esbuild',
      terserOptions:
        mode === 'production'
          ? {
              compress: {
                // Remove console.log/debug/info but keep console.error/warn for error tracking
                drop_console: false,
                pure_funcs: ['console.log', 'console.debug', 'console.info'],
              },
            }
          : undefined,
    },
    // WASM mode and demo mode: handle sql.js WASM files
    ...((isWasmMode || isDemoMode) && {
      optimizeDeps: {
        // Include sql.js for proper CommonJS->ESM conversion
        include: ['sql.js'],
      },
      assetsInclude: ['**/*.wasm'],
    }),
    // Vitest configuration
    test: {
      globals: true,
      // Use 'node' for pure logic tests. Switch to 'jsdom' when testing React components
      // or code that needs real DOM/browser APIs (e.g., document, window, localStorage).
      // Per-file override: add `// @vitest-environment jsdom` at top of file.
      environment: 'node',
      include: ['src/**/*.test.ts'],
    },
  }
})
