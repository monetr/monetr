import '@fontsource-variable/inter';
import '@monetr/site/theme/index.scss';

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

// Make the theme switcher not render at all!
function SwitchAppearance() {
  return null;
}

export * from '@rspress/core/theme-original';
export { HomeLayout, SwitchAppearance };

