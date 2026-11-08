import { describe, expect, it } from 'vitest';
import type { Rhythm } from '@/shared/api/client';
import { grooveChanges } from './grooveChanges';

const part = (id: string, name: string) => ({
  id,
  name,
  description: '',
  hits: 4,
  onBeat: 4,
  offBeat: 0,
  finer: 0,
  syncopated: false,
  explanation: '',
});

const celtic: Rhythm = {
  accompaniment: part('reel', 'リールの裏打ち'),
  drums: part('bodhran', 'ボーラン風'),
  style: { id: 'celtic', name: 'ケルト' },
};

describe('grooveChanges (BL-087)', () => {
  it('names only the parts that changed', () => {
    const jazz: Rhythm = {
      accompaniment: part('comp', '4 ビートのコンピング'),
      drums: part('bodhran', 'ボーラン風'),
      feel: { id: 'swing', name: 'スウィング', explanation: '' },
    };
    expect(grooveChanges('ジャズ', celtic, jazz)).toBe(
      'ジャズで鳴らし直しました。和声はそのままです。伴奏: リールの裏打ち → 4 ビートのコンピング、ノリ: ストレート → スウィング。',
    );
  });

  it('reads a missing part as block chords, no drums, straight', () => {
    expect(grooveChanges('ジャズ', celtic, undefined)).toBe(
      'ジャズで鳴らし直しました。和声はそのままです。伴奏: リールの裏打ち → ブロック、ドラム: ボーラン風 → なし。',
    );
  });

  it('says so when the groove is the same', () => {
    expect(grooveChanges('ケルト', celtic, celtic)).toContain('伴奏・ドラム・ノリも前と同じです');
  });
});
