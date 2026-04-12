import styles from './BlogMeme.module.scss';

import { normalizeImagePath } from '@rspress/core/runtime';

export interface BlogMemeProps {
  alt: string;
  src: string;
}

export default function BlogMeme(props: BlogMemeProps): React.JSX.Element {
  return (
    <div className={styles.root}>
      <div className={styles.inner}>
        <img alt={props.alt} className='medium-zoom-image' src={normalizeImagePath(props.src)} />
      </div>
    </div>
  );
}
