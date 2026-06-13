import { Fragment } from 'react';

import Divider from '@monetr/interface/components/Divider';
import Typography from '@monetr/interface/components/Typography';
import mergeClasses from '@monetr/interface/util/mergeClasses';

import styles from './MStepper.module.scss';

export interface MStepperProps {
  steps: Array<string>;
  activeIndex: number;
}

export default function MStepper(props: MStepperProps): React.JSX.Element {
  const { steps, activeIndex } = props;

  const states = [MStepState.Inactive, MStepState.Active, MStepState.Complete];

  const items = steps.map((name, index) => {
    // There will only be a single truthy value in this array.
    // That truthy value's index corresponds to the index of states above.
    // Find the truthy index, and you have the state of the item.
    const state =
      states[
        [
          index > activeIndex, // Not there yet.
          index === activeIndex, // Active.
          index < activeIndex, // Completed
        ].indexOf(true)
      ] ?? MStepState.Inactive;

    // We will show the divider when we are not the last item.
    const divider = index !== steps.length - 1;

    return <MStep currentIndex={activeIndex} divider={divider} index={index} key={name} name={name} state={state} />;
  });

  return <div className={styles.stepper}>{items}</div>;
}

enum MStepState {
  Inactive,
  Active,
  Complete,
}

interface MStepProps {
  state: MStepState;
  name: string;
  currentIndex: number;
  index: number;
  divider?: boolean;
}

function MStep(props: MStepProps): React.JSX.Element {
  const lineClass = mergeClasses(
    {
      [MStepState.Inactive]: styles.lineDashed,
      [MStepState.Active]: styles.lineDashed,
      [MStepState.Complete]: styles.lineComplete,
    }[props.state],
    styles.line,
  );

  const numberClass = mergeClasses(
    {
      [MStepState.Inactive]: styles.numberInactive,
      [MStepState.Active]: styles.numberActive,
      [MStepState.Complete]: styles.numberComplete,
    }[props.state],
    styles.number,
  );

  const textClass = mergeClasses(
    // On smaller screens hide the text for items that are not the current step or not the next step. There isn't a
    // stable base class on this element to hang a data attribute off of, so we keep the conditional class here.
    ![props.currentIndex, props.currentIndex + 1].includes(props.index) && styles.textHidden,
    {
      [MStepState.Inactive]: undefined,
      [MStepState.Active]: styles.textActive,
      [MStepState.Complete]: styles.textComplete,
    }[props.state],
  );

  return (
    <Fragment>
      <Typography className={styles.row}>
        <Typography className={numberClass}>{props.index + 1}</Typography>
        <Typography className={textClass}>{props.name}</Typography>
      </Typography>
      {props.divider && <Divider className={lineClass} />}
    </Fragment>
  );
}
