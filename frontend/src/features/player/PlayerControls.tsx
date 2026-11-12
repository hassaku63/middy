import type { Timeline } from '@/shared/api/client';
import { formatPosition } from '@/shared/lib/timing';
import type { UsePlayer } from './usePlayer';
import './PlayerControls.css';

type Props = {
  timeline: Timeline | null;
  player: UsePlayer;
};

/**
 * Play/pause, stop and a bar.beat read-out (BL-334, as a recorder): ▶ plays
 * from the playhead, ⏸ stops where it is, ■ stops and goes back to the
 * beginning. Purely presentational.
 */
export function PlayerControls({ timeline, player }: Props) {
  const playing = player.state === 'playing';
  const position = timeline
    ? formatPosition(player.positionTick, timeline.ppq, timeline.timeSignature)
    : '-';

  return (
    <div role="group" aria-label="再生" className="player">
      {/* The main operation once an example is open (experience.md): it
          stands out from the toolbar's other buttons (#279). One button
          plays and pauses, in the same place and size, so the operation
          that can be done now is the one that stands out (#402 B). */}
      <button
        type="button"
        className="player__play"
        onClick={() => {
          if (playing) player.pause();
          else if (timeline) void player.play(timeline);
        }}
        disabled={!timeline}
      >
        <span className="player__icon" aria-hidden="true">
          {playing ? '⏸' : '▶'}
        </span>
        {playing ? '一時停止' : '再生'}
      </button>
      {/* Back to the beginning: quieter than play, and off when there is
          nothing to stop or rewind. */}
      <button
        type="button"
        className="player__stop"
        onClick={player.stop}
        disabled={!timeline || (!playing && player.positionTick === 0)}
      >
        <span className="player__icon" aria-hidden="true">
          ■
        </span>
        停止
      </button>
      <output aria-label="再生位置">{position}</output>
    </div>
  );
}
