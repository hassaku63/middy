import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { explainedIndex } from './explainedIndex';
import { ExplanationPanel } from './ExplanationPanel';

const timeline: Timeline = {
  ppq: 480,
  tempoBpm: 120,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [],
  harmony: [
    {
      start: 0,
      duration: 1920,
      chord: { root: 7, quality: 'dom7' },
      romanNumeral: 'V7',
      analysis: { degree: 5, function: 'D', explanation: 'V7（G7）はドミナント（D）。' },
    },
    {
      start: 1920,
      duration: 1920,
      chord: { root: 0, quality: 'maj' },
      romanNumeral: 'I',
      analysis: {
        degree: 1,
        function: 'T',
        cadence: 'authentic',
        explanation: 'I（C）はトニック（T）。 直前の V から I へ解決する正格終止。',
      },
    },
  ],
};

describe('explainedIndex', () => {
  it('prefers the pinned chord, else the one under the playhead', () => {
    expect(explainedIndex(timeline.harmony, 0, null)).toBe(0);
    expect(explainedIndex(timeline.harmony, 2000, null)).toBe(1);
    expect(explainedIndex(timeline.harmony, 2000, 0)).toBe(0);
    expect(explainedIndex(timeline.harmony, 9999, null)).toBe(-1);
    expect(explainedIndex(timeline.harmony, 0, 7)).toBe(0); // out-of-range pin is ignored
  });
});

describe('ExplanationPanel', () => {
  it('follows the playhead and shows function, cadence and text', () => {
    const { rerender } = render(
      <ExplanationPanel timeline={timeline} positionTick={0} selectedHarmony={null} />,
    );
    expect(screen.getByRole('heading')).toHaveTextContent('V7G7');
    expect(screen.getByText('D · ドミナント')).toBeInTheDocument();
    expect(screen.queryByText('正格終止')).not.toBeInTheDocument();

    rerender(<ExplanationPanel timeline={timeline} positionTick={2500} selectedHarmony={null} />);
    expect(screen.getByRole('heading')).toHaveTextContent('IC');
    expect(screen.getByText('T · トニック')).toBeInTheDocument();
    expect(screen.getByText('正格終止')).toBeInTheDocument();
    expect(screen.getByText(/直前の V から I へ解決する正格終止/)).toBeInTheDocument();
  });

  it('stays on the pinned chord and offers to unpin', async () => {
    const onClear = vi.fn();
    const user = userEvent.setup();
    render(
      <ExplanationPanel
        timeline={timeline}
        positionTick={2500}
        selectedHarmony={0}
        onClearSelection={onClear}
      />,
    );
    expect(screen.getByRole('heading')).toHaveTextContent('V7');
    await user.click(screen.getByRole('button', { name: /固定を解除/ }));
    expect(onClear).toHaveBeenCalledTimes(1);
  });

  it('explains the empty states', () => {
    const { rerender } = render(
      <ExplanationPanel timeline={null} positionTick={0} selectedHarmony={null} />,
    );
    expect(screen.getByText('タイムラインがありません。')).toBeInTheDocument();

    rerender(<ExplanationPanel timeline={timeline} positionTick={99999} selectedHarmony={null} />);
    expect(screen.getByText(/コードをクリックすると解説/)).toBeInTheDocument();
  });

  it('copes with a timeline that was not analysed', () => {
    const plain: Timeline = {
      ...timeline,
      harmony: timeline.harmony.map((h) => ({
        start: h.start,
        duration: h.duration,
        chord: h.chord,
        romanNumeral: h.romanNumeral,
      })),
    };
    render(<ExplanationPanel timeline={plain} positionTick={0} selectedHarmony={null} />);
    expect(screen.getByText('このコードの分析はまだありません。')).toBeInTheDocument();
  });
});
