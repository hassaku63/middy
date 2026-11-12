import { useRef, useState, type KeyboardEvent } from 'react';
import type { Project } from '@/shared/lib/storage/projects';

type Props = {
  project: Project;
  onRename: (name: string) => void;
};

/**
 * A kept example's name, renamed in place (BL-297, #440): 「名前を変える」
 * turns it into a field. Enter or leaving the field keeps the new name, Esc
 * puts the old one back, and an empty name changes nothing. Focus goes into
 * the field and back to the button, so the keyboard alone does it all.
 */
export function ProjectName({ project, onRename }: Props) {
  const [draft, setDraft] = useState<string | null>(null);
  const button = useRef<HTMLButtonElement>(null);

  // Set once Enter, Esc or leaving the field has decided: removing the
  // field can blur it again, and that blur must not undo an Esc.
  const decided = useRef(false);

  const finish = (keep: boolean) => {
    if (draft === null || decided.current) return;
    decided.current = true;
    if (keep) onRename(draft);
    setDraft(null);
    // Back to where the user started, once the button is drawn again.
    requestAnimationFrame(() => button.current?.focus());
  };

  const onKeyDown = (e: KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' && !e.nativeEvent.isComposing) {
      e.preventDefault();
      finish(true);
    } else if (e.key === 'Escape') {
      e.preventDefault();
      finish(false);
    }
  };

  if (draft !== null) {
    return (
      <input
        className="projects__rename"
        aria-label={`${project.name} の新しい名前`}
        value={draft}
        maxLength={100}
        autoFocus
        onChange={(e) => setDraft(e.target.value)}
        onKeyDown={onKeyDown}
        onBlur={() => finish(true)}
      />
    );
  }
  return (
    <>
      <span className="projects__name">{project.name}</span>
      <button
        ref={button}
        type="button"
        onClick={() => {
          decided.current = false;
          setDraft(project.name);
        }}
        aria-label={`${project.name} の名前を変える`}
      >
        名前を変える
      </button>
    </>
  );
}
