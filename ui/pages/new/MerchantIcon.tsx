import React from 'react';
import { Avatar } from '@mui/material';

import { useIconSearch } from 'hooks/useIconSearch';

export interface MerchantIconProps {
  name?: string;
  size?: number; // TODO this doesn't really work.
}

export default function MerchantIcon(props: MerchantIconProps): JSX.Element {
  const icon = useIconSearch(props?.name);
  const size = props?.size || 30;
  if (icon?.svg) {
    // It is possible for colors to be missing for a given icon. When this happens just fall back to a black color.
    const colorStyles = icon?.colors?.length > 0 ?
      { backgroundColor: `#${icon.colors[0]}` } :
      { backgroundColor: '#000000' };

    const styles = {
      WebkitMaskImage: `url(data:image/svg+xml;base64,${icon.svg})`,
      WebkitMaskRepeat: 'no-repeat',
      height: `${size}px`,
      width: `${size}px`,
      ...colorStyles,
    };

    return (
      <div className='bg-white flex items-center justify-center h-10 w-10 rounded-full'>
        <div style={ styles } />
      </div>
    );
  }

  // If we have no icon to work with then create an avatar with the first character of the transaction name.
  const letter = props?.name?.toUpperCase().charAt(0) || '?';
  return (
    <Avatar className='bg-zinc-800 h-10 w-10 text-zinc-200'>
      {letter}
    </Avatar>
  );
}
