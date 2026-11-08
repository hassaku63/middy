import type { PatternsResponse } from '@/shared/api/client';
import './StyleSelect.css';

type Props = {
  styles: PatternsResponse['styles'];
  /** The open example's style; undefined when it was rendered without one. */
  current: string | undefined;
  busy: boolean;
  error: string | null;
  note: string | null;
  onChange: (id: string) => void;
};

/** Picks the open example's style; the harmony stays, the groove changes (BL-087). */
export function StyleSelect({ styles, current, busy, error, note, onChange }: Props) {
  if (styles.length === 0) return null;
  return (
    <div className="style-select">
      <label>
        スタイル
        <select value={current ?? ''} disabled={busy} onChange={(e) => onChange(e.target.value)}>
          {current === undefined && (
            <option value="" disabled>
              —
            </option>
          )}
          {styles.map((s) => (
            <option key={s.id} value={s.id}>
              {s.name}
            </option>
          ))}
        </select>
      </label>
      {error && (
        <span role="alert" className="style-select__error">
          {error}
        </span>
      )}
      {note && (
        <p role="status" className="style-select__note">
          {note}
        </p>
      )}
    </div>
  );
}
