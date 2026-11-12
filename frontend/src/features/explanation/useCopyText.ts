import { useCallback, useEffect, useRef, useState } from 'react';

export type CopyState = 'idle' | 'copied' | 'failed';

/** How long "コピーしました" stays before the button rests again. */
const SHOWN_MS = 2000;

/**
 * Writes text to the clipboard and says how it went (BL-300): "copied"
 * for a while, then idle again; "failed" when the browser refuses (no
 * permission, an insecure page), until the next try.
 */
export function useCopyText(write: (text: string) => Promise<void> = writeClipboard) {
  const [state, setState] = useState<CopyState>('idle');
  const timer = useRef<ReturnType<typeof setTimeout> | undefined>(undefined);
  useEffect(() => () => clearTimeout(timer.current), []);

  const copy = useCallback(
    async (text: string) => {
      clearTimeout(timer.current);
      try {
        await write(text);
        setState('copied');
        timer.current = setTimeout(() => setState('idle'), SHOWN_MS);
      } catch {
        setState('failed');
      }
    },
    [write],
  );
  return { state, copy };
}

// Async, so a browser without the Clipboard API (an insecure page) rejects
// instead of throwing.
async function writeClipboard(text: string): Promise<void> {
  await navigator.clipboard.writeText(text);
}
