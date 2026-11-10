import type { UsePlayer } from '@/features/player/usePlayer';
import type { IdeaCandidate, Timeline } from '@/shared/api/client';
import { useHearOnOpen, useSpaceToPlay } from './useHearOnOpen';

type Options = {
  openExample: IdeaCandidate | null;
  timeline: Timeline | null;
  player: UsePlayer;
  /** Opens a chosen example (the screen's own opening). */
  open: (candidate: IdeaCandidate) => void;
};

/**
 * Choosing an example by the learner's own hand (a card, a line, a digit)
 * also plays it, and choosing the playing one again stops it;
 * Space plays or stops (#315). An arriving answer still opens silently.
 */
export function useChooseToHear({ openExample, timeline, player, open }: Options) {
  const hear = useHearOnOpen();
  const playing = player.state === 'playing';
  const openChosen = (c: IdeaCandidate) => {
    const plays = hear.on;
    if (plays && c === openExample && playing) {
      player.stop();
      return;
    }
    if (c !== openExample) open(c);
    if (plays) void player.play(c.timeline);
  };
  useSpaceToPlay(timeline !== null, () => {
    if (playing) player.stop();
    else if (timeline) void player.play(timeline);
  });
  return { openChosen, hear };
}
