from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

openai.api_key = ''

def extract_quick_glance_information(text):
    prompt = "Based on this website home page, share what this company does, categorized in a form that makes intuitive sense. The goal is to make it obvious as to what the company does, or what categories of products they are known for. For example, on seeing the salesforce home page, we mentally categorize it as 'B2B Software', 'CRM', 'Sales Software', 'Marketing Automation', and so on. In the same way, share the top 5 categories that the following website falls into, as an array with strings: \n\n" + text + "\n\n. "
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.2
    )

    # print('original data')
    # print('')
    # print(response)
    # print('')

    quick_glance_info = response.choices[0].text.strip()
    
    print('OPENAI DETECTED QUICK GLANCE')    
    print(quick_glance_info)
    print('')

    return quick_glance_info


def find_quick_glance(text):

    # Find the pricing page URL
    quick_glance_info = extract_quick_glance_information(text)

    if quick_glance_info:
        # Scrape the pricing buckets from the pricing page
        
        return quick_glance_info
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No ideal company assumption possible.')

