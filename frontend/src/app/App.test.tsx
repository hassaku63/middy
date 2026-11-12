import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { AudioEngine } from '@/features/player/usePlayer';
import type { Timeline } from '@/shared/api/client';
import { MemoryProjectStore } from '@/shared/lib/storage/projects';
import { App } from './App';

const analysed: Timeline = {
  ppq: 480,
  tempoBpm: 90,
  timeSignature: { numerator: 4, denominator: 4 },
  key: { tonic: 0, mode: 'major' },
  tracks: [
    {
      name: 'chords',
      channel: 0,
      program: 0,
      notes: [{ pitch: 60, velocity: 80, start: 0, duration: 1920 }],
    },
  ],
  harmony: [
    {
      start: 0,
      duration: 1920,
      chord: { root: 0, quality: 'maj' },
      romanNumeral: 'I',
      analysis: { degree: 1, function: 'T', explanation: 'I（C）はトニック（T）。' },
    },
    {
      start: 1920,
      duration: 1920,
      chord: { root: 7, quality: 'maj' },
      romanNumeral: 'V',
      analysis: {
        degree: 5,
        function: 'D',
        cadence: 'half',
        explanation: 'V（G）はドミナント（D）。',
      },
    },
  ],
};

function json(status: number, body: unknown) {
  return new Response(JSON.stringify(body), { status });
}

const fakeEngine = (): AudioEngine => ({
  synth: { play: () => undefined, allOff: () => undefined },
  clock: { now: () => 0 },
  resume: () => Promise.resolve(),
});

const ideasReply = {
  reply: '静かな雨。',
  mood: ['静か'],
  candidates: [
    {
      title: '窓辺の雨',
      key: { tonic: 0, mode: 'major' },
      tempoBpm: 90,
      progression: ['I', 'V'],
      beatsPerChord: 4,
      reason: 'テスト用。',
      timeline: analysed,
    },
  ],
};

function stubBackend() {
  vi.stubGlobal(
    'fetch',
    vi.fn((url: string) =>
      Promise.resolve(
        url.endsWith('/health')
          ? json(200, { status: 'ok', version: 'dev' })
          : url.endsWith('/patterns')
            ? json(200, { accompaniments: [], drums: [], feels: [], styles: [] })
            : json(200, ideasReply),
      ),
    ),
  );
}

/** Sends an image; the first example opens by itself (BL-064). */
/** Sends an image and opens its first example, as the user does (BL-219). */
async function sendImage(user: ReturnType<typeof userEvent.setup>, text = '雨の日曜') {
  await user.type(screen.getByLabelText('イメージ'), text);
  await user.click(screen.getByRole('button', { name: '送る' }));
  const card = await screen.findByRole('article', { name: '窓辺の雨' });
  await user.click(within(card).getByRole('button', { name: '窓辺の雨' }));
  return { card, result: await screen.findByRole('region', { name: '結果' }) };
}

