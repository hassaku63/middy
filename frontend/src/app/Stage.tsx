import { ChordFrame } from '@/features/explanation/ChordFrame';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import type { UsePlayer } from '@/features/player/usePlayer';
import { TimelineOverview } from '@/features/timeline/TimelineOverview';
import type { Timeline } from '@/shared/api/client';
import { useChordKeys } from './useChordKeys';

type Props = {
  timeline: Timeline;
  player: UsePlayer;
  /** How the detail of a long timeline moves (BL-127). */
  follow: boolean;
  /** The desk under trial (BL-184): the chord's explanation in a frame above the timeline. */
  desk: boolean;
  selectedHarmony: number | null;
  onSelectHarmony: (index: number | null) => void;
};

/** What the example looks like: the timeline and its explanation (BL-061). */
export function Stage({ timeline, player, follow, desk, selectedHarmony, onSelectHarmony }: Props) {
  useChordKeys({
    enabled: desk,
    harmony: timeline.harmony,
    positionTick: player.positionTick,
    selected: selectedHarmony,
    onSelect: onSelectHarmony,
  });
  const clear = () => onSelectHarmony(null);
  const overview = (
    <TimelineOverview
      follow={follow}
      timeline={timeline}
      positionTick={player.positionTick}
      playing={player.state === 'playing'}
      selectedHarmony={selectedHarmony}
      onSelectHarmony={onSelectHarmony}
      marks={desk}
    />
  );
  return (
    // Side by side on wide screens: hear, watch and read at once (BL-061).
    <div className="app__stage">
      {desk ? (
        <div className="app__score">
          <ChordFrame
            timeline={timeline}
            positionTick={player.positionTick}
            selectedHarmony={selectedHarmony}
            onClearSelection={clear}
          />
          {overview}
        </div>
      ) : (
        overview
      )}
      <ExplanationPanel
        timeline={timeline}
        positionTick={player.positionTick}
        selectedHarmony={selectedHarmony}
        onClearSelection={clear}
        chord={!desk}
      />
    </div>
  );
}
