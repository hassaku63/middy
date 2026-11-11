import { useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import type { ProjectSource } from '@/shared/lib/storage/projects';
import type { Projects } from './useProjects';
import './projects.css';

type Props = {
  projects: Projects;
  /** The open example; the button exists only while one is open (BL-060). */
  current: { timeline: Timeline; source: ProjectSource };
};

/**
 * Saves the open example in one click, under the candidate's title (#440):
 * the name is given, not asked for. Follows 再生 in the result toolbar.
 */
export function SaveProjectForm({ projects, current }: Props) {
  const [saving, setSaving] = useState(false);

  const save = async () => {
    setSaving(true);
    try {
      await projects.save('', current.timeline, current.source);
    } finally {
      setSaving(false);
    }
  };

  return (
    <button type="button" className="projects__save" onClick={() => void save()} disabled={saving}>
      保存
    </button>
  );
}
