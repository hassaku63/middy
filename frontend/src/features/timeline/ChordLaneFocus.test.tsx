import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { useState } from 'react';
import { describe, expect, it } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { TimelineView } from './TimelineView';

/** I | V | vi | IV, a bar each. */
const timeline: Timeline = {
  ppq: 480,
  tempoBpm: 120,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [],
  harmony: [
    { start: 0, duration: 1920, chord: { root: 0, quality: 'maj' }, romanNumeral: 'I' },
    { start: 1920, duration: 1920, chord: { root: 7, quality: 'maj' }, romanNumeral: 'V' },
    { start: 3840, duration: 1920, chord: { root: 9, quality: 'min' }, romanNumeral: 'vi' },
    { start: 5760, duration: 1920, chord: { root: 5, quality: 'maj' }, romanNumeral: 'IV' },
  ],
};

function Lane({ pinned = null, at = 0 }: { pinned?: number | null; at?: number }) {
  const [selected, setSelected] = useState<number | null>(pinned);
  return (
    <>
      <button type="button">before</button>
      <TimelineView
        timeline={timeline}
        positionTick={at}
        selectedHarmony={selected}
        onSelectHarmony={setSelected}
      />
      <button type="button">after</button>
    </>
  );
}

const tabStops = () =>
  screen.getAllByRole('option').filter((o) => o.getAttribute('tabindex') === '0');

describe('the chord lane is one tab stop (#500, APG listbox)', () => {
  it('takes the focus once, on the first chord when none is pinned or sounding', async () => {
    const user = userEvent.setup();
    render(<Lane />);
    expect(tabStops()).toHaveLength(1);
    await user.click(screen.getByRole('button', { name: 'before' }));
    await user.tab();
    expect(screen.getByRole('option', { name: 'I C' })).toHaveFocus();
    await user.tab();
    expect(screen.getByRole('button', { name: 'after' })).toHaveFocus();
  });

  it('stops on the pinned chord', () => {
    render(<Lane pinned={2} />);
    expect(tabStops()).toEqual([screen.getByRole('option', { name: 'vi Am' })]);
  });

  it('stops on the sounding chord while playing', () => {
    render(<Lane at={1920 + 10} />);
    expect(tabStops()).toEqual([screen.getByRole('option', { name: 'V G' })]);
  });

  it('Home and End pin the first and the last chord, and the focus goes with the pin', async () => {
    const user = userEvent.setup();
    render(<Lane />);
    await user.click(screen.getByRole('button', { name: 'before' }));
    await user.tab();
    await user.keyboard('{End}');
    const last = screen.getByRole('option', { name: 'IV F' });
    expect(last).toHaveAttribute('aria-selected', 'true');
    expect(last).toHaveFocus();
    expect(tabStops()).toEqual([last]);
    await user.keyboard('{Home}');
    expect(screen.getByRole('option', { name: 'I C' })).toHaveFocus();
  });
});
