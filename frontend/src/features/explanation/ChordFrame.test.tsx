import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { DegreeChord, Timeline } from '@/shared/api/client';
import { ChordFrame } from './ChordFrame';

const timeline = {
  key: { tonic: 0, mode: 'major' },
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

/** C major's seven triads, as the engine lays them out (BL-099). */
const cMajor: DegreeChord[] = [
  { degree: 1, romanNumeral: 'I', symbol: 'C', function: 'T', used: true },
  { degree: 2, romanNumeral: 'ii', symbol: 'Dm', function: 'S' },
  { degree: 3, romanNumeral: 'iii', symbol: 'Em', function: 'T' },
  { degree: 4, romanNumeral: 'IV', symbol: 'F', function: 'S' },
  { degree: 5, romanNumeral: 'V', symbol: 'G', function: 'D', used: true },
  { degree: 6, romanNumeral: 'vi', symbol: 'Am', function: 'T' },
  { degree: 7, romanNumeral: 'vii°', symbol: 'Bdim', function: 'D' },
];

const withMap = (mode: Timeline['key']['mode']): Timeline => ({
  ...timeline,
  key: { tonic: 0, mode },
  overview: { whole: '', flow: '', degrees: cMajor },
});

describe('ChordFrame and the degree map (#360, ADR-0044)', () => {
  it('lays the map out as the flow of functions in major and minor, lighting the chord', () => {
    render(
      <ChordFrame
        timeline={withMap('major')}
        positionTick={0}
        selectedHarmony={null}
        onClearSelection={vi.fn()}
      />,
    );
    const frame = screen.getByRole('region', { name: 'コードの解説' });
    const map = within(frame).getByRole('group', { name: '度数の地図（機能の流れ）' });
    const column = (name: string) =>
      within(within(map).getByRole('list', { name }))
        .getAllByRole('listitem')
        .map((li) => li.textContent);
    expect(column('トニック')).toEqual(['IC', 'iiiEm', 'viAm']);
    expect(column('サブドミナント')).toEqual(['iiDm', 'IVF']);
    expect(column('ドミナント')).toEqual(['VG', 'vii°Bdim']);
    expect(within(map).getByLabelText('V（G）、ドミナント、この例で使う')).toHaveAttribute(
      'aria-current',
      'true',
    );
  });

  it('keeps degree order in a mode and labels nothing by function (ADR-0045)', () => {
    render(
      <ChordFrame
        timeline={withMap('dorian')}
        positionTick={0}
        selectedHarmony={null}
        onClearSelection={vi.fn()}
      />,
    );
    const frame = screen.getByRole('region', { name: 'コードの解説' });
    const map = within(frame).getByRole('list', { name: '度数の地図' });
    expect(within(frame).queryByRole('group', { name: /機能の流れ/ })).toBeNull();
    // No function badge, no function in the live line, no function colour;
    // the tonic is named and marked instead.
    // (The text is the engine's and is shown as it is; ADR-0046 keeps functions out of it.)
    expect(within(frame).queryByText('D · ドミナント')).toBeNull();
    expect(within(frame).getByText('V7 G7')).toHaveAttribute('aria-live', 'polite');
    const cells = within(map).getAllByRole('listitem');
    expect(cells.some((li) => /\bis-[TSD]\b/.test(li.className))).toBe(false);
    expect(within(map).getByLabelText('I（C）、主和音、この例で使う')).toHaveClass('is-tonic');
  });
});

/** D dorian's triads with the engine's roles (ADR-0047). */
const dDorian: DegreeChord[] = [
  { degree: 1, romanNumeral: 'i', symbol: 'Dm', function: 'T', role: 'tonic', used: true },
  { degree: 2, romanNumeral: 'ii', symbol: 'Em', function: 'S', role: 'characteristic' },
  { degree: 3, romanNumeral: 'III', symbol: 'F', function: 'T', role: 'other' },
  {
    degree: 4,
    romanNumeral: 'IV',
    symbol: 'G',
    function: 'S',
    role: 'characteristic',
    used: true,
  },
  { degree: 5, romanNumeral: 'v', symbol: 'Am', function: 'D', role: 'other' },
  { degree: 6, romanNumeral: 'vi°', symbol: 'Bdim', function: 'T', role: 'tritone' },
  { degree: 7, romanNumeral: 'VII', symbol: 'C', function: 'D', role: 'other' },
];

const dorianTimeline = {
  key: { tonic: 2, mode: 'dorian' },
  harmony: [
    {
      start: 0,
      duration: 1920,
      chord: { root: 7, quality: 'maj' },
      romanNumeral: 'IV',
      analysis: {
        degree: 4,
        function: 'S',
        role: 'characteristic',
        explanation: 'ドリアンの 4 度の和音。',
      },
    },
    {
      start: 1920,
      duration: 1920,
      chord: { root: 5, quality: 'maj' },
      romanNumeral: 'III',
      analysis: { degree: 3, function: 'T', role: 'other', explanation: 'ドリアンの 3 度の和音。' },
    },
  ],
  overview: { whole: '', flow: '', degrees: dDorian },
} as Timeline;

describe('ChordFrame in a mode (#366, ADR-0048)', () => {
  it('groups the map by role and names the chord by its role, in words', () => {
    render(
      <ChordFrame
        timeline={dorianTimeline}
        positionTick={0}
        selectedHarmony={null}
        onClearSelection={vi.fn()}
      />,
    );
    const frame = screen.getByRole('region', { name: 'コードの解説' });
    const map = within(frame).getByRole('group', { name: '度数の地図（和音の役割）' });
    const group = (name: string) =>
      within(within(map).getByRole('list', { name }))
        .getAllByRole('listitem')
        .map((li) => li.textContent);
    expect(
      within(map)
        .getAllByRole('list')
        .map((l) => l.getAttribute('aria-label')),
    ).toEqual(['主和音', '特徴和音', 'そのほかの和音', '三全音の和音']);
    expect(group('主和音')).toEqual(['iDm']);
    expect(group('特徴和音')).toEqual(['iiEm', 'IVG']);
    expect(group('そのほかの和音')).toEqual(['IIIF', 'vAm', 'VIIC']);
    expect(group('三全音の和音')).toEqual(['vi°Bdim']);
    expect(within(map).getByLabelText('IV（G）、特徴和音、この例で使う')).toHaveAttribute(
      'aria-current',
      'true',
    );
    // The badge and the live line say the role, not a function.
    expect(within(frame).getByText('特徴和音', { selector: '.explanation__role' })).toBeVisible();
    expect(within(frame).getByText('IV G 特徴和音')).toHaveAttribute('aria-live', 'polite');
    expect(within(frame).queryByText(/サブドミナント/)).toBeNull();
  });

  it('shows no role badge for a chord with no named role', () => {
    render(
      <ChordFrame
        timeline={dorianTimeline}
        positionTick={1920}
        selectedHarmony={null}
        onClearSelection={vi.fn()}
      />,
    );
    const frame = screen.getByRole('region', { name: 'コードの解説' });
    expect(within(frame).queryByText(/./, { selector: '.explanation__role' })).toBeNull();
    expect(within(frame).getByText('III F')).toHaveAttribute('aria-live', 'polite');
  });
});
