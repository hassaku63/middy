import { keyName } from '@/shared/lib/notation';
import type { Project } from '@/shared/lib/storage/projects';
import type { Projects } from './useProjects';
import './projects.css';

type Props = {
  projects: Projects;
  onOpen: (project: Project) => void;
};

/**
 * Saved examples, folded until asked for (docs/product/experience.md). Shown
 * only when there is something saved, or a storage error to report.
 */
export function ProjectList({ projects, onOpen }: Props) {
  const count = projects.projects.length;
  if (count === 0 && !projects.error) return null;
  return (
    <section className="projects" aria-label="保存した例">
      {projects.error && (
        <p role="alert" className="projects__error">
          {projects.error}
        </p>
      )}
      {count > 0 && (
        <details className="projects__fold">
          <summary>保存した例（{count}）</summary>
          <ul className="projects__list" aria-label="保存した例の一覧">
            {projects.projects.map((p) => (
              <li key={p.id} className="projects__item">
                <span className="projects__name">{p.name}</span>
                <span className="projects__meta">
                  {keyName(p.timeline.key)} · {p.timeline.tempoBpm} BPM ·{' '}
                  {new Date(p.updatedAt).toLocaleString('ja-JP')}
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
        </details>
      )}
    </section>
  );
}
