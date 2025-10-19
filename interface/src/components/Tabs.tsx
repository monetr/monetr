import * as React from 'react';
import * as TabsPrimitive from '@radix-ui/react-tabs';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

const Tabs = TabsPrimitive.Root;

const TabsList = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.List>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.List
    ref={ref}
    className={mergeTailwind(
      'inline-flex justify-center items-center',
      'rounded-md p-1 h-10',
      'text-dark-monetr-content-subtle',
      'bg-dark-monetr-background-subtle',
      className,
    )}
    {...props}
  />
));
TabsList.displayName = TabsPrimitive.List.displayName;

const TabsTrigger = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Trigger>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Trigger>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Trigger
    ref={ref}
    className={mergeTailwind(
      'inline-flex items-center justify-center',
      'whitespace-nowrap',
      'rounded-md px-3 py-1.5',
      'text-sm font-medium',
      'ring-offset-background transition-all',
      // Focused
      'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
      // Disabled
      'disabled:pointer-events-none disabled:opacity-50',
      // Active state
      'data-[state=active]:bg-dark-monetr-background data-[state=active]:text-dark-monetr-content-emphasis',
      'data-[state=active]:shadow-sm',
      className,
    )}
    {...props}
  />
));
TabsTrigger.displayName = TabsPrimitive.Trigger.displayName;

const TabsContent = React.forwardRef<
  React.ElementRef<typeof TabsPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof TabsPrimitive.Content>
>(({ className, ...props }, ref) => (
  <TabsPrimitive.Content
    ref={ref}
    className={mergeTailwind(
      'mt-2',
      'ring-offset-background',
      // TODO Ring probably doesn't properly work here, need to redo it.
      'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2',
      className,
    )}
    {...props}
  />
));
TabsContent.displayName = TabsPrimitive.Content.displayName;

export { Tabs, TabsContent, TabsList, TabsTrigger };
