from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

openai.api_key = ''

def extract_ideal_buying_companies(text):
    prompt = "Based on this website text, share an array with strings with the types of companies that they sell to, that they consider as their ideal customers to have. Share 5 at max. \n\n" + text + "\n\n. "
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.3
    )
    ideal_companies = response.choices[0].text.strip()
    
    print('OPENAI DETECTED IDEAL COMPANIES')
    print(ideal_companies)
    print('')

    return ideal_companies


def find_main_buying_companies(text):

    # Find the pricing page URL
    ideal_companies = extract_ideal_buying_companies(text)

    if ideal_companies:
        # Scrape the pricing buckets from the pricing page
        
        return ideal_companies
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No ideal company assumption possible.')

