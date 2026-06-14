import React from 'react';
import { OTPInput, OTPInputContext } from 'input-otp';
import { Dot } from 'lucide-react';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './InputOTP.module.scss';

const InputOTP = React.forwardRef<React.ElementRef<typeof OTPInput>, React.ComponentPropsWithoutRef<typeof OTPInput>>(
  ({ className, containerClassName, ...props }, ref) => (
    <OTPInput
      className={mergeClasses(styles.input, className)}
      containerClassName={mergeClasses(styles.container, containerClassName)}
      ref={ref}
      {...props}
    />
  ),
);
InputOTP.displayName = 'InputOTP';

const InputOTPGroup = React.forwardRef<React.ElementRef<'div'>, React.ComponentPropsWithoutRef<'div'>>(
  ({ className, ...props }, ref) => <div className={mergeClasses(styles.group, className)} ref={ref} {...props} />,
);
InputOTPGroup.displayName = 'InputOTPGroup';

const InputOTPSlot = React.forwardRef<
  React.ElementRef<'div'>,
  React.ComponentPropsWithoutRef<'div'> & { index: number }
>(({ index, className, ...props }, ref) => {
  const inputOTPContext = React.useContext(OTPInputContext);
  // slots is indexed by a prop so noUncheckedIndexedAccess treats it as possibly undefined, fall back to an empty
  // inactive slot if we ever get handed an index that is out of range.
  const { char, hasFakeCaret, isActive } = inputOTPContext.slots[index] ?? {
    char: null,
    placeholderChar: null,
    hasFakeCaret: false,
    isActive: false,
  };

  const finalClassName = mergeClasses(styles.slot, className);

  return (
    <div className={finalClassName} data-active={isActive} ref={ref} {...props}>
      {char}
      {hasFakeCaret && (
        <div className={styles.caretWrapper}>
          <div className={styles.caret} />
        </div>
      )}
    </div>
  );
});
InputOTPSlot.displayName = 'InputOTPSlot';

const InputOTPSeparator = React.forwardRef<React.ElementRef<'div'>, React.ComponentPropsWithoutRef<'div'>>(
  ({ ...props }, ref) => (
    <div ref={ref} {...props}>
      <Dot />
    </div>
  ),
);
InputOTPSeparator.displayName = 'InputOTPSeparator';

export { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot };
