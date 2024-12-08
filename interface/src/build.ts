declare var RELEASE: string;
declare var REVISION: string;
declare var NODE_VERSION: string;

// Based on https://www.chatwoot.com/hc/user-guide/articles/1677587234-how-to-send-additional-user-information-to-chatwoot-using-sdk#set-the-user-in-the-widget
declare var $chatwoot: null | {
  reset: () => void;
  setUser: (userId: string, details: {
    name: string;
    email: string;
    identifier_hash: string;
  }) => void;
  setCustomAttributes: (attributes: { [key: string]: any }) => void;
  setLabel: (label: string) => void;
  removeLabel: (label: string) => void;
  toggle: (state?: 'open' | 'close') => void;
  toggleBubbleVisibility: (state?: 'show' | 'hide') => void;
};
declare var chatwootSettings: null | {
  hideMessageBubble: boolean;
  showUnreadMessagesDialoge: boolean;
  position: 'left' | 'right';
  locale: 'en';
  useBrowserLanguage: false;
  type: 'standard' | 'expanded_bubble';
  darkMode: 'auto' | 'light';
  baseDomain: string;
};
