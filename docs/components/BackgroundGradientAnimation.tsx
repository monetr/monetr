import { useEffect, useState } from 'react';

import mergeClasses from '@monetr/docs/util/mergeClasses';

import styles from './BackgroundGradientAnimation.module.scss';

function hexToRgb(hex: string) {
  var result = /^#?([a-f\d]{2})([a-f\d]{2})([a-f\d]{2})$/i.exec(hex);
  return result
    ? [
        parseInt(result[1], 16), // Red
        parseInt(result[2], 16), // Green
        parseInt(result[3], 16), // Blue
      ].join(', ')
    : '0, 0, 0';
}

export function BackgroundGradientAnimation(): JSX.Element {
  const gradientBackgroundStart = hexToRgb('#111111');
  const gradientBackgroundEnd = hexToRgb('#111111');
  const firstColor = hexToRgb('#3b82f6');
  const secondColor = hexToRgb('#f056a3');
  const thirdColor = hexToRgb('#2cedff');
  const fourthColor = hexToRgb('#ef4444');
  const fifthColor = hexToRgb('#22c55e');
  const pointerColor = hexToRgb('#4E1AA0');
  const size = '80%';
  const blendingValue = 'hard-light';

  useEffect(() => {
    document.body.style.setProperty('--gradient-background-start', gradientBackgroundStart);
    document.body.style.setProperty('--gradient-background-end', gradientBackgroundEnd);
    document.body.style.setProperty('--first-color', firstColor);
    document.body.style.setProperty('--second-color', secondColor);
    document.body.style.setProperty('--third-color', thirdColor);
    document.body.style.setProperty('--fourth-color', fourthColor);
    document.body.style.setProperty('--fifth-color', fifthColor);
    document.body.style.setProperty('--pointer-color', pointerColor);
    document.body.style.setProperty('--size', size);
    document.body.style.setProperty('--blending-value', blendingValue);
  }, [fifthColor, firstColor, fourthColor, gradientBackgroundEnd, pointerColor, secondColor, thirdColor]);

  const [isSafari, setIsSafari] = useState(false);
  useEffect(() => {
    setIsSafari(/^((?!chrome|android).)*safari/i.test(navigator.userAgent));
  }, []);

  return (
    <div className={styles.root}>
      <svg className={styles.svg}>
        <title>Background animation</title>
        <defs>
          <filter id='blurMe'>
            <feGaussianBlur in='SourceGraphic' result='blur' stdDeviation='10' />
            <feColorMatrix in='blur' mode='matrix' result='goo' values='1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 18 -8' />
            <feBlend in='SourceGraphic' in2='goo' />
          </filter>
        </defs>
      </svg>
      <div className={mergeClasses(styles.container, isSafari ? styles.containerSafari : styles.containerChromium)}>
        <div className={mergeClasses(styles.gradientBlob, styles.first)}></div>
        <div className={mergeClasses(styles.gradientBlob, styles.second)}></div>
        <div className={mergeClasses(styles.gradientBlob, styles.third)}></div>
        <div className={mergeClasses(styles.gradientBlob, styles.fourth)}></div>
        <div className={mergeClasses(styles.gradientBlob, styles.fifth)}></div>
      </div>
    </div>
  );
}
