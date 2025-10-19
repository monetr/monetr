import MerchantIcon, { type MerchantIconProps } from '@monetr/interface/components/MerchantIcon';

export interface TransactionMerchantIconProps extends MerchantIconProps {
  pending?: boolean;
}

export default function TransactionMerchantIcon(props: TransactionMerchantIconProps): JSX.Element {
  const { pending, ...merchantIconProps } = props;

  if (pending) {
    return (
      <div className='relative'>
        <MerchantIcon {...merchantIconProps} />
        <span className='absolute flex h-3 w-3 right-0 bottom-0'>
          <span className='animate-ping-slow absolute inline-flex h-full w-full rounded-full bg-blue-400' />
          <span className='relative inline-flex rounded-full h-3 w-3 bg-blue-500' />
        </span>
      </div>
    );
  }

  return <MerchantIcon {...merchantIconProps} />;
}
