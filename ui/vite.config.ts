import { HttpProxy, defineConfig, splitVendorChunkPlugin } from 'vite'
import react from '@vitejs/plugin-react'

const target1: HttpProxy.ProxyTarget = "http://127.0.0.1:8080";
// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react(), splitVendorChunkPlugin()],
  server: {
    port: 3000,
    host: true,
    proxy: {
      "/api": {
        target: `${target1}`,
        changeOrigin: true,
        secure: false,
        ws: true,
        rewrite: (path: string) => path.replace(/^\/api/, ""),
        cookieDomainRewrite: `${target1}`,        
      },
    }
  }
})
