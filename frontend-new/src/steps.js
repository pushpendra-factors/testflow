const steps = [
  {
    selector: '[data-tour="step-1"]',
    content: 'Vivamus sed dui nisi',
    action: node => {
      node.click()
    },
  },
  {
    selector: '[data-tour="step-2"]',
    content: 'Vivamus sed dui nisi',
    position: 'center',
    action: node => {
      node.focus()
    },
  },
  {
    selector: '[data-tour="step-2.1"]',
    content: 'Vivamus sed dui nisi',
    position: 'center',
  },
  {
    selector: '[data-tour="step-3"]',
    content: 'Vivamus sed dui nisi',
  },
  {
    selector: '[data-tour="step-4"]',
    content: 'Vivamus sed dui nisi',
    position: 'right',
    action: node => {
      node.click()
    },
  },
  {
    selector: '[data-tour="step-4.1"]',
    content: 'Vivamus sed dui nisi',
    position: 'right',
  },
  {
    selector: '[data-tour="step-5"]',
    content: 'Vivamus sed dui nisi',
    content: 'Vivamus sed dui nisi',
    action: node => {
      node.click()
    },
  },
  {
    selector: '[data-tour="step-6"]',
    content: 'Vivamus sed dui nisi',
    content: 'Vivamus sed dui nisi',
    action: node => {
      node.click()
    },
  },
  {
    selector: '[data-tour="step-7"]',
    content: 'Vivamus sed dui nisi',
    action: node => {
      node.click()
    },
  },
  {
    selector: '[data-tour="step-8"]',
    content: 'Vivamus sed dui nisi',
    action: node => {
      node.click()
    },
  },
  {
    selector: '[data-tour="step-9"]',
    content: 'Vivamus sed dui nisi',
  },
    {
    selector: '[data-tour="step-9.1"]',
    content: 'Vivamus sed dui nisi',
  },
];

export default steps;
