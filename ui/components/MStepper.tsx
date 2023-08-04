import React, { Fragment } from 'react';

import MDivider from './MDivider';
import MSpan from './MSpan';

import mergeTailwind from 'util/mergeTailwind';

export interface MStepperProps {
  steps: Array<string>;
  activeIndex: number;
}

export default function MStepper(props: MStepperProps): JSX.Element {
  const { steps, activeIndex } = props;

  const states = [
    MStepState.Inactive,
    MStepState.Active,
    MStepState.Complete,
  ];

  const items = steps.map((name, index) => {
    // There will only be a single truthy value in this array.
    // That truthy value's index corresponds to the index of states above.
    // Find the truthy index, and you have the state of the item.
    const state = states[[
      index > activeIndex, // Not there yet.
      index === activeIndex, // Active.
      index < activeIndex, // Completed
    ].indexOf(true)];

    // We will show the divider when we are not the last item.
    const divider = index != steps.length - 1;

    return (
      <MStep
        key={ index }
        state={ state }
        name={ name }
        currentIndex={ activeIndex }
        index={ index }
        divider={ divider }
      />
    );
  });

  return (
    <div className='w-full flex gap-2 items-center bg-dark-monetr-background-focused p-2 rounded-xl'>
      { items }
    </div>
  );
}

enum MStepState {
  Inactive,
  Active,
  Complete,
}

interface MStepProps {
  state: MStepState,
  name: string;
  currentIndex: number;
  index: number;
  divider?: boolean;
}

function MStep(props: MStepProps): JSX.Element {
  const lineClass = mergeTailwind(
    {
      [MStepState.Inactive]: [
        'dark:border-dark-monetr-background-emphasis',
        'border-dashed',
      ],
      [MStepState.Active]: [
        'dark:border-dark-monetr-background-emphasis',
        'border-dashed',
      ],
      [MStepState.Complete]: [
        'dark:border-green-600',
      ],
    }[props.state],
    'flex-grow',
  );

  const numberClass = mergeTailwind(
    {
      [MStepState.Inactive]: [
        'dark:bg-dark-monetr-background-emphasis',
      ],
      [MStepState.Active]: [
        'dark:bg-dark-monetr-brand-bright',
        'dark:text-black',
      ],
      [MStepState.Complete]: [
        'dark:bg-green-600',
        'dark:text-white',
      ],
    }[props.state],
    'rounded-full',
    'w-5',
    'h-5',
    'flex',
    'text-center',
    'align-middle',
    'items-center',
    'justify-center',
  );

  const textClass = mergeTailwind(
    {
      // On smaller screens hide the text for items that are not the current step or not the next step.
      'sm:inline hidden': ![props.currentIndex, props.currentIndex + 1].includes(props.index),
    },
    {
      [MStepState.Inactive]: [],
      [MStepState.Active]: [
        'dark:text-dark-monetr-brand-bright',
      ],
      [MStepState.Complete]: [
        'dark:text-green-600',
      ],
    }[props.state],
  );

  return (
    <Fragment>
      <MSpan className='flex gap-1 items-center h-6'>
        <MSpan className={ numberClass }>
          { props.index + 1 }
        </MSpan>
        <MSpan className={ textClass }>
          { props.name }
        </MSpan>
      </MSpan>
      { props.divider &&
        <MDivider className={ lineClass } />
      }
    </Fragment>
  );
}
