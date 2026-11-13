import { useCallback, useEffect, useRef, useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import { positionStore, type PlayerPosition } from './position';
import { TimelinePlayer, type Clock, type PlayerState, type Synth } from './TimelinePlayer';
import { WebAudioSynth } from './WebAudioSynth';

export type AudioEngine = {
  synth: Synth;
  clock: Clock;
  resume(): Promise<void>;
  /** Releases the audio device (optional: test engines have none). */
  close?(): Promise<void>;
};

/**
 * Builds the Web Audio engine. Called lazily from a user gesture (the play
 * button) because browsers refuse to start an AudioContext otherwise.
 */
export function createWebAudioEngine(): AudioEngine {
  const ctx = new AudioContext();
  return {
    synth: new WebAudioSynth(ctx),
    clock: { now: () => ctx.currentTime },
    resume: () => ctx.resume(),
    close: () => ctx.close(),
  };
}

/**
 * What the player last did that a screen reader should hear (#500): only
 * what leaves it silent, so nothing is said over the music. A seek while
 * playing and play itself say nothing.
 */
export type PlayerNotice = {
  kind: 'pause' | 'stop' | 'end' | 'seek';
  tick: number;
  /** Counts the notices, so the same words twice are said twice. */
  n: number;
};

export type UsePlayer = {
  state: PlayerState;
  /** Where it is; read it with usePosition, in what shows it (#500). */
  position: PlayerPosition;
  notice: PlayerNotice | null;
  /** Why the last play made no sound (the audio device refused), or null (#500). */
  error: string | null;
  /** Plays from the playhead: where it paused or was moved to, else the beginning. */
  play(timeline: Timeline): Promise<void>;
  /**
   * Stops where it is; play goes on from there (BL-334). Quiet, it says
   * nothing (a drag of the playhead holds the sound, #500).
   */
  pause: (options?: { quiet?: boolean }) => void;
  /** Stops and goes back to the beginning, as a recorder's ■ (BL-334). */
  stop: () => void;
  /**
   * Moves the position (BL-327): playing, play goes on from tick; stopped,
   * the next play starts there.
   */
  seek: (tick: number) => void;
};

const CANNOT_PLAY = '音を鳴らせませんでした。もう一度「再生」を押してください。';

/**
 * React binding for TimelinePlayer. The engine is created on the first
 * play() and reused afterwards; `createEngine` is injectable for tests.
 */
export function usePlayer(createEngine: () => AudioEngine = createWebAudioEngine): UsePlayer {
  const engineRef = useRef<AudioEngine | null>(null);
  const playerRef = useRef<TimelinePlayer | null>(null);
  const [state, setState] = useState<PlayerState>('stopped');
  const [position] = useState(positionStore);
  // Counts play requests; stop() and unmounting count too, so a play()
  // still waiting for the audio device does not start once it is outdated.
  const requests = useRef(0);
  // Where play starts; kept here too because the player is made on the
  // first play, after a seek may already have set it.
  const cue = useRef(0);
  const [notice, setNotice] = useState<PlayerNotice | null>(null);
  const [error, setError] = useState<string | null>(null);
  const tell = useCallback((kind: PlayerNotice['kind'], tick: number) => {
    setNotice((prev) => ({ kind, tick, n: (prev?.n ?? 0) + 1 }));
  }, []);
  /** Why the player is about to stop; none means it played to the end. */
  const why = useRef<PlayerNotice['kind'] | 'quiet' | null>(null);
  const wasPlaying = useRef(false);

  /** The player, or null when the hook unmounted while the device resumed. */
  const ensurePlayer = useCallback(async () => {
    engineRef.current ??= createEngine();
    const engine = engineRef.current;
    await engine.resume();
    if (engineRef.current !== engine) return null; // closed while resuming
    if (!playerRef.current) {
      const player = new TimelinePlayer({
        synth: engine.synth,
        clock: engine.clock,
      });
      player.subscribe((tick, s) => {
        position.set(tick);
        setState(s);
        if (s === 'stopped') cue.current = tick; // back to the beginning at the end
        if (s === 'stopped' && wasPlaying.current) {
          const kind = why.current ?? 'end';
          if (kind !== 'quiet') tell(kind, tick);
        }
        if (s === 'stopped') why.current = null;
        wasPlaying.current = s === 'playing';
      });
      playerRef.current = player;
    }
    return playerRef.current;
  }, [createEngine, tell, position]);

  const play = useCallback(
    async (timeline: Timeline) => {
      const request = ++requests.current;
      setError(null);
      let player: TimelinePlayer | null;
      try {
        player = await ensurePlayer();
      } catch {
        // No AudioContext, or the browser would not start it: say so, and
        // the next play asks again (the engine is kept or made anew).
        if (request === requests.current) setError(CANNOT_PLAY);
        return;
      }
      if (!player || request !== requests.current) return; // stopped or replaced meanwhile
      const from = cue.current; // load() reports the stop of what played before
      why.current = 'quiet'; // another example takes over: nothing ended
      player.load(timeline);
      why.current = null;
      player.play(from);
    },
    [ensurePlayer],
  );

  const pause = useCallback((options?: { quiet?: boolean }) => {
    requests.current += 1;
    why.current = options?.quiet ? 'quiet' : 'pause';
    playerRef.current?.pause(); // reports where it paused, which sets the cue
  }, []);

  const stop = useCallback(() => {
    requests.current += 1;
    const playing = playerRef.current?.isPlaying ?? false;
    why.current = 'stop';
    playerRef.current?.stop();
    if (!playing) {
      why.current = null;
      tell('stop', 0); // already silent: the player reports no stop
    }
    cue.current = 0;
    position.set(0);
  }, [tell, position]);

  const seek = useCallback(
    (tick: number) => {
      // Stopped, the player may still hold the timeline shown before.
      const player = playerRef.current;
      if (player?.isPlaying) {
        player.seek(tick); // pausing later reports where it stopped, which sets the cue
        return;
      }
      cue.current = Math.max(0, Math.round(tick));
      position.set(cue.current);
      tell('seek', cue.current);
    },
    [tell, position],
  );

  // Unmounting stops the sound and releases the AudioContext: browsers
  // cap how many may be open at once.
  useEffect(
    () => () => {
      requests.current += 1;
      playerRef.current?.stop();
      void engineRef.current?.close?.();
      engineRef.current = null;
      playerRef.current = null;
    },
    [],
  );

  return { state, position, notice, error, play, pause, stop, seek };
}
