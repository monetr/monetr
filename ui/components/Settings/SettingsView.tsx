import { Info, Lock, Password, Payment, Security, Settings } from '@mui/icons-material';
import { Box, Button, TextField, Typography } from '@mui/material';
import Tabs from '@mui/material/Tabs';
import Tab from '@mui/material/Tab';
import ChangePassword from 'components/Settings/ChangePassword';
import TOTPSettings from 'components/Settings/TOTPSettings';
import React, { useState } from 'react';

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
  className?: string;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={ value !== index }
      id={ `simple-tabpanel-${ index }` }
      aria-labelledby={ `simple-tab-${ index }` }
      { ...other }
    >
      { value === index && (
        <Box sx={ { p: 3 } }>
          <Typography>{ children }</Typography>
        </Box>
      ) }
    </div>
  );
}

export default function SettingsView(): JSX.Element {
  const [value, setValue] = useState(0);

  const handleChange = (event: React.SyntheticEvent, newValue: number) => {
    setValue(newValue);
  };

  return (
    <div className="w-full h-full flex flex-col">
      <div>
        <Tabs
          value={ value }
          onChange={ handleChange }
          aria-label="scrollable force tabs example"
          className=""
        >
          <Tab className="h-12 min-h-0" label="General" iconPosition="start" icon={ <Settings/> }/>
          <Tab className="h-12 min-h-0" label="Security" iconPosition="start" icon={ <Security/> }/>
          <Tab className="h-12 min-h-0" label="About" iconPosition="start" icon={ <Info/> }/>
        </Tabs>
      </div>
      <TabPanel value={ value } index={ 0 }>
        <h1>General</h1>
      </TabPanel>
      <TabPanel value={ value } index={ 1 } className="w-full h-full">
        <div className="w-full 2xl:w-1/2">
          <div className="grid gap-16">
            <ChangePassword/>
          </div>
        </div>
      </TabPanel>
      <TabPanel value={ value } index={ 2 }>
        <h1>About</h1>
      </TabPanel>
    </div>
  )
}
