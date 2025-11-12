import Flex from '@monetr/interface/components/Flex';
import MLink from '@monetr/interface/components/MLink';
import Typography from '@monetr/interface/components/Typography';

export default function LogoutFooter(): JSX.Element {
  return (
    <Flex gap='sm' justify='center'>
      <Typography className='text-sm' color='subtle' size='sm'>
        Not ready to continue?
      </Typography>
      <MLink size='sm' to='/logout'>
        Logout for now
      </MLink>
    </Flex>
  );
}
