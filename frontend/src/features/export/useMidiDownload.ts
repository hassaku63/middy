import { useCallback, useState } from 'react';
import { ApiError, exportMidi, type Timeline } from '@/shared/api/client';
import { midiFileName } from './midiFileName';
import { saveBlobAsFile } from './saveBlobAsFile';

export type MidiDownload = {
  busy: boolean;
  error: string | null;
  download: (timeline: Timeline) => Promise<void>;
};

/** Fetches the SMF for a timeline and hands it to the browser's download. */
export function useMidiDownload(saveFile: typeof saveBlobAsFile = saveBlobAsFile): MidiDownload {
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const download = useCallback(
    async (timeline: Timeline) => {
      setBusy(true);
      setError(null);
      try {
        saveFile(await exportMidi(timeline), midiFileName(timeline));
      } catch (e) {
        setError(e instanceof ApiError ? e.message : 'MIDI の書き出しに失敗しました');
      } finally {
        setBusy(false);
      }
    },
    [saveFile],
  );

  return { busy, error, download };
}
