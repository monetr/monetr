import React from 'react';
import { Avatar } from '@mui/material';

import { useIconSearch } from 'hooks/useIconSearch';
import Transaction from 'models/Transaction';

interface Props {
  transaction: Transaction;
}

export default function TransactionIcon(props: Props): JSX.Element {
  // Try to retrieve the icon. If the icon cannot be retrieved or icons are not currently enabled in the application
  // config then this will simply return null.
  const icon = useIconSearch(props.transaction.name);
  if (icon?.svg) {
    // It is possible for colors to be missing for a given icon. When this happens just fall back to a black color.
    const colorStyles = icon?.colors?.length > 0 ?
      { backgroundColor: `#${ icon.colors[0] }` } :
      { backgroundColor: '#000000' };

    const styles = {
      WebkitMaskImage: `url(data:image/svg+xml;base64,${ icon.svg })`,
      WebkitMaskRepeat: 'no-repeat',
      height: '40px',
      width: '40px',
      ...colorStyles,
    };

    return (
      <div style={ styles } />
    );
  }

  // If we have no icon to work with then create an avatar with the first character of the transaction name.
  const letter = props.transaction.name.toUpperCase().charAt(0);
  return (
    <Avatar>
      { letter }
    </Avatar>
  );
}
