import { useEffect, useState } from 'react';
import { fetchHealth } from '@/shared/api/client';

export type BackendStatus = 'checking' | 'ok' | 'down';

/**
 * Probes the API on mount and again whenever the browser comes back online,
 * so a notice about an unreachable backend never outlives the outage.
 * While offline it does not probe: the PWA banner already explains that.
 */
export function useBackendHealth(online = true): BackendStatus {
  const [status, setStatus] = useState<BackendStatus>('checking');
  useEffect(() => {
    if (!online) return;
    let cancelled = false;
    fetchHealth()
      .then(() => !cancelled && setStatus('ok'))
      .catch(() => !cancelled && setStatus('down'));
    return () => {
      cancelled = true;
    };
  }, [online]);
  return status;
}
