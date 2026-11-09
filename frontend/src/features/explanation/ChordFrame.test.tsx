import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { ChordFrame } from './ChordFrame';

const timeline = {
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
      analysis: { degree: 1, function: 'T', explanation: 'I（C）はトニック（T）。' },
    },
  ],
} as Timeline;

describe('ChordFrame (BL-185)', () => {
  it('follows the playhead and announces only the chord and its function', () => {
    render(
      <ChordFrame
        timeline={timeline}
        positionTick={0}
        selectedHarmony={null}
        onClearSelection={vi.fn()}
      />,
    );
    const frame = screen.getByRole('region', { name: 'コードの解説' });
    expect(frame).toHaveClass('is-following');
    expect(frame).not.toHaveAttribute('aria-live');
    expect(screen.getByText('V7 G7 ドミナント')).toHaveAttribute('aria-live', 'polite');
    expect(screen.getByText('V7（G7）はドミナント（D）。')).toBeInTheDocument();
  });

  it('shows the pinned chord, and lets it follow again', async () => {
    const onClear = vi.fn();
    render(
      <ChordFrame
        timeline={timeline}
        positionTick={0}
        selectedHarmony={1}
        onClearSelection={onClear}
      />,
    );
    expect(screen.getByRole('region', { name: 'コードの解説' })).toHaveClass('is-pinned');
    expect(screen.getByText('I（C）はトニック（T）。')).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', { name: /固定を解除/ }));
    expect(onClear).toHaveBeenCalled();
  });

  it('says how to move when no chord is under the playhead', () => {
    render(
      <ChordFrame
        timeline={timeline}
        positionTick={9999}
        selectedHarmony={null}
        onClearSelection={vi.fn()}
      />,
    );
    expect(screen.getByText(/← → で前後のコードへ/)).toBeInTheDocument();
  });
});
