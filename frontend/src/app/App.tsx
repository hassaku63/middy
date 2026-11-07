import { useEffect, useState } from 'react';
import { fetchHealth, type Health } from '@/shared/api/client';

type HealthState = { kind: 'loading' } | { kind: 'ok'; health: Health } | { kind: 'error' };

/**
 * App shell. Today it only proves the frontend-backend wiring; feature
 * screens (timeline, player, chat) will mount here in later milestones.
 */
export function App() {
  const [state, setState] = useState<HealthState>({ kind: 'loading' });

  useEffect(() => {
    let cancelled = false;
    fetchHealth()
      .then((health) => {
        if (!cancelled) setState({ kind: 'ok', health });
      })
      .catch(() => {
        if (!cancelled) setState({ kind: 'error' });
      });
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <main>
      <h1>middy</h1>
      <p>作曲と音楽理論の学習を助ける、MIDI 生成・再生・チャットアプリ。</p>
      <p role="status">{describe(state)}</p>
    </main>
  );
}

function describe(state: HealthState): string {
  switch (state.kind) {
    case 'loading':
      return 'バックエンドに接続中…';
    case 'ok':
      return `バックエンド: ${state.health.status} (${state.health.version})`;
    case 'error':
      return 'バックエンドに接続できません';
  }
}
