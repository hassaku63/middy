import { useState } from 'react';
import { ChatPanel } from '@/features/chat/ChatPanel';
import { useIdeas } from '@/features/chat/useIdeas';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import { DownloadMidiButton } from '@/features/export/DownloadMidiButton';
import { PlayerControls } from '@/features/player/PlayerControls';
import { createWebAudioEngine, usePlayer, type AudioEngine } from '@/features/player/usePlayer';
import { usePatterns } from '@/features/chat/usePatterns';
import { ProjectsPanel } from '@/features/projects/ProjectsPanel';
import { PwaBanner } from '@/features/pwa/PwaBanner';
import { usePwa } from '@/features/pwa/usePwa';
import { useProjects } from '@/features/projects/useProjects';
import { TimelineView } from '@/features/timeline/TimelineView';
import type { IdeaCandidate, Timeline } from '@/shared/api/client';
import type { Project, ProjectSource } from '@/shared/lib/storage/projects';
import { useBackendHealth } from './useBackendHealth';

type AppProps = {
  /** Injected in tests; defaults to a real Web Audio engine. */
  createEngine?: () => AudioEngine;
  /** Injected in tests; defaults to IndexedDB (memory when unavailable). */
  projectStore?: Parameters<typeof useProjects>[0];
};

const BACKEND_LABEL = {
  checking: 'バックエンドに接続中…',
  ok: 'バックエンド: ok',
  down: 'バックエンドに接続できません',
} as const;

/**
 * The one screen of docs/product/experience.md: describe an image, hear the
 * examples it becomes, see how they are built, and take one home. The image
 * is the only entry; saved projects reopen the same timeline.
 */
export function App({ createEngine = createWebAudioEngine, projectStore }: AppProps) {
  const [timeline, setTimeline] = useState<Timeline | null>(null);
  const [selectedHarmony, setSelectedHarmony] = useState<number | null>(null);
  const [openTitle, setOpenTitle] = useState<string | null>(null);
  /** Where the loaded timeline came from; what "保存" stores alongside it. */
  const [origin, setOrigin] = useState<ProjectSource | null>(null);
  const [projectName, setProjectName] = useState<string | null>(null);
  const backend = useBackendHealth();
  const patterns = usePatterns();
  const player = usePlayer(createEngine);
  const projects = useProjects(projectStore);
  const pwa = usePwa();

  const show = (tl: Timeline) => {
    player.stop();
    setTimeline(tl);
    setSelectedHarmony(null);
  };

  const openCandidate = (c: IdeaCandidate, image: string) => {
    show(c.timeline);
    setOpenTitle(c.title);
    setProjectName(null);
    setOrigin({ kind: 'idea', image, candidateTitle: c.title });
  };

  // An answer arrives ready to hear: its first example opens at once (BL-064).
  const ideas = useIdeas((answer) => {
    const first = answer.candidates[0];
    if (first) openCandidate(first, answer.image);
  });
  const lastImage = () => [...ideas.turns].reverse().find((t) => t.role === 'user')?.content ?? '';

  const openProject = (p: Project) => {
    show(p.timeline);
    setOpenTitle(null);
    setProjectName(p.name);
    setOrigin(p.source);
  };

  return (
    <main className="app">
      <header className="app__header">
        <h1>middy</h1>
        <p>イメージや進行から曲の骨組みを作って、鳴らして、成り立ちを見る。</p>
      </header>
      <PwaBanner pwa={pwa} />

      <ChatPanel
        ideas={ideas}
        onOpenCandidate={(c) => openCandidate(c, lastImage())}
        openTitle={openTitle}
        patterns={patterns}
      />

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
