import { useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import type { ProjectSource } from '@/shared/lib/storage/projects';
import { keptAs } from './keptAs';
import type { Projects } from './useProjects';
import './projects.css';

type Props = {
  projects: Projects;
  /** The open example; the toggle exists only while one is open (BL-060). */
  current: { timeline: Timeline; source: ProjectSource };
};

/**
 * Keeps the open example, or lets it go (#440, BL-296): a star beside the
 * example's title, ☆ until kept and ★ after. The star alone is the visible
 * label; the button is named 「とっておく」 and says whether it is pressed
 * (WCAG 4.1.2), and its tooltip says what pressing will do.
 */
export function KeepToggle({ projects, current }: Props) {
  const [busy, setBusy] = useState(false);
  const kept = keptAs(projects.projects, current.timeline);

  const toggle = async () => {
    setBusy(true);
    try {
      if (kept) await projects.remove(kept.id);
      else await projects.save('', current.timeline, current.source);
    } finally {
      setBusy(false);
    }
  };

  return (
    <button
      type="button"
      className="projects__keep"
      aria-label="とっておく"
      aria-pressed={kept !== undefined}
      title={kept ? 'とっておいた例から外す' : 'とっておく'}
      onClick={() => void toggle()}
      disabled={busy}
    >
      <span aria-hidden="true">{kept ? '★' : '☆'}</span>
    </button>
  );
}
