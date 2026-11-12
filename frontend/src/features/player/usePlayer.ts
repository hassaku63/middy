import { useCallback, useEffect, useRef, useState } from 'react';
import type { Timeline } from '@/shared/api/client';
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

export type UsePlayer = {
  state: PlayerState;
  positionTick: number;
  /** Plays from the playhead: where it paused or was moved to, else the beginning. */
  play(timeline: Timeline): Promise<void>;
  /** Stops where it is; play goes on from there (BL-334). */
  pause: () => void;
  /** Stops and goes back to the beginning, as a recorder's ■ (BL-334). */
  stop: () => void;
  /**
   * Moves the position (BL-327): playing, play goes on from tick; stopped,
   * the next play starts there.
   */
  seek: (tick: number) => void;
};

/**
 * React binding for TimelinePlayer. The engine is created on the first
 * play() and reused afterwards; `createEngine` is injectable for tests.
 */
export function usePlayer(createEngine: () => AudioEngine = createWebAudioEngine): UsePlayer {
  const engineRef = useRef<AudioEngine | null>(null);
  const playerRef = useRef<TimelinePlayer | null>(null);
  const [state, setState] = useState<PlayerState>('stopped');
  const [positionTick, setPositionTick] = useState(0);
  // Counts play requests; stop() and unmounting count too, so a play()
  // still waiting for the audio device does not start once it is outdated.
  const requests = useRef(0);
  // Where play starts; kept here too because the player is made on the
  // first play, after a seek may already have set it.
  const cue = useRef(0);

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
        setPositionTick(tick);
        setState(s);
        if (s === 'stopped') cue.current = tick; // back to the beginning at the end
      });
      playerRef.current = player;
    }
    return playerRef.current;
  }, [createEngine]);

  const play = useCallback(
    async (timeline: Timeline) => {
      const request = ++requests.current;
      const player = await ensurePlayer();
      if (!player || request !== requests.current) return; // stopped or replaced meanwhile
      const from = cue.current; // load() reports the stop of what played before
      player.load(timeline);
      player.play(from);
    },
    [ensurePlayer],
  );

  const pause = useCallback(() => {
    requests.current += 1;
    playerRef.current?.pause(); // reports where it paused, which sets the cue
  }, []);

  const stop = useCallback(() => {
    requests.current += 1;
    playerRef.current?.stop();
    cue.current = 0;
    setPositionTick(0);
  }, []);

  const seek = useCallback((tick: number) => {
    // Stopped, the player may still hold the timeline shown before.
    const player = playerRef.current;
    if (player?.isPlaying) {
      player.seek(tick); // pausing later reports where it stopped, which sets the cue
      return;
    }
    cue.current = Math.max(0, Math.round(tick));
    setPositionTick(cue.current);
  }, []);

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

  return { state, positionTick, play, pause, stop, seek };
}
