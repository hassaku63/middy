import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
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

  it('reads empty beats as one length for every chord (review)', () => {
    render(<CandidateCard candidate={{ ...base, beats: [] }} open={false} onOpen={vi.fn()} />);
    expect(screen.getByText('4 拍/コード')).toBeInTheDocument();
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
  it('puts the reason last: title (the switch), progression, reason', () => {
    render(
      <CandidateCard candidate={{ ...base, reason: '理由の文' }} open={true} onOpen={vi.fn()} />,
    );
    const card = within(screen.getByRole('article'));
    const order = [
      card.getByRole('button', { name: base.title }),
      card.getByText(base.progression.join(' - ')),
      card.getByText('理由の文'),
    ];
    // Each part comes after the one before it in the document.
    for (const [i, el] of order.slice(1).entries()) {
      expect(order[i]!.compareDocumentPosition(el) & Node.DOCUMENT_POSITION_FOLLOWING).toBeTruthy();
    }
  });
});

describe('CandidateCard as a whole (#313)', () => {
  it('opens from its title, which is the one button of the card and says whether it is open', async () => {
    const onOpen = vi.fn();
    const { rerender } = render(<CandidateCard candidate={base} open={false} onOpen={onOpen} />);
    const card = within(screen.getByRole('article'));
    expect(card.getAllByRole('button')).toHaveLength(1);
    const button = card.getByRole('button', { name: base.title });
    expect(button).toHaveAttribute('aria-pressed', 'false');
    await userEvent.click(button);
    expect(onOpen).toHaveBeenCalledTimes(1);
    rerender(<CandidateCard candidate={base} open={true} onOpen={onOpen} />);
    expect(card.getByRole('button', { name: base.title })).toHaveAttribute('aria-pressed', 'true');
    // The frame and its fill mark it (BL-216); no word is added to the card.
    expect(card.queryByText('開いています')).not.toBeInTheDocument();
  });
});

describe('CandidateCard progression (#451)', () => {
  const harmony = [
    {
      start: 0,
      duration: 1920,
      chord: { root: 10, quality: 'maj7' as const },
      romanNumeral: 'I',
      symbol: 'Bbmaj7',
    },
    {
      start: 1920,
      duration: 1920,
      chord: { root: 10, quality: 'maj7' as const },
      bass: 9,
      romanNumeral: 'I/7',
      symbol: 'Bbmaj7/A',
    },
    {
      start: 3840,
      duration: 1920,
      chord: { root: 8, quality: 'dom7' as const },
      romanNumeral: '♭VII7',
      symbol: 'Ab7',
    },
  ];
  const withHarmony: IdeaCandidate = {
    ...base,
    key: { tonic: 10, mode: 'major' },
    progression: ['I/1', 'I/7', 'bVII7'],
    timeline: { ...base.timeline, harmony },
  };

  it('shows the numerals as the chord lane spells them, and the chords by name', () => {
    render(<CandidateCard candidate={withHarmony} open={false} onOpen={vi.fn()} />);
    expect(screen.getByText('I - I/7 - ♭VII7')).toBeInTheDocument();
    expect(screen.getByText('Bbmaj7 - Bbmaj7/A - Ab7')).toBeInTheDocument();
    // Three bars of 4/4 at 480 ticks a quarter (BL-305).
    expect(screen.getByText('3 小節')).toBeInTheDocument();
    expect(screen.queryByText(/bVII7/)).not.toBeInTheDocument();
  });

  it('falls back to the labels as written while there is no harmony', () => {
    render(<CandidateCard candidate={base} open={false} onOpen={vi.fn()} />);
    expect(screen.getByText('i - IV - i - VII')).toBeInTheDocument();
  });
});
