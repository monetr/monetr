'use client';

import React, { useCallback, useEffect, useRef } from 'react';

export default function InterfaceExample(): JSX.Element {
  const ref = useRef<HTMLIFrameElement>(null);

  const onLoad = useCallback(() => {
    ref.current?.classList.remove('opacity-0');
    ref.current?.classList.remove('scale-90');
    ref.current?.classList.add('opacity-90');
    ref.current?.classList.add('scale-100');
  }, []);;

  useEffect(() => {
    if (!ref.current) {
      return undefined;
    }

    const current = ref.current;
    ref.current.addEventListener('load', onLoad);

    return () => {
      current.removeEventListener('load', onLoad);
    };
  });

  return (
    <iframe
      ref={ ref }
      title='monetr interface'
      loading='lazy'
      className='w-full h-full translate-x-0 translate-y-0 scale-90 delay-150 duration-500 ease-in-out rounded-2xl mt-8 shadow-2xl z-10 backdrop-blur-md bg-black/90 transition-all opacity-0 pointer-events-none select-none max-w-[1280px] max-h-[720px] aspect-video-vertical md:aspect-video'
      src='/_storybook/iframe.html?viewMode=story&id=new-ui--transactions&shortcuts=false&singleStory=true&args='
    />
  );
}
