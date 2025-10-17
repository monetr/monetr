

import MLink from '@monetr/interface/components/MLink';
import MSpan from '@monetr/interface/components/MSpan';

export default function LogoutFooter(): JSX.Element {
  return (
    <div className='flex justify-center gap-1'>
      <MSpan color='subtle' className='text-sm'>
        Not ready to continue?
      </MSpan>
      <MLink to='/logout' size='sm'>
        Logout for now
      </MLink>
    </div>
  );
}
