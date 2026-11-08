import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import type { PatternsResponse } from '@/shared/api/client';
import { StyleSelect } from './StyleSelect';

const style = (id: string, name: string): PatternsResponse['styles'][number] => ({
  id,
  name,
  description: '',
  accompaniments: [],
  drums: [],
  feels: [],
  grooves: [],
});
const styles = [style('rock', 'ロック'), style('jazz', 'ジャズ')];

describe('StyleSelect', () => {
  it('shows a style the catalogue does not list as unchosen, so the first can be picked (review)', async () => {
    const onChange = vi.fn();
    const user = userEvent.setup();
    render(
      <StyleSelect
        styles={styles}
        current="gone"
        busy={false}
        error={null}
        note={null}
        onChange={onChange}
      />,
    );
    const select = screen.getByLabelText('スタイル');
    expect(select).toHaveValue('');
    await user.selectOptions(select, 'rock');
    expect(onChange).toHaveBeenCalledWith('rock');
  });

  it('shows the listed style as chosen', () => {
    render(
      <StyleSelect
        styles={styles}
        current="jazz"
        busy={false}
        error={null}
        note={null}
        onChange={vi.fn()}
      />,
    );
    expect(screen.getByLabelText('スタイル')).toHaveValue('jazz');
    expect(screen.queryByRole('option', { name: '—' })).not.toBeInTheDocument();
  });
});
