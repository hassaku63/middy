import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { ProgressionForm } from './ProgressionForm';

describe('ProgressionForm', () => {
  it('submits the default pop progression in C major at 90 BPM', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    render(<ProgressionForm onSubmit={onSubmit} />);

    await user.click(screen.getByRole('button', { name: '生成' }));

    expect(onSubmit).toHaveBeenCalledWith({
      key: { tonic: 0, mode: 'major' },
      tempoBpm: 90,
      progression: ['I', 'V', 'vi', 'IV'].map((numeral) => ({ numeral, beats: 4 })),
    });
  });

  it('lets the user change key, mode, progression, beats and tempo', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    render(<ProgressionForm onSubmit={onSubmit} />);

    await user.selectOptions(screen.getByLabelText('調'), '9');
    await user.selectOptions(screen.getByLabelText('モード'), 'minor');
    await user.clear(screen.getByLabelText('進行（ローマ数字）'));
    await user.type(screen.getByLabelText('進行（ローマ数字）'), 'i iv V7 i');
    await user.clear(screen.getByLabelText('1 コードの拍数'));
    await user.type(screen.getByLabelText('1 コードの拍数'), '2');
    await user.clear(screen.getByLabelText('テンポ (BPM)'));
    await user.type(screen.getByLabelText('テンポ (BPM)'), '120');
    await user.click(screen.getByRole('button', { name: '生成' }));

    expect(onSubmit).toHaveBeenCalledWith({
      key: { tonic: 9, mode: 'minor' },
      tempoBpm: 120,
      progression: ['i', 'iv', 'V7', 'i'].map((numeral) => ({ numeral, beats: 2 })),
    });
  });

  it('refuses an empty progression without calling the API', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    render(<ProgressionForm onSubmit={onSubmit} />);

    await user.clear(screen.getByLabelText('進行（ローマ数字）'));
    await user.click(screen.getByRole('button', { name: '生成' }));

    expect(onSubmit).not.toHaveBeenCalled();
    expect(screen.getByRole('alert')).toHaveTextContent('進行を入力してください');
  });

  it('offers patterns when a catalogue is given and sends the chosen ids', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    render(
      <ProgressionForm
        onSubmit={onSubmit}
        patterns={{
          accompaniments: [
            { id: 'block', name: 'ブロック', description: 'b', program: 0 },
            { id: 'arpeggio', name: 'アルペジオ', description: 'a', program: 0 },
          ],
          drums: [{ id: 'eight-beat', name: '8 ビート', description: 'd' }],
        }}
      />,
    );
    await user.selectOptions(screen.getByLabelText('伴奏'), 'arpeggio');
    await user.selectOptions(screen.getByLabelText('ドラム'), 'eight-beat');
    await user.click(screen.getByRole('button', { name: '生成' }));
    expect(onSubmit).toHaveBeenCalledWith(
      expect.objectContaining({ accompaniment: 'arpeggio', drums: 'eight-beat' }),
    );
  });

  it('hides the pattern selects without a catalogue and omits the ids', async () => {
    const onSubmit = vi.fn();
    const user = userEvent.setup();
    render(<ProgressionForm onSubmit={onSubmit} />);
    expect(screen.queryByLabelText('伴奏')).not.toBeInTheDocument();
    await user.click(screen.getByRole('button', { name: '生成' }));
    const req = onSubmit.mock.calls[0]?.[0] as Record<string, unknown>;
    expect(req).not.toHaveProperty('accompaniment');
    expect(req).not.toHaveProperty('drums');
  });

  it('is disabled while busy', () => {
    render(<ProgressionForm onSubmit={vi.fn()} busy />);
    expect(screen.getByRole('button', { name: '生成中…' })).toBeDisabled();
  });
});
