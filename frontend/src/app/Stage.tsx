import { ChordFrame } from '@/features/explanation/ChordFrame';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import { usePosition } from '@/features/player/position';
import type { UsePlayer } from '@/features/player/usePlayer';
import { TimelineOverview } from '@/features/timeline/TimelineOverview';
import type { ReactNode } from 'react';
import type { Timeline } from '@/shared/api/client';
import { useChordKeys } from './useChordKeys';
import { useScrub } from './useScrub';

type Props = {
  timeline: Timeline;
  player: UsePlayer;
  selectedHarmony: number | null;
  onSelectHarmony: (index: number | null) => void;
  /** The open example's card, at the head of the explanation (BL-298). */
  example?: ReactNode;
  /** The explanation as text, for its copy button (BL-300). */
  copyText?: string | undefined;
};

/**
 * What the example looks like (BL-061): the timeline across the whole
 * stage, the chord's frame and the other explanations side by side below
 * it (the wide layout the owner chose, #279).
 */
export function Stage({
  timeline,
  player,
  selectedHarmony,
  onSelectHarmony,
  example,
  copyText,
}: Props) {
  // The stage shows the position; the screen around it is not drawn again (#500).
  const positionTick = usePosition(player.position);
  useChordKeys({
    enabled: true,
    harmony: timeline.harmony,
    positionTick,
    selected: selectedHarmony,
    onSelect: onSelectHarmony,
  });
  const clear = () => onSelectHarmony(null);
  // Stopped, the seek sets where play starts; playing, it jumps there.
  const scrub = useScrub(player, timeline);
  const playFrom = (tick: number) => {
    player.seek(tick);
    if (player.state !== 'playing') void player.play(timeline);
  };
  const overview = (
    <TimelineOverview
      timeline={timeline}
      positionTick={positionTick}
      playing={player.state === 'playing'}
      selectedHarmony={selectedHarmony}
      onSelectHarmony={onSelectHarmony}
      onSeek={player.seek}
      onScrub={scrub}
      marks
      fit
    />
  );
  const frame = (
    <ChordFrame
      timeline={timeline}
      positionTick={positionTick}
      selectedHarmony={selectedHarmony}
      onClearSelection={clear}
      onPlayFrom={playFrom}
    />
  );
  const notes = (
    <ExplanationPanel
      timeline={timeline}
      positionTick={positionTick}
      selectedHarmony={selectedHarmony}
      onClearSelection={clear}
      chord={false}
      example={example}
      copyText={copyText}
    />
  );
  return (
    <div className="app__stage">
      {overview}
      <div className="app__below">
        {frame}
        {notes}
      </div>
    </div>
  );
}
