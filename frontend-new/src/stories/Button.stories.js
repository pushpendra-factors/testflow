import React from 'react';
import { Button  } from 'antd';  
import 'antd/dist/antd.css';
import '../assets/tailwind.output.css';
import '../assets/index.scss';
import '../styles/factors-ai.main.scss';

export default {
  title: 'Components/Button',
  component: Button, 
  parameters: {
    docs: { 
        description: { 
            component: 'Primary UI Component for User Interaction.' 
        } 
    },
  },
  argTypes: { 
    label: {  
      control: false }, 
    type: {control: false}
  }, 
};

const Template = (args) => <Button {...args} >{args.label}</Button> ; 

export const Primary = Template.bind({});
Primary.args = { 
  type: 'primary',
  label: 'Primary Button', 
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
