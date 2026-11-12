import { keyName } from '@/shared/lib/notation';
import type { Project } from '@/shared/lib/storage/projects';
import { ProjectName } from './ProjectName';
import type { Projects } from './useProjects';
import './projects.css';

type Props = {
  projects: Projects;
  onOpen: (project: Project) => void;
};

/**
 * Kept examples, folded until asked for (docs/product/experience.md). Shown
 * only when there is something saved, or a storage error to report.
 */
export function ProjectList({ projects, onOpen }: Props) {
  const count = projects.projects.length;
  if (count === 0 && !projects.error) return null;
  return (
    <section className="projects" aria-label="とっておいた例">
      {projects.error && (
        <p role="alert" className="projects__error">
          {projects.error}
        </p>
      )}
      {count > 0 && (
        <details className="projects__fold">
          <summary>とっておいた例（{count}）</summary>
          <ul className="projects__list" aria-label="とっておいた例の一覧">
            {projects.projects.map((p) => (
              <li key={p.id} className="projects__item">
                <ProjectName project={p} onRename={(name) => void projects.rename(p.id, name)} />
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
                <CardLines project={p} />
              </li>
            ))}
          </ul>
        </details>
      )}
    </section>
  );
}

/**
 * What the example is, below its row (BL-299, #441): the progression as it
 * sounds, and the AI's reason when the card was kept with it.
 */
function CardLines({ project }: { project: Project }) {
  const numerals = project.timeline.harmony.map((h) => h.romanNumeral);
  const reason = project.source.kind === 'idea' ? project.source.candidate?.reason : undefined;
  if (numerals.length === 0 && !reason) return null;
  return (
    <div className="projects__card">
      {numerals.length > 0 && <p>{numerals.join(' - ')}</p>}
      {reason && <p>{reason}</p>}
    </div>
  );
}
