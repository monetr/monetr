import React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';
import Logout from 'views/Authentication/Logout';
import AfterCheckout from 'views/Subscriptions/AfterCheckout';
import Subscribe from 'views/Subscriptions/Subscribe';

export default function BillingRequiredRouter(): JSX.Element {
  return (
    <Routes>
      <Route path="/logout" element={ <Logout/> }/>
      <Route path="/account/subscribe" element={ <Subscribe/> }/>
      <Route path="/account/subscribe/after" element={ <AfterCheckout/> }/>
      <Route path="*" element={ <Navigate replace to="/account/subscribe"/> }/>
    </Routes>
  );
}
