import React from 'react';
import APIKeysCard from '@monetr/interface/components/settings/security/APIKeysCard';

export default function SettingsAPIKeys(): JSX.Element {
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold mb-4 text-dark-monetr-content-emphasis">API Keys</h1>
      <div className="mb-6 p-4 bg-dark-monetr-background-subtle border border-dark-monetr-border rounded-lg">
        <h2 className="text-lg font-semibold mb-2 text-dark-monetr-content-emphasis">About API Keys</h2>
        <p className="mb-2 text-dark-monetr-content">
          API keys allow you to access your monetr data programmatically through the REST API.
        </p>
        <p className="mb-2 text-dark-monetr-content">
          <strong className="text-dark-monetr-content-emphasis">Important:</strong> API keys provide the same level of access as your account, so keep them secure.
          Never share your API keys or store them in publicly accessible areas such as GitHub or client-side code.
        </p>
        <p className="text-dark-monetr-content">
          When you create a new API key, you'll only be shown the key once. Make sure to copy it and store it securely.
        </p>
      </div>
      <APIKeysCard />
    </div>
  );
}
