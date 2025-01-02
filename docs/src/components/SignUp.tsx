import { ArrowRight } from 'lucide-react';
import Link from 'next/link';

export default function SignUp(): JSX.Element {
  return (
    <Link
      className='btn-sm hidden sm:block text-slate-300 hover:text-white transition duration-150 ease-in-out group relative'
      href='https://my.monetr.app/register'
    >
      <span className='relative inline-flex items-center'>
        Sign Up
        <ArrowRight className='h-5 w-5' />
      </span>
    </Link>
  );
}
