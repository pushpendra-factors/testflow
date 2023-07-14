from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

openai.api_key = ''

def extract_CTAs_from_text(text):
    prompt = "Extract the top 4 CTAs or actions that this page has, and Share the output as an array with strings:\n\n" + text + "\n\n CTAs are conversion actions that are promoted from a marketing website. Usually, companies have between 1-3 CTAs. Common ones include 'Start free trial', 'Book a Demo', 'Contact Us', etc. "
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.2
    )
    ctas = response.choices[0].text.strip()
    
    print('OPENAI DETECTED CTAs')
    print(ctas)
    print('')

    return ctas


def find_main_CTAs(text):

    # Find the pricing page URL
    main_ctas = extract_CTAs_from_text(text)

    if main_ctas:
        # Scrape the pricing buckets from the pricing page
        
        return main_ctas
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No CTAs found.')

