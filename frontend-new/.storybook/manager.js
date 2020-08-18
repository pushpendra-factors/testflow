import { addons } from '@storybook/addons';
import { themes } from '@storybook/theming';
import factorsTheme from './factorsTheme'; 

addons.setConfig({
  theme: factorsTheme,
//   showRoots: false
});