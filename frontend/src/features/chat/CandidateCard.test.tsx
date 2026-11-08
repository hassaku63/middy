import { render, screen } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import type { IdeaCandidate } from '@/shared/api/client';
import { CandidateCard } from './CandidateCard';

const base: IdeaCandidate = {
  title: '緑の丘のリール',
  key: { tonic: 2, mode: 'dorian' },
  tempoBpm: 112,
  progression: ['i', 'IV', 'i', 'VII'],
  beatsPerChord: 4,
  reason: 'r',
  timeline: {
    ppq: 480,
    tempoBpm: 112,
    timeSignature: { numerator: 4, denominator: 4 },
    key: { tonic: 2, mode: 'dorian' },
    tracks: [],
    harmony: [],
  },
};

describe('CandidateCard', () => {
  it('shows one length for every chord when there are no per-chord beats', () => {
    render(<CandidateCard candidate={base} open={false} onOpen={vi.fn()} />);
    expect(screen.getByText('4 拍/コード')).toBeInTheDocument();
  });

  it('shows each chord its length when the candidate has beats (BL-109)', () => {
    render(
      <CandidateCard candidate={{ ...base, beats: [8, 4, 2, 2] }} open={false} onOpen={vi.fn()} />,
    );
    expect(screen.getByText('長さ: 8・4・2・2 拍')).toBeInTheDocument();
    expect(screen.queryByText('4 拍/コード')).not.toBeInTheDocument();
  });
});

describe('CandidateCard style (BL-119)', () => {
  it('names the style from the catalogue, falling back to its id', () => {
    const patterns = {
      accompaniments: [],
      drums: [],
      feels: [],
      styles: [
        {
          id: 'celtic',
          name: 'ケルト',
          description: '',
          accompaniments: ['reel'],
          drums: ['bodhran'],
          feels: ['straight' as const],
          grooves: [],
        },
      ],
    };
    const { rerender } = render(
      <CandidateCard
        candidate={{ ...base, style: 'celtic' }}
        open={false}
        onOpen={vi.fn()}
        patterns={patterns}
      />,
    );
    expect(screen.getByText('スタイル: ケルト')).toBeInTheDocument();
    rerender(
      <CandidateCard candidate={{ ...base, style: 'jazz' }} open={false} onOpen={vi.fn()} />,
    );
    expect(screen.getByText('スタイル: jazz')).toBeInTheDocument();
  });
});

describe('CandidateCard order (BL-073)', () => {
  it('puts the switch before the reason: title, progression, switch, reason', () => {
    const { container } = render(
      <CandidateCard candidate={{ ...base, reason: '理由の文' }} open={false} onOpen={vi.fn()} />,
    );
    const article = container.querySelector('article');
    const order = Array.from(article?.children ?? []).map((el) =>
      el.tagName === 'BUTTON' ? 'switch' : el.className,
    );
    expect(order.indexOf('candidate__title')).toBeLessThan(order.indexOf('candidate__progression'));
    expect(order.indexOf('candidate__progression')).toBeLessThan(order.indexOf('switch'));
    expect(order.indexOf('switch')).toBeLessThan(order.indexOf('candidate__reason'));
    expect(screen.getByText('理由の文')).toBeInTheDocument();
  });
});
