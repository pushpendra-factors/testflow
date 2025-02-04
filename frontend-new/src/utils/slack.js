const host = BUILD_CONFIG.backend_host;

export const sendSlackNotification = (email, projectname, name, text) => {
  let webhookURL =
    'https://hooks.slack.com/services/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';
  let data = {
    text: text
      ? text
      : `User ${email} from Project "${projectname}" Activated Integration: ${name}`,
    username: 'Signup User Actions',
    icon_emoji: ':golf:'
  };
  let params = {
    method: 'POST',
    body: JSON.stringify(data)
  };

  if (host === 'https://api.factors.ai') {
    fetch(webhookURL, params)
      .then((response) => response.json())
      .then((response) => {
        console.log(response);
      })
      .catch((err) => {
        console.log('err', err);
      });
  }
};
