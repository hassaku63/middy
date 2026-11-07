import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { registerSW } from 'virtual:pwa-register';
import { App } from '@/app/App';
import '@/app/global.css';

// Auto-update service worker; a "new version available" prompt is a backlog item.
registerSW({ immediate: true });

const root = document.getElementById('root');
if (!root) {
  throw new Error('#root element not found');
}
createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
