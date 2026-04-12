import { normalizeImagePath } from '@rspress/core/runtime';

export interface BlogMemeProps {
  alt: string;
  src: string;
}

export default function BlogMeme(props: BlogMemeProps): React.JSX.Element {
  return (
    <div className='flex justify-center'>
      <div className='w-2/3 md:w-1/2 lg:w-1/4'>
        <img alt={props.alt} className='medium-zoom-image' src={normalizeImagePath(props.src)} />
      </div>
    </div>
  );
}
