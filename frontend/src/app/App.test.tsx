import { render, screen, within } from '@testing-library/react';
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

describe('App (scenario S1)', () => {
  afterEach(() => vi.unstubAllGlobals());

  it('generates a timeline from the form, shows chords and explanation, and lets the user pin a chord', async () => {
    const fetchMock = vi.fn((url: string) =>
      Promise.resolve(
        url.endsWith('/health')
          ? json(200, { status: 'ok', version: 'dev' })
          : url.endsWith('/patterns')
            ? json(200, { accompaniments: [], drums: [] })
            : json(200, { timeline: analysed }),
      ),
    );
    vi.stubGlobal('fetch', fetchMock);
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} />);

    expect(await screen.findByRole('status')).toHaveTextContent('バックエンド: ok');
    expect(screen.queryByRole('region', { name: '結果' })).not.toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: '進行から' }));
    await user.click(screen.getByRole('button', { name: '生成' }));

    const result = await screen.findByRole('region', { name: '結果' });
    expect(
      within(result)
        .getAllByRole('option')
        .map((o) => o.getAttribute('aria-label')),
    ).toEqual(['I (C)', 'V (G)']);
    // Playhead at 0: the explanation follows the first chord.
    expect(within(result).getByRole('heading')).toHaveTextContent('IC');

    await user.click(within(result).getByRole('option', { name: 'V (G)' }));
    expect(within(result).getByRole('heading')).toHaveTextContent('VG');
    expect(within(result).getByText('半終止')).toBeInTheDocument();

    await user.click(screen.getByRole('button', { name: /固定を解除/ }));
    expect(within(result).getByRole('heading')).toHaveTextContent('IC');

    const postCall = fetchMock.mock.calls.find(([u]) => u.endsWith('/timelines/from-progression'));
    expect(postCall).toBeDefined();
  });

  it('shows API validation errors from the envelope', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn((url: string) =>
        Promise.resolve(
          url.endsWith('/health')
            ? json(200, { status: 'ok', version: 'dev' })
            : json(400, {
                error: {
                  code: 'invalid_request',
                  message: 'progression[1]: "X" is not a roman numeral',
                },
              }),
        ),
      ),
    );
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} />);

    await user.click(screen.getByRole('button', { name: '進行から' }));
    await user.click(screen.getByRole('button', { name: '生成' }));

    expect(await screen.findByRole('alert')).toHaveTextContent('"X" is not a roman numeral');
    expect(screen.queryByRole('region', { name: '結果' })).not.toBeInTheDocument();
  });

  it('opens an AI candidate into the player, view and panel (scenario S2)', async () => {
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
    vi.stubGlobal(
      'fetch',
      vi.fn((url: string) =>
        Promise.resolve(
          url.endsWith('/health')
            ? json(200, { status: 'ok', version: 'dev' })
            : url.endsWith('/patterns')
              ? json(200, { accompaniments: [], drums: [] })
              : json(200, ideasReply),
        ),
      ),
    );
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} />);

    await user.type(screen.getByLabelText('イメージ'), '雨の日曜');
    await user.click(screen.getByRole('button', { name: '送る' }));
    const card = await screen.findByRole('article', { name: '窓辺の雨' });
    await user.click(within(card).getByRole('button', { name: 'この候補を開く' }));

    const result = screen.getByRole('region', { name: '結果' });
    expect(within(result).getAllByRole('option')).toHaveLength(2);
    expect(within(result).getByText('候補: 窓辺の雨')).toBeInTheDocument();
    expect(within(card).getByRole('button', { name: '開いています' })).toBeInTheDocument();
  });

  it('saves the loaded timeline and reopens it from the project list', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn((url: string) =>
        Promise.resolve(
          url.endsWith('/health')
            ? json(200, { status: 'ok', version: 'dev' })
            : url.endsWith('/patterns')
              ? json(200, { accompaniments: [], drums: [] })
              : json(200, { timeline: analysed }),
        ),
      ),
    );
    const store = new MemoryProjectStore();
    const user = userEvent.setup();
    render(<App createEngine={fakeEngine} projectStore={store} />);

    expect(screen.getByRole('button', { name: '保存' })).toBeDisabled();
    await user.click(screen.getByRole('button', { name: '進行から' }));
    await user.click(screen.getByRole('button', { name: '生成' }));
    await screen.findByRole('region', { name: '結果' });

    await user.type(screen.getByLabelText('プロジェクト名'), '練習 1');
    await user.click(screen.getByRole('button', { name: '保存' }));
    const list = await screen.findByRole('list', { name: 'プロジェクト一覧' });
    const saved = await store.list();
    expect(saved).toHaveLength(1);
    expect(saved[0]?.source).toMatchObject({ kind: 'progression' });

    await user.click(within(list).getByRole('button', { name: '開く' }));
    expect(screen.getByText('プロジェクト: 練習 1')).toBeInTheDocument();
  });

  it('reports an unreachable backend', async () => {
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new TypeError('network')));
    render(<App createEngine={fakeEngine} />);
    expect(await screen.findByRole('status')).toHaveTextContent('バックエンドに接続できません');
  });
});
