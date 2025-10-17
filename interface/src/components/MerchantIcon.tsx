

import { Avatar, AvatarFallback } from '@monetr/interface/components/Avatar';
import { useIconSearch } from '@monetr/interface/hooks/useIconSearch';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';

export interface MerchantIconProps {
  name?: string;
  className?: string;
}

export default function MerchantIcon(props: MerchantIconProps): JSX.Element {
  const icon = useIconSearch(props?.name);
  const size = 30;
  if (icon?.svg) {
    const classNames = mergeTailwind(
      'dark:bg-dark-monetr-background-bright',
      'flex',
      'items-center',
      'justify-center',
      'h-10',
      'w-10',
      'rounded-full',
      'flex-none',
      props.className,
    );

    // It is possible for colors to be missing for a given icon. When this happens just fall back to a black color.
    const colorStyles =
      icon?.colors?.length > 0 ? { backgroundColor: `#${icon.colors[0]}` } : { backgroundColor: '#000000' };

    const styles = {
      // TODO Add mask image things for other browsers.
      WebkitMaskImage: `url(data:image/svg+xml;base64,${icon.svg})`,
      WebkitMaskRepeat: 'no-repeat',
      height: `${size}px`,
      width: `${size}px`,
      ...colorStyles,
    };

    return (
      <div className={classNames}>
        <div style={styles} />
      </div>
    );
  }

  const classNames = mergeTailwind('h-10 w-10', props.className);

  // If we have no icon to work with then create an avatar with the first character of the transaction name.
  const letter = props?.name?.toUpperCase().charAt(0) || '?';
  return (
    <Avatar className={classNames}>
      <AvatarFallback className='dark:bg-dark-monetr-background-subtle dark:text-dark-monetr-content'>
        {letter}
      </AvatarFallback>
    </Avatar>
  );
}
