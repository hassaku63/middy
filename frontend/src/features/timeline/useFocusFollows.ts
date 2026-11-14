import { useEffect, type RefObject } from 'react';

/**
 * Moves the focus to the option at index when it changes while the focus
 * is inside the list (#500): the keys move the pinned chord, and the one
 * tab stop of the lane goes with it.
 */
export function useFocusFollows(list: RefObject<SVGGElement | null>, index: number | null) {
  useEffect(() => {
    const g = list.current;
    if (index === null || !g?.contains(document.activeElement)) return;
    const option = g.querySelectorAll<SVGGElement>('[role=option]')[index];
    if (option && option !== document.activeElement) option.focus();
  }, [list, index]);
}
