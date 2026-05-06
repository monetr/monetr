import * as React from 'react';

import mergeTailwind from '@monetr/interface/util/mergeTailwind';
import { Drawer as DrawerPrimitive } from '@monetr/vaul';

import styles from './Drawer.module.scss';

// iOS Safari/WebView ignores `overflow: hidden` on the body. The correct workaround is `position: fixed` with a
// negative `top` equal to the current scroll offset. We opt out of vaul's own body style handling (noBodyStyles) and do
// it ourselves so we control timing.
function useIosScrollLock() {
  const scrollY = React.useRef(0);

  const lock = React.useCallback(() => {
    if (document.body.style.position === 'fixed') {
      return;
    }
    scrollY.current = window.scrollY;
    document.body.style.position = 'fixed';
    document.body.style.top = `-${scrollY.current}px`;
    document.body.style.width = '100%';
  }, []);

  const unlock = React.useCallback(() => {
    if (document.body.style.position !== 'fixed') {
      return;
    }
    document.body.style.position = '';
    document.body.style.top = '';
    document.body.style.width = '';
    window.scrollTo(0, scrollY.current);
  }, []);

  return { lock, unlock };
}

const Drawer = ({
  shouldScaleBackground = false,
  onOpenChange,
  open,
  ...props
}: React.ComponentProps<typeof DrawerPrimitive.Root>) => {
  const { lock, unlock } = useIosScrollLock();

  // When the drawer is used as a controlled component, the parent can change `open` directly (e.g. closing after
  // selecting a value). In that case vaul does NOT fire onOpenChange — doing so would be circular — so we watch the
  // prop ourselves to catch that path.
  React.useEffect(() => {
    if (open === true) {
      lock();
    } else if (open === false) {
      unlock();
    }
  }, [open, lock, unlock]);

  const handleOpenChange = React.useCallback(
    (newOpen: boolean) => {
      if (newOpen) {
        lock();
      } else {
        unlock();
      }
      onOpenChange?.(newOpen);
    },
    [lock, unlock, onOpenChange],
  );

  return (
    <DrawerPrimitive.Root
      noBodyStyles
      onOpenChange={handleOpenChange}
      open={open}
      shouldScaleBackground={shouldScaleBackground}
      {...props}
    />
  );
};
Drawer.displayName = 'Drawer';

const DrawerTrigger = DrawerPrimitive.Trigger;

const DrawerPortal = DrawerPrimitive.Portal;

const DrawerClose = DrawerPrimitive.Close;

const DrawerOverlay = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Overlay>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Overlay>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Overlay className={mergeTailwind(styles.drawerOverlay, className)} ref={ref} {...props} />
));
DrawerOverlay.displayName = DrawerPrimitive.Overlay.displayName;

const DrawerContent = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Content>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Content>
>(({ className, children, ...props }, ref) => (
  <DrawerPortal>
    <DrawerOverlay />
    <DrawerPrimitive.Content className={mergeTailwind(styles.drawerContent, className)} ref={ref} {...props}>
      <div className={styles.drawerHandle} />
      {children}
    </DrawerPrimitive.Content>
  </DrawerPortal>
));
DrawerContent.displayName = 'DrawerContent';

// Wrapper component goes after the header and before the footer in the content and makes the drawer scrollable
// properly.
const DrawerWrapper = React.forwardRef<HTMLDivElement, React.ButtonHTMLAttributes<HTMLDivElement>>(
  ({ className, children, ...props }, ref) => (
    <div className={mergeTailwind(styles.drawerWrapper, className)} ref={ref} {...props}>
      {children}
    </div>
  ),
);
DrawerWrapper.displayName = 'DrawerWrapper';

const DrawerHeader = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={mergeTailwind(styles.drawerHeader, className)} {...props} />
);
DrawerHeader.displayName = 'DrawerHeader';

const DrawerFooter = ({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) => (
  <div className={mergeTailwind(styles.drawerFooter, className)} {...props} />
);
DrawerFooter.displayName = 'DrawerFooter';

const DrawerTitle = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Title>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Title>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Title className={mergeTailwind(styles.drawerTitle, className)} ref={ref} {...props} />
));
DrawerTitle.displayName = DrawerPrimitive.Title.displayName;

const DrawerDescription = React.forwardRef<
  React.ElementRef<typeof DrawerPrimitive.Description>,
  React.ComponentPropsWithoutRef<typeof DrawerPrimitive.Description>
>(({ className, ...props }, ref) => (
  <DrawerPrimitive.Description className={mergeTailwind(styles.drawerDescription, className)} ref={ref} {...props} />
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
