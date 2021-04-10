import { Card, List } from '@material-ui/core';
import React, { Component } from 'react';


export class GoalsView extends Component<any, any> {

  render() {
    return (
      <div className="minus-nav">
        <div className="flex flex-col h-full p-10 max-h-full">
          <div className="grid grid-cols-3 gap-4 flex-grow">
            <div className="col-span-2">
              <Card elevation={ 4 } className="w-full transaction-list">
                <List disablePadding className="w-full">

                </List>
              </Card>
            </div>
            <div>
              <Card elevation={ 4 } className="h-full w-full">

              </Card>
            </div>
          </div>
        </div>
      </div>
    );
  }
}
