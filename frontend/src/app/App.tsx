import { useState } from 'react';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import { DownloadMidiButton } from '@/features/export/DownloadMidiButton';
import { PlayerControls } from '@/features/player/PlayerControls';
import { createWebAudioEngine, usePlayer, type AudioEngine } from '@/features/player/usePlayer';
import { ProgressionForm } from '@/features/progression-input/ProgressionForm';
import { TimelineView } from '@/features/timeline/TimelineView';
import type { FromProgressionRequest } from '@/shared/api/client';
import { useBackendHealth } from './useBackendHealth';
import { useTimelineGeneration } from './useTimelineGeneration';

type AppProps = {
  /** Injected in tests; defaults to a real Web Audio engine. */
  createEngine?: () => AudioEngine;
};

const BACKEND_LABEL = {
  checking: 'バックエンドに接続中…',
  ok: 'バックエンド: ok',
  down: 'バックエンドに接続できません',
} as const;

/**
 * Scenario S1 (docs/product/vision.md): choose a key and a progression,
 * hear it, watch the chord and notes light up, read why it works.
 */
export function App({ createEngine = createWebAudioEngine }: AppProps) {
  const [selectedHarmony, setSelectedHarmony] = useState<number | null>(null);
  const backend = useBackendHealth();
  const generation = useTimelineGeneration();
  const player = usePlayer(createEngine);
  const { timeline } = generation;

  const generate = async (req: FromProgressionRequest) => {
    player.stop();
    if (await generation.generate(req)) setSelectedHarmony(null);
  };

  return (
    <main className="app">
      <header className="app__header">
        <h1>middy</h1>
        <p>調と進行を選んで、鳴らして、成り立ちを見る。</p>
      </header>

      <ProgressionForm onSubmit={(req) => void generate(req)} busy={generation.busy} />
      {generation.error && (
        <p role="alert" className="app__error">
          {generation.error}
        </p>
      )}

      {timeline && (
        <section className="app__timeline" aria-label="結果">
          <div className="app__toolbar">
            <PlayerControls timeline={timeline} player={player} />
            <DownloadMidiButton timeline={timeline} />
          </div>
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
        <span role="status">{BACKEND_LABEL[backend]}</span>
      </footer>
    </main>
  );
}
