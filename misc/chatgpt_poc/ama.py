from flask import Flask, request, render_template_string
from chat import chat_once_mode

app = Flask(__name__)

template = """
<!DOCTYPE html>
<html>
  <head>
    <title>Factors.Ai: Ask Me Anything</title>
    <script>
      function onSubmit() {
        var xhr = new XMLHttpRequest();
        xhr.open('POST', '/', true);
        xhr.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
        xhr.onreadystatechange = function() {
          if (xhr.readyState === 4 && xhr.status === 200) {
            document.getElementById('result').innerHTML += xhr.responseText;
          }
        };
        var question = document.getElementById('question').value;
        var data = 'question=' + encodeURIComponent(question);
        xhr.send(data);
        return false;
      }
    </script>
  </head>
  <body>
    <h1>Factors.Ai: Ask Me Anything</h1>
    <form onsubmit="return onSubmit()">
      <label for="question">Question:</label>
      <input type="text" id="question" name="question" style="width: 800px;">
      <input type="submit" value="Ask!">
    </form>
    <div id="result"></div>
  </body>
</html>
"""

@app.route('/', methods=['GET', 'POST'])
def ask():
    if request.method == 'POST':
        question = request.form['question']
        # answer = ask_gpt(question=question, prepend_question=True)
        resp = chat_once_mode(question, 'ft', silent=True, return_answer=True, return_prompt=False, reduce=False)
        answer = resp['answer']
        return answer
    else:
        return render_template_string(template)

if __name__ == '__main__':
    app.run()
