import { keyframes } from '@emotion/core';
import React from 'react';

const keyframesRotate = keyframes`
  50% {
    transform: translateY(-5px);
  }
}`

const style = {
  popover: base => ({
    ...base,
    boxShadow: '0 0 3em rgba(0, 0, 0, 0.5)',
    color: '#007aff',
    borderRadius: 10,
    'p' : {
      fontSize: 15,
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
    lineHeight: 2,
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
        <p><strong>Let’s start — here is your dashboard.</strong></p>
        <p>All previously built and saved reports can be added to this view for quick and easy access.</p>
      </div>,
    styles: style,
  },
  {
    selector: '[data-tour="step-2"]',
    content: <p>The search bar on top gives you easy access to recently saved reports from anywhere.</p>,
    action: node => {
      node.focus()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-3"]',
    content: <p>You can also run new queries from here, if you’re in a rush!</p>,
    styles: style,
  },
  {
    selector: '[data-tour="step-4"]',
    content: <p>You can add new dashboard views, to represent different categories of reports you need to look at</p>,
    styles: style,
  },
  {
    selector: '[data-tour="step-5"]',
    content: 
      <div>
        <p><strong>The heart of your work lies here — the Analyse engine.</strong></p>
        <p>Here you can run deep analyses and charts for events, funnels, and campaigns, as well as model attribution analyses across all your marketing touchpoints</p>
      </div>,
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-6"]',
    content: 
      <div>
        <p><strong>Ah, the Explain engine!</strong></p>
        <p>Here, we periodically track conversion goals and journeys you define, to help you understand what factors are impacting them the most.</p>
        <p>You’ll have periodic and actionable insights sent to you.</p>
      </div>,
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-7"]',
    content: 
      <div>
        <p>Set up your custom events and properties,</p>
        <p>as well as configure your UTM parameters to match the standards used across the platform.</p>
      </div>,
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-8"]',
    content: <p>All projects you have access to appear here.</p>,
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-9"]',
    content: <p>You can also view and edit account settings form this panel</p>,
    position: 'right',
    styles: style,
  },
  {
    selector: '[data-tour="step-10"]',
    content: <p>Connect to all your data sources here!</p>,
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-11"]',
    content: 
      <div>
        <p><strong>For each integration you need,</strong></p>
        <p>use the primary blue button for an easy SSO (Single Sign-On) process, or view documentation from the adjacent button.</p>
      </div>,
    position: 'right',
    styles: style,
  },
];

export default steps;
