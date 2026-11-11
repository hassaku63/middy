import type { ChatMessage, IdeaCandidate } from '@/shared/api/client';
import { keyName } from '@/shared/lib/notation';
import type { ChatTurn } from './useIdeas';

/** The most characters a message may carry (the API's `content.maxLength`). */
const MAX_CONTENT = 1000;

/**
 * A turn as the next request carries it. An answer goes back as its reply
 * and the gist of its examples (key, meter, tempo, chords with their
 * beats, patterns, colour, length), so the model sees what it proposed and
 * can keep or change it (BL-310, #452). backend/cmd/converse writes the
 * same text; keep the two in step.
 */
export function historyMessage(turn: ChatTurn): ChatMessage {
  if (turn.role === 'user' || turn.candidates.length === 0) {
    return { role: turn.role, content: turn.content };
  }
  const lines = turn.candidates.map(candidateGist);
  const content = `${turn.content}\n前回の候補:\n${lines.join('\n')}`;
  return { role: 'assistant', content: content.slice(0, MAX_CONTENT) };
}

/** One example in a line: "- 窓辺の雨: A minor, 4/4, 72 BPM, 12 bars, i:4 VI:4 …, style pop …". */
export function candidateGist(c: IdeaCandidate): string {
  const beats = c.beats?.length ? c.beats : c.progression.map(() => c.beatsPerChord);
  const chords = c.progression.map((l, i) => `${l}:${String(beats[i] ?? c.beatsPerChord)}`);
  const parts = [
    keyName(c.key),
    c.meter ?? '4/4',
    `${String(c.tempoBpm)} BPM`,
    ...(c.bars ? [`${String(c.bars)} bars`] : []),
    chords.join(' '),
    ...(c.style ? [`style ${c.style}`] : []),
    ...(c.accompaniment ? [`accompaniment ${c.accompaniment}`] : []),
    ...(c.drums ? [`drums ${c.drums}`] : []),
    ...(c.feel ? [`feel ${c.feel}`] : []),
    ...(c.tensions ? [`tensions ${c.tensions}`] : []),
  ];
  return `- ${c.title}: ${parts.join(', ')}`;
}
