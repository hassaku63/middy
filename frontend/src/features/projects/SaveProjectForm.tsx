import { useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import type { ProjectSource } from '@/shared/lib/storage/projects';
import type { Projects } from './useProjects';
import './projects.css';

type Props = {
  projects: Projects;
  /** The open example; the form exists only while one is open (BL-060). */
  current: { timeline: Timeline; source: ProjectSource };
};

/** Saves the open example under a name. Follows 再生 in the result toolbar. */
export function SaveProjectForm({ projects, current }: Props) {
  const [name, setName] = useState('');
  const [saving, setSaving] = useState(false);

  const save = async () => {
    setSaving(true);
    try {
      // Keep the name when the store refused, so the user can try again.
      if (await projects.save(name, current.timeline, current.source)) setName('');
    } finally {
      setSaving(false);
    }
  };

  return (
    <form
      className="projects__save"
      aria-label="この例を保存"
      onSubmit={(e) => {
        e.preventDefault();
        void save();
      }}
    >
      <input
        type="text"
        value={name}
        onChange={(e) => setName(e.target.value)}
        placeholder="名前（空なら自動）"
        aria-label="例の名前"
        disabled={saving}
      />
      <button type="submit" disabled={saving}>
        保存
      </button>
    </form>
  );
}
