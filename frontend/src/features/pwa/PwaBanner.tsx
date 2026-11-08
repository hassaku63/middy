import type { Pwa } from './usePwa';
import './PwaBanner.css';

/**
 * One-line notices about the app itself: offline, a new version and the
 * install offer. Renders nothing when there is nothing to say; being ready
 * to work offline is normal and says nothing (principle 7, BL-071).
 */
export function PwaBanner({ pwa }: { pwa: Pwa }) {
  const notices = [
    !pwa.online && (
      <p key="offline" className="pwa__notice is-offline" role="status">
        オフラインです。保存した例は開いて再生できます。生成と AI
        の提案はオンラインに戻ってから使えます。
      </p>
    ),
    pwa.updateReady && (
      <p key="update" className="pwa__notice" role="status">
        新しいバージョンがあります。
        <button type="button" onClick={() => void pwa.applyUpdate()}>
          更新して再読み込み
        </button>
      </p>
    ),
    pwa.installable && (
      <p key="install" className="pwa__notice">
        middy をアプリとして使えます。
        <button type="button" onClick={() => void pwa.promptInstall()}>
          インストール
        </button>
      </p>
    ),
  ].filter(Boolean);

  if (notices.length === 0) return null;
  return (
    <div className="pwa" aria-label="アプリの状態">
      {notices}
    </div>
  );
}
