import React from 'react';
import type { DialogProps } from '@radix-ui/react-dialog';
import { Command as CommandPrimitive } from 'cmdk';
import { Search } from 'lucide-react';

import { Dialog, DialogContent } from '@monetr/interface/components/Dialog';
import mergeClasses from '@monetr/interface/util/mergeClasses';
import type { ExtractProps } from '@monetr/interface/util/typescriptEvils';

import styles from './Command.module.scss';

const Command = React.forwardRef<
  React.ElementRef<typeof CommandPrimitive>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive>
>(({ className, ...props }, ref) => (
  <CommandPrimitive className={mergeClasses(styles.command, className)} ref={ref} {...props} />
));
Command.displayName = CommandPrimitive.displayName;

interface CommandDialogProps extends DialogProps {}

const CommandDialog = ({ children, ...props }: CommandDialogProps) => {
  return (
    <Dialog {...props}>
      <DialogContent className={styles.dialogContentOverride}>
        <Command className={styles.commandDialog}>{children}</Command>
      </DialogContent>
    </Dialog>
  );
};

interface CommandInputProps extends ExtractProps<typeof CommandPrimitive.Input> {
  wrapperClassName?: string;
}

const CommandInput = React.forwardRef<React.ElementRef<typeof CommandPrimitive.Input>, CommandInputProps>(
  ({ className, wrapperClassName, ...props }, ref) => (
    <div className={mergeClasses(styles.inputWrapper, wrapperClassName)} cmdk-input-wrapper=''>
      <Search className={styles.inputIcon} />
      <CommandPrimitive.Input className={mergeClasses(styles.input, className)} ref={ref} {...props} />
    </div>
  ),
);

CommandInput.displayName = CommandPrimitive.Input.displayName;

const CommandList = React.forwardRef<
  React.ElementRef<typeof CommandPrimitive.List>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.List>
>(({ className, ...props }, ref) => (
  <CommandPrimitive.List className={mergeClasses(styles.list, className)} ref={ref} {...props} />
));

CommandList.displayName = CommandPrimitive.List.displayName;

const CommandEmpty = React.forwardRef<
  React.ElementRef<typeof CommandPrimitive.Empty>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.Empty>
>((props, ref) => <CommandPrimitive.Empty className={styles.empty} ref={ref} {...props} />);

CommandEmpty.displayName = CommandPrimitive.Empty.displayName;

const CommandGroup = React.forwardRef<
  React.ElementRef<typeof CommandPrimitive.Group>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.Group>
>(({ className, ...props }, ref) => (
  <CommandPrimitive.Group className={mergeClasses(styles.group, className)} ref={ref} {...props} />
));

CommandGroup.displayName = CommandPrimitive.Group.displayName;

const CommandSeparator = React.forwardRef<
  React.ElementRef<typeof CommandPrimitive.Separator>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.Separator>
>(({ className, ...props }, ref) => (
  <CommandPrimitive.Separator className={mergeClasses(styles.separator, className)} ref={ref} {...props} />
));
CommandSeparator.displayName = CommandPrimitive.Separator.displayName;

const CommandItem = React.forwardRef<
  React.ElementRef<typeof CommandPrimitive.Item>,
  React.ComponentPropsWithoutRef<typeof CommandPrimitive.Item>
>(({ className, ...props }, ref) => (
  <CommandPrimitive.Item className={mergeClasses(styles.item, className)} ref={ref} {...props} />
));

CommandItem.displayName = CommandPrimitive.Item.displayName;

const CommandShortcut = ({ className, ...props }: React.HTMLAttributes<HTMLSpanElement>) => {
  return <span className={mergeClasses(styles.shortcut, className)} {...props} />;
};
CommandShortcut.displayName = 'CommandShortcut';

export {
  Command,
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
  CommandShortcut,
};
