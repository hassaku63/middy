import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { Timeline } from '@/shared/api/client';
import { MemoryProjectStore, type Project } from '@/shared/lib/storage/projects';
import { ProjectList } from './ProjectList';
import { KeepToggle } from './KeepToggle';
import { keptAs } from './keptAs';
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
      <KeepToggle
        projects={projects}
        current={{
          timeline,
          source: { kind: 'idea', image: '雨', candidateTitle: '窓辺の雨' },
        }}
      />
      <ProjectList projects={projects} onOpen={onOpen} />
    </>
  );
}

describe('KeepToggle + ProjectList + useProjects', () => {
  it('keeps in one click under the candidate title, lists, opens and deletes (#440)', async () => {
    const store = new MemoryProjectStore();
    const onOpen = vi.fn();
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={onOpen} />);

    // No name is asked for; the star says it is not kept yet.
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument();
    const star = screen.getByRole('button', { name: 'とっておく' });
    expect(star).toHaveAttribute('aria-pressed', 'false');
    expect(star).toHaveAttribute('title', 'とっておく');
    expect(star).toHaveTextContent('☆');
    await user.click(star);
    await waitFor(() => expect(star).toHaveAttribute('aria-pressed', 'true'));
    expect(star).toHaveAttribute('title', 'とっておいた例から外す');
    expect(star).toHaveTextContent('★');

    // The list stays folded until asked for.
    await user.click(await screen.findByText('とっておいた例（1）'));
    const list = screen.getByRole('list', { name: 'とっておいた例の一覧' });
    const item = within(list).getByRole('listitem');
    expect(item).toHaveTextContent('窓辺の雨');
    expect(item).toHaveTextContent('A minor · 90 BPM');

    await user.click(within(item).getByRole('button', { name: '開く' }));
    expect(onOpen).toHaveBeenCalledWith(expect.objectContaining({ name: '窓辺の雨', timeline }));

    await user.click(within(item).getByRole('button', { name: '窓辺の雨 を削除' }));
    await waitFor(() =>
      expect(screen.queryByRole('list', { name: 'とっておいた例の一覧' })).not.toBeInTheDocument(),
    );
    expect(await store.list()).toHaveLength(0);
    expect(star).toHaveAttribute('aria-pressed', 'false');
  });

  it('lets a kept example go with a second press, and never keeps it twice (BL-296)', async () => {
    const store = new MemoryProjectStore();
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={vi.fn()} />);
    const star = screen.getByRole('button', { name: 'とっておく' });
    await user.click(star);
    await waitFor(() => expect(star).toHaveAttribute('aria-pressed', 'true'));
    expect(await store.list()).toHaveLength(1);
    await user.click(star);
    await waitFor(() => expect(star).toHaveAttribute('aria-pressed', 'false'));
    expect(await store.list()).toHaveLength(0);
  });

  it('treats the same candidate in another key as another example (BL-296)', () => {
    const kept: Project = {
      id: 'a',
      name: '窓辺の雨',
      createdAt: 0,
      updatedAt: 0,
      timeline,
      source: { kind: 'idea', image: '雨', candidateTitle: '窓辺の雨' },
    };
    // A copy that sounds the same is the kept one; another key is not.
    expect(keptAs([kept], structuredClone(timeline))).toBe(kept);
    expect(keptAs([kept], { ...timeline, key: { tonic: 0, mode: 'minor' } })).toBeUndefined();
  });

  it('renames a kept example by keyboard: Enter keeps, Esc and an empty name do not (BL-297)', async () => {
    const store = new MemoryProjectStore();
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={vi.fn()} />);
    await user.click(screen.getByRole('button', { name: 'とっておく' }));
    await user.click(await screen.findByText('とっておいた例（1）'));
    const list = screen.getByRole('list', { name: 'とっておいた例の一覧' });

    // Enter keeps the new name; focus comes back to the button.
    const rename = within(list).getByRole('button', { name: '窓辺の雨 の名前を変える' });
    rename.focus();
    await user.keyboard('{Enter}');
    const field = within(list).getByRole('textbox', { name: '窓辺の雨 の新しい名前' });
    expect(field).toHaveFocus();
    expect(field).toHaveValue('窓辺の雨');
    await user.clear(field);
    await user.type(field, '雨の練習{Enter}');
    expect(await within(list).findByText('雨の練習')).toBeInTheDocument();
    await waitFor(() =>
      expect(within(list).getByRole('button', { name: '雨の練習 の名前を変える' })).toHaveFocus(),
    );
    expect((await store.list())[0]?.name).toBe('雨の練習');

    // Esc puts the old name back.
    await user.keyboard('{Enter}');
    await user.type(within(list).getByRole('textbox'), 'を書き換え{Escape}');
    expect(within(list).getByText('雨の練習')).toBeInTheDocument();

    // An empty name changes nothing, even when leaving the field.
    await user.click(within(list).getByRole('button', { name: '雨の練習 の名前を変える' }));
    await user.clear(within(list).getByRole('textbox'));
    await user.tab();
    expect(await within(list).findByText('雨の練習')).toBeInTheDocument();
    expect((await store.list())[0]?.name).toBe('雨の練習');

    // Leaving the field with a name keeps it.
    await user.click(within(list).getByRole('button', { name: '雨の練習 の名前を変える' }));
    await user.type(within(list).getByRole('textbox'), '（2）');
    await user.tab();
    expect(await within(list).findByText('雨の練習（2）')).toBeInTheDocument();
  });

  it("lists a kept example's progression and the reason kept with its card (BL-299)", async () => {
    const store = new MemoryProjectStore();
    const withHarmony: Timeline = {
      ...timeline,
      harmony: [
        { romanNumeral: 'i', start: 0, duration: 1920 },
        { romanNumeral: 'VI', start: 1920, duration: 1920 },
      ] as Timeline['harmony'],
    };
    await store.save({
      id: 'p1',
      name: '窓辺の雨',
      createdAt: 1,
      updatedAt: 1,
      timeline: withHarmony,
      source: {
        kind: 'idea',
        image: '雨',
        candidateTitle: '窓辺の雨',
        candidate: {
          title: '窓辺の雨',
          key: timeline.key,
          tempoBpm: 90,
          progression: ['i', 'VI'],
          beatsPerChord: 4,
          reason: '短調で落ち着いた切なさ。',
          timeline: withHarmony,
        },
      },
    });
    // A card kept before cards were stored still lists, without a reason.
    await store.save({
      id: 'p0',
      name: '前の例',
      createdAt: 0,
      updatedAt: 0,
      timeline,
      source: { kind: 'idea', image: '雪', candidateTitle: '前の例' },
    });
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={vi.fn()} />);
    await user.click(await screen.findByText('とっておいた例（2）'));
    const items = within(screen.getByRole('list', { name: 'とっておいた例の一覧' })).getAllByRole(
      'listitem',
    );
    expect(items[0]).toHaveTextContent('i - VI');
    expect(items[0]).toHaveTextContent('短調で落ち着いた切なさ。');
    expect(items[1]).toHaveTextContent('前の例');
  });

  it('shows no list while nothing is saved', () => {
    function Empty() {
      const projects = useProjects(new MemoryProjectStore());
      return <ProjectList projects={projects} onOpen={vi.fn()} />;
    }
    render(<Empty />);
    expect(screen.queryByRole('region', { name: 'とっておいた例' })).not.toBeInTheDocument();
  });

  it('says so when the store refuses a save, and can be tried again (review)', async () => {
    const store = new MemoryProjectStore();
    store.save = () => Promise.reject(new Error('QuotaExceededError'));
    const user = userEvent.setup();
    render(<Harness store={store} onOpen={vi.fn()} />);
    await user.click(screen.getByRole('button', { name: 'とっておく' }));
    expect(await screen.findByRole('alert')).toHaveTextContent('とっておけませんでした');
    expect(screen.getByRole('button', { name: 'とっておく' })).toBeEnabled();
    expect(screen.getByRole('button', { name: 'とっておく' })).toHaveAttribute(
      'aria-pressed',
      'false',
    );
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
    await user.click(screen.getByRole('button', { name: 'とっておく' }));
    expect(await screen.findByText('とっておいた例（1）')).toBeInTheDocument();
    firstLoad([]); // the first load answers last, with the list from before the save
    await new Promise((r) => setTimeout(r, 0));
    expect(screen.getByText('とっておいた例（1）')).toBeInTheDocument();
  });
});
