import Flex from '@monetr/interface/components/Flex';
import MLink from '@monetr/interface/components/MLink';
import Typography from '@monetr/interface/components/Typography';

export default function LogoutFooter(): JSX.Element {
  return (
    <Flex justify='center' gap='sm'>
      <Typography size='sm' color='subtle' className='text-sm'>
        Not ready to continue?
      </Typography>
      <MLink to='/logout' size='sm'>
        Logout for now
      </MLink>
    </Flex>
  );
}
