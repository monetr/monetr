import { ChevronRight } from 'lucide-react';

import Badge from '@monetr/interface/components/Badge';
import type { SelectAction } from '@monetr/interface/components/Select';
import Typography from '@monetr/interface/components/Typography';
import useIsMobile from '@monetr/interface/hooks/useIsMobile';

import styles from './SelectActionOption.module.scss';

export interface SelectActionOptionProps {
  action: SelectAction;
}

export default function SelectActionOption(props: SelectActionOptionProps): React.JSX.Element {
  const isMobile = useIsMobile();
  const Icon = props.action.icon;
  return (
    <div className={styles.actionOptionRow}>
      <Badge className={styles.actionOptionIcon} size='sm' variant='brand'>
        <Icon />
      </Badge>
      <div className={styles.actionOptionText}>
        <Typography color='emphasis' ellipsis size='md' weight='semibold'>
          {props.action.label}
        </Typography>
        {props.action.description && (
          <Typography color='subtle' ellipsis size='sm'>
            {props.action.description}
          </Typography>
        )}
      </div>
      {/* In the drawer this takes you somewhere else instead of picking something, the chevron hints at that */}
      {isMobile && <ChevronRight className={styles.actionOptionChevron} />}
    </div>
  );
}
