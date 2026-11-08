import { useState } from 'react';
import { ChatPanel } from '@/features/chat/ChatPanel';
import { useIdeas } from '@/features/chat/useIdeas';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import { KeySelect } from '@/features/key-change/KeySelect';
import { useKeyChange } from '@/features/key-change/useKeyChange';
import { DownloadMidiButton } from '@/features/export/DownloadMidiButton';
import { PlayerControls } from '@/features/player/PlayerControls';
import { createWebAudioEngine, usePlayer, type AudioEngine } from '@/features/player/usePlayer';
import { usePatterns } from '@/features/chat/usePatterns';
import { ProjectList } from '@/features/projects/ProjectList';
import { SaveProjectForm } from '@/features/projects/SaveProjectForm';
import { PwaBanner } from '@/features/pwa/PwaBanner';
import { usePwa } from '@/features/pwa/usePwa';
import { useProjects } from '@/features/projects/useProjects';
import { TimelineStage } from '@/features/timeline/lab/TimelineStage';
import { TIMELINE_LAYOUTS, isTimelineLayout } from '@/features/timeline/lab/variants';
import type { IdeaCandidate, PitchClass, Timeline } from '@/shared/api/client';
import type { Project, ProjectSource } from '@/shared/lib/storage/projects';
import { useBackendHealth } from './useBackendHealth';
import { useTimelineLayout } from './useTimelineLayout';

type AppProps = {
  /** Injected in tests; defaults to a real Web Audio engine. */
  createEngine?: () => AudioEngine;
  /** Injected in tests; defaults to IndexedDB (memory when unavailable). */
  projectStore?: Parameters<typeof useProjects>[0];
};

/**
 * The one screen of docs/product/experience.md: describe an image, hear the
 * examples it becomes, see how they are built, and take one home. The image
 * is the only entry; saved projects reopen the same timeline.
 */
export function App({ createEngine = createWebAudioEngine, projectStore }: AppProps) {
  const [timeline, setTimeline] = useState<Timeline | null>(null);
  const [selectedHarmony, setSelectedHarmony] = useState<number | null>(null);
  /** The open example itself: titles can repeat across answers (BL-070). */
  const [openExample, setOpenExample] = useState<IdeaCandidate | null>(null);
  /** Where the loaded timeline came from; what "保存" stores alongside it. */
  const [origin, setOrigin] = useState<ProjectSource | null>(null);
  const [projectName, setProjectName] = useState<string | null>(null);
  const patterns = usePatterns();
  const player = usePlayer(createEngine);
  const projects = useProjects(projectStore);
  const pwa = usePwa();
  const keyChange = useKeyChange();
  const timelineLayout = useTimelineLayout();
  // Normal is silent; only an unreachable backend is worth a word (BL-065).
  const backend = useBackendHealth(pwa.online);

  const show = (tl: Timeline) => {
    player.stop();
    setTimeline(tl);
    setSelectedHarmony(null);
  };

  const openCandidate = (c: IdeaCandidate, image: string) => {
    show(c.timeline);
    setOpenExample(c);
    setProjectName(null);
    setOrigin({ kind: 'idea', image, candidateTitle: c.title });
  };

  // An answer arrives ready to hear: its first example opens at once (BL-064).
  const ideas = useIdeas((answer) => {
    const first = answer.candidates[0];
    if (first) openCandidate(first, answer.image);
  });
  const lastImage = () => [...ideas.turns].reverse().find((t) => t.role === 'user')?.content ?? '';

  // Same example, another key: the engine re-renders it, the LLM is not asked (BL-078).
  const changeKey = async (tonic: PitchClass) => {
    if (!timeline) return;
    const next = await keyChange.change(timeline, tonic);
    if (next) show(next);
  };

  const openProject = (p: Project) => {
    show(p.timeline);
    setOpenExample(null);
    setProjectName(p.name);
    setOrigin(p.source);
  };

  return (
    <main className="app">
      <header className="app__header">
        <h1>middy</h1>
        <p>曲のイメージから例を鳴らして、その成り立ちを見る。</p>
      </header>
      <PwaBanner pwa={pwa} />
      {backend === 'down' && pwa.online && (
        <p role="status" className="app__notice">
          バックエンドに接続できません。少し待ってから再読み込みしてください。
        </p>
      )}
      <ProjectList projects={projects} onOpen={openProject} />

      <ChatPanel
        ideas={ideas}
        onOpenCandidate={(c) => openCandidate(c, lastImage())}
        openExample={openExample}
        patterns={patterns}
      />

      {timeline && (
        <section className="app__timeline" aria-label="結果">
          <div className="app__toolbar">
            <PlayerControls timeline={timeline} player={player} />
            <KeySelect
              tonic={timeline.key.tonic}
              mode={timeline.key.mode}
              busy={keyChange.busy}
              error={keyChange.error}
              onChange={(t) => void changeKey(t)}
            />
            <DownloadMidiButton timeline={timeline} />
            {origin && (
              <SaveProjectForm projects={projects} current={{ timeline, source: origin }} />
            )}
            {openExample && <span className="app__open-title">候補: {openExample.title}</span>}
            {projectName && <span className="app__open-title">プロジェクト: {projectName}</span>}
            {timelineLayout.lab && (
              <label className="app__lab">
                表示（試験）
                <select
                  value={timelineLayout.layout}
                  onChange={(e) => {
                    if (isTimelineLayout(e.target.value)) timelineLayout.choose(e.target.value);
                  }}
                >
                  {TIMELINE_LAYOUTS.map((l) => (
                    <option key={l.id} value={l.id}>
                      {l.name}
                    </option>
                  ))}
                </select>
              </label>
            )}
          </div>
          {/* Side by side on wide screens: hear, watch and read at once (BL-061). */}
          <div className="app__stage">
            <TimelineStage
              layout={timelineLayout.layout}
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
          </div>
        </section>
      )}
    </main>
  );
}
