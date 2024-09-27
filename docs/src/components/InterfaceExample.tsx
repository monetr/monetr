import React from 'react';

export default function InterfaceExample(): JSX.Element {
  return (
    <iframe
      title='monetr interface'
      loading='lazy'
      className='w-full h-full rounded-2xl mt-8 shadow-2xl z-10 backdrop-blur-md bg-black/90 pointer-events-none select-none max-w-[1280px] max-h-[720px] aspect-video-vertical md:aspect-video'
      src='/_storybook/iframe.html?viewMode=story&id=new-ui--transactions&shortcuts=false&singleStory=true&args='
    />
  );
}
