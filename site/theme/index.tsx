import '@fontsource-variable/inter';
import '@monetr/site/theme/index.scss';
import '@monetr/site/theme/base.css';

import Logo from './logo.svg';

import {
  HomeLayout as OriginalHomeLayout,
} from '@rspress/core/theme-original';

function HomeLayout() {
  return (
    <OriginalHomeLayout
      afterFeatures={<h1> Testing!</h1>}
      afterHeroActions={
        <h1>Another!</h1>
      }
    />
  );
}

function NavTitle() {
  return (
    <a className='flex hover:opacity-75' href='/'>
      <img alt='monetr logo' className='w-8 h-8 lg:w-10 lg:h-10' src={ Logo } />
      <div className='flex items-center justify-center ml-3'>
        <span className='absolute mx-auto flex border w-fit bg-gradient-to-r blur-xl opacity-50 from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-2xl lg:text-3xl box-content font-extrabold text-transparent text-center select-none'>
          monetr
        </span>
        <h1 className='relative top-0 justify-center flex bg-gradient-to-r items-center from-purple-100 via-purple-200 to-purple-300 bg-clip-text text-2xl lg:text-3xl font-extrabold text-transparent text-center select-auto'>
          monetr
        </h1>
      </div>
    </a>
  )
}

function HomeFooter() {
  return (
    <footer className="rp-home-footer">
      <div className="rp-home-footer__container">
        <div
          className="rp-home-footer__message"
        >
          <a className='rp-link' href="/policy/privacy">Privacy Policy</a>
        </div>
      </div>
    </footer>
  )
}

// Make the theme switcher not render at all!
function SwitchAppearance() {
  return null;
}

export * from '@rspress/core/theme-original';
export { HomeFooter, HomeLayout, SwitchAppearance, NavTitle };

