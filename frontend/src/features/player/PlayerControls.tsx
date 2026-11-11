import type { Timeline } from '@/shared/api/client';
import { formatPosition } from '@/shared/lib/timing';
import type { UsePlayer } from './usePlayer';
import './PlayerControls.css';

type Props = {
  timeline: Timeline | null;
  player: UsePlayer;
};

/** A play/stop button with a bar.beat read-out. Purely presentational. */
export function PlayerControls({ timeline, player }: Props) {
  const playing = player.state === 'playing';
  const position = timeline
    ? formatPosition(player.positionTick, timeline.ppq, timeline.timeSignature)
    : '-';

  return (
    <div role="group" aria-label="再生" className="player">
      {/* The main operation once an example is open (experience.md): it
          stands out from the toolbar's other buttons (#279). One button
          plays and stops, in the same place and size, so the operation
          that can be done now is the one that stands out (#402 B). */}
      <button
        type="button"
        className="player__play"
        onClick={() => {
          if (playing) player.stop();
          else if (timeline) void player.play(timeline);
        }}
        disabled={!timeline}
      >
        <span className="player__icon" aria-hidden="true">
          {playing ? '■' : '▶'}
        </span>
        {playing ? '停止' : '再生'}
      </button>
      <output aria-label="再生位置">{position}</output>
    </div>
  );
}
