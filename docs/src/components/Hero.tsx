import Particles from './Particles';

import Image from 'next/image';

import Logo from '../assets/logo.svg';

export default function Hero(): JSX.Element {
  return (
    <div className="w-full">
      <div className="absolute inset-0 overflow-hidden pointer-events-none -z-10" aria-hidden="true">
        <div className="absolute flex items-center justify-center top-0 -translate-y-1/2 left-1/2 -translate-x-1/2 w-1/3 aspect-square">
          <div className="absolute inset-0 translate-z-0 bg-purple-500 rounded-full blur-[120px] opacity-50" />
        </div>
      </div>
      <div className="max-md:hidden absolute bottom-0 -mb-20 left-2/3 -translate-x-1/2 blur-2xl opacity-70 pointer-events-none" aria-hidden="true">
        <svg xmlns="http://www.w3.org/2000/svg" width="434" height="427">
          <defs>
            <linearGradient id="bs5-a" x1="19.609%" x2="50%" y1="14.544%" y2="100%">
              <stop offset="0%" stopColor="#A855F7" />
              <stop offset="100%" stopColor="#6366F1" stopOpacity="0" />
            </linearGradient>
          </defs>
          <path fill="url(#bs5-a)" fillRule="evenodd" d="m661 736 461 369-284 58z" transform="matrix(1 0 0 -1 -661 1163)" />
        </svg>
      </div>
      <Particles className="absolute inset-0 -z-10" />
      <div className="m-view-height m-view-width flex flex-col py-8 mx-auto items-center justify-center">
        <div className="max-w-3xl flex flex-col">
          <Image src={ Logo } alt="monetr logo" width={ 75 } height={ 75 } />
          <h1 className="text-5xl font-bold">monetr</h1>
          <h2 className="text-xl font-medium">
            monetr is currently in a <b>closed beta</b>! We are building a source-visible financial planning application
            focused on helping you plan and budget for recurring expenses, or future goals.
          </h2>
        </div>
      </div>
    </div>
  );
}
