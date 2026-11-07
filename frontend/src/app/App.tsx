import { useState } from 'react';
import { ChatPanel } from '@/features/chat/ChatPanel';
import { useIdeas } from '@/features/chat/useIdeas';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import { DownloadMidiButton } from '@/features/export/DownloadMidiButton';
import { PlayerControls } from '@/features/player/PlayerControls';
import { createWebAudioEngine, usePlayer, type AudioEngine } from '@/features/player/usePlayer';
import { ProgressionForm } from '@/features/progression-input/ProgressionForm';
import { TimelineView } from '@/features/timeline/TimelineView';
import type { FromProgressionRequest, IdeaCandidate } from '@/shared/api/client';
import { useBackendHealth } from './useBackendHealth';
import { useTimelineGeneration } from './useTimelineGeneration';

type AppProps = {
  /** Injected in tests; defaults to a real Web Audio engine. */
  createEngine?: () => AudioEngine;
};

type Source = 'ideas' | 'progression';

const BACKEND_LABEL = {
  checking: 'バックエンドに接続中…',
  ok: 'バックエンド: ok',
  down: 'バックエンドに接続できません',
} as const;

/**
 * Scenarios S1 and S2 (docs/product/vision.md): get a timeline either from
 * an image interpreted by the AI or from a key and progression typed in;
 * then hear it, watch it, read why it works, and take the MIDI home.
 */
export function App({ createEngine = createWebAudioEngine }: AppProps) {
  const [source, setSource] = useState<Source>('ideas');
  const [selectedHarmony, setSelectedHarmony] = useState<number | null>(null);
  const [openTitle, setOpenTitle] = useState<string | null>(null);
  const backend = useBackendHealth();
  const generation = useTimelineGeneration();
  const ideas = useIdeas();
  const player = usePlayer(createEngine);
  const { timeline } = generation;

  const generate = async (req: FromProgressionRequest) => {
    player.stop();
    setOpenTitle(null);
    if (await generation.generate(req)) setSelectedHarmony(null);
  };

  const openCandidate = (c: IdeaCandidate) => {
    player.stop();
    generation.open(c.timeline);
    setOpenTitle(c.title);
    setSelectedHarmony(null);
  };

  return (
    <main className="app">
      <header className="app__header">
        <h1>middy</h1>
        <p>イメージや進行から曲の骨組みを作って、鳴らして、成り立ちを見る。</p>
      </header>

      <nav className="app__source" aria-label="入力方法">
        <button type="button" onClick={() => setSource('ideas')} aria-pressed={source === 'ideas'}>
          イメージから（AI）
        </button>
        <button
          type="button"
          onClick={() => setSource('progression')}
          aria-pressed={source === 'progression'}
        >
          進行から
        </button>
      </nav>

      {source === 'ideas' ? (
        <ChatPanel ideas={ideas} onOpenCandidate={openCandidate} openTitle={openTitle} />
      ) : (
        <ProgressionForm onSubmit={(req) => void generate(req)} busy={generation.busy} />
      )}
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
            {openTitle && <span className="app__open-title">候補: {openTitle}</span>}
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
