import { Fragment } from 'react';
import type { DegreeChord, HarmonicFunction, HarmonyEvent } from '@/shared/api/client';
import { FUNCTION_NAME } from '@/shared/lib/notation';
import { currentDegree } from './currentDegree';
import './DegreeMap.css';

export type DegreeMapProps = {
  degrees: DegreeChord[];
  /** The chord being explained (playing or pinned), if any. */
  event: HarmonyEvent | undefined;
  /**
   * Lay the chords out as the flow of functions, T → S → D → T, instead of
   * in degree order (#360). Functions are a theory of major and minor;
   * the caller asks for the flow only there (ADR-0044).
   */
  flow?: boolean;
};

const FLOW: HarmonicFunction[] = ['T', 'S', 'D'];

/**
 * The degree map (BL-099): the seven chords of the example's mode in a
 * band, coloured by function. The ones the example plays stand out, and
 * the one being explained lights up, so the learner sees where in the mode
 * the music is. The data is the engine's (Overview.degrees). It takes no
 * place of its own, so any screen can put it where it fits. In major and
 * minor it can lay the chords out as the flow of functions (#360).
 */
export function DegreeMap({ degrees, event, flow = false }: DegreeMapProps) {
  if (degrees.length === 0) return null;
  const current = currentDegree(event);
  const cell = (d: DegreeChord) => <DegreeCell key={d.degree} d={d} current={current} />;
  if (!flow) {
    return (
      <ol className="degree-map" aria-label="度数の地図">
        {degrees.map(cell)}
      </ol>
    );
  }
  // The flow of functions (#360): each function's chords in a column, in
  // degree order, the columns joined left to right and back to T.
  return (
    <div className="degree-map degree-map--flow" role="group" aria-label="度数の地図（機能の流れ）">
      {FLOW.map((f, i) => (
        <Fragment key={f}>
          {i > 0 && (
            <span className="degree-map__arrow" aria-hidden="true">
              →
            </span>
          )}
          <div className={`degree-map__column is-${f}`}>
            <span className="degree-map__function" aria-hidden="true">
              {f}
            </span>
            <ul className="degree-map__chords" aria-label={FUNCTION_NAME[f]}>
              {degrees.filter((d) => d.function === f).map(cell)}
            </ul>
          </div>
        </Fragment>
      ))}
      <span className="degree-map__arrow" aria-hidden="true">
        → T
      </span>
    </div>
  );
}

function DegreeCell({ d, current }: { d: DegreeChord; current: number | null }) {
  return (
    <li
      className={[
        'degree-map__chord',
        `is-${d.function}`,
        d.used ? 'is-used' : '',
        d.degree === current ? 'is-current' : '',
      ]
        .filter(Boolean)
        .join(' ')}
      aria-current={d.degree === current ? 'true' : undefined}
      aria-label={`${d.romanNumeral}（${d.symbol}）、${FUNCTION_NAME[d.function]}${usedText(d)}`}
    >
      <span className="degree-map__numeral">{d.romanNumeral}</span>
      <span className="degree-map__symbol">{d.symbol}</span>
      {d.altered?.map((a) => (
        <span key={a} className="degree-map__altered">
          {a}
        </span>
      ))}
    </li>
  );
}

/**
 * What the example does with the degree: plays its chord, or plays it
 * altered (#296: E minor's harmonic V is B7, not the map's Bm).
 */
function usedText(d: DegreeChord): string {
  if (d.altered?.length)
    return `、この例では ${d.altered.join('、')}（音階の外の音を含む形）で使う`;
  return d.used ? '、この例で使う' : '';
}
