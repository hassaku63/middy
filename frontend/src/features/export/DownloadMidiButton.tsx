import type { Timeline } from '@/shared/api/client';
import type { saveBlobAsFile } from './saveBlobAsFile';
import { useMidiDownload } from './useMidiDownload';

export type DownloadMidiButtonProps = {
  timeline: Timeline | null;
  /** Injected in tests; defaults to the browser's anchor-click download. */
  saveFile?: typeof saveBlobAsFile;
};

/** "Save MIDI" button with inline error; the work happens in useMidiDownload. */
export function DownloadMidiButton({ timeline, saveFile }: DownloadMidiButtonProps) {
  const { busy, error, download } = useMidiDownload(saveFile);

  return (
    <span>
      <button
        type="button"
        onClick={() => {
          if (timeline) void download(timeline);
        }}
        disabled={!timeline || busy}
      >
        {busy ? '書き出し中…' : 'MIDI 書き出し'}
      </button>
      {error && <span role="alert"> {error}</span>}
    </span>
  );
}
