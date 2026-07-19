import { useId } from 'react';
import { Copy } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import ErrorText from '@monetr/interface/components/ErrorText';
import Label, { type LabelDecorator } from '@monetr/interface/components/Label';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Code.module.scss';
import errorTextStyles from './ErrorText.module.scss';
import inputStyles from './FormTextField.module.scss';
import selectStyles from './Select.module.scss';

export interface CodeProps extends React.ComponentPropsWithoutRef<'code'> {
  label?: string;
  labelDecorator?: LabelDecorator;
  isLoading?: boolean;
  icon?: React.FC;
  copy?: string;
}

export default function Code(props: CodeProps): React.JSX.Element {
  const id = useId();
  props = {
    id,
    ...props,
  };
  const { labelDecorator, icon, className, ...otherProps } = props;
  const LabelDecorator = labelDecorator ?? (() => null);
  const Icon = icon ?? (() => null);

  if (props.isLoading) {
    return (
      <div className={mergeClasses(errorTextStyles.errorTextPadding, props.className)}>
        <Label disabled htmlFor={props.id} label={props.label}>
          <LabelDecorator disabled />
        </Label>
        <div>
          <div aria-disabled='true' className={mergeClasses(inputStyles.input, selectStyles.selectLoading)}>
            <Skeleton className={selectStyles.loadingSkeleton} />
          </div>
        </div>
        <ErrorText />
      </div>
    );
  }

  return (
    <div className={mergeClasses(errorTextStyles.errorTextPadding, props.className)}>
      <Label htmlFor={props.id} label={props.label}>
        <LabelDecorator />
      </Label>
      <div>
        <div className={mergeClasses(inputStyles.input, styles.root)}>
          <Icon />
          <code className={mergeClasses(styles.code, className)} {...otherProps}>
            {props.children}
          </code>
          {Boolean(props.copy) && (
            <Button aria-label='Copy' variant='outlined'>
              <Copy />
            </Button>
          )}
        </div>
      </div>
      <ErrorText />
    </div>
  );
}
