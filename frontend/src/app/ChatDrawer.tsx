import { CHAT_CLOSE_ID, CHAT_TOGGLE_ID, type ChatDrawer } from './useChatDrawer';

/** The id of the conversation's column, which the toggle controls. */
export const CONVERSATION_ID = 'conversation';

/** 「会話」 in the toolbar: shown only on a short, wide screen (global.css). */
export function ChatToggle({ drawer }: { drawer: ChatDrawer }) {
  return (
    <button
      type="button"
      className="app__chat-toggle"
      id={CHAT_TOGGLE_ID}
      aria-expanded={drawer.open}
      aria-controls={CONVERSATION_ID}
      onClick={drawer.toggle}
    >
      会話
    </button>
  );
}

/** 「閉じる」 at the head of the drawer, shown only while it is open there. */
export function ChatClose({ drawer }: { drawer: ChatDrawer }) {
  return (
    <button type="button" className="app__chat-close" id={CHAT_CLOSE_ID} onClick={drawer.close}>
      会話を閉じる
    </button>
  );
}
