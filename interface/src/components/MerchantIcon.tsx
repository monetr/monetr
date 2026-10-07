import type { VariantProps } from 'class-variance-authority';

import { Avatar, AvatarFallback, avatarVariants } from '@monetr/interface/components/Avatar';
import { useIconSearch } from '@monetr/interface/hooks/useIconSearch';
import avatarLetter from '@monetr/interface/util/avatarLetter';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import merchantIconStyles from './MerchantIcon.module.scss';

export interface MerchantIconProps extends VariantProps<typeof avatarVariants> {
  name?: string;
  className?: string;
}

export default function MerchantIcon(props: MerchantIconProps): React.JSX.Element {
  const icon = useIconSearch(props?.name ?? '');
  // The icon takes up 75% of the avatar no matter how big the avatar is, which is 30px in a 40px avatar
  const size = '75%';
  if (icon?.svg) {
    // It is possible for colors to be missing for a given icon. When this happens just fall back to a black color.
    const colorStyles =
      icon?.colors?.length > 0 ? { backgroundColor: `#${icon.colors[0]}` } : { backgroundColor: '#000000' };

    const styles = {
      // TODO Add mask image things for other browsers.
      WebkitMaskImage: `url(data:image/svg+xml;base64,${icon.svg})`,
      WebkitMaskRepeat: 'no-repeat',
      height: size,
      width: size,
      ...colorStyles,
    };

    return (
      <div
        className={mergeClasses(avatarVariants({ size: props.size }), merchantIconStyles.merchantIcon, props.className)}
      >
        <div style={styles} />
      </div>
    );
  }

  // If we have no icon to work with then create an avatar with the first character of the transaction name.
  const letter = avatarLetter(props?.name);
  return (
    <Avatar className={props.className} size={props.size}>
      <AvatarFallback>{letter}</AvatarFallback>
    </Avatar>
  );
}
