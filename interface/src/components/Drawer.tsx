import * as React from 'react';
import { Drawer as DrawerPrimitive } from 'vaul';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';

import styles from './Drawer.module.scss';

const Drawer = ({ shouldScaleBackground = true, ...props }: React.ComponentProps<typeof DrawerPrimitive.Root>) => (
  <DrawerPrimitive.Root shouldScaleBackground={shouldScaleBackground} {...props} />
);
Drawer.displayName = 'Drawer';

const DrawerTrigger = DrawerPrimitive.Trigger;

const DrawerPortal = DrawerPrimitive.Portal;

const DrawerClose = DrawerPrimitive.Close;

const DrawerOverlay = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Overlay
    className={mergeTailwind('fixed inset-0 z-50 bg-black/80', className)}
    ref={ref}
    {...props}
  />
));
DrawerOverlay.displayName = DrawerPrimitive.Overlay.displayName;

const DrawerContent = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <DrawerPortal>
    <DrawerOverlay />
    <DrawerPrimitive.Content
      className={mergeTailwind(
        'fixed',
        'inset-x-0',
        'z-50',
        'flex flex-col',
        'rounded-t-[10px]',
        'bottom-0 left-0 right-0',
        'border border-b-0 border-dark-monetr-border bg-dark-monetr-background',
        'max-h-[60%]',
        'gap-4',
        styles.drawerContent,
        className,
      )}
      ref={ref}
      {...props}
    >
      <div className='mx-auto mt-4 h-2 min-h-2 w-[100px] rounded-full bg-dark-monetr-content-muted' />
      {children}
    </DrawerPrimitive.Content>
  </DrawerPortal>
));
DrawerContent.displayName = 'DrawerContent';

// Wrapper component goes after the header and before the footer in the content and makes the drawer scrollable
// properly.
const DrawerWrapper = React.forwardRef<HTMLDivElement, React.ButtonHTMLAttributes<HTMLDivElement>>(
  ({ className, children, ...props }, ref) => (
    <div className={mergeTailwind('flex-shrink overflow-y-auto', className)} ref={ref} {...props}>
      {children}
    </div>
  ),
);
DrawerWrapper.displayName = 'DrawerWrapper';

const DrawerHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={mergeTailwind('grid gap-1.5 p-4 text-center sm:text-left', className)} {...props} />
);
DrawerHeader.displayName = 'DrawerHeader';

const DrawerFooter = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={mergeTailwind('mt-auto flex flex-col gap-2 p-4', className)} {...props} />
);
DrawerFooter.displayName = 'DrawerFooter';

const DrawerTitle = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Title
    className={mergeTailwind('text-lg font-semibold leading-none tracking-tight', className)}
    ref={ref}
    {...props}
  />
));
DrawerTitle.displayName = DrawerPrimitive.Title.displayName;

const DrawerDescription = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Description
    className={mergeTailwind('text-sm text-muted-foreground', className)}
    ref={ref}
    {...props}
  />
));
DrawerDescription.displayName = DrawerPrimitive.Description.displayName;

export {
  Drawer,
  DrawerClose,
  DrawerContent,
  DrawerDescription,
  DrawerFooter,
  DrawerHeader,
  DrawerOverlay,
  DrawerPortal,
  DrawerTitle,
  DrawerTrigger,
  DrawerWrapper,
};
