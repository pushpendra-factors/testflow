import React from 'react';

import { Button  } from 'antd'; 
import '../styles/factors-ai.main.scss';
export default {
  title: 'Example/Button',
  component: Button, 
};

const Template = (args) => <Button {...args} >{args.label}</Button> ; 

export const Primary = Template.bind({});
Primary.args = { 
  type: 'primary',
  label: 'Button 123',
};

export const Secondary = Template.bind({});
Secondary.args = { 
  label: 'Secondary Button',
};

export const Dashed = Template.bind({});
Dashed.args = {
  type:"dashed",
  label: 'Dashed Button',
};

export const Text = Template.bind({});
Text.args = {
  type:"text",
  label: 'Text Button',
};

export const Link = Template.bind({});
Link.args = {
  type:"link",
  label: 'Link Button',
};
