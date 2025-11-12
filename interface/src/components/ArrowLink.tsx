import { ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';

export interface ArrowRedirectProps {
  to: string;
}

export default function ArrowLink(props: ArrowRedirectProps): JSX.Element {
  return (
    <Link
      className='flex-none dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis md:cursor-pointer'
      tabIndex={-1}
      to={props.to}
    >
      <ChevronRight />
    </Link>
  );
}
