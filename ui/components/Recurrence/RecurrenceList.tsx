import React, { useState } from 'react';
import { Checkbox, List as ListUI, ListItem, ListItemIcon, ListItemText } from '@mui/material';
import moment from 'moment';

import getRecurrencesForDate from 'components/Recurrence/getRecurrencesForDate';
import Recurrence from 'components/Recurrence/Recurrence';

export interface Props {
  // TODO Add a way to pass a current value to the RecurrenceList component.
  date: moment.Moment;
  onChange: { (value: Recurrence): void };
  disabled?: boolean;
}

// RecurrenceList generates a list of possible recurrence rules based on the provided date. When a recurrence is
// selected the onChange function will be called with a string value representing an RRule.
export default function RecurrenceList(props: Props): JSX.Element {
  const [selectedIndex, setSelectedIndex] = useState<number | null>(null);
  const rules = getRecurrencesForDate(props.date);

  const selectItem = (index: number) => () => {
    const { onChange } = props;
    setSelectedIndex(index);
    onChange(rules[index]);
  };

  return (
    <ListUI dense>
      { rules
        .map((rule, index) => (
          <ListItem key={ rule.name } dense button onClick={ selectItem(index) }>
            <ListItemIcon>
              <Checkbox
                edge="start"
                checked={ selectedIndex === index }
                tabIndex={ -1 }
                color="primary"
                disabled={ !!props.disabled }
              />
            </ListItemIcon>
            <ListItemText>
              { rule.name }
            </ListItemText>
          </ListItem>
        )) }
    </ListUI>
  );
}
