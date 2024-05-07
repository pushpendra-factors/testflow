# ChatGPT PoC

## Setup
> Note: Use `python` or `python3` (correspondingly `pip` or `pip3`) based on whichever points to Python 3.
- Step 1: Move into this directory
  - `cd factors/misc/chatgpt_poc`
- Step 2: Check if you have Python 3
  - `python --version`
  should return `Python 3.\*.\*`
  - If the above command gives an error or returns `Python 2.\*.\*`, try `python3 --version`.
  - If you still get a not-found error, you might have to install Python. Check [here](https://realpython.com/installing-python/).
- Step 3: Install requirements
  - `pip install -r requirements.txt`
  - If `pip` doesn't work or points to Python 2, use `pip3`
  - Ignore the `pip` upgrade warning, if any.

## Usage
### Step 1: Set up OpenAI API Key
Create a new file (or edit if exists), `factors/misc/chatgpt_poc/key.json` with the following content:
  ```
  {"key": "OPENAI_API_KEY"}
  ```
  and replace `OPENAI_API_KEY` with your actual API key that can be downloaded from [here](https://platform.openai.com/account/api-keys).
### Step 2: Run main.py
- `python main.py -h` or `python main.py --help` for more help:
  ```
  python main.py [-h] [--model MODEL] [--scratch] [-s] prompt

  positional arguments:
    prompt         The question you want to ask about your data (E.g., "How many users visited our website last month?")

  optional arguments:
    -h, --help     show this help message and exit
    --model MODEL  The model to use -- fine-tuned ("ft") or information-retrieval ("ir").
    --scratch      If "ft", fine-tune model again, and if "ir", generate embeddings from scratch.
    -s, --silent   Silent
  ```

   (Use `python3 main.py` if it doesn't work)

