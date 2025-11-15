import Badge from '@monetr/interface/components/Badge';
import Typography from '@monetr/interface/components/Typography';
import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import { Link, useLocation } from 'react-router-dom';

export interface BudgetingSidebarItemProps {
  to: string;
  icon: React.FC<{ className?: string }>;
  children: string;
  badge?: string;
}

export default function BudgetingSidebarItem(props: BudgetingSidebarItemProps): React.JSX.Element {
  const Icon = props.icon;
  const location = useLocation();
  const active = location.pathname.endsWith(props.to.replaceAll('.', ''));

  const className = mergeTailwind(
    {
      'dark:bg-dark-monetr-background-emphasis': active,
      'dark:text-dark-monetr-content-emphasis': active,
      'dark:text-dark-monetr-content-subtle': !active,
    },
    [
      'align-middle',
      'cursor-pointer',
      'flex',
      'text-lg',
      'gap-2',
      'dark:hover:bg-dark-monetr-background-emphasis',
      'dark:hover:text-dark-monetr-content-emphasis',
      'items-center',
      'px-2',
      'py-1',
      'rounded-md',
      'w-full',
    ],
  );

  return (
    <Link className={className} to={props.to}>
      <Icon />
      <Typography color={active ? 'emphasis' : 'subtle'} ellipsis size='lg' weight={active ? 'semibold' : 'medium'}>
        {props.children}
      </Typography>
      {props.badge && (
        <Badge className='ml-auto' size='sm'>
          {props.badge}
        </Badge>
      )}
    </Link>
  );
}
