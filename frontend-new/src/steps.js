const style = {
  popover: base => ({
    ...base,
    boxShadow: '0 0 3em rgba(0, 0, 0, 0.5)',
    color: '#007aff',
    borderRadius: 10,
    top: 10,
  }),
  button: base => ({ ...base, transform: `scale(1.3)`}),
  navigation: base => ({ ...base, margin: 10}),
  maskWrapper: base => ({
    ...base,
    opacity: 0.5,
  }),
  // highlightedArea: (base, { x, y }) => ({
  //   ...base,
  //   x: x + 10,
  //   y: y + 10,
  // }),
  // badge: base => ({ ...base, color: 'red' }),
  close: base => ({ ...base, color: 'blue', width: 11, height: 11}),
}

const steps = [
  {
    selector: '[data-tour="step-1"]',
    content: "Let’s start — here is your dashboard. All previously built and saved reports can be added to this view for quick and easy access.",
    position: 'center',
    styles: style,
  },
  {
    selector: '[data-tour="step-2"]',
    content: 'The search bar on top gives you easy access to recently saved reports from anywhere.',
    action: node => {
      node.focus()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-3"]',
    content: 'You can also run new queries from here, if you’re in a rush!',
    styles: style,
  },
  {
    selector: '[data-tour="step-4"]',
    content: 'You can add new dashboard views, to represent different categories of reports you need to look at',
    styles: style,
  },
  {
    selector: '[data-tour="step-5"]',
    content: 'The heart of your work lies here — the Analyse engine.Here you can run deep analyses and charts for events, funnels, and campaigns, as well as model attribution analyses across all your marketing touchpoints',
    position: 'center',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-6"]',
    content: 'Ah, the Explain engine! Here, we periodically track conversion goals and journeys you define, to help you understand what factors are impacting them the most. You’ll have periodic and actionable insights sent to you.',
    position: 'center',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-7"]',
    content: 'Set up your custom events and properties, as well as configure your UTM parameters to match the standards used across the platform.',
    position: 'center',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-8"]',
    content: 'All projects you have access to appear here.',
    action: node => {
      node.click()
    },
    styles: style,
  },
   {
    selector: '[data-tour="step-9"]',
    content: 'You can also view and edit account settings form this panel',
    position: 'right',
    styles: style,
  },
  {
    selector: '[data-tour="step-10"]',
    content: 'Connect to all your data sources here!',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-11"]',
    content: 'For each integration you need, use the primary blue button for an easy SSO (Single Sign-On) process, or view documentation from the adjacent button.',
    styles: style,
  },
];

export default steps;
