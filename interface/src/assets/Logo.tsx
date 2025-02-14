/* eslint-disable max-len */
import React from 'react';

interface LogoProps {
  className?: string;
}

export default function Logo(props: LogoProps): JSX.Element {
  return (
    <svg 
      className={ props.className } 
      xmlns='http://www.w3.org/2000/svg' 
      xmlnsXlink='http://www.w3.org/1999/xlink' 
      viewBox='0 0 156.17 133.82'
    >
      <defs>
        <style> {`
        .a {
          fill: #0093a6;
        }

        .b {
          fill: #231f20;
        }

        .c {
          fill: url(#a);
        }

        .d {
          fill: url(#b);
        }

        .e {
          fill: #ff5798;
        }

        .f {
          fill: url(#c);
        }

        .g {
          fill: url(#d);
        }

        .h {
          fill: #4e1aa0;
        }
        `} </style>
        <linearGradient id='a' x1='441.63' y1='444.08' x2='448.03' y2='521.51' gradientUnits='userSpaceOnUse'>
          <stop offset='0' stopColor='#829aff' />
          <stop offset='1' stopColor='#2cedff' />
        </linearGradient>
        <linearGradient id='b' x1='561.57' y1='446.54' x2='551.09' y2='519.31' xlinkHref='#a' />
        <linearGradient id='c' x1='433.74' y1='408.17' x2='504.48' y2='503.36' gradientUnits='userSpaceOnUse'>
          <stop offset='0' stopColor='#7200e1' />
          <stop offset='1' stopColor='#f056a3' />
        </linearGradient>
        <linearGradient id='d' x1='569.82' y1='404.95' x2='480.31' y2='511.49' xlinkHref='#c' />
      </defs>
      <g>
        <path className='a' d='M578,404.79a14.08,14.08,0,0,0-3.51-11.12c-4.85-5.16-11.65-4.14-12.2-4.05H539.9c.44,2-.61,4.25-3.19,4.06-8.17-.61-12.2,6.85-18,11.32-1.83,1.4-3.62.45-4.44-1.16L500,416.7l-19.35-17.41a2.84,2.84,0,0,1-2.62.16,43.85,43.85,0,0,1-8.19-3.88,12.93,12.93,0,0,0-3.29-1.95,16.34,16.34,0,0,0-4.53-.11,3.25,3.25,0,0,1-3.27-3.89h-21a14.7,14.7,0,0,0-7.35.91l-.51.21-.11,0a12.86,12.86,0,0,0-4.56,3.35A14.11,14.11,0,0,0,422,404.79h0V509.23a12.84,12.84,0,0,0,20.35,11.66l28.48-28.45h0c8.69,8.94,13.69,16.91,27.66,16.91,17.3,0,23.23-8,30.67-16.91l28.47,28.45A12.85,12.85,0,0,0,578,509.23Z' transform='translate(-421.91 -389.48)' />
        <path className='b' d='M429.85,390.74l.51-.21Z' transform='translate(-421.91 -389.48)' />
        <path className='c' d='M422,404.79a14.25,14.25,0,0,1,4-11.46,13.48,13.48,0,0,1,3.71-2.55,12.86,12.86,0,0,0-4.56,3.35A14.18,14.18,0,0,0,422,404.79V509.23a12.84,12.84,0,0,0,20.35,11.66l28.48-28.45L422,444.53Z' transform='translate(-421.91 -389.48)' />
        <path className='b' d='M458.71,389.62Z' transform='translate(-421.91 -389.48)' />
        <path className='b' d='M562.27,389.62h0Z' transform='translate(-421.91 -389.48)' />
        <path className='d' d='M578,444.53l-48.83,47.91h0l28.47,28.45A12.85,12.85,0,0,0,578,509.23Z' transform='translate(-421.91 -389.48)' />
        <path className='b' d='M574.81,394.13A14.18,14.18,0,0,1,578,404.79h0a14.08,14.08,0,0,0-3.51-11.12c-4.85-5.16-11.65-4.14-12.2-4.05C562.79,389.54,570,388.6,574.81,394.13Z' transform='translate(-421.91 -389.48)' />
        <path className='b' d='M429.74,390.78l.11,0Z' transform='translate(-421.91 -389.48)' />
        <path className='e' d='M470.83,492.44h0c8.69,8.94,13.69,16.91,27.66,16.91,17.3,0,23.23-8,30.67-16.91L500,463.31Z' transform='translate(-421.91 -389.48)' />
        <path className='b' d='M437.71,389.62a14.7,14.7,0,0,0-7.35.91A15.91,15.91,0,0,1,437.71,389.62Z' transform='translate(-421.91 -389.48)' />
        <path className='f' d='M449.94,413.31a13.73,13.73,0,0,1,8.77-23.69h-21a15.91,15.91,0,0,0-7.35.91l-.51.21-.11,0a13.48,13.48,0,0,0-3.71,2.55,14.25,14.25,0,0,0-4,11.46v39.74l48.83,47.91L500,463.31Z' transform='translate(-421.91 -389.48)' />
        <path className='g' d='M578,444.53V404.79h0a14.18,14.18,0,0,0-3.18-10.66c-4.79-5.53-12-4.59-12.53-4.51h-21a13.73,13.73,0,0,1,8.77,23.69l-50.06,50,29.17,29.13h0Z' transform='translate(-421.91 -389.48)' />
        <path className='h' d='M446.43,398.77a13.77,13.77,0,0,0,3.51,14.54l50.05,50,50.06-50a13.73,13.73,0,0,0-8.77-23.69h0a29.19,29.19,0,0,0-19.52,7.49L500,416.7l-21.78-19.6a29.14,29.14,0,0,0-19.51-7.48h0A13.79,13.79,0,0,0,446.43,398.77Z' transform='translate(-421.91 -389.48)' />
      </g>
    </svg>
  );
}
