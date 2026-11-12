import { useCopyText } from './useCopyText';

/**
 * Copies the whole explanation as text (BL-300, #442): a small button in
 * the panel's top right corner, drawn faint and always there. Its icon is
 * two sheets of paper; after copying it shows ✓, and the words are read
 * out (aria-live). A refusal says how to copy by hand.
 */
export function CopyButton({ text }: { text: string }) {
  const { state, copy } = useCopyText();
  return (
    <div className="explanation__copy">
      <span className="explanation__copy-status" aria-live="polite">
        {state === 'copied' && 'コピーしました'}
        {state === 'failed' && 'コピーできませんでした。文字を選んでコピーしてください'}
      </span>
      <button
        type="button"
        className="explanation__copy-button"
        aria-label="解説をコピー"
        title="解説をコピー"
        onClick={() => void copy(text)}
      >
        {state === 'copied' ? <CheckIcon /> : <SheetsIcon />}
      </button>
    </div>
  );
}

function SheetsIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
      <rect x="5" y="1.5" width="9" height="10" rx="1.5" fill="none" stroke="currentColor" />
      <path
        d="M3 4.5h-.5A1.5 1.5 0 0 0 1 6v7.5A1.5 1.5 0 0 0 2.5 15H9a1.5 1.5 0 0 0 1.5-1.5V13"
        fill="none"
        stroke="currentColor"
      />
    </svg>
  );
}

function CheckIcon() {
  return (
    <svg width="16" height="16" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
      <path d="M3 8.5l3 3 7-7" fill="none" stroke="currentColor" strokeWidth="1.5" />
    </svg>
  );
}
