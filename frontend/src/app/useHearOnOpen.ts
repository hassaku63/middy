import { useCallback, useEffect, useRef, useState } from 'react';
import { typing } from './useChordKeys';

const KEY = 'middy.hearOnOpen';

function stored(): boolean {
  try {
    return window.localStorage.getItem(KEY) !== 'off';
  } catch {
    return true; // storage blocked: the default
  }
}

/**
 * Whether choosing an example also plays it, on the desks under trial
 * (#315). On by default; the choice is remembered in this browser.
 */
export function useHearOnOpen() {
  const [on, setOn] = useState(stored);
  const choose = useCallback((next: boolean) => {
    setOn(next);
    try {
      window.localStorage.setItem(KEY, next ? 'on' : 'off');
    } catch {
      // storage blocked: the choice lasts for this visit
    }
  }, []);
  return { on, choose };
}

/**
 * Space plays or stops the open example (#315), unless the focus is on
 * something Space already operates (a field, a button, a chord).
 */
export function useSpaceToPlay(enabled: boolean, toggle: () => void) {
  const toggleRef = useRef(toggle);
  useEffect(() => {
    toggleRef.current = toggle;
  });
  useEffect(() => {
    if (!enabled) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== ' ' || e.defaultPrevented || e.altKey || e.ctrlKey || e.metaKey) return;
      if (typing(e.target) || operable(e.target)) return;
      e.preventDefault();
      toggleRef.current();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [enabled]);
}

/** An element Space presses or selects already. */
function operable(target: EventTarget | null) {
  if (!(target instanceof Element)) return false;
  return target.closest('button, a, summary, [role="option"], [role="button"]') !== null;
}
