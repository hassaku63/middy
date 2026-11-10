import { Fragment, useId } from 'react';
import type { DegreeChord, HarmonicFunction, HarmonyEvent, ModalRole } from '@/shared/api/client';
import { FUNCTION_NAME, ROLE_NAME, roleName } from '@/shared/lib/notation';
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
  /**
   * Colour and name the chords by function. False in a mode, where
   * functions do not hold: the tonic is marked instead (ADR-0045).
   */
  functions?: boolean;
  /**
   * Group the chords by their role in the mode (#366, ADR-0048): the
   * tonic, the characteristic chords, the others, then the chords with the
   * tritone. The caller asks for it in a mode, where the engine gives roles.
   */
  roles?: boolean;
};

const FLOW: HarmonicFunction[] = ['T', 'S', 'D'];

/**
 * The role groups in reading order: the centre, what makes the mode, the
 * rest, and last what blurs it. Each has a visible heading, so the group
 * is told by words, not by colour (WCAG 1.4.1).
 */
const ROLE_GROUPS: { role: ModalRole; heading: string; name: string }[] = [
  { role: 'tonic', heading: '主和音', name: ROLE_NAME.tonic },
  { role: 'characteristic', heading: '特徴和音', name: ROLE_NAME.characteristic },
  { role: 'other', heading: 'そのほか', name: 'そのほかの和音' },
  { role: 'tritone', heading: '三全音', name: ROLE_NAME.tritone },
];

/**
 * The degree map (BL-099): the seven chords of the example's mode in a
 * band, coloured by function. The ones the example plays stand out, and
 * the one being explained lights up, so the learner sees where in the mode
 * the music is. The data is the engine's (Overview.degrees). It takes no
 * place of its own, so any screen can put it where it fits. In major and
 * minor it can lay the chords out as the flow of functions (#360).
 */
export function DegreeMap({
  degrees,
  event,
  flow = false,
  functions = true,
  roles = false,
}: DegreeMapProps) {
  const captionId = useId();
  if (degrees.length === 0) return null;
  const current = currentDegree(event);
  const cell = (d: DegreeChord) => (
    <DegreeCell key={d.degree} d={d} current={current} functions={functions} />
  );
  if (roles && degrees.some((d) => d.role)) {
    return (
      <div
        className="degree-map degree-map--flow"
        role="group"
        aria-label="度数の地図（和音の役割）"
      >
        {ROLE_GROUPS.map(({ role, heading, name }) => {
          const chords = degrees.filter((d) => d.role === role);
          if (chords.length === 0) return null;
          return (
            <div key={role} className={`degree-map__column is-role-${role}`}>
              <span className="degree-map__function" aria-hidden="true">
                {heading}
              </span>
              <ul className="degree-map__chords" aria-label={name}>
                {chords.map(cell)}
              </ul>
            </div>
          );
        })}
      </div>
    );
  }
  if (!flow) {
    return (
      <ol className="degree-map" aria-label="度数の地図">
        {degrees.map(cell)}
      </ol>
    );
  }
  // The flow of functions (#360): each function's chords in a column, in
  // degree order, the columns joined left to right and back to T. The
  // caption says it is the basic flow, not a rule (#371): music also goes
  // D → S (V → IV in blues and rock), and the arrows alone read as a law.
  return (
    <div
      className="degree-map degree-map--flow degree-map--captioned"
      role="group"
      aria-label="度数の地図（機能の流れ）"
      aria-describedby={captionId}
    >
      <div className="degree-map__row">
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
      <p id={captionId} className="degree-map__caption">
        基本の流れ（いつもこの順とは限らない）
      </p>
    </div>
  );
}

function DegreeCell({
  d,
  current,
  functions,
}: {
  d: DegreeChord;
  current: number | null;
  functions: boolean;
}) {
  // In a mode the chord is named by its role (ADR-0048); the tonic is
  // named even where the engine gives no role.
  const role = functions
    ? FUNCTION_NAME[d.function]
    : roleName(d.role) || (d.degree === 1 ? ROLE_NAME.tonic : '');
  return (
    <li
      className={[
        'degree-map__chord',
        functions ? `is-${d.function}` : '',
        !functions && d.degree === 1 ? 'is-tonic' : '',
        d.used ? 'is-used' : '',
        d.degree === current ? 'is-current' : '',
      ]
        .filter(Boolean)
        .join(' ')}
      aria-current={d.degree === current ? 'true' : undefined}
      aria-label={`${d.romanNumeral}（${d.symbol}）${role ? `、${role}` : ''}${usedText(d)}`}
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
