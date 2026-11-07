import { useCallback, useEffect, useRef, useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import { TimelinePlayer, type Clock, type PlayerState, type Synth } from './TimelinePlayer';
import { WebAudioSynth } from './WebAudioSynth';

export type AudioEngine = { synth: Synth; clock: Clock; resume(): Promise<void> };

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
  };
}

export type UsePlayer = {
  state: PlayerState;
  positionTick: number;
  play(timeline: Timeline): Promise<void>;
  stop(): void;
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

  const ensurePlayer = useCallback(async () => {
    if (!engineRef.current) engineRef.current = createEngine();
    await engineRef.current.resume();
    if (!playerRef.current) {
      const player = new TimelinePlayer({
        synth: engineRef.current.synth,
        clock: engineRef.current.clock,
      });
      player.subscribe((tick, s) => {
        setPositionTick(tick);
        setState(s);
      });
      playerRef.current = player;
    }
    return playerRef.current;
  }, [createEngine]);

  const play = useCallback(
    async (timeline: Timeline) => {
      const player = await ensurePlayer();
      player.load(timeline);
      player.play();
    },
    [ensurePlayer],
  );

  const stop = useCallback(() => {
    playerRef.current?.stop();
  }, []);

  useEffect(() => () => playerRef.current?.stop(), []);

  return { state, positionTick, play, stop };
}
