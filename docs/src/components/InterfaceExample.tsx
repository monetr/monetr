'use client';

import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

import Login from '@monetr/interface/pages/login';

export default function InterfaceExample(): JSX.Element {

  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        initialData: {
          ['test']: 'test',
        },
      },
    },
  });

  return (
    <div className='w-full h-full translate-x-0 translate-y-0 scale-100 delay-150 duration-500 ease-in-out rounded-2xl mt-8 shadow-2xl z-10 backdrop-blur-md transition-all opacity-95 pointer-events-none select-none max-w-[1280px] max-h-[720px] aspect-video-vertical md:aspect-video bg-[#19161f]'>
      <MemoryRouter>
        <QueryClientProvider client={ queryClient }> 
          <Login />
        </QueryClientProvider>
      </MemoryRouter>
    </div>
  );
}
