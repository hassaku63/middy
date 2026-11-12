import type { Pwa } from './usePwa';
import './PwaBanner.css';

/**
 * One-line notices about the app itself: offline and a new version. Renders
 * nothing when there is nothing to say; being ready to work offline is
 * normal and says nothing (principle 7, BL-071). The install offer is a
 * small button by the title instead (InstallButton, #439).
 */
export function PwaBanner({ pwa }: { pwa: Pwa }) {
  const notices = [
    !pwa.online && (
      <p key="offline" className="pwa__notice is-offline" role="status">
        オフラインです。とっておいた例は開いて再生できます。生成と AI
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
  ].filter(Boolean);

  if (notices.length === 0) return null;
  return (
    <div className="pwa" aria-label="アプリの状態">
      {notices}
    </div>
  );
}

/**
 * The install offer, by the title (#439, web.dev's header pattern): shown
 * only while the browser offers to install, taking no line of its own.
 */
export function InstallButton({ pwa }: { pwa: Pwa }) {
  if (!pwa.installable) return null;
  return (
    <button
      type="button"
      className="pwa__install"
      title="middy をアプリとしてインストールする"
      onClick={() => void pwa.promptInstall()}
    >
      <span aria-hidden="true">⤓ </span>アプリにする
    </button>
  );
}
