import { useState } from 'react';
import type { Timeline } from '@/shared/api/client';
import { pitchClassName } from '@/shared/lib/notation';
import type { Project, ProjectSource } from '@/shared/lib/storage/projects';
import type { Projects } from './useProjects';
import './ProjectsPanel.css';

type Props = {
  projects: Projects;
  /** What would be saved right now, or null when nothing is loaded. */
  current: { timeline: Timeline; source: ProjectSource } | null;
  onOpen: (project: Project) => void;
};

/** Save the current timeline under a name; list, open and delete saved ones. */
export function ProjectsPanel({ projects, current, onOpen }: Props) {
  const [name, setName] = useState('');
  const [saving, setSaving] = useState(false);

  const save = async () => {
    if (!current) return;
    setSaving(true);
    try {
      await projects.save(name, current.timeline, current.source);
      setName('');
    } finally {
      setSaving(false);
    }
  };

  return (
    <section className="projects" aria-label="保存したプロジェクト">
      <form
        className="projects__save"
        onSubmit={(e) => {
          e.preventDefault();
          void save();
        }}
      >
        <input
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="プロジェクト名（空なら自動）"
          aria-label="プロジェクト名"
          disabled={!current || saving}
        />
        <button type="submit" disabled={!current || saving}>
          保存
        </button>
      </form>
      {projects.error && (
        <p role="alert" className="projects__error">
          {projects.error}
        </p>
      )}
      {projects.projects.length > 0 && (
        <ul className="projects__list" aria-label="プロジェクト一覧">
          {projects.projects.map((p) => (
            <li key={p.id} className="projects__item">
              <span className="projects__name">{p.name}</span>
              <span className="projects__meta">
                {pitchClassName(p.timeline.key.tonic)} {p.timeline.key.mode} · {p.timeline.tempoBpm}{' '}
                BPM · {new Date(p.updatedAt).toLocaleString('ja-JP')}
              </span>
              <button type="button" onClick={() => onOpen(p)}>
                開く
              </button>
              <button
                type="button"
                onClick={() => void projects.remove(p.id)}
                aria-label={`${p.name} を削除`}
              >
                削除
              </button>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}
