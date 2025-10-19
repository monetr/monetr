import { KeyboardArrowRight } from '@mui/icons-material';
import { Link } from 'react-router-dom';

export interface ArrowRedirectProps {
  to: string;
}

export default function ArrowLink(props: ArrowRedirectProps): JSX.Element {
  return (
    <Link
      to={props.to}
      className='flex-none dark:text-dark-monetr-content-subtle dark:group-hover:text-dark-monetr-content-emphasis md:cursor-pointer'
    >
      <KeyboardArrowRight />
    </Link>
  );
}
