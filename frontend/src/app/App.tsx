import { useState } from 'react';
import { ChatPanel } from '@/features/chat/ChatPanel';
import { ExplanationPanel } from '@/features/explanation/ExplanationPanel';
import { KeySelect } from '@/features/key-change/KeySelect';
import { StyleSelect } from '@/features/regroove/StyleSelect';
import { DownloadMidiButton } from '@/features/export/DownloadMidiButton';
import { PlayerControls } from '@/features/player/PlayerControls';
import { createWebAudioEngine, usePlayer, type AudioEngine } from '@/features/player/usePlayer';
import { usePatterns } from '@/features/chat/usePatterns';
import { ProjectList } from '@/features/projects/ProjectList';
import { SaveProjectForm } from '@/features/projects/SaveProjectForm';
import { PwaBanner } from '@/features/pwa/PwaBanner';
import { usePwa } from '@/features/pwa/usePwa';
import { useProjects } from '@/features/projects/useProjects';
import { TimelineOverview } from '@/features/timeline/TimelineOverview';
import type { IdeaCandidate, Timeline } from '@/shared/api/client';
import type { Project, ProjectSource } from '@/shared/lib/storage/projects';
import { useBackendHealth } from './useBackendHealth';
import { LabControls } from './LabControls';
import { useDegreeMapTrial } from './useDegreeMapTrial';
import { useExampleChanges } from './useExampleChanges';
import { imageFor } from './imageFor';
import { useAnswerOpening } from './useAnswerOpening';

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
  const pwa = usePwa();
  // Normal is silent; only an unreachable backend is worth a word (BL-065).
  const backend = useBackendHealth(pwa.online);
  const patterns = usePatterns(backend !== 'down');
  const player = usePlayer(createEngine);
  const projects = useProjects(projectStore);
  const degreeMap = useDegreeMapTrial();

  const show = (tl: Timeline) => {
    player.stop();
    setTimeline(tl);
    setSelectedHarmony(null);
  };
  const { keyChange, regroove, changeKey, changeStyle, forget } = useExampleChanges(
    timeline,
    show,
    patterns.styles,
  );

  const openCandidate = (c: IdeaCandidate, image: string) => {
    show(c.timeline);
    forget();
    setOpenExample(c);
    setProjectName(null);
    setOrigin({ kind: 'idea', image, candidateTitle: c.title });
  };

  // An answer arrives ready to hear, unless the user chose another meanwhile.
  const { ideas, opened } = useAnswerOpening(openCandidate);

  const openProject = (p: Project) => {
    opened();
    show(p.timeline);
    forget();
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
        onOpenCandidate={(c) => {
          opened();
          openCandidate(c, imageFor(ideas.turns, c));
        }}
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
              busy={keyChange.busy || regroove.busy}
              error={keyChange.error}
              onChange={(t) => void changeKey(t)}
            />
            <StyleSelect
              styles={patterns.styles}
              current={timeline.rhythm?.style?.id}
              busy={regroove.busy || keyChange.busy}
              error={regroove.error}
              note={regroove.note}
              onChange={(id) => void changeStyle(id)}
            />
            <DownloadMidiButton timeline={timeline} />
            {origin && (
              <SaveProjectForm projects={projects} current={{ timeline, source: origin }} />
            )}
            {openExample && <span className="app__open-title">候補: {openExample.title}</span>}
            {projectName && <span className="app__open-title">保存した例: {projectName}</span>}
            <LabControls degreeMap={degreeMap} />
          </div>
          {/* Side by side on wide screens: hear, watch and read at once (BL-061). */}
          <div className="app__stage">
            <TimelineOverview
              timeline={timeline}
              positionTick={player.positionTick}
              playing={player.state === 'playing'}
              selectedHarmony={selectedHarmony}
              onSelectHarmony={setSelectedHarmony}
            />
            <ExplanationPanel
              timeline={timeline}
              positionTick={player.positionTick}
              selectedHarmony={selectedHarmony}
              onClearSelection={() => setSelectedHarmony(null)}
              showDegreeMap={degreeMap.on}
            />
          </div>
        </section>
      )}
    </main>
  );
}
