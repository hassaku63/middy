import { useEffect, useRef, useState } from 'react';

/** Where focus goes when the drawer is moved by hand. */
export const CHAT_TOGGLE_ID = 'chat-toggle';
export const CHAT_CLOSE_ID = 'chat-close';

/**
 * The conversation as a drawer on a short, wide screen (a phone held
 * sideways, #498): opening an example folds the conversation away so the
 * timeline and the chord's frame take the screen; 「会話」 brings it back.
 * Elsewhere the drawer's state shows nowhere: CSS decides where it
 * matters (global.css), so the tree and its state stay the same when the
 * phone turns (ADR-0042).
 */
export function useChatDrawer() {
  const [open, setOpen] = useState(true);
  /** Focus follows a drawer moved by hand, not one folded by opening an example. */
  const byHand = useRef(false);
  useEffect(() => {
    if (!byHand.current) return;
    byHand.current = false;
    document.getElementById(open ? CHAT_CLOSE_ID : CHAT_TOGGLE_ID)?.focus();
  }, [open]);
  const set = (next: boolean) => {
    byHand.current = true;
    setOpen(next);
  };
  return {
    open,
    toggle: () => {
      set(!open);
    },
    close: () => {
      set(false);
    },
    /** An example opened: the conversation steps aside. */
    fold: () => {
      setOpen(false);
    },
  };
}

export type ChatDrawer = ReturnType<typeof useChatDrawer>;
