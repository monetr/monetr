import React from 'react';
import { Navigate, Route, Routes } from 'react-router-dom';

import SubscribePage from 'pages/account/subscribe';
import AfterCheckoutPage from 'pages/account/subscribe/after';
import LogoutPage from 'pages/logout';

export default function BillingRequiredRouter(): JSX.Element {
  return (
    <Routes>
      <Route path="/logout" element={ <LogoutPage /> } />
      <Route path="/account/subscribe" element={ <SubscribePage /> } />
      <Route path="/account/subscribe/after" element={ <AfterCheckoutPage /> } />
      <Route path="*" element={ <Navigate replace to="/account/subscribe" /> } />
    </Routes>
  );
}
