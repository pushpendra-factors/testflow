from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

openai.api_key = ''

def extract_customerbase_from_text(text):
    prompt = "Based on this website, share an array with strings with their top 6 customer types identified. The customer base is defined as the job titles within companies that make up their ideal buyers and users. \n\n" + text + "\n\n. Example values could be 'Founders', 'Demand Gen', 'Sales Leaders', 'Demand Gen' "
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.1
    )
    ctas = response.choices[0].text.strip()
    
    print('OPENAI DETECTED CUSTOMER BASE')
    print(ctas)
    print('')

    return ctas


def find_main_customer_base(text):

    # Find the pricing page URL
    customer_base = extract_customerbase_from_text(text)

    if customer_base:
        # Scrape the pricing buckets from the pricing page
        
        return customer_base
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No Customer Base assumption possible.')

