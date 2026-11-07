import type { PatternsResponse } from '@/shared/api/client';

type Props = {
  patterns: PatternsResponse;
  accompaniment: string;
  drums: string;
  onAccompaniment: (id: string) => void;
  onDrums: (id: string) => void;
};

/** Accompaniment / drums selects; hidden until the catalogue is loaded. */
export function PatternSelects({
  patterns,
  accompaniment,
  drums,
  onAccompaniment,
  onDrums,
}: Props) {
  return (
    <>
      {patterns.accompaniments.length > 0 && (
        <label>
          伴奏
          <select
            value={accompaniment}
            onChange={(e) => onAccompaniment(e.target.value)}
            title={patterns.accompaniments.find((a) => a.id === accompaniment)?.description}
          >
            <option value="">ブロック（既定）</option>
            {patterns.accompaniments
              .filter((a) => a.id !== 'block')
              .map((a) => (
                <option key={a.id} value={a.id}>
                  {a.name}
                </option>
              ))}
          </select>
        </label>
      )}
      {patterns.drums.length > 0 && (
        <label>
          ドラム
          <select
            value={drums}
            onChange={(e) => onDrums(e.target.value)}
            title={patterns.drums.find((d) => d.id === drums)?.description}
          >
            <option value="">なし</option>
            {patterns.drums.map((d) => (
              <option key={d.id} value={d.id}>
                {d.name}
              </option>
            ))}
          </select>
        </label>
      )}
    </>
  );
}
