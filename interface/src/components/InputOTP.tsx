import React from 'react';
import { OTPInput, OTPInputContext } from 'input-otp';
import { Dot } from 'lucide-react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

const InputOTP = React.forwardRef<React.ElementRef<typeof OTPInput>, React.ComponentPropsWithoutRef<typeof OTPInput>>(
  ({ className, containerClassName, ...props }, ref) => (
    <OTPInput
      className={mergeTailwind('disabled:cursor-not-allowed', className)}
      containerClassName={mergeTailwind('flex items-center gap-2 has-[:disabled]:opacity-50', containerClassName)}
      ref={ref}
      {...props}
    />
  ),
);
InputOTP.displayName = 'InputOTP';

const InputOTPGroup = React.forwardRef<React.ElementRef<'div'>, React.ComponentPropsWithoutRef<'div'>>(
  ({ className, ...props }, ref) => (
    <div className={mergeTailwind('flex items-center', className)} ref={ref} {...props} />
  ),
);
InputOTPGroup.displayName = 'InputOTPGroup';

const InputOTPSlot = React.forwardRef<
  React.ElementRef<'div'>,
  React.ComponentPropsWithoutRef<'div'> & { index: number }
>(({ index, className, ...props }, ref) => {
  const inputOTPContext = React.useContext(OTPInputContext);
  const { char, hasFakeCaret, isActive } = inputOTPContext.slots[index];

  const finalClassName = mergeTailwind(
    'relative flex h-10 w-10 items-center justify-center',
    'border-y border-r border-input text-sm transition-all first:rounded-l-md first:border-l last:rounded-r-md',
    { 'z-10 ring-2 ring-ring ring-offset-background': isActive },
    className,
  );

  return (
    <div className={finalClassName} ref={ref} {...props}>
      {char}
      {hasFakeCaret && (
        <div className='pointer-events-none absolute inset-0 flex items-center justify-center'>
          <div className='h-4 w-px animate-caret-blink bg-foreground duration-1000' />
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
