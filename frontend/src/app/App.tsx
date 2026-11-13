import { useState } from 'react';
import { ChatPanel } from '@/features/chat/ChatPanel';
import { KeySelect } from '@/features/key-change/KeySelect';
import { StyleSelect } from '@/features/regroove/StyleSelect';
import { DownloadMidiButton } from '@/features/export/DownloadMidiButton';
import { PlayerControls } from '@/features/player/PlayerControls';
import { createWebAudioEngine, usePlayer, type AudioEngine } from '@/features/player/usePlayer';
import { usePatterns } from '@/features/chat/usePatterns';
import { ExampleNotes } from '@/features/chat/ExampleNotes';
import { cardLines } from '@/features/chat/cardText';
import { explanationLines } from '@/features/explanation/explanationText';
import { ProjectList } from '@/features/projects/ProjectList';
import { KeepToggle } from '@/features/projects/KeepToggle';
import { keptAs } from '@/features/projects/keptAs';
import { InstallButton, PwaBanner } from '@/features/pwa/PwaBanner';
import { usePwa } from '@/features/pwa/usePwa';
import { useProjects } from '@/features/projects/useProjects';
import type { IdeaCandidate, PatternsResponse, Timeline } from '@/shared/api/client';
import type { Project, ProjectSource } from '@/shared/lib/storage/projects';
import { useIdeas, type ChatTurn } from '@/features/chat/useIdeas';
import { useBackendHealth } from './useBackendHealth';
import { useExampleChanges } from './useExampleChanges';
import { imageFor } from './imageFor';
import { Stage } from './Stage';
import { usePinMemory } from './usePinMemory';
import { useChooseToHear } from './useChooseToHear';
import { HearOnOpenToggle } from './HearOnOpenToggle';
import { useCandidateKeys } from './useCandidateKeys';
import { CONVERSATION_ID, ChatClose, ChatToggle } from './ChatDrawer';
import { useChatDrawer } from './useChatDrawer';

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
  /** Where the loaded timeline came from; what keeping it stores alongside it. */
  const [origin, setOrigin] = useState<ProjectSource | null>(null);
  const [projectName, setProjectName] = useState<string | null>(null);
  const pwa = usePwa();
  // Normal is silent; only an unreachable backend is worth a word (BL-065).
  const backend = useBackendHealth(pwa.online);
  const patterns = usePatterns(backend !== 'down');
  const player = usePlayer(createEngine);
  const projects = useProjects(projectStore);
  const drawer = useChatDrawer();

  const show = (tl: Timeline) => {
    player.stop();
    drawer.fold();
    setTimeline(tl);
    setSelectedHarmony(null);
  };
  const { keyChange, regroove, changeKey, changeStyle, forget } = useExampleChanges(
    timeline,
    show,
    patterns.styles,
  );

  const recallPin = usePinMemory(openExample, selectedHarmony);
  const openCandidate = (c: IdeaCandidate, image: string) => {
    show(c.timeline);
    setSelectedHarmony(recallPin(c));
    forget();
    setOpenExample(c);
    setProjectName(null);
    setOrigin({ kind: 'idea', image, candidateTitle: c.title, candidate: c });
  };

  // An answer opens nothing by itself: the user chooses (BL-219). What the
  // user chose before stays open when a new answer arrives.
  const ideas = useIdeas();

  const openProject = (p: Project) => {
    show(p.timeline);
    forget();
    setOpenExample(null);
    setProjectName(p.name);
    setOrigin(p.source);
  };

  const { openChosen, hear } = useChooseToHear({
    openExample,
    timeline,
    player,
    open: (c) => openCandidate(c, imageFor(ideas.turns, c)),
  });
  useCandidateKeys(true, latestCandidates(ideas.turns), openChosen);

  const conversation = (
    <>
      <ProjectList projects={projects} onOpen={openProject} />

      <ChatPanel
        ideas={ideas}
        onOpenCandidate={openChosen}
        openExample={openExample}
        patterns={patterns}
        list
        options={<HearOnOpenToggle hear={hear} />}
      />
    </>
  );

  // A kept example opened from the list goes by its current name (BL-297).
  const exampleTitle =
    openExample?.title ?? (timeline && keptAs(projects.projects, timeline)?.name) ?? projectName;
  const result = timeline && (
    <section className="app__timeline" aria-label="結果">
      {exampleTitle && (
        <div className="app__example-head">
          <h2 className="app__example-title">{exampleTitle}</h2>
          {origin && <KeepToggle projects={projects} current={{ timeline, source: origin }} />}
        </div>
      )}
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
        <ChatToggle drawer={drawer} />
      </div>
      <Stage
        timeline={timeline}
        player={player}
        selectedHarmony={selectedHarmony}
        onSelectHarmony={setSelectedHarmony}
        example={exampleNotes(openExample, origin, timeline, patterns)}
        copyText={copyTextOf(openExample, origin, timeline, patterns)}
      />
    </section>
  );

  return (
    <main className="app app--desk" data-chat={drawer.open ? 'open' : 'closed'}>
      {/* The conversation in a left column beside the stage on wide screens;
          stacked below 1024 px (BL-184, #279). The title heads the column,
          so the stage starts at the top of the page (#436). */}
      <div className="app__columns">
        <div className="app__side" id={CONVERSATION_ID}>
          <ChatClose drawer={drawer} />
          <AppHeader pwa={pwa} backendDown={backend === 'down'} />
          {conversation}
        </div>
        <div className="app__main">{result}</div>
      </div>
    </main>
  );
}

/**
 * The title with the install offer beside it, the app's notices and, when
 * the backend is unreachable, a word.
 */
function AppHeader({ pwa, backendDown }: { pwa: ReturnType<typeof usePwa>; backendDown: boolean }) {
  return (
    <>
      <header className="app__header">
        <div className="app__title">
          <h1>middy</h1>
          <InstallButton pwa={pwa} />
        </div>
        <p>曲のイメージから例を鳴らして、その成り立ちを見る。</p>
      </header>
      <PwaBanner pwa={pwa} />
      {/* Normal is silent; only an unreachable backend is worth a word (BL-065). */}
      {backendDown && pwa.online && (
        <p role="status" className="app__notice">
          バックエンドに接続できません。少し待ってから再読み込みしてください。
        </p>
      )}
    </>
  );
}

/** The examples of the latest answer, the ones 1, 2 and 3 open on the desk. */
function latestCandidates(turns: ChatTurn[]) {
  const last = turns.at(-1);
  return last?.role === 'assistant' ? last.candidates : [];
}

/**
 * The card for the head of the explanation (BL-298): the open candidate's,
 * or the one kept with an example opened from the list (BL-299). None for
 * an example kept before cards were stored.
 */
function cardOf(open: IdeaCandidate | null, origin: ProjectSource | null) {
  if (origin?.kind !== 'idea') return undefined;
  const candidate = open ?? origin.candidate;
  return candidate && { candidate, image: origin.image };
}

/** Whether the example sounds otherwise than its card (a new key or style). */
function changedFrom(candidate: IdeaCandidate, timeline: Timeline): boolean {
  return JSON.stringify(timeline) !== JSON.stringify(candidate.timeline);
}

function exampleNotes(
  open: IdeaCandidate | null,
  origin: ProjectSource | null,
  timeline: Timeline,
  patterns: PatternsResponse,
) {
  const card = cardOf(open, origin);
  if (!card) return undefined;
  return (
    <ExampleNotes
      candidate={card.candidate}
      image={card.image}
      patterns={patterns}
      changed={changedFrom(card.candidate, timeline)}
    />
  );
}

/**
 * The explanation as text for its copy button (BL-300, #442): the card,
 * then the engine's notes on the whole example, a blank line between.
 */
function copyTextOf(
  open: IdeaCandidate | null,
  origin: ProjectSource | null,
  timeline: Timeline,
  patterns: PatternsResponse,
): string {
  const card = cardOf(open, origin);
  const head = card
    ? [
        changedFrom(card.candidate, timeline)
          ? 'この例（AI の提案。キーやスタイルを変える前の、元の例）'
          : 'この例（AI の提案）',
        ...cardLines(card.candidate, card.image, patterns),
      ].join('\n')
    : '';
  return [head, explanationLines(timeline).join('\n')].filter(Boolean).join('\n\n');
}
