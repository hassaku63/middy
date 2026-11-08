import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { MemoryProjectStore, type Project } from '@/shared/lib/storage/projects';
import { ProjectList } from './ProjectList';
import { SaveProjectForm } from './SaveProjectForm';
import { useProjects } from './useProjects';

const timeline: Timeline = {
  ppq: 480,
  tempoBpm: 90,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 9, mode: 'minor' },
  tracks: [],
  harmony: [],
};

function Harness({ store, onOpen }: { store: MemoryProjectStore; onOpen: (p: Project) => void }) {
  const projects = useProjects(store);
  return (
    <>
      <SaveProjectForm
        projects={projects}
        current={{ timeline, source: { kind: 'idea', image: '雨', candidateTitle: '窓辺の雨' } }}
      />
      <ProjectList projects={projects} onOpen={onOpen} />
    </>
  );
}

describe('SaveProjectForm + ProjectList + useProjects', () => {
  it('saves with a typed name, lists, opens and deletes', async () => {
    const store = new MemoryProjectStore();
    const onOpen = vi.fn();
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={onOpen} />);

    await user.type(screen.getByLabelText('例の名前'), '雨の曲');
    await user.click(screen.getByRole('button', { name: '保存' }));

    // The list stays folded until asked for.
    await user.click(await screen.findByText('保存した例（1）'));
    const list = screen.getByRole('list', { name: '保存した例の一覧' });
    const item = within(list).getByRole('listitem');
    expect(item).toHaveTextContent('雨の曲');
    expect(item).toHaveTextContent('A minor · 90 BPM');
    expect(screen.getByLabelText('例の名前')).toHaveValue('');

    await user.click(within(item).getByRole('button', { name: '開く' }));
    expect(onOpen).toHaveBeenCalledWith(expect.objectContaining({ name: '雨の曲', timeline }));

    await user.click(within(item).getByRole('button', { name: '雨の曲 を削除' }));
    await waitFor(() =>
      expect(screen.queryByRole('list', { name: '保存した例の一覧' })).not.toBeInTheDocument(),
    );
    expect(await store.list()).toHaveLength(0);
  });

  it('falls back to the candidate title when no name is typed', async () => {
    const store = new MemoryProjectStore();
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={vi.fn()} />);
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByText('窓辺の雨')).toBeInTheDocument();
  });

  it('shows no list while nothing is saved', () => {
    function Empty() {
      const projects = useProjects(new MemoryProjectStore());
      return <ProjectList projects={projects} onOpen={vi.fn()} />;
    }
    render(<Empty />);
    expect(screen.queryByRole('region', { name: '保存した例' })).not.toBeInTheDocument();
  });
});
