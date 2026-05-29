import * as React from 'react';
import * as TooltipPrimitive from '@radix-ui/react-tooltip';

import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './Tooltip.module.scss';

const TooltipProvider = TooltipPrimitive.Provider;

const Tooltip = TooltipPrimitive.Root;

const TooltipTrigger = TooltipPrimitive.Trigger;

const TooltipContent = React.forwardRef<
  React.ElementRef<typeof TooltipPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TooltipPrimitive.Content>
>(({ className, sideOffset = 4, ...props }, ref) => (
  <TooltipPrimitive.Content
    className={mergeClasses(styles.tooltipContent, className)}
    ref={ref}
    sideOffset={sideOffset}
    {...props}
  />
));
TooltipContent.displayName = TooltipPrimitive.Content.displayName;

export { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger };
