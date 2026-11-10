import type { useHearOnOpen } from './useHearOnOpen';

/** The switch for playing on choosing, beside the player (#315). */
export function HearOnOpenToggle({ hear }: { hear: ReturnType<typeof useHearOnOpen> }) {
  return (
    <label className="app__lab">
      <input type="checkbox" checked={hear.on} onChange={(e) => hear.choose(e.target.checked)} />
      候補を選んだら再生
    </label>
  );
}
