import { useSyncExternalStore } from 'react';

/**
 * Where the player is, as a value to subscribe to (#500): it changes many
 * times a second while playing, so only what shows it reads it, and the
 * rest of the screen is not drawn again on every beat.
 */
export type PlayerPosition = {
  get: () => number;
  subscribe: (onChange: () => void) => () => void;
};

/** A position that tells its subscribers when it changes. */
export function positionStore(): PlayerPosition & { set: (tick: number) => void } {
  let tick = 0;
  const listeners = new Set<() => void>();
  return {
    get: () => tick,
    set: (next) => {
      if (next === tick) return;
      tick = next;
      listeners.forEach((l) => l());
    },
    subscribe: (onChange) => {
      listeners.add(onChange);
      return () => listeners.delete(onChange);
    },
  };
}

/** The player's position, drawing the calling component again when it moves. */
export function usePosition(position: PlayerPosition): number {
  return useSyncExternalStore(position.subscribe, position.get);
}
