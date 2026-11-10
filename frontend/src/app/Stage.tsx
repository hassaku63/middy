import { ChordFrame } from '@/features/explanation/ChordFrame';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import type { UsePlayer } from '@/features/player/usePlayer';
import { TimelineOverview } from '@/features/timeline/TimelineOverview';
import type { Timeline } from '@/shared/api/client';
import { useChordKeys } from './useChordKeys';

type Props = {
  timeline: Timeline;
  player: UsePlayer;
  selectedHarmony: number | null;
  onSelectHarmony: (index: number | null) => void;
};

/**
 * What the example looks like (BL-061): the timeline across the whole
 * stage, the chord's frame and the other explanations side by side below
 * it (the wide layout the owner chose, #279).
 */
export function Stage({ timeline, player, selectedHarmony, onSelectHarmony }: Props) {
  useChordKeys({
    enabled: true,
    harmony: timeline.harmony,
    positionTick: player.positionTick,
    selected: selectedHarmony,
    onSelect: onSelectHarmony,
  });
  const clear = () => onSelectHarmony(null);
  const overview = (
    <TimelineOverview
      timeline={timeline}
      positionTick={player.positionTick}
      playing={player.state === 'playing'}
      selectedHarmony={selectedHarmony}
      onSelectHarmony={onSelectHarmony}
      marks
    />
  );
  const frame = (
    <ChordFrame
      timeline={timeline}
      positionTick={player.positionTick}
      selectedHarmony={selectedHarmony}
      onClearSelection={clear}
    />
  );
  const notes = (
    <ExplanationPanel
      timeline={timeline}
      positionTick={player.positionTick}
      selectedHarmony={selectedHarmony}
      onClearSelection={clear}
      chord={false}
    />
  );
  return (
    <div className="app__stage app__stage--wide">
      {overview}
      <div className="app__below">
        {frame}
        {notes}
      </div>
    </div>
  );
}
