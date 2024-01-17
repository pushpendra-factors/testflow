# Chat Factors

## Setup
> Note: Use `python` or `python3` (correspondingly `pip` or `pip3`) based on whichever points to Python 3.
- Step 1: Move into this directory
  - `cd factors/python_backend`
- Step 2: Check if you have Python 3
  - `python --version`
  should return `Python 3.\*.\*`
  - If the above command gives an error or returns `Python 2.\*.\*`, try `python3 --version`.
  - If you still get a not-found error, you might have to install Python. Check [here](https://realpython.com/installing-python/).
- Step 3: Install requirements
  - `pip install -r adwords_requirements.txt`
  - `pip install -r chat_requirements.txt`
  - If `pip` doesn't work or points to Python 2, use `pip3`
  - Ignore the `pip` upgrade warning, if any.

## Usage
### Step 1: Set up OpenAI API Key
Create a new file (or edit if exists), `factors/python_backend/chat_factors/chatgpt_poc/key.json ` with the following content:
  ```
  {"key": "OPENAI_API_KEY"}
  ```
  and replace `OPENAI_API_KEY` with your actual API key that can be downloaded from [here](https://platform.openai.com/account/api-keys).
### Step 2: Run app
  ```
  python app.py --env development --developer_token <ADS_DEVELOPER_TOKEN> --oauth_secret  $(cat <GOOGLE_OAUTH_CLIENT_JSON_FILEPATH>) --chat_bucket_name <BUCKET_FOR_CHAT_SERVER>
  ```

