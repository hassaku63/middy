import { useState } from 'react';
import { ChatPanel } from '@/features/chat/ChatPanel';
import { useIdeas } from '@/features/chat/useIdeas';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import { DownloadMidiButton } from '@/features/export/DownloadMidiButton';
import { PlayerControls } from '@/features/player/PlayerControls';
import { createWebAudioEngine, usePlayer, type AudioEngine } from '@/features/player/usePlayer';
import { ProgressionForm } from '@/features/progression-input/ProgressionForm';
import { usePatterns } from '@/features/progression-input/usePatterns';
import { ProjectsPanel } from '@/features/projects/ProjectsPanel';
import { useProjects } from '@/features/projects/useProjects';
import { TimelineView } from '@/features/timeline/TimelineView';
import type { FromProgressionRequest, IdeaCandidate } from '@/shared/api/client';
import type { Project, ProjectSource } from '@/shared/lib/storage/projects';
import { useBackendHealth } from './useBackendHealth';
import { useTimelineGeneration } from './useTimelineGeneration';

type AppProps = {
  /** Injected in tests; defaults to a real Web Audio engine. */
  createEngine?: () => AudioEngine;
  /** Injected in tests; defaults to IndexedDB (memory when unavailable). */
  projectStore?: Parameters<typeof useProjects>[0];
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
export function App({ createEngine = createWebAudioEngine, projectStore }: AppProps) {
  const [source, setSource] = useState<Source>('ideas');
  const [selectedHarmony, setSelectedHarmony] = useState<number | null>(null);
  const [openTitle, setOpenTitle] = useState<string | null>(null);
  /** Where the loaded timeline came from; what "保存" stores alongside it. */
  const [origin, setOrigin] = useState<ProjectSource | null>(null);
  const [projectName, setProjectName] = useState<string | null>(null);
  const backend = useBackendHealth();
  const generation = useTimelineGeneration();
  const ideas = useIdeas();
  const patterns = usePatterns();
  const player = usePlayer(createEngine);
  const projects = useProjects(projectStore);
  const { timeline } = generation;

  const generate = async (req: FromProgressionRequest) => {
    player.stop();
    setOpenTitle(null);
    setProjectName(null);
    if (await generation.generate(req)) {
      setSelectedHarmony(null);
      setOrigin({ kind: 'progression', request: req });
    }
  };

  const openCandidate = (c: IdeaCandidate) => {
    player.stop();
    generation.open(c.timeline);
    setOpenTitle(c.title);
    setProjectName(null);
    setSelectedHarmony(null);
    const image = [...ideas.turns].reverse().find((t) => t.role === 'user')?.content ?? '';
    setOrigin({ kind: 'idea', image, candidateTitle: c.title });
  };

  const openProject = (p: Project) => {
    player.stop();
    generation.open(p.timeline);
    setOpenTitle(null);
    setProjectName(p.name);
    setSelectedHarmony(null);
    setOrigin(p.source);
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
        <ChatPanel
          ideas={ideas}
          onOpenCandidate={openCandidate}
          openTitle={openTitle}
          patterns={patterns}
        />
      ) : (
        <ProgressionForm
          onSubmit={(req) => void generate(req)}
          busy={generation.busy}
          patterns={patterns}
        />
      )}
      {generation.error && (
        <p role="alert" className="app__error">
          {generation.error}
        </p>
      )}

      <ProjectsPanel
        projects={projects}
        current={timeline && origin ? { timeline, source: origin } : null}
        onOpen={openProject}
      />

      {timeline && (
        <section className="app__timeline" aria-label="結果">
          <div className="app__toolbar">
            <PlayerControls timeline={timeline} player={player} />
            <DownloadMidiButton timeline={timeline} />
            {openTitle && <span className="app__open-title">候補: {openTitle}</span>}
            {projectName && <span className="app__open-title">プロジェクト: {projectName}</span>}
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
