from urllib.parse import urlparse, urljoin
import requests
from bs4 import BeautifulSoup
import openai

openai.api_key = ''

def extract_logos_and_testimonials(text):
    prompt = "Based on this website home page, extract all logos and testimonials displayed on the page that represent the company's customers. Only return the name of the person, the name of the copmany, and the domain if foound: \n\n" + text + "\n\n. "
    response = openai.Completion.create(
        engine='text-davinci-003',
        prompt=prompt,
        max_tokens=100,
        n=1,
        stop=None,
        temperature=0.2
    )
    logos_and_testimonials = response.choices[0].text.strip()
    
    print('OPENAI DETECTED LOGOS AND TESTIMONIALS')
    print(logos_and_testimonials)
    print('')



    return logos_and_testimonials


def find_logos_and_testimonials(text):

    # Find the pricing page URL
    logos_and_testimonials = extract_logos_and_testimonials(text)

    if logos_and_testimonials:
        # Scrape the pricing buckets from the pricing page
        
        return logos_and_testimonials
        # print('Pricing Buckets:', pricing_buckets)
    else:
        print('No logos or testimonials found.')

