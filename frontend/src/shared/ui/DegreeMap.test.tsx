import { render, screen } from '@testing-library/react';
import { describe, expect, it } from 'vitest';
import type { DegreeChord, HarmonyEvent } from '@/shared/api/client';
import { DegreeMap } from './DegreeMap';
import { currentDegree } from './currentDegree';

const dorian: DegreeChord[] = [
  { degree: 1, romanNumeral: 'i', symbol: 'Dm', function: 'T', used: true },
  { degree: 2, romanNumeral: 'ii', symbol: 'Em', function: 'S' },
  { degree: 3, romanNumeral: 'III', symbol: 'F', function: 'T' },
  { degree: 4, romanNumeral: 'IV', symbol: 'G', function: 'S', used: true },
  { degree: 5, romanNumeral: 'v', symbol: 'Am', function: 'D' },
  { degree: 6, romanNumeral: 'vi°', symbol: 'Bdim', function: 'T' },
  { degree: 7, romanNumeral: 'VII', symbol: 'C', function: 'D', used: true },
];

const chord = (analysis: NonNullable<HarmonyEvent['analysis']>): HarmonyEvent => ({
  start: 0,
  duration: 1920,
  chord: { root: 7, quality: 'maj' },
  romanNumeral: 'IV',
  analysis,
});

describe('DegreeMap', () => {
  it('lists the seven chords in order, marks the used ones and lights the current one', () => {
    render(
      <DegreeMap degrees={dorian} event={chord({ degree: 4, function: 'S', explanation: '' })} />,
    );
    const items = screen.getAllByRole('listitem');
    expect(items.map((li) => li.textContent)).toEqual([
      'iDm',
      'iiEm',
      'IIIF',
      'IVG',
      'vAm',
      'vi°Bdim',
      'VIIC',
    ]);
    expect(items.filter((li) => li.classList.contains('is-used'))).toHaveLength(3);
    expect(screen.getByLabelText('IV（G）、サブドミナント、この例で使う')).toHaveAttribute(
      'aria-current',
      'true',
    );
    expect(items.filter((li) => li.getAttribute('aria-current'))).toHaveLength(1);
  });

  it('lights nothing when the explained chord is not one of the key (borrowed, applied)', () => {
    expect(
      currentDegree(chord({ degree: 4, function: 'S', borrowed: true, explanation: '' })),
    ).toBeNull();
    expect(
      currentDegree(chord({ degree: 3, function: 'D', appliedTo: 6, explanation: '' })),
    ).toBeNull();
    expect(currentDegree(undefined)).toBeNull();
    render(<DegreeMap degrees={dorian} event={undefined} />);
    expect(screen.getAllByRole('listitem').some((li) => li.getAttribute('aria-current'))).toBe(
      false,
    );
  });
});
