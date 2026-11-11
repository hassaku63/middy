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
  it('saves in one click under the candidate title, lists, opens and deletes (#440)', async () => {
    const store = new MemoryProjectStore();
    const onOpen = vi.fn();
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={onOpen} />);

    // No name is asked for.
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '保存' }));

    // The list stays folded until asked for.
    await user.click(await screen.findByText('保存した例（1）'));
    const list = screen.getByRole('list', { name: '保存した例の一覧' });
    const item = within(list).getByRole('listitem');
    expect(item).toHaveTextContent('窓辺の雨');
    expect(item).toHaveTextContent('A minor · 90 BPM');

    await user.click(within(item).getByRole('button', { name: '開く' }));
    expect(onOpen).toHaveBeenCalledWith(expect.objectContaining({ name: '窓辺の雨', timeline }));

    await user.click(within(item).getByRole('button', { name: '窓辺の雨 を削除' }));
    await waitFor(() =>
      expect(screen.queryByRole('list', { name: '保存した例の一覧' })).not.toBeInTheDocument(),
    );
    expect(await store.list()).toHaveLength(0);
  });

  it('shows no list while nothing is saved', () => {
    function Empty() {
      const projects = useProjects(new MemoryProjectStore());
      return <ProjectList projects={projects} onOpen={vi.fn()} />;
    }
    render(<Empty />);
    expect(screen.queryByRole('region', { name: '保存した例' })).not.toBeInTheDocument();
  });

  it('says so when the store refuses a save, and can be tried again (review)', async () => {
    const store = new MemoryProjectStore();
    store.save = () => Promise.reject(new Error('QuotaExceededError'));
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={vi.fn()} />);
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('保存できませんでした');
    expect(screen.getByRole('button', { name: '保存' })).toBeEnabled();
  });

  it('keeps a save that lands before a slow first load (review)', async () => {
    const store = new MemoryProjectStore();
    const list = store.list.bind(store);
    let firstLoad: (projects: Project[]) => void = () => undefined;
    let calls = 0;
    store.list = () => {
      calls += 1;
      return calls === 1 ? new Promise<Project[]>((r) => (firstLoad = r)) : list();
    };
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={vi.fn()} />);
    await user.click(screen.getByRole('button', { name: '保存' }));
    expect(await screen.findByText('保存した例（1）')).toBeInTheDocument();
    firstLoad([]); // the first load answers last, with the list from before the save
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.getByText('保存した例（1）')).toBeInTheDocument();
  });
});
