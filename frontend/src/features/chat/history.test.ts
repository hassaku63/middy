import { describe, expect, it } from 'vitest';
import type { IdeaCandidate } from '@/shared/api/client';
import { candidateGist, historyMessage } from './history';

const candidate: IdeaCandidate = {
  title: '窓辺の雨',
  key: { tonic: 9, mode: 'minor' },
  tempoBpm: 72,
  progression: ['i', 'VI', 'III', 'VII'],
  beatsPerChord: 4,
  beats: [4, 4, 2, 2],
  bars: 3,
  style: 'pop',
  accompaniment: 'arpeggio',
  drums: 'ballad',
  tensions: '9',
  reason: 'r',
  timeline: {
    ppq: 480,
    tempoBpm: 72,
    timeSignature: { numerator: 4, denominator: 4 },
    key: { tonic: 9, mode: 'minor' },
    tracks: [],
    harmony: [],
  },
};

describe('the history a turn sends (BL-310, #452)', () => {
  it('writes an example as one line the model can read back', () => {
    expect(candidateGist(candidate)).toBe(
      '- 窓辺の雨: A minor, 4/4, 72 BPM, 3 bars, i:4 VI:4 III:2 VII:2, style pop, accompaniment arpeggio, drums ballad, tensions 9',
    );
  });

  it('sends an answer as its reply and the gist of its examples', () => {
    const msg = historyMessage({
      role: 'assistant',
      content: '静かで少し切ない情景。',
      mood: [],
      candidates: [candidate],
    });
    expect(msg.content).toBe(`静かで少し切ない情景。\n前回の候補:\n${candidateGist(candidate)}`);
  });

  it('sends the user words as they are, and keeps within the API limit', () => {
    expect(historyMessage({ role: 'user', content: '雨' })).toEqual({
      role: 'user',
      content: '雨',
    });
    const long = historyMessage({
      role: 'assistant',
      content: 'あ'.repeat(990),
      mood: [],
      candidates: [candidate],
    });
    expect(long.content.length).toBeLessThanOrEqual(1000);
  });
});
