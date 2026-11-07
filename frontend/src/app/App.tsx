import { useCallback, useEffect, useState } from 'react';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import { PlayerControls } from '@/features/player/PlayerControls';
import { createWebAudioEngine, usePlayer, type AudioEngine } from '@/features/player/usePlayer';
import { ProgressionForm } from '@/features/progression-input/ProgressionForm';
import { TimelineView } from '@/features/timeline/TimelineView';
import {
  ApiError,
  createTimelineFromProgression,
  fetchHealth,
  type FromProgressionRequest,
  type Timeline,
} from '@/shared/api/client';

type AppProps = {
  /** Injected in tests; defaults to a real Web Audio engine. */
  createEngine?: () => AudioEngine;
};

/**
 * Scenario S1 (docs/product/vision.md): choose a key and a progression,
 * hear it, watch the chord and notes light up, read why it works.
 */
export function App({ createEngine = createWebAudioEngine }: AppProps) {
  const [timeline, setTimeline] = useState<Timeline | null>(null);
  const [selectedHarmony, setSelectedHarmony] = useState<number | null>(null);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [backend, setBackend] = useState<'checking' | 'ok' | 'down'>('checking');
  const player = usePlayer(createEngine);

  useEffect(() => {
    let cancelled = false;
    fetchHealth()
      .then(() => !cancelled && setBackend('ok'))
      .catch(() => !cancelled && setBackend('down'));
    return () => {
      cancelled = true;
    };
  }, []);

  const generate = useCallback(
    async (req: FromProgressionRequest) => {
      setBusy(true);
      setError(null);
      player.stop();
      try {
        const tl = await createTimelineFromProgression(req);
        setTimeline(tl);
        setSelectedHarmony(null);
      } catch (e) {
        setError(e instanceof ApiError ? e.message : 'バックエンドに接続できません');
      } finally {
        setBusy(false);
      }
    },
    [player],
  );

  return (
    <main className="app">
      <header className="app__header">
        <h1>middy</h1>
        <p>調と進行を選んで、鳴らして、成り立ちを見る。</p>
      </header>

      <ProgressionForm onSubmit={(req) => void generate(req)} busy={busy} />
      {error && (
        <p role="alert" className="app__error">
          {error}
        </p>
      )}

      {timeline && (
        <section className="app__timeline" aria-label="結果">
          <PlayerControls timeline={timeline} player={player} />
          <TimelineView
            timeline={timeline}
            positionTick={player.positionTick}
            selectedHarmony={selectedHarmony}
            onSelectHarmony={setSelectedHarmony}
          />
          <ExplanationPanel
            timeline={timeline}
            positionTick={player.positionTick}
            selectedHarmony={selectedHarmony}
            onClearSelection={() => setSelectedHarmony(null)}
          />
        </section>
      )}

      <footer className="app__footer">
        <span role="status">
          {backend === 'checking' && 'バックエンドに接続中…'}
          {backend === 'ok' && 'バックエンド: ok'}
          {backend === 'down' && 'バックエンドに接続できません'}
        </span>
      </footer>
    </main>
  );
}
