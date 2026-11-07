import { useState } from 'react';
import { ApiError, exportMidi, type Timeline } from '@/shared/api/client';
import { midiFileName } from './midiFileName';
import { saveBlobAsFile } from './saveBlobAsFile';

export type DownloadMidiButtonProps = {
  timeline: Timeline | null;
  /** Injected in tests; defaults to the browser's anchor-click download. */
  saveFile?: (blob: Blob, fileName: string) => void;
};

/** Fetches the SMF for the current timeline and saves it with a descriptive name. */
export function DownloadMidiButton({
  timeline,
  saveFile = saveBlobAsFile,
}: DownloadMidiButtonProps) {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const download = async () => {
    if (!timeline) return;
    setBusy(true);
    setError(null);
    try {
      const blob = await exportMidi(timeline);
      saveFile(blob, midiFileName(timeline));
    } catch (e) {
      setError(e instanceof ApiError ? e.message : 'MIDI の書き出しに失敗しました');
    } finally {
      setBusy(false);
    }
  };

  return (
    <span>
      <button type="button" onClick={() => void download()} disabled={!timeline || busy}>
        {busy ? '書き出し中…' : 'MIDI を保存'}
      </button>
      {error && <span role="alert"> {error}</span>}
    </span>
  );
}
