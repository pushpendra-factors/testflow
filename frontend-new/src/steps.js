import React from 'react';
import { keyframes } from '@emotion/core';
import { Text } from 'factorsComponents';

const keyframesRotate = keyframes`
  50% {
    transform: translateY(-5px);
  }
}`

const style = {
  popover: base => ({
    ...base,
    boxShadow: '0 0 3em rgba(0, 0, 0, 0.5)',
    borderRadius: 10,
    '.fai-text' : {
      color: '#007aff',
    }
  }),
  button: base => ({ 
    ...base,
    transform: `scale(0.9)`,
    '&:focus':{
      outline:'none',
    }
  }),
  arrow: (base,{ disabled }) => ({
    ...base,
    width: 25,
    height: 25,
    flex: '0 0 16px',
    '&:hover': {
      color: disabled ? '#caccce' : '#000',
    },
  }),
  navigation: base => ({ ...base, margin: 10}),
  maskArea: base => ({ ...base, rx: 10 }),
  maskWrapper: base => ({
    ...base,
    opacity: 0.3,
  }),
  close: base => ({ 
    ...base,
    background: 'var(--reactour-accent,#007aff)',
    height: '2.1em',
    width: '2.6em',
    paddingLeft: '0.8125em',
    paddingRight: '0.8125em',
    borderRadius: '1.625em',
    color: 'white',
    textAlign: 'center',
    boxShadow: '0 0.25em 0.5em rgb(0 0 0 / 30%)',
    top: '-0.8125em',
    right: '-0.8125em',
  }),
  dot: base => ({
    ...base,
    animationDuration: '1s',
    animationName: keyframesRotate,
    animationIterationCount: 'infinite',
    '&:nth-of-type(1)': {
      animationDelay: '.3s',
    },
    '&:nth-of-type(2)': {
      animationDelay: '.6s',
    },
  }),
}

const steps = [
  {
    selector: '[data-tour="step-1"]',
    content: 
      <div>
        <Text type={'title'} level={6} weight={'bold'}>Let’s start — here is your dashboard.</Text>
        <Text type={'title'} level={7}>All previously built and saved reports can be added to this view for quick and easy access.</Text>
      </div>,
    styles: style,
    position: 'right',
  },
  {
    selector: '[data-tour="step-2"]',
    content: <Text type={'title'} level={7}>The search bar on top gives you easy access to recently saved reports from anywhere.</Text>,
    action: node => {
      node.focus()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-3"]',
    content: <Text type={'title'} level={7}>You can also run new queries from here, if you’re in a rush!</Text>,
    styles: style,
  },
  {
    selector: '[data-tour="step-4"]',
    content: <Text type={'title'} level={7}>You can add new dashboard views, to represent different categories of reports you need to look at</Text>,
    styles: style,
  },
  {
    selector: '[data-tour="step-5"]',
    content: 
      <div>
        <Text type={'title'} level={6} weight={'bold'}>The heart of your work lies here — the Analyse engine.</Text>
        <Text type={'title'} level={7}>Here you can run deep analyses and charts for events, funnels, and campaigns, as well as model attribution analyses across all your marketing touchpoints</Text>
      </div>,
    action: node => {
      node.click()
    },
    styles: style,
    position: 'right',
  },
  {
    selector: '[data-tour="step-6"]',
    content: 
      <div>
        <Text type={'title'} level={6} weight={'bold'}>Ah, the Explain engine!</Text>
        <Text type={'title'} level={7}>Here, we periodically track conversion goals and journeys you define, to help you understand what factors are impacting them the most.</Text>
        <Text type={'title'} level={7}>You’ll have periodic and actionable insights sent to you.</Text>
      </div>,
    action: node => {
      node.click()
    },
    styles: style,
    position: 'right',
  },
  {
    selector: '[data-tour="step-7"]',
    content: 
      <div>
        <Text type={'title'} level={6} weight={'bold'}>Set up your custom events and properties,</Text>
        <Text type={'title'} level={7}>as well as configure your UTM parameters to match the standards used across the platform.</Text>
      </div>,
    action: node => {
      node.click()
    },
    styles: style,
    position: 'right',
  },
  {
    selector: '[data-tour="step-8"]',
    content: <Text type={'title'} level={7}>All projects you have access to appear here.</Text>,
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-9"]',
    content: <Text type={'title'} level={7}>You can also view and edit account settings form this panel</Text>,
    position: 'right',
    styles: style,
  },
  {
    selector: '[data-tour="step-10"]',
    content: <Text type={'title'} level={7}>Connect to all your data sources here!</Text>,
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-11"]',
    content: 
      <div>
        <Text type={'title'} level={6} weight={'bold'}>For each integration you need,</Text>
        <Text type={'title'} level={7}>use the primary blue button for an easy SSO (Single Sign-On) process, or view documentation from the adjacent button.</Text>
      </div>,
    position: 'right',
    styles: style,
  },
];

export default steps;
