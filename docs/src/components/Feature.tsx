import React from 'react';
import Image, { StaticImageData } from 'next/image';

import { twMerge } from 'tailwind-merge';

interface FeatureProps {
  title: React.ReactNode;
  description?: React.ReactNode;
  className?: string;
  image?: StaticImageData;
}

export default function Feature(props: FeatureProps): JSX.Element {
  const className = twMerge('rounded-xl bg-dark-monetr-background shadow-lg col-span-2', props.className);

  return (
    <div className={ className }>
      { props.title }
      { props.description && props.description }
      { props.image && 
        <Image
          src={ props.image }
          className=''
          alt='TODO'
        />
      }
    </div>
  );
}
