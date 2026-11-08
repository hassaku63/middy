import { useEffect, useState } from 'react';
import { fetchHealth } from '@/shared/api/client';

export type BackendStatus = 'checking' | 'ok' | 'down';

/** How long to wait before asking an unreachable backend again. */
export const RETRY_MS = 15_000;

/**
 * Probes the API on mount, again whenever the browser comes back online,
 * and every RETRY_MS while it does not answer, so a notice about an
 * unreachable backend never outlives the outage. While offline it does not
 * probe: the PWA banner already explains that.
 */
export function useBackendHealth(online = true): BackendStatus {
  const [status, setStatus] = useState<BackendStatus>('checking');
  useEffect(() => {
    if (!online) return;
    let cancelled = false;
    let retry: ReturnType<typeof setTimeout> | undefined;
    const probe = () => {
      fetchHealth()
        .then(() => !cancelled && setStatus('ok'))
        .catch(() => {
          if (cancelled) return;
          setStatus('down');
          retry = setTimeout(probe, RETRY_MS);
        });
    };
    probe();
    return () => {
      cancelled = true;
      clearTimeout(retry);
    };
  }, [online]);
  return status;
}
