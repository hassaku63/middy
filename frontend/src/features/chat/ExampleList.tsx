import type { IdeaCandidate } from '@/shared/api/client';
import { keyName } from '@/shared/lib/notation';
import type { Exchange } from './exchanges';

type Props = {
  exchanges: Exchange[];
  openExample: IdeaCandidate | null;
  onOpen: (candidate: IdeaCandidate) => void;
};

/**
 * Earlier exchanges on the desk under trial (BL-186, #279): one line per
 * example (name, key, BPM, progression), so four exchanges fit in the
 * column without scrolling. The exchange's number tells apart examples of
 * the same name.
 */
export function ExampleList({ exchanges, openExample, onOpen }: Props) {
  return (
    <ol className="chat__exchanges" aria-label="これまでの例">
      {exchanges.map((ex) => (
        <li key={ex.number} className="chat__exchange">
          <p className="chat__exchange-image">
            <span className="chat__exchange-number">{ex.number}</span> {ex.image}
          </p>
          <ul className="chat__lines">
            {ex.candidates.map((c, i) => (
              <li key={`${i}-${c.title}`}>
                <button
                  type="button"
                  className={`chat__line${openExample === c ? ' is-open' : ''}`}
                  aria-pressed={openExample === c}
                  onClick={() => onOpen(c)}
                >
                  <span className="visually-hidden">{ex.number} 回目：</span>
                  <span className="chat__line-title">{c.title}</span>{' '}
                  <span className="chat__line-spec">
                    {keyName(c.key)} {c.tempoBpm} BPM {c.progression.join(' - ')}
                  </span>
                </button>
              </li>
            ))}
          </ul>
        </li>
      ))}
    </ol>
  );
}
