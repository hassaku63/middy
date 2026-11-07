import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { registerSW } from 'virtual:pwa-register';
import { App } from '@/app/App';
import { pwaStatus } from '@/shared/lib/pwa/status';
import { wirePwa } from '@/shared/lib/pwa/wire';
import '@/app/global.css';

// Service worker in "prompt" mode: a new version waits until the user
// accepts it in the banner, so nobody loses work to a surprise reload.
wirePwa(pwaStatus, registerSW);

const root = document.getElementById('root');
if (!root) {
  throw new Error('#root element not found');
}
createRoot(root).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
