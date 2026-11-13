import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it } from 'vitest';
import { Fold } from './Fold';

describe('Fold (#511)', () => {
  it('starts closed, says its title and gist, and opens and closes by click', async () => {
    render(
      <Fold title="リズム" gist="伴奏: アルペジオ · ドラム: バラード">
        <p>構成音を 8 分音符で弾く</p>
      </Fold>,
    );
    const fold = screen.getByRole('group');
    expect(fold).not.toHaveAttribute('open');
    expect(screen.getByText('伴奏: アルペジオ · ドラム: バラード')).toBeInTheDocument();
    await userEvent.click(screen.getByText('リズム'));
    expect(fold).toHaveAttribute('open');
    await userEvent.click(screen.getByText('リズム'));
    expect(fold).not.toHaveAttribute('open');
  });
});
