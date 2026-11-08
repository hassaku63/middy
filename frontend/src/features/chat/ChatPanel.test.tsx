import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { IdeaCandidate } from '@/shared/api/client';
import { ChatPanel } from './ChatPanel';
import type { Ideas } from './useIdeas';

const candidate: IdeaCandidate = {
  title: '窓辺の雨',
  key: { tonic: 9, mode: 'minor' },
  tempoBpm: 72,
  progression: ['i', 'VI', 'III', 'VII'],
  beatsPerChord: 4,
  reason: '短調で落ち着いた切なさ。',
  timeline: {
    ppq: 480,
    tempoBpm: 72,
    timeSignature: { numerator: 4, denominator: 4 },
    key: { tonic: 9, mode: 'minor' },
    tracks: [],
    harmony: [],
  },
};

function ideasWith(partial: Partial<Ideas>): Ideas {
  return {
    turns: [],
    busy: false,
    error: null,
    send: vi.fn().mockResolvedValue(undefined),
    reset: vi.fn(),
    ...partial,
  };
}

describe('ChatPanel', () => {
  it('shows hints and sends the typed image', async () => {
    const ideas = ideasWith({});
    const user = userEvent.setup();
    render(<ChatPanel ideas={ideas} onOpenCandidate={vi.fn()} />);

    expect(screen.getByRole('button', { name: '送る' })).toBeDisabled();
    await user.click(screen.getByRole('button', { name: '雨の日曜の午後' }));
    expect(screen.getByLabelText('イメージ')).toHaveValue('雨の日曜の午後');
    await user.click(screen.getByRole('button', { name: '送る' }));

    expect(ideas.send).toHaveBeenCalledWith('雨の日曜の午後');
    expect(screen.getByLabelText('イメージ')).toHaveValue('');
  });

  it('renders the conversation with candidate cards and opens one', async () => {
    const onOpen = vi.fn();
    const ideas = ideasWith({
      turns: [
        { role: 'user', content: '雨の日曜' },
        { role: 'assistant', content: '静かな雨。', mood: ['静か'], candidates: [candidate] },
      ],
    });
    const user = userEvent.setup();
    render(<ChatPanel ideas={ideas} onOpenCandidate={onOpen} openExample={null} />);

    const log = screen.getByRole('list', { name: '会話' });
    expect(within(log).getByText('雨の日曜')).toBeInTheDocument();
    expect(within(log).getByText('静かな雨。')).toBeInTheDocument();
    expect(
      within(screen.getByRole('list', { name: '雰囲気' })).getByText('静か'),
    ).toBeInTheDocument();

    const card = screen.getByRole('article', { name: '窓辺の雨' });
    expect(within(card).getByText('A minor')).toBeInTheDocument();
    expect(within(card).getByText('72 BPM')).toBeInTheDocument();
    expect(within(card).getByText('i - VI - III - VII')).toBeInTheDocument();
    await user.click(within(card).getByRole('button', { name: 'この候補を開く' }));
    expect(onOpen).toHaveBeenCalledWith(candidate);
    expect(screen.getByRole('button', { name: '最初から' })).toBeInTheDocument();
  });

  it('names the candidate patterns from the catalogue, falling back to ids', () => {
    const ideas = ideasWith({
      turns: [
        {
          role: 'assistant',
          content: 'x',
          mood: [],
          candidates: [
            { ...candidate, accompaniment: 'arpeggio', drums: 'mystery', feel: 'swing' },
          ],
        },
      ],
    });
    render(
      <ChatPanel
        ideas={ideas}
        onOpenCandidate={vi.fn()}
        patterns={{
          accompaniments: [
            { id: 'arpeggio', name: 'アルペジオ', description: '', program: 0, hits: [] },
          ],
          drums: [],
          feels: [{ id: 'swing', name: 'スウィング', description: '跳ねる' }],
        }}
      />,
    );
    const card = screen.getByRole('article', { name: '窓辺の雨' });
    expect(within(card).getByText('伴奏: アルペジオ')).toBeInTheDocument();
    expect(within(card).getByText('ドラム: mystery')).toBeInTheDocument();
    expect(within(card).getByText('ノリ: スウィング')).toBeInTheDocument();
  });

  it('marks the open candidate and shows busy and error states', () => {
    const ideas = ideasWith({
      turns: [{ role: 'assistant', content: 'x', mood: [], candidates: [candidate] }],
      busy: true,
      error: 'AI に接続できませんでした。',
    });
    render(<ChatPanel ideas={ideas} onOpenCandidate={vi.fn()} openExample={candidate} />);
    expect(screen.getByRole('button', { name: '開いています' })).toHaveAttribute(
      'aria-pressed',
      'true',
    );
    expect(screen.getByText('考え中…')).toBeInTheDocument();
    expect(screen.getByRole('alert')).toHaveTextContent('AI に接続できませんでした');
    expect(screen.getByLabelText('イメージ')).toBeDisabled();
  });

  it('keeps the latest exchange open and folds the earlier ones (BL-068)', () => {
    const ideas = ideasWith({
      turns: [
        { role: 'user', content: '雨の日曜の午後' },
        { role: 'assistant', content: '静かな雨。', mood: [], candidates: [candidate] },
        { role: 'user', content: 'もう少し明るく' },
        { role: 'assistant', content: '少し明るく。', mood: [], candidates: [] },
      ],
    });
    render(<ChatPanel ideas={ideas} onOpenCandidate={vi.fn()} />);

    const latest = screen.getByRole('list', { name: '会話' });
    expect(within(latest).getByText('もう少し明るく')).toBeInTheDocument();
    expect(within(latest).queryByText('雨の日曜の午後')).not.toBeInTheDocument();
    expect(screen.getByText('これまでのやり取り（1）')).toBeInTheDocument();
    const earlier = screen.getByRole('list', { name: 'これまでの会話' });
    expect(within(earlier).getByText('雨の日曜の午後')).toBeInTheDocument();
  });

  it('shows no fold for the first exchange', () => {
    const ideas = ideasWith({ turns: [{ role: 'user', content: '雨' }], busy: true });
    render(<ChatPanel ideas={ideas} onOpenCandidate={vi.fn()} />);
    expect(screen.queryByText(/これまでのやり取り/)).not.toBeInTheDocument();
    expect(screen.getByText('考え中…')).toBeInTheDocument();
  });

  it('marks only the open example when titles repeat across answers (BL-070)', async () => {
    const again: IdeaCandidate = { ...candidate };
    const ideas = ideasWith({
      turns: [
        { role: 'user', content: '雨' },
        { role: 'assistant', content: 'x', mood: [], candidates: [candidate] },
        { role: 'user', content: 'やっぱり夜' },
        { role: 'assistant', content: 'y', mood: [], candidates: [again] },
      ],
    });
    const user = userEvent.setup();
    render(<ChatPanel ideas={ideas} onOpenCandidate={vi.fn()} openExample={again} />);
    await user.click(screen.getByText('これまでのやり取り（1）'));
    expect(screen.getAllByRole('button', { name: '開いています' })).toHaveLength(1);
    expect(screen.getAllByRole('button', { name: 'この候補を開く' })).toHaveLength(1);
    const latest = screen.getByRole('list', { name: '会話' });
    expect(within(latest).getByRole('button', { name: '開いています' })).toBeInTheDocument();
  });
});
