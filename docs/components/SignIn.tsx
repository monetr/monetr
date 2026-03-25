import { ArrowRight } from 'lucide-react';

import { Link } from '@rspress/core/theme-original';

export default function SignIn(): JSX.Element {
  return (
    <Link
      className='btn-sm hidden sm:block text-slate-300 hover:text-white transition duration-150 ease-in-out group relative no-underline'
      href='https://my.monetr.app/'
    >
      <span className='relative inline-flex items-center text-nowrap'>
        Sign In
        <ArrowRight className='size-5' />
      </span>
    </Link>
  );
}
