import React from 'react';

export default function GeneralSettings(): JSX.Element {

  return (
    <div className="p-2.5 w-full h-full bg-fuchsia-200">
      <span className="text-2xl">
        General Settings
      </span>
      <div className="grid grid-cols-2 gap-2.5">
        <span>
          Email
        </span>
        <span>
          Input
        </span>
      </div>
    </div>
  );
}
