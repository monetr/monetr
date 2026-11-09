import { flexVariants } from '@monetr/interface/components/Flex';
import Typography from '@monetr/interface/components/Typography';
import type { SimpleFinLinkError } from '@monetr/interface/models/SimpleFinLink';

export interface SimpleFinLinkErrorsProps {
  errors: Array<SimpleFinLinkError>;
}

export default function SimpleFinLinkErrors(props: SimpleFinLinkErrorsProps): React.JSX.Element | null {
  if (props.errors.length === 0) {
    return null;
  }

  return (
    <ul className={flexVariants({ orientation: 'column', gap: 'sm' })} data-testid='simplefin-link-errors'>
      {props.errors.map(item => (
        <li key={`${item.code}-${item.connectionId}-${item.simpleFinId}-${item.message}`}>
          <Typography color='negative' size='sm'>
            {item.message}
          </Typography>
        </li>
      ))}
    </ul>
  );
}
