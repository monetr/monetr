import React from 'react';
import { useQuery } from 'react-query';
import { Avatar } from '@mui/material';

import Transaction from 'models/Transaction';

interface Props {
  transaction: Transaction;
}

interface IconResponse {
  svg: string;
  colors: Array<string>;
}

export default function TransactionIcon(props: Props): JSX.Element {
  const letter = props.transaction.name.toUpperCase().charAt(0);

  const { data } = useQuery<IconResponse>(`/api/icons/search?name=${ props.transaction.name }`);

  if (data?.svg) {
    const styles = {
      webkitMaskImage: `url(data:image/svg+xml;base64,${data.svg})`,
      webkitMaskRepeat: 'no-repeat',
      height: '40px',
      width: '40px',
      ...(data.colors.length > 0 && { backgroundColor: `#${data.colors[0]}` }),
    };

    return (
      <div style={ styles } />
    );
  }

  return (
    <Avatar>
      { letter }
    </Avatar>
  );
}
