import { useState, type FormEvent } from 'react';
import type { FromProgressionRequest, Mode } from '@/shared/api/client';
import { pitchClassName } from '@/shared/lib/notation';
import { parseProgressionText } from './parseProgressionText';
import './ProgressionForm.css';

export type ProgressionFormProps = {
  onSubmit: (request: FromProgressionRequest) => void;
  busy?: boolean;
};

const DEFAULTS = { tonic: 0, mode: 'major' as Mode, text: 'I - V - vi - IV', beats: 4, tempo: 90 };

/** Key, mode, roman-numeral progression and tempo. Emits an API request. */
export function ProgressionForm({ onSubmit, busy = false }: ProgressionFormProps) {
  const [tonic, setTonic] = useState(DEFAULTS.tonic);
  const [mode, setMode] = useState<Mode>(DEFAULTS.mode);
  const [text, setText] = useState(DEFAULTS.text);
  const [beats, setBeats] = useState(DEFAULTS.beats);
  const [tempo, setTempo] = useState(DEFAULTS.tempo);
  const [error, setError] = useState<string | null>(null);

  const submit = (e: FormEvent) => {
    e.preventDefault();
    const progression = parseProgressionText(text, beats);
    if (progression.length === 0) {
      setError('進行を入力してください（例: I - V - vi - IV）');
      return;
    }
    setError(null);
    onSubmit({ key: { tonic, mode }, tempoBpm: tempo, progression });
  };

  return (
    <form className="progression-form" onSubmit={submit} aria-label="進行の入力">
      <label>
        調
        <select value={tonic} onChange={(e) => setTonic(Number(e.target.value))}>
          {Array.from({ length: 12 }, (_, pc) => (
            <option key={pc} value={pc}>
              {pitchClassName(pc)}
            </option>
          ))}
        </select>
      </label>
      <label>
        モード
        <select value={mode} onChange={(e) => setMode(e.target.value as Mode)}>
          <option value="major">major</option>
          <option value="minor">minor</option>
        </select>
      </label>
      <label className="progression-form__text">
        進行（ローマ数字）
        <input
          type="text"
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="I - V - vi - IV"
          spellCheck={false}
        />
      </label>
      <label>
        1 コードの拍数
        <input
          type="number"
          min={1}
          max={16}
          value={beats}
          onChange={(e) => setBeats(Number(e.target.value))}
        />
      </label>
      <label>
        テンポ (BPM)
        <input
          type="number"
          min={20}
          max={300}
          value={tempo}
          onChange={(e) => setTempo(Number(e.target.value))}
        />
      </label>
      <button type="submit" disabled={busy}>
        {busy ? '生成中…' : '生成'}
      </button>
      {error && (
        <p role="alert" className="progression-form__error">
          {error}
        </p>
      )}
    </form>
  );
}
