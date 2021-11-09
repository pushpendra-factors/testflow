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
    content: 'Vivamus sed dui nisi',
    position: 'right',
    styles: style,
  },
  {
    selector: '[data-tour="step-2"]',
    content: 'Vivamus sed dui nisi',
    position: 'center',
    action: node => {
      node.focus()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-2.1"]',
    content: 'Vivamus sed dui nisi',
    position: 'center',
    styles: style,
  },
  {
    selector: '[data-tour="step-3"]',
    content: 'Vivamus sed dui nisi',
    styles: style,
  },
  {
    selector: '[data-tour="step-4"]',
    content: 'Vivamus sed dui nisi',
    position: 'right',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-4.1"]',
    content: 'Vivamus sed dui nisi',
    position: 'right',
    styles: style,
  },
  {
    selector: '[data-tour="step-5"]',
    content: 'Vivamus sed dui nisi',
    position: 'right',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-6"]',
    content: 'Vivamus sed dui nisi',
    position: 'right',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-7"]',
    content: 'Vivamus sed dui nisi',
    position: 'right',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-8"]',
    content: 'Vivamus sed dui nisi',
    action: node => {
      node.click()
    },
    styles: style,
  },
  {
    selector: '[data-tour="step-9"]',
    content: 'Vivamus sed dui nisi',
    styles: style,
  },
    {
    selector: '[data-tour="step-9.1"]',
    content: 'Vivamus sed dui nisi',
    styles: style,
  },
];

export default steps;