describe('App docs/product/experience.md', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('offers the image as the only entry', async () => {
    stubBackend();
    render(<App createEngine={fakeEngine} />);

    // Healthy is silent (BL-065): once the probe has answered, no status line remains.
    await waitFor(() => expect(fetch).toHaveBeenCalledWith('/api/v1/health', expect.anything()));
    await waitFor(() => expect(screen.queryByRole('status')).not.toBeInTheDocument());
    expect(screen.getAllByRole('textbox')).toHaveLength(1); // the image only
    expect(screen.getByLabelText('イメージ')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: '進行から' })).not.toBeInTheDocument();
    expect(screen.queryByRole('navigation', { name: '入力方法' })).not.toBeInTheDocument();
    expect(screen.queryByRole('region', { name: '結果' })).not.toBeInTheDocument();
  });

  it('opens the example the user chooses, and lets the user pin a chord', async () => {
    stubBackend();
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} />);

    const { card, result } = await sendImage(user);
    expect(
      within(within(result).getByRole('listbox', { name: 'コード' }))
        .getAllByRole('option')
        .map((o) => o.getAttribute('aria-label')),
    ).toEqual(['I C', 'V G']);
    expect(within(result).getByRole('heading', { name: '窓辺の雨' })).toBeInTheDocument();
    expect(within(card).getByRole('button', { name: '窓辺の雨' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    // Playhead at 0: the explanation follows the first chord.
    expect(chordHeading(result)).toHaveTextContent('IC');

    await user.click(within(result).getByRole('option', { name: 'V G' }));
    expect(chordHeading(result)).toHaveTextContent('VG');
    expect(within(result).getByText('半終止')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /固定を解除/ }));
    expect(chordHeading(result)).toHaveTextContent('IC');
  });

  it('keeps the open example with the star and reopens it from the list (BL-296)', async () => {
    stubBackend();
    const store = new MemoryProjectStore();
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} projectStore={store} />);

    // Keeping belongs to an open example (BL-060).
    expect(screen.queryByRole('button', { name: 'とっておく' })).not.toBeInTheDocument();
    await sendImage(user);

    await user.click(screen.getByRole('button', { name: 'とっておく' }));
    await user.click(await screen.findByText('とっておいた例（1）'));
    const list = screen.getByRole('list', { name: 'とっておいた例の一覧' });
    const saved = await store.list();
    expect(saved).toHaveLength(1);
    expect(saved[0]?.source).toMatchObject({
      kind: 'idea',
      image: '雨の日曜',
      candidateTitle: '窓辺の雨',
    });

    await user.click(within(list).getByRole('button', { name: '開く' }));
    // Opened from the list, it is the kept example: its star is filled.
    expect(resultHeading('窓辺の雨')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'とっておく' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );

    // Renaming it in the list renames the title above the result (BL-297).
    await user.click(within(list).getByRole('button', { name: '窓辺の雨 の名前を変える' }));
    await user.type(within(list).getByRole('textbox'), '（練習）{Enter}');
    expect(await findResultHeading('窓辺の雨（練習）')).toBeInTheDocument();
  });

  it('opens nothing by itself when an answer arrives, and keeps what the user chose (BL-219)', async () => {
    stubBackend();
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} />);

    await user.type(screen.getByLabelText('イメージ'), '雨の日曜');
    await user.click(screen.getByRole('button', { name: '送る' }));
    const card = await screen.findByRole('article', { name: '窓辺の雨' });
    // The user has chosen nothing yet: nothing is marked and nothing is open.
    expect(within(card).getByRole('button', { name: '窓辺の雨' })).toHaveAttribute(
      'aria-pressed',
      'false',
    );
    expect(screen.queryByRole('region', { name: '結果' })).not.toBeInTheDocument();

    await user.click(within(card).getByRole('button', { name: '窓辺の雨' }));
    expect(await findResultHeading('窓辺の雨')).toBeInTheDocument();

    // A later answer adds its cards; the example the user chose stays open.
    await user.type(screen.getByLabelText('イメージ'), 'もう少し明るく');
    await user.click(screen.getByRole('button', { name: '送る' }));
    await screen.findAllByRole('article', { name: '窓辺の雨' });
    expect(resultHeading('窓辺の雨')).toBeInTheDocument();
    expect(screen.getAllByRole('button', { pressed: true })).toHaveLength(1);
  });

  it('keeps what the user opened while the answer was on its way (BL-064)', async () => {
    let answer: (r: Response) => void = () => undefined;
    vi.stubGlobal(
      'fetch',
      vi.fn((url: string) =>
        url.endsWith('/health')
          ? Promise.resolve(json(200, { status: 'ok', version: 'dev' }))
          : url.endsWith('/patterns')
            ? Promise.resolve(json(200, { accompaniments: [], drums: [], feels: [], styles: [] }))
            : new Promise<Response>((resolve) => (answer = resolve)),
      ),
    );
    const store = new MemoryProjectStore();
    await store.save({
      id: 'p1',
      name: '前の例',
      createdAt: 1,
      updatedAt: 1,
      timeline: analysed,
      source: { kind: 'idea', image: '雪', candidateTitle: '雪' },
    });
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} projectStore={store} />);

    await user.type(screen.getByLabelText('イメージ'), '雨の日曜');
    await user.click(screen.getByRole('button', { name: '送る' }));
    await user.click(await screen.findByText('とっておいた例（1）'));
    await user.click(screen.getByRole('button', { name: '開く' }));
    expect(resultHeading('前の例')).toBeInTheDocument();

    answer(json(200, ideasReply));
    expect(await screen.findByRole('article', { name: '窓辺の雨' })).toBeInTheDocument();
    expect(resultHeading('前の例')).toBeInTheDocument();
    expect(
      within(screen.getByRole('region', { name: '結果' })).queryByRole('heading', {
        name: '窓辺の雨',
      }),
    ).not.toBeInTheDocument();
  });

  it('reports an unreachable backend', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('network')));
    render(<App createEngine={fakeEngine} />);
    expect(await screen.findByRole('status')).toHaveTextContent('バックエンドに接続できません');
  });

  it("changes the open example's key through the engine, keeping the numerals (BL-078)", async () => {
    const inD: Timeline = {
      ...analysed,
      key: { tonic: 2, mode: 'major' },
      harmony: analysed.harmony.map((h, i) => ({
        ...h,
        chord: { ...h.chord, root: i === 0 ? 2 : 9 },
      })),
    };
    const fetchMock = vi.fn<(url: string, init?: RequestInit) => Promise<Response>>((url) =>
      Promise.resolve(
        url.endsWith('/health')
          ? json(200, { status: 'ok', version: 'dev' })
          : url.endsWith('/patterns')
            ? json(200, { accompaniments: [], drums: [], feels: [], styles: [] })
            : url.endsWith('/timelines/from-progression')
              ? json(200, { timeline: inD })
              : json(200, ideasReply),
      ),
    );
    vi.stubGlobal('fetch', fetchMock);
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} />);
    const { result } = await sendImage(user);

    await user.selectOptions(within(result).getByLabelText('キー'), '2');
    const chords = within(result).getByRole('listbox', { name: 'コード' });
    await waitFor(() =>
      expect(
        within(chords)
          .getAllByRole('option')
          .map((o) => o.getAttribute('aria-label')),
      ).toEqual(['I D', 'V A']),
    );
    const call = fetchMock.mock.calls.find(([u]) => u.endsWith('/timelines/from-progression'));
    expect(JSON.parse(call?.[1]?.body as string)).toMatchObject({
      key: { tonic: 2, mode: 'major' },
      progression: [
        { numeral: 'I', beats: 4 },
        { numeral: 'V', beats: 4 },
      ],
    });
    // Still the same example: the AI was not asked again.
    expect(fetchMock.mock.calls.filter(([u]) => u.endsWith('/ideas'))).toHaveLength(1);
    // The explanation still shows the card, as the original example (BL-298).
    const notes = within(result).getByLabelText('この例（AI の提案）');
    expect(notes).toHaveTextContent('テスト用。');
    expect(notes).toHaveTextContent('元の例');
    expect(within(result).getByRole('heading', { name: '窓辺の雨' })).toBeInTheDocument();
  });
});

/** The heading of the one-chord explanation inside the result. */
function chordHeading(result: HTMLElement) {
  return within(within(result).getByRole('region', { name: 'コードの解説' })).getByRole('heading');
}

/** The open example's title above the result (BL-296). */
function resultHeading(name: string) {
  return within(screen.getByRole('region', { name: '結果' })).getByRole('heading', { name });
}

async function findResultHeading(name: string) {
  return within(await screen.findByRole('region', { name: '結果' })).findByRole('heading', {
    name,
  });
}
