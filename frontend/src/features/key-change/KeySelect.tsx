import type { PitchClass } from '@/shared/api/client';
import { pitchClassName } from '@/shared/lib/notation';
import './KeySelect.css';

type Props = {
  tonic: PitchClass;
  mode: string;
  busy: boolean;
  error: string | null;
  onChange: (tonic: PitchClass) => void;
};

const TONICS = Array.from({ length: 12 }, (_, i) => i);

/** Picks the open example's key; the numerals stay, the sound moves (BL-078). */
export function KeySelect({ tonic, mode, busy, error, onChange }: Props) {
  return (
    <label className="key-select">
      キー
      <select value={tonic} disabled={busy} onChange={(e) => onChange(Number(e.target.value))}>
        {TONICS.map((t) => (
          <option key={t} value={t}>
            {pitchClassName(t)} {mode}
          </option>
        ))}
      </select>
      {error && (
        <span role="alert" className="key-select__error">
          {error}
        </span>
      )}
    </label>
  );
}
