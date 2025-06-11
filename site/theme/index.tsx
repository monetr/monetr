import { Fragment } from 'react';

import { BackgroundGradientAnimation } from '@monetr/site/components/BackgroundGradientAnimation';
import Hero from '@monetr/site/components/Hero';

import { NoSSR } from 'rspress/runtime';
import Theme, {
  HomeLayout as BasicHomeLayout,
  Layout as BasicLayout,
} from 'rspress/theme';

function HomeLayout() {
  return (
    <BasicHomeLayout
      beforeHero={ (
        <Fragment>
          <BackgroundGradientAnimation />
          <Hero />
        </Fragment>
      ) }
      afterFeatures={ <h1>TESTING</h1> }
    />
  );
}

const Layout = () => {
  return (
    <BasicLayout
      beforeNav={
        <NoSSR>
          <h1>testing #2</h1>
        </NoSSR>
      }
    />
  );
};

export default { ...Theme, HomeLayout, Layout };
export * from 'rspress/theme';
