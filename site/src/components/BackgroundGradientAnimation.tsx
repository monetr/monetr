/* eslint-disable max-len */
import { useEffect, useState } from 'react';

import { twMerge } from 'tailwind-merge';

function hexToRgb(hex: string) {
  var result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result ? [
    parseInt(result[1], 16), // Red
    parseInt(result[2], 16), // Green
    parseInt(result[3], 16), // Blue
  ].join(', ') : '0, 0, 0';
}

export function BackgroundGradientAnimation(): JSX.Element {
  const gradientBackgroundStart = '--rp-c-bg';
  const gradientBackgroundEnd = hexToRgb('#19161f');
  const firstColor = hexToRgb('#3b82f6');
  const secondColor = hexToRgb('#f056a3');
  const thirdColor = hexToRgb('#2cedff');
  const fourthColor = hexToRgb('#ef4444');
  const fifthColor = hexToRgb('#22c55e');
  const pointerColor = hexToRgb('#4E1AA0');
  const size = '80%';
  const blendingValue = 'hard-light';

  useEffect(() => {
    document.body.style.setProperty(
      '--gradient-background-start',
      gradientBackgroundStart
    );
    document.body.style.setProperty(
      '--gradient-background-end',
      gradientBackgroundEnd
    );
    document.body.style.setProperty('--first-color', firstColor);
    document.body.style.setProperty('--second-color', secondColor);
    document.body.style.setProperty('--third-color', thirdColor);
    document.body.style.setProperty('--fourth-color', fourthColor);
    document.body.style.setProperty('--fifth-color', fifthColor);
    document.body.style.setProperty('--pointer-color', pointerColor);
    document.body.style.setProperty('--size', size);
    document.body.style.setProperty('--blending-value', blendingValue);
  }, [blendingValue, fifthColor, firstColor, fourthColor, gradientBackgroundEnd, gradientBackgroundStart, pointerColor, secondColor, size, thirdColor]);

  const [isSafari, setIsSafari] = useState(false);
  useEffect(() => {
    setIsSafari(/^((?!chrome|android).)*safari/i.test(navigator.userAgent));
  }, []);

  return (
    <div
      className={ twMerge(
        'h-screen w-screen',
        'fixed overflow-hidden',
        'inset-0 top-0 left-0 -z-20',
        'bg-[linear-gradient(40deg,rgb(var(--gradient-background-start)),rgb(var(--gradient-background-end)))]',
        'opacity-20',
      ) }
    >
      <svg className='hidden'>
        <defs>
          <filter id='blurMe'>
            <feGaussianBlur
              in='SourceGraphic'
              stdDeviation='10'
              result='blur'
            />
            <feColorMatrix
              in='blur'
              mode='matrix'
              values='1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 18 -8'
              result='goo'
            />
            <feBlend in='SourceGraphic' in2='goo' />
          </filter>
        </defs>
      </svg>
      <div
        className={ twMerge(
          'gradients-container h-full w-full blur-lg',
          isSafari ? 'blur-2xl' : '[filter:url(#blurMe)_blur(40px)]'
        ) }
      >
        <div
          className={ twMerge(
            'absolute',
            '[background:radial-gradient(circle_at_center,_var(--first-color)_0,_var(--first-color)_50%)_no-repeat]',
            '[mix-blend-mode:var(--blending-value)]',
            'w-[var(--size)] h-[var(--size)]',
            'top-[calc(50%-var(--size)/2)] left-[calc(50%-var(--size)/2)]',
            '[transform-origin:center_center]',
            'animate-first',
            'opacity-100'
          ) }
        ></div>
        <div
          className={ twMerge(
            'absolute',
            '[background:radial-gradient(circle_at_center,_rgba(var(--second-color),_0.8)_0,_rgba(var(--second-color),_0)_50%)_no-repeat]',
            '[mix-blend-mode:var(--blending-value)]',
            'w-[var(--size)] h-[var(--size)]',
            'top-[calc(50%-var(--size)/2)] left-[calc(50%-var(--size)/2)]',
            '[transform-origin:calc(50%-400px)]',
            'animate-second',
            'opacity-100'
          ) }
        ></div>
        <div
          className={ twMerge(
            'absolute',
            '[background:radial-gradient(circle_at_center,_rgba(var(--third-color),_0.8)_0,_rgba(var(--third-color),_0)_50%)_no-repeat]',
            '[mix-blend-mode:var(--blending-value)]',
            'w-[var(--size)] h-[var(--size)]',
            'top-[calc(50%-var(--size)/2)] left-[calc(50%-var(--size)/2)]',
            '[transform-origin:calc(50%+400px)]',
            'animate-third',
            'opacity-100'
          ) }
        ></div>
        <div
          className={ twMerge(
            'absolute',
            '[background:radial-gradient(circle_at_center,_rgba(var(--fourth-color),_0.8)_0,_rgba(var(--fourth-color),_0)_50%)_no-repeat]',
            '[mix-blend-mode:var(--blending-value)]',
            'w-[var(--size)] h-[var(--size)]',
            'top-[calc(50%-var(--size)/2)] left-[calc(50%-var(--size)/2)]',
            '[transform-origin:calc(50%-200px)]',
            'animate-fourth',
            'opacity-70'
          ) }
        ></div>
        <div
          className={ twMerge(
            'absolute',
            '[background:radial-gradient(circle_at_center,_rgba(var(--fifth-color),_0.8)_0,_rgba(var(--fifth-color),_0)_50%)_no-repeat]',
            '[mix-blend-mode:var(--blending-value)]',
            'w-[var(--size)] h-[var(--size)]',
            'top-[calc(50%-var(--size)/2)] left-[calc(50%-var(--size)/2)]',
            '[transform-origin:calc(50%-800px)_calc(50%+800px)]',
            'animate-fifth',
            'opacity-100'
          ) }
        ></div>
      </div>
    </div>
  );
};
