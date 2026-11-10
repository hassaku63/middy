import type { Layout, useLayoutTrial } from './useLayoutTrial';

/**
 * The switch of the screen trial, shown with ?lab: where the conversation
 * sits (BL-184). It goes away with the trial.
 */
export function LabControls({ layout }: { layout: ReturnType<typeof useLayoutTrial> }) {
  if (!layout.lab) return null;
  return (
    <label className="app__lab">
      画面の配置（試験）
      <select value={layout.mode} onChange={(e) => layout.choose(e.target.value as Layout)}>
        <option value="stacked">会話が上、結果が下</option>
        <option value="desk">会話を左の列に</option>
        <option value="wide">会話を左の列に、タイムラインを横いっぱいに</option>
      </select>
    </label>
  );
}
