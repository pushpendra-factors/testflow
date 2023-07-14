from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

openai.api_key = ''

def extract_exact_category(text):
    prompt = "Based on this website home page, categorize this company amongst the following options only: 'B2B Software', 'B2C App', 'Edtech', 'Consulting Firm', 'Services & Agencies', 'Fintech', 'Media & Publishing'. If it doesn't fall into these categories, give it a new category name that is short and sensible : \n\n" + text + "\n\n. "
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.3
        # top_p
    )
    category = response.choices[0].text.strip()
    
    print('OPENAI DETECTED EXACT CATEGORY')
    print(category)
    print('')



    return category


def find_exact_category(text):

    # Find the pricing page URL
    category = extract_exact_category(text)

    if category:
        # Scrape the pricing buckets from the pricing page
        
        return category
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No ideal company entry angle found.')

