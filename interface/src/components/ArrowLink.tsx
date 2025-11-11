import { ChevronRight } from 'lucide-react';
import { Link } from 'react-router-dom';

export interface ArrowRedirectProps {
  to: string;
}

export default function ArrowLink(props: ArrowRedirectProps): JSX.Element {
  return (
    <Link
      to={props.to}
      tabIndex={-1}
      className='flex-none dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis md:cursor-pointer'
    >
      <ChevronRight />
    </Link>
  );
}
