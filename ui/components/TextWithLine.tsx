import React from 'react';

export interface TextWithLineProps {
  children: React.ReactNode | JSX.Element;
}

export default function TextWithLine(props: TextWithLineProps): JSX.Element {
  return (
    <div className='w-full flex items-center'>
      <div className="flex-grow border-t border-gray-400" style={{
        top: '1.2em',
      }} />
      <span className="text-center relative p-1.5 dark:text-white basis-auto">
        {props.children}
      </span>
      <div className="flex-grow border-t border-gray-400" style={{
        top: '1.2em',
      }} />
    </div>
  );
}

