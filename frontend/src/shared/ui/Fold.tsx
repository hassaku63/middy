import type { ReactNode } from 'react';
import './Fold.css';

type Props = {
  /** What is folded, as its heading. */
  title: ReactNode;
  /** The gist, beside the title, so the closed fold still says something. */
  gist?: string | undefined;
  children: ReactNode;
};

/**
 * A part of the explanation that is read once per example, not while it
 * plays (#511): closed at first, one line with its title and gist, opened
 * and closed by a click, a tap or the keyboard. A native details element,
 * so assistive tech says whether it is open; it stays as it was set when
 * another example opens, as the element is kept.
 */
export function Fold({ title, gist, children }: Props) {
  return (
    <details className="fold">
      <summary className="fold__summary">
        <span className="fold__title">{title}</span>
        {gist && <span className="fold__gist">{gist}</span>}
      </summary>
      <div className="fold__body">{children}</div>
    </details>
  );
}
