import { render, screen, within } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { IdeaCandidate } from '@/shared/api/client';
import { ExampleNotes } from './ExampleNotes';

const candidate: IdeaCandidate = {
  title: '緑の丘のリール',
  key: { tonic: 2, mode: 'dorian' },
  tempoBpm: 112,
  progression: ['i', 'IV', 'i', 'VII'],
  beatsPerChord: 4,
  style: 'celtic',
  accompaniment: 'reel',
  reason: '主和音の Dm を保ってから IV へ。',
  timeline: {
    ppq: 480,
    tempoBpm: 112,
    timeSignature: { numerator: 4, denominator: 4 },
    key: { tonic: 2, mode: 'dorian' },
    tracks: [],
    harmony: [],
  },
};

const patterns = {
  accompaniments: [{ id: 'reel', name: 'リールの裏打ち' }],
  drums: [],
  feels: [],
  styles: [{ id: 'celtic', name: 'ケルト' }],
} as never;

describe('ExampleNotes (BL-298)', () => {
  it("lays out the card's content as plain text, the AI's proposal", () => {
    render(
      <ExampleNotes
        candidate={candidate}
        image="アイルランドのパブ"
        patterns={patterns}
        changed={false}
      />,
    );
    const notes = screen.getByLabelText('この例（AI の提案）');
    expect(notes).toHaveTextContent('この例（AI の提案）');
    expect(notes).toHaveTextContent('イメージ: アイルランドのパブ');
    expect(notes).toHaveTextContent('D dorian · 112 BPM · 4 拍/コード');
    expect(notes).toHaveTextContent('i - IV - i - VII');
    expect(notes).toHaveTextContent('スタイル: ケルト · 伴奏: リールの裏打ち');
    expect(notes).toHaveTextContent('主和音の Dm を保ってから IV へ。');
    // Text to read and copy, nothing to press.
    expect(within(notes).queryByRole('button')).not.toBeInTheDocument();
    expect(notes).not.toHaveTextContent('元の例');
  });

  it('says it describes the original once the key or style has changed', () => {
    render(<ExampleNotes candidate={candidate} image="" changed />);
    const notes = screen.getByLabelText('この例（AI の提案）');
    expect(notes).toHaveTextContent('キーやスタイルを変える前の、元の例');
    expect(notes).not.toHaveTextContent('イメージ:');
  });
});
