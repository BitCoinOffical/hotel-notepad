import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const fileEnv = loadEnv(mode, process.cwd(), '');
  // Docker Compose environment variables take precedence over local .env values.
  const authTarget = process.env.AUTH_PROXY_TARGET || fileEnv.AUTH_PROXY_TARGET || 'http://localhost:8080';
  const guestbookTarget = process.env.GUESTBOOK_PROXY_TARGET || fileEnv.GUESTBOOK_PROXY_TARGET || 'http://localhost:8081';

  return {
    plugins: [react()],
    server: {
      host: '0.0.0.0',
      port: 5173,
      proxy: {
        '/api/v1/auth': {
          target: authTarget,
          changeOrigin: true,
        },
        '/api/v1/guestbook': {
          target: guestbookTarget,
          changeOrigin: true,
        },
      },
    },
  };
});
