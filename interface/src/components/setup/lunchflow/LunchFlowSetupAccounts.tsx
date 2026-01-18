import { useParams } from 'react-router-dom';

export default function LunchFlowSetupAccounts(): React.JSX.Element {
  const { lunchFlowLinkId } = useParams();
  console.log({
    lunchFlowLinkId,
  });

  return <div />;
}
