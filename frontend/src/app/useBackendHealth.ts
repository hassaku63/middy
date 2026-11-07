import { useEffect, useState } from 'react';
import { fetchHealth } from '@/shared/api/client';

export type BackendStatus = 'checking' | 'ok' | 'down';

/** Probes the API once on mount; purely informational footer state. */
export function useBackendHealth(): BackendStatus {
  const [status, setStatus] = useState<BackendStatus>('checking');
  useEffect(() => {
    let cancelled = false;
    fetchHealth()
      .then(() => !cancelled && setStatus('ok'))
      .catch(() => !cancelled && setStatus('down'));
    return () => {
      cancelled = true;
    };
  }, []);
  return status;
}
