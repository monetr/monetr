import { Fragment, useCallback, useId, useState } from 'react';
import { Copy } from 'lucide-react';

import { Button } from '@monetr/interface/components/Button';
import ErrorText from '@monetr/interface/components/ErrorText';
import Label, { type LabelDecorator } from '@monetr/interface/components/Label';
import { Skeleton } from '@monetr/interface/components/Skeleton';
import mergeClasses from '@monetr/interface/util/mergeClasses';
import { useSnackbar } from '@monetr/notify';

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
    // Even if the prop is set, if the browser doesn't have the capabilty for some reason then don't even show the
    // button.
    copy: props.copy && typeof navigator.clipboard?.writeText === 'function' ? props.copy : undefined,
  };
  const { labelDecorator, icon, className, ...otherProps } = props;
  const LabelDecorator = labelDecorator ?? (() => null);
  const Icon = icon ?? (() => null);
  const [copied, setCopied] = useState(false);
  const { enqueueSnackbar } = useSnackbar();

  const onCopy = useCallback(async () => {
    if (props.copy) {
      return await navigator.clipboard
        .writeText(props.copy)
        .then(() => setCopied(true))
        .catch(() =>
          enqueueSnackbar('Failed to copy to the clipboard.', {
            variant: 'error',
            disableWindowBlurListener: true,
          }),
        );
    }
  }, [enqueueSnackbar, props.copy]);

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
            <Button aria-label='Copy' onClick={onCopy} variant={copied ? 'primary' : 'outlined'}>
              {copied ? <Fragment>Copied!</Fragment> : <Copy />}
            </Button>
          )}
        </div>
      </div>
      <ErrorText />
    </div>
  );
}
