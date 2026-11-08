import { render, screen, within } from '@testing-library/react';
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

describe('OverviewNotes inside the panel', () => {
  it('shows the whole and the flow above the chord, and keeps them while playback moves', () => {
    const withOverview: Timeline = {
      ...timeline,
      overview: {
        whole: 'C major（長調）。明るく安定して聞こえる。',
        flow: '2 個の和音で 2 小節。最後は V7 から I への正格終止で閉じる。',
        degrees: [],
      },
    };
    const { rerender } = render(
      <ExplanationPanel timeline={withOverview} positionTick={0} selectedHarmony={null} />,
    );
    const overview = screen.getByLabelText('全体と流れ');
    expect(overview).toHaveTextContent('全体C major（長調）');
    expect(overview).toHaveTextContent('流れ2 個の和音で 2 小節');
    rerender(
      <ExplanationPanel timeline={withOverview} positionTick={2000} selectedHarmony={null} />,
    );
    expect(screen.getByLabelText('全体と流れ')).toHaveTextContent('最後は V7 から I への正格終止');
  });

  it('is absent for a timeline without an overview', () => {
    render(<ExplanationPanel timeline={timeline} positionTick={0} selectedHarmony={null} />);
    expect(screen.queryByLabelText('全体と流れ')).toBeNull();
  });
});

describe('RhythmNotes inside the panel', () => {
  it('describes applied patterns and flags syncopation', () => {
    const withRhythm: Timeline = {
      ...timeline,
      rhythm: {
        accompaniment: {
          id: 'strum',
          name: 'ストローク',
          description: 'ギターのように刻む。',
          hits: 5,
          onBeat: 3,
          offBeat: 2,
          finer: 0,
          syncopated: true,
          explanation: '1 小節に伴奏が 5 回。',
        },
        drums: {
          id: 'eight-beat',
          name: '8 ビート',
          description: 'ポップスの土台。',
          hits: 4,
          onBeat: 4,
          offBeat: 0,
          finer: 0,
          syncopated: false,
          explanation: '1 小節にキックとスネアが 4 回。',
        },
        feel: { id: 'swing', name: 'スウィング', explanation: '裏拍を少し遅らせている。' },
      },
    };
    render(<ExplanationPanel timeline={withRhythm} positionTick={0} selectedHarmony={null} />);
    const rhythm = screen.getByRole('heading', { name: 'リズム' }).parentElement!;
    expect(within(rhythm).getByRole('region', { name: '伴奏: ストローク' })).toHaveTextContent(
      'シンコペーション',
    );
    expect(within(rhythm).getByRole('region', { name: 'ドラム: 8 ビート' })).not.toHaveTextContent(
      'シンコペーション',
    );
    expect(within(rhythm).getByText('1 小節にキックとスネアが 4 回。')).toBeInTheDocument();
    expect(within(rhythm).getByRole('region', { name: 'ノリ: スウィング' })).toHaveTextContent(
      '裏拍を少し遅らせている。',
    );
  });

  it('shows nothing about rhythm for a block-chord timeline', () => {
    render(<ExplanationPanel timeline={timeline} positionTick={0} selectedHarmony={null} />);
    expect(screen.queryByRole('heading', { name: 'リズム' })).not.toBeInTheDocument();
  });
});

describe('RhythmNotes style (BL-110)', () => {
  it('says which instrument the style gave the chords', () => {
    const styled: Timeline = {
      ...timeline,
      rhythm: { style: { id: 'celtic', name: 'ケルト', instrument: 'ハープ' } },
    };
    render(<ExplanationPanel timeline={styled} positionTick={0} selectedHarmony={null} />);
    expect(screen.getByText('ケルトのスタイルなので、和音はハープで鳴らす。')).toBeInTheDocument();
  });
});
